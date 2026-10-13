package nuzengine

import "testing"

func TestStatelessChooseMatchesReplayCandidates(t *testing.T) {
	tests := map[string]struct {
		candidates    []actionCandidate
		wantKey       string
		wantSuccess   bool
		wantReplayIdx int
		wantSequence  []string
		wantError     error
	}{
		"selects the expected replay action": {
			candidates:    []actionCandidate{{key: "move:other"}, {key: "move:expected"}},
			wantKey:       "move:expected",
			wantSuccess:   true,
			wantReplayIdx: 1,
			wantSequence:  []string{"move:expected"},
		},
		"fails when the replay action is unavailable": {
			candidates: []actionCandidate{{key: "move:available"}},
			wantError:  errStatelessReplayFailed,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			bs := &singleBattleState{}
			la := &learningAiStateless{
				replay:    []string{"move:expected"},
				replaying: true,
			}

			got, ok := la.choose(bs, tc.candidates)
			if ok != tc.wantSuccess {
				t.Fatalf("choose() success = %t, want %t", ok, tc.wantSuccess)
			}
			if ok && got.key != tc.wantKey {
				t.Errorf("chosen action = %q, want %q", got.key, tc.wantKey)
			}
			if la.replayIndex != tc.wantReplayIdx {
				t.Errorf("replay index = %d, want %d", la.replayIndex, tc.wantReplayIdx)
			}
			if len(la.sequence) != len(tc.wantSequence) {
				t.Fatalf("sequence = %v, want %v", la.sequence, tc.wantSequence)
			}
			for i, action := range tc.wantSequence {
				if la.sequence[i] != action {
					t.Errorf("sequence[%d] = %q, want %q", i, la.sequence[i], action)
				}
			}
			if bs.err != tc.wantError {
				t.Errorf("battle error = %v, want %v", bs.err, tc.wantError)
			}
		})
	}
}
