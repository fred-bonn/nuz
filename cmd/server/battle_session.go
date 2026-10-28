package main

import (
	"encoding/json"
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
)

// battle is an interactive battle driven through guided AI prompts.
type battle interface {
	Run(out io.Writer) error
	Send(line string) error
	Close()
	PlayerWon() bool
	Snapshot() nuzengine.BattleSnapshot
}

type session struct {
	battle   battle
	started  bool // guarded by server.mu
	awaiting atomic.Bool
	inputs   chan struct{}
}

func newSession(b battle) *session {
	return &session{battle: b, inputs: make(chan struct{}, 1)}
}

type event struct {
	Type    string   `json:"type"`
	Text    string   `json:"text,omitempty"`
	Options []string `json:"options,omitempty"`
	// State is the full battle state when the event was emitted.
	State *nuzengine.BattleSnapshot `json:"state,omitempty"`
	Won   *bool                     `json:"won,omitempty"`
	Error string                    `json:"error,omitempty"`
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
	pw.sess.awaiting.Store(true)
	text = strings.TrimSuffix(text, "> ")
	// Write runs on the engine goroutine, which is blocked on input, so reading the state here is safe.
	state := pw.sess.battle.Snapshot()
	select {
	case pw.events <- event{Type: "prompt", Text: text, Options: parsePromptOptions(text), State: &state}:
	case <-pw.done:
	}
	return len(p), nil
}

// expireIfUnstarted drops a battle the client never started.
func (s *server) expireIfUnstarted(id string) {
	s.mu.Lock()
	sess, ok := s.battles[id]
	expired := ok && !sess.started
	if expired {
		delete(s.battles, id)
	}
	s.mu.Unlock()
	if expired {
		sess.battle.Close()
	}
}

func (s *server) remove(id string) {
	s.mu.Lock()
	delete(s.battles, id)
	s.mu.Unlock()
}

func (s *server) lookup(w http.ResponseWriter, r *http.Request) (string, *session, bool) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		respondWithJSON(w, http.StatusMethodNotAllowed, errorResponse{"method not allowed"})
		return "", nil, false
	}
	id := r.Header.Get(battleIDHeader)
	if id == "" {
		respondWithJSON(w, http.StatusBadRequest, errorResponse{battleIDHeader + " header is required"})
		return "", nil, false
	}
	s.mu.Lock()
	sess, ok := s.battles[id]
	s.mu.Unlock()
	if !ok {
		respondWithJSON(w, http.StatusNotFound, errorResponse{"battle not found"})
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
	s.mu.Lock()
	if sess.started {
		s.mu.Unlock()
		respondWithJSON(w, http.StatusConflict, errorResponse{"battle already started"})
		return
	}
	sess.started = true
	s.mu.Unlock()

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
		s.remove(id)
		<-done
	}
	defer s.remove(id)

	timer := time.NewTimer(s.timeout)
	timer.Stop()
	defer timer.Stop()
	var timerC <-chan time.Time
	for {
		select {
		case ev := <-events:
			timer.Reset(s.timeout)
			timerC = timer.C
			send(ev)
		case <-sess.inputs:
			timer.Stop()
			timerC = nil
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
		respondWithJSON(w, http.StatusBadRequest, errorResponse{"invalid JSON body: " + err.Error()})
		return
	}
	if !sess.awaiting.CompareAndSwap(true, false) {
		respondWithJSON(w, http.StatusConflict, errorResponse{"battle is not waiting for input"})
		return
	}
	select {
	case sess.inputs <- struct{}{}:
	default:
	}
	if err := sess.battle.Send(strconv.Itoa(req.Choice)); err != nil {
		respondWithJSON(w, http.StatusGone, errorResponse{"battle is no longer running"})
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
