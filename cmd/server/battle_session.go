package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fred-bonn/nuz/nuzengine"
)

const (
	battleIDHeader = "X-Battle-ID"
	inputTimeout   = 2 * time.Minute
	// hintInterval is how often the best known option is pushed while a prompt is pending.
	hintInterval = 1 * time.Second
)

// battle is an interactive battle driven through guided AI prompts.
type battle interface {
	Run(out io.Writer) error
	Send(line string) error
	Close()
	PlayerWon() bool
	Snapshot() nuzengine.BattleSnapshot
	// BestOption is the 1-based number of the option currently rated best for the pending prompt.
	BestOption() (int, bool)
}

type session struct {
	battle  battle
	started atomic.Bool

	mu      sync.Mutex
	waiting bool // a prompt has been shown and not yet answered
	options int  // number of options of the latest prompt
}

var (
	errNotWaiting = errors.New("battle is not waiting for input")
	errBadChoice  = errors.New("choice is not one of the prompt options")
)

func newSession(b battle) *session {
	return &session{battle: b}
}

// setPending records a shown prompt. A prompt without options is a re-prompt, which keeps the previous options.
func (s *session) setPending(options int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.waiting = true
	if options > 0 {
		s.options = options
	}
}

func (s *session) pending() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.waiting
}

// claim answers the pending prompt; an invalid choice leaves the prompt pending.
func (s *session) claim(choice int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.waiting {
		return errNotWaiting
	}
	if choice < 1 || choice > s.options {
		return errBadChoice
	}
	s.waiting = false
	return nil
}

type event struct {
	Type    string   `json:"type"`
	Text    string   `json:"text,omitempty"`
	Options []string `json:"options,omitempty"`
	// State is the full battle state when the event was emitted.
	State *nuzengine.BattleSnapshot `json:"state,omitempty"`
	// Best is the 1-based number of the best option in a hint event.
	Best  int    `json:"best,omitempty"`
	Won   *bool  `json:"won,omitempty"`
	Error string `json:"error,omitempty"`
}

var optionPattern = regexp.MustCompile(`^[0-9]+\. (.+)$`)

// parsePromptOptions extracts the numbered options ("1. tackle") from prompt text, in order.
func parsePromptOptions(text string) []string {
	var options []string
	for _, line := range strings.Split(text, "\n") {
		if m := optionPattern.FindStringSubmatch(strings.TrimRight(line, "\r")); m != nil {
			options = append(options, m[1])
		}
	}
	return options
}

type inputRequest struct {
	Choice int `json:"choice"`
}

// promptWriter turns guided AI output into prompt events; every prompt ends with "> ".
type promptWriter struct {
	sess   *session
	events chan<- event
	done   <-chan struct{}
	buf    strings.Builder
}

func (pw *promptWriter) Write(p []byte) (int, error) {
	pw.buf.Write(p)
	text := pw.buf.String()
	if !strings.HasSuffix(text, "> ") {
		return len(p), nil
	}
	pw.buf.Reset()
	text = strings.TrimSuffix(text, "> ")
	// Write runs on the engine goroutine, which is blocked on input, so reading the state here is safe.
	state := pw.sess.battle.Snapshot()
	select {
	case pw.events <- event{Type: "prompt", Text: text, Options: parsePromptOptions(text), State: &state}:
	case <-pw.done:
	}
	return len(p), nil
}

func (s *server) lookup(w http.ResponseWriter, r *http.Request) (string, *session, bool) {
	id := r.Header.Get(battleIDHeader)
	if id == "" {
		respondWithError(w, http.StatusBadRequest, battleIDHeader+" header is required", nil)
		return "", nil, false
	}
	sess, ok := s.battles.get(id)
	if !ok {
		respondWithError(w, http.StatusNotFound, "battle not found", nil)
		return "", nil, false
	}
	return id, sess, true
}

// handleBattleStart runs the battle and streams NDJSON events until it ends,
// fails, or the client stays silent for longer than the input timeout.
func (s *server) handleBattleStart(w http.ResponseWriter, r *http.Request) {
	id, sess, ok := s.lookup(w, r)
	if !ok {
		return
	}
	if !sess.started.CompareAndSwap(false, true) {
		respondWithError(w, http.StatusConflict, "battle already started", nil)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)
	send := func(ev event) {
		_ = enc.Encode(ev)
		if flusher != nil {
			flusher.Flush()
		}
	}

	events := make(chan event, 16)
	stop := make(chan struct{})
	var stopOnce sync.Once
	closeStop := func() { stopOnce.Do(func() { close(stop) }) }
	defer closeStop()
	done := make(chan error, 1)
	go func() { done <- sess.battle.Run(&promptWriter{sess: sess, events: events, done: stop}) }()

	abort := func() {
		closeStop()
		sess.battle.Close()
		s.battles.remove(id)
		<-done
	}
	defer s.battles.remove(id)

	timer := time.NewTimer(s.timeout)
	timer.Stop()
	defer timer.Stop()
	var timerC <-chan time.Time
	hints := time.NewTicker(s.hintInterval)
	defer hints.Stop()
	for {
		select {
		case ev := <-events:
			timer.Reset(s.timeout)
			timerC = timer.C
			// Marked pending before the client can see the prompt, so an answer can never arrive early.
			if ev.Type == "prompt" {
				sess.setPending(len(ev.Options))
			}
			send(ev)
		case <-hints.C:
			if !sess.pending() {
				continue
			}
			if best, ok := sess.battle.BestOption(); ok {
				send(event{Type: "hint", Best: best})
			}
		case err := <-done:
			for len(events) > 0 {
				send(<-events)
			}
			sess.battle.Close()
			if err != nil {
				send(event{Type: "error", Error: err.Error()})
				return
			}
			won := sess.battle.PlayerWon()
			state := sess.battle.Snapshot()
			send(event{Type: "end", Won: &won, State: &state})
			return
		case <-timerC:
			// An answered prompt means the engine is working, not that the client went silent.
			if !sess.pending() {
				timerC = nil
				continue
			}
			// Abort first so the battle is already gone when the client sees the timeout.
			abort()
			send(event{Type: "timeout", Error: "no input received in time"})
			return
		case <-r.Context().Done():
			abort()
			return
		}
	}
}

// handleBattleInput answers the pending prompt; the next prompt arrives on the start stream.
func (s *server) handleBattleInput(w http.ResponseWriter, r *http.Request) {
	_, sess, ok := s.lookup(w, r)
	if !ok {
		return
	}
	var req inputRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid JSON body: "+err.Error(), err)
		return
	}
	switch err := sess.claim(req.Choice); {
	case errors.Is(err, errNotWaiting):
		respondWithError(w, http.StatusConflict, err.Error(), nil)
		return
	case errors.Is(err, errBadChoice):
		respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}
	if err := sess.battle.Send(strconv.Itoa(req.Choice)); err != nil {
		respondWithError(w, http.StatusGone, "battle is no longer running", err)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
