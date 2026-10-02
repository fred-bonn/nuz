package nuzengine

import (
	"bytes"
	"strings"
	"testing"
)

func TestGuidedAIChoosesForcedReplacement(t *testing.T) {
	mons := []*pokemon{
		{base: BasePokemon{Name: "first"}, hp: 5, stats: []int{10}},
		{base: BasePokemon{Name: "second"}, hp: 8, stats: []int{12}},
	}
	var output bytes.Buffer
	ga := newGuidedAI(strings.NewReader("2\n"), &output)

	got := ga.evaluteSwitchIns(nil, mons, nil)
	if got != mons[1] {
		t.Fatalf("replacement = %v, want %v", got, mons[1])
	}
	if !strings.Contains(output.String(), "Choose a replacement:") {
		t.Fatalf("replacement prompt was not printed: %q", output.String())
	}
}
