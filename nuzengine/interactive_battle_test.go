package nuzengine

import (
	"context"
	"testing"
	"time"
)

func TestInteractiveBattleRunAbortsWhenContextCancelled(t *testing.T) {
	ib, err := InitInteractiveBattle(context.Background(), 0, cancelTestPlayerParty, cancelTestOpponentParty, 0)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runErr := make(chan error, 1)
	go func() {
		runErr <- ib.Run(ctx, func(title string, options []string) {
			// Cancel while the first prompt waits for an answer.
			cancel()
		})
	}()

	select {
	case err := <-runErr:
		if err == nil {
			t.Fatal("Run returned no error after context cancellation")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after context cancellation")
	}

	// Run closes the battle before returning, so the close must already be visible.
	select {
	case <-ib.closed:
	default:
		t.Fatal("battle was not closed after context cancellation")
	}

	// The live learner is stopped by the deferred close before Run returns.
	if ib.live == nil {
		t.Fatal("live learner was not initialized")
	}
	select {
	case <-ib.live.done:
	default:
		t.Fatal("live learner was not stopped after context cancellation")
	}
}
