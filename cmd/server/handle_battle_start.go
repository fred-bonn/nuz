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
