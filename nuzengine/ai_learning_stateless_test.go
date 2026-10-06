package nuzengine

import "testing"

func TestStatelessChooseFailsWhenReplayActionUnavailable(t *testing.T) {
	bs := &singleBattleState{}
	la := &learningAiStateless{
		replay:    []string{"move:expected"},
		replaying: true,
	}

	if _, ok := la.choose(bs, []actionCandidate{{key: "move:available"}}); ok {
		t.Fatal("choose() succeeded when the replay action was unavailable")
	}
	if bs.err != errStatelessReplayFailed {
		t.Fatalf("battle error = %v, want %v", bs.err, errStatelessReplayFailed)
	}
}
