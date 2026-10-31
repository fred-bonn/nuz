package main

import (
	"errors"
	"io"
	"regexp"
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
