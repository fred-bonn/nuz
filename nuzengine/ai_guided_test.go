package nuzengine

import (
	"bytes"
	"strings"
	"testing"
)

func TestGuidedAIChoosesForcedReplacement(t *testing.T) {
	tests := map[string]struct {
		input        string
		wantName     string
		wantMessages []string
	}{
		"chooses the requested replacement": {
			input:        "2\n",
			wantName:     "second",
			wantMessages: []string{"Choose a replacement:"},
		},
		"retries invalid and out-of-range choices": {
			input:        "invalid\n3\n1\n",
			wantName:     "first",
			wantMessages: []string{"invalid input", "choice out of range"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mons := []*pokemon{
				{base: BasePokemon{Name: "first"}, hp: 5, stats: []int{10}},
				{base: BasePokemon{Name: "second"}, hp: 8, stats: []int{12}},
			}
			var output bytes.Buffer
			ga := newGuidedAI(strings.NewReader(tc.input), &output)

			got := ga.evaluteSwitchIns(nil, mons, nil)
			if got != mons[0] && got != mons[1] || got.base.Name != tc.wantName {
				t.Fatalf("replacement = %v, want %s", got, tc.wantName)
			}
			for _, message := range tc.wantMessages {
				if !strings.Contains(output.String(), message) {
					t.Errorf("output %q does not contain %q", output.String(), message)
				}
			}
		})
	}
}

type errRecordingState struct {
	*dummyBattleState
	err error
}

func (s *errRecordingState) setError(err error) { s.err = err }

func TestGuidedAIEvaluateActionsInput(t *testing.T) {
	newSlot := func() (*slot, []*moveAction) {
		active := &pokemon{base: BasePokemon{Name: "active"}, stats: []int{10}}
		bench := &pokemon{base: BasePokemon{Name: "bench"}, hp: 5, stats: []int{10}}
		sl := &slot{mon: active, Trainer: &trainer{pokemonParty: []*pokemon{active, bench}}}
		target := &slot{mon: &pokemon{base: BasePokemon{Name: "foe"}}}
		return sl, []*moveAction{{userSlot: sl, targetSlot: target, move: &Move{Move: "Tackle"}}}
	}

	tests := map[string]struct {
		input      string
		aborted    func() bool
		wantScore  int
		wantErr    bool
		wantOutput []string
	}{
		"chooses a move":                    {input: "1\n", wantScore: 1, wantOutput: []string{"1. Tackle against foe", "2. switch to bench"}},
		"chooses a switch":                  {input: "2\n", wantScore: -1},
		"retries invalid then out of range": {input: "x\n9\n1\n", wantScore: 1, wantOutput: []string{"invalid input", "choice out of range"}},
		"aborted input end sets error":      {input: "", aborted: func() bool { return true }, wantScore: 1, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			sl, actions := newSlot()
			bs := &errRecordingState{dummyBattleState: &dummyBattleState{}}
			var output bytes.Buffer
			ga := newGuidedAI(strings.NewReader(tc.input), &output)
			ga.aborted = tc.aborted

			got, score := ga.evaluateActions(bs, sl, actions)
			if got != actions[0] || score != tc.wantScore {
				t.Fatalf("got action %v score %d", got, score)
			}
			if (bs.err != nil) != tc.wantErr {
				t.Fatalf("error = %v, want error %v", bs.err, tc.wantErr)
			}
			if tc.wantErr && ga.shouldSwitch(bs, sl, 0, nil) {
				t.Fatal("aborted guided AI must not switch")
			}
			for _, want := range tc.wantOutput {
				if !strings.Contains(output.String(), want) {
					t.Errorf("output %q missing %q", output.String(), want)
				}
			}
		})
	}
}
