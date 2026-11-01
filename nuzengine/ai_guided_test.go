package nuzengine

import (
	"bytes"
	"errors"
	"fmt"
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

			got := ga.evaluateSwitchIns(nil, mons, nil)
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
func (s *errRecordingState) getError() error    { return s.err }

type failingWriter struct {
	err error
}

func (w failingWriter) Write([]byte) (int, error) {
	return 0, w.err
}

type countingAI struct {
	evaluateCalls int
}

func (ai *countingAI) evaluateActions(_ battleState, _ *slot, actions []*moveAction) (*moveAction, int) {
	ai.evaluateCalls++
	return actions[0], 1
}

func (ai *countingAI) evaluateSwitchIns(_ battleState, mons []*pokemon, _ *slot) *pokemon {
	return mons[0]
}

func (ai *countingAI) shouldSwitch(battleState, *slot, int, []*pokemon) bool {
	return false
}

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

func TestGuidedAIOutputErrorsSetBattleError(t *testing.T) {
	outputErr := errors.New("writer failed")
	newSlot := func() (*slot, []*moveAction) {
		active := &pokemon{base: BasePokemon{Name: "active"}, stats: []int{10}}
		sl := &slot{mon: active, Trainer: &trainer{pokemonParty: []*pokemon{active}}}
		target := &slot{mon: &pokemon{base: BasePokemon{Name: "foe"}}}
		return sl, []*moveAction{{userSlot: sl, targetSlot: target, move: &Move{Move: "Tackle"}}}
	}

	tests := map[string]struct {
		run func(*errRecordingState, *guidedAi) error
	}{
		"action prompt": {
			run: func(bs *errRecordingState, ga *guidedAi) error {
				sl, actions := newSlot()
				ga.evaluateActions(bs, sl, actions)
				return bs.err
			},
		},
		"replacement prompt": {
			run: func(bs *errRecordingState, ga *guidedAi) error {
				mons := []*pokemon{{base: BasePokemon{Name: "replacement"}, stats: []int{10}}}
				ga.evaluateSwitchIns(bs, mons, nil)
				return bs.err
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			bs := &errRecordingState{dummyBattleState: &dummyBattleState{}}
			ga := newGuidedAI(strings.NewReader("1\n"), failingWriter{err: outputErr})
			if err := tc.run(bs, ga); !errors.Is(err, outputErr) {
				t.Fatalf("error = %v, want wrapped writer error %v", err, outputErr)
			}
		})
	}
}

func TestGuidedAIOutputErrorStopsGatheringActions(t *testing.T) {
	outputErr := errors.New("writer failed")
	ga := newGuidedAI(strings.NewReader(""), failingWriter{err: outputErr})
	bs := testRootBattle(t, ga)
	opponentAI := &countingAI{}
	bs.opponent.ai = opponentAI

	if err := bs.execute(); !errors.Is(err, outputErr) {
		t.Fatalf("execute error = %v, want wrapped writer error %v", err, outputErr)
	}
	if opponentAI.evaluateCalls != 0 {
		t.Fatalf("opponent evaluated %d actions, want none after player output failure", opponentAI.evaluateCalls)
	}
}

func TestParseChoice(t *testing.T) {
	tests := []struct {
		line    string
		want    int
		wantErr bool
	}{
		{line: "1\n", want: 1},
		{line: "12\r\n", want: 12},
		{line: "7", want: 7},
		{line: "0\n", want: 0},
		{line: "", wantErr: true},
		{line: "\n", wantErr: true},
		{line: " 1\n", wantErr: true},
		{line: "1 \n", wantErr: true},
		{line: "+1\n", wantErr: true},
		{line: "-1\n", wantErr: true},
		{line: "1 2\n", wantErr: true},
		{line: "1abc\n", wantErr: true},
		{line: "1\n2\n", wantErr: true},
		{line: "1.0\n", wantErr: true},
		{line: "0x1\n", wantErr: true},
		{line: "١\n", wantErr: true},
		{line: "99999999999999999999\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.line), func(t *testing.T) {
			got, err := parseChoice(tt.line)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("parseChoice = %d, %v; want %d, err=%v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestGuidedAIDiscardsNonNumericInput(t *testing.T) {
	mons := []*pokemon{{base: BasePokemon{Name: "first"}, stats: []int{10}}, {base: BasePokemon{Name: "second"}, stats: []int{10}}}
	var output bytes.Buffer
	ga := newGuidedAI(strings.NewReader("1 2\n+2\n -1\n2abc\n2\n"), &output)
	if got := ga.evaluateSwitchIns(nil, mons, nil); got != mons[1] {
		t.Fatalf("got %v, want second", got)
	}
	if n := strings.Count(output.String(), "invalid input"); n != 4 {
		t.Fatalf("invalid input messages = %d, want 4", n)
	}
}
