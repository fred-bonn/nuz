package main

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// handleBattleStart runs the battle and streams NDJSON events until it ends,
// fails, or the client stays silent for longer than the input timeout.
func (s *server) handleBattleStart(w http.ResponseWriter, r *http.Request) {
	id, sess := sessionFrom(r)
	if !sess.started.CompareAndSwap(false, true) {
		respondWithError(w, http.StatusConflict, "battle already started", nil)
		return
	}

	send := startStream(w)
	events, done, stop := runBattle(sess)
	defer stop()

	abort := func() {
		stop()
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
	lastHint := 0 // best option last sent for the current prompt; 0 means none yet
	for {
		select {
		case ev := <-events:
			timer.Reset(s.timeout)
			timerC = timer.C
			// Marked pending before the client can see the prompt, so an answer can never arrive early.
			if ev.Type == "prompt" {
				sess.setPending(len(ev.Options))
				lastHint = 0
			}
			send(ev)
		case <-hints.C:
			if !sess.pending() {
				continue
			}
			if best, ok := sess.battle.BestOption(); ok && best != lastHint {
				lastHint = best
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

// startStream writes the NDJSON response header and returns a function that sends and flushes one event.
func startStream(w http.ResponseWriter) func(event) {
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)
	flusher, _ := w.(http.Flusher)
	enc := json.NewEncoder(w)
	return func(ev event) {
		_ = enc.Encode(ev)
		if flusher != nil {
			flusher.Flush()
		}
	}
}

// runBattle runs the battle on its own goroutine. Prompts arrive on events and the result on done;
// stop releases a prompt writer blocked on events and is safe to call more than once.
func runBattle(sess *session) (events chan event, done chan error, stop func()) {
	events = make(chan event, 16)
	done = make(chan error, 1)
	stopped := make(chan struct{})
	stop = sync.OnceFunc(func() { close(stopped) })
	go func() { done <- sess.battle.Run(&promptWriter{sess: sess, events: events, done: stopped}) }()
	return events, done, stop
}
