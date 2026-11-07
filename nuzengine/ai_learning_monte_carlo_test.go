package nuzengine

import (
	"context"
	"testing"
)

// The parties use only cached Pokemon and moves, so no network access is needed.
const (
	cancelTestPlayerParty   = "Horsea @ Oran Berry\nLevel: 17\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n"
	cancelTestOpponentParty = "Ponyta @ Oran Berry\nLevel: 17\nAdamant Nature\nAbility: Flame Body\n- Ember\n"
)

func TestRunWorkerStopsWhenContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	episodes := 0
	runWorker(ctx, &config{}, 0, cancelTestPlayerParty, cancelTestOpponentParty, 0, 100, func(episode) {
		episodes++
	})
	if episodes != 0 {
		t.Fatalf("episodes = %d, want 0 after context cancellation", episodes)
	}
}
