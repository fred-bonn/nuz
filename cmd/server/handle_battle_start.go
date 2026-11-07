package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"
)

// errInputTimeout is the cause recorded on the battle context when a shown prompt
// stays unanswered for longer than the input timeout.
var errInputTimeout = errors.New("no input received in time")

// handleBattleStart runs the battle and streams NDJSON events until it ends,
// fails, or the client stays silent for longer than the input timeout.
func (s *server) handleBattleStart(w http.ResponseWriter, r *http.Request) {
	id, sess := sessionFrom(r)
	if !sess.started.CompareAndSwap(false, true) {
		respondWithError(w, http.StatusConflict, "battle already started", nil)
		return
	}

	send := startStream(w)
	// The battle context is cancelled on the input timeout and on client disconnects;
	// cancelling it aborts the battle and stops the live learner.
	battleCtx, cancelBattle := context.WithCancelCause(r.Context())
	defer cancelBattle(nil)
	events, done, stop := runBattle(battleCtx, sess)
	defer stop()

	abort := func() {
		stop()
		sess.battle.Close()
		s.battles.remove(id)
		<-done
	}
	defer s.battles.remove(id)

	idle := time.AfterFunc(s.timeout, func() {
		// An answered prompt means the engine is working, not that the client went silent.
		if sess.pending() {
			cancelBattle(errInputTimeout)
		}
	})
	defer idle.Stop()
	hints := time.NewTicker(s.hintInterval)
	defer hints.Stop()
	lastHint := 0 // best option last sent for the current prompt; 0 means none yet
	for {
		select {
		case ev := <-events:
			idle.Reset(s.timeout)
			// Marked pending before the client can see the prompt, so an answer can never arrive early.
			sess.setPending(len(ev.Options))
			lastHint = 0
			send(ev)
		case <-hints.C:
			if !sess.pending() {
				continue
			}
			if best, ok := sess.battle.BestOption(); ok && best != lastHint {
				lastHint = best
				send(newHintEvent(best))
			}
		case err := <-done:
			if err != nil && battleCtx.Err() != nil {
				// The battle stopped because its context was cancelled; close it and
				// report the timeout, if that was the cause. The deferred remove drops it.
				sess.battle.Close()
				if errors.Is(context.Cause(battleCtx), errInputTimeout) {
					send(newTimeoutEvent())
				}
				return
			}
			for len(events) > 0 {
				send(<-events)
			}
			sess.battle.Close()
			if err != nil {
				send(newErrorEvent(err))
				return
			}
			send(newEndEvent(sess.battle.PlayerWon(), sess.battle.Snapshot()))
			return
		case <-battleCtx.Done():
			// Abort first so the battle is already gone when the client sees the timeout.
			abort()
			if errors.Is(context.Cause(battleCtx), errInputTimeout) {
				send(newTimeoutEvent())
			}
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
// stop releases a prompt callback blocked on events and is safe to call more than once.
func runBattle(ctx context.Context, sess *session) (events chan promptEvent, done chan error, stop func()) {
	events = make(chan promptEvent, 16)
	done = make(chan error, 1)
	stopped := make(chan struct{})
	stop = sync.OnceFunc(func() { close(stopped) })
	go func() {
		// Runs on the engine goroutine, which is blocked on input afterwards, so reading the state is safe.
		done <- sess.battle.Run(ctx, func(title string, options []string) {
			select {
			case events <- newPromptEvent(title, options, sess.battle.Snapshot()):
			case <-stopped:
			}
		})
	}()
	return events, done, stop
}
