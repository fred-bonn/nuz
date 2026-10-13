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
