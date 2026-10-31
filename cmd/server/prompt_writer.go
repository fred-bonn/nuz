package main

import (
	"regexp"
	"strings"
)

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
