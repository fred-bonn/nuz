package main

import (
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/fred-bonn/nuz/internal/pokeapi"
)

func TestMoveParsingAllowsMovesWithNoAilment(t *testing.T) {
	move, err := toMove(pokeapi.MoveJSON{
		Name:     "low-sweep",
		Power:    65,
		PP:       20,
		Accuracy: 100,
		Priority: 0,
		Type: struct {
			Name string `json:"name"`
		}{Name: "fighting"},
		DamageClass: struct {
			Name string `json:"name"`
		}{Name: "physical"},
		Meta: struct {
			Ailment struct {
				Name string `json:"name"`
			} `json:"ailment"`
			AilmentChance int `json:"ailment_chance"`
			Category      struct {
				Name string `json:"name"`
			} `json:"category"`
			CritRate     int `json:"crit_rate"`
			Drain        int `json:"drain"`
			FlinchChance int `json:"flinch_chance"`
			Heal         int `json:"healing"`
			MaxHits      int `json:"max_hits"`
			MaxTurns     int `json:"max_turns"`
			MinHits      int `json:"min_hits"`
			MinTurns     int `json:"min_turns"`
			StatChance   int `json:"stat_chance"`
		}{
			Ailment: struct {
				Name string `json:"name"`
			}{Name: "none"},
		},
	})
	if err != nil {
		t.Fatalf("toMove should accept a move with no ailment: %v", err)
	}
	if move.Name != "low-sweep" {
		t.Fatalf("unexpected move parsed: %s", move.Name)
	}
	if move.Ailment != noneAilment {
		t.Fatalf("move ailment = %v, want %v", move.Ailment, noneAilment)
	}
}

func TestStringToAilmentStateParsesKnownAilmentNames(t *testing.T) {
	tests := map[string]struct {
		input string
		want  ailmentState
	}{
		"parses paralysis":               {input: "paralysis", want: paralysisAilment},
		"parses poison":                  {input: "poison", want: poisonAilment},
		"parses toxic":                   {input: "toxic", want: toxicAilment},
		"parses burn":                    {input: "burn", want: burnAilment},
		"parses freeze":                  {input: "freeze", want: freezeAilment},
		"parses sleep":                   {input: "sleep", want: sleepAilment},
		"parses infatuation":             {input: "infatuation", want: infatuationAilment},
		"parses confusion":               {input: "confusion", want: confusionAilment},
		"parses trap":                    {input: "trap", want: trapAilment},
		"parses bound":                   {input: "bound", want: boundAilment},
		"parses leech seed":              {input: "leech seed", want: leechSeedAilment},
		"parses yawn":                    {input: "yawn", want: yawnAilment},
		"defaults unknown input to none": {input: "", want: noneAilment},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := stringToAilmentState(tc.input); got != tc.want {
				t.Errorf("stringToAilmentState(%s) = %s, want %s", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsNonVolatileStatusIdentifiesNonVolatileAilments(t *testing.T) {
	tests := map[string]struct {
		state ailmentState
		want  bool
	}{
		"treats paralysis as non-volatile":       {state: paralysisAilment, want: true},
		"treats none as not non-volatile":        {state: noneAilment, want: false},
		"treats sleep as non-volatile":           {state: sleepAilment, want: true},
		"treats infatuation as not non-volatile": {state: infatuationAilment, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.state.isNonVolatileStatus(); got != tc.want {
				t.Errorf("tc.%s.isNonVolatileStatus() = %t, want %t", tc.state, got, tc.want)
			}
		})
	}
}

func TestAilmentIteratorsIncludeTheExpectedAilmentGroups(t *testing.T) {
	tests := map[string]struct {
		iterator func(func(ailmentState) bool)
		want     bool
	}{
		"includes non-volatile statuses": {iterator: nonVolatileStatuses, want: true},
		"includes volatile statuses":     {iterator: volatileStatuses, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for state := range tc.iterator {
				if got := state.isNonVolatileStatus(); got != tc.want {
					t.Errorf("%s := range %s, tc.%s.isNonVolatileStatus() = %t, want %t", state, FunctionName(tc.iterator), state, got, tc.want)
				}
			}
		})
	}
}

const (
	generateAilmentIterations int = 500
)

func TestGenerateAilmentProducesAValidAilmentState(t *testing.T) {
	tests := map[string]struct {
		state    ailmentState
		want     ailmentState
		minTurns int
		maxTurns int
	}{
		"generates sleep with a short duration":     {state: sleepAilment, want: sleepAilment, minTurns: 1, maxTurns: 3},
		"generates confusion with a short duration": {state: confusionAilment, want: confusionAilment, minTurns: 1, maxTurns: 4},
		"generates yawn with a fixed duration":      {state: yawnAilment, want: yawnAilment, minTurns: 2, maxTurns: 2},
		"generates paralysis with zero duration":    {state: paralysisAilment, want: paralysisAilment, minTurns: 0, maxTurns: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < generateAilmentIterations; i++ {
				if got := generateAilment(tc.state, nil); got.state != tc.want || got.turns < tc.minTurns || got.turns > tc.maxTurns {
					t.Fatalf("generateAilment(%s, _) = {state: %s, turns: %d, _}, want {state: %s, turns: %d-%d, _}", tc.state, got.state, got.turns, tc.want, tc.minTurns, tc.maxTurns)
				}
			}
		})
	}
}

func TestGenerateTrapProducesAValidTrapAilment(t *testing.T) {
	tests := map[string]struct {
		minTurns int
		maxTurns int
	}{
		"generates a trap with 1 to 5 turns":  {minTurns: 1, maxTurns: 5},
		"generates a trap with 2 turns":       {minTurns: 2, maxTurns: 2},
		"generates a trap with 0 to 1 turns":  {minTurns: 0, maxTurns: 1},
		"generates a trap with 0 turns":       {minTurns: 0, maxTurns: 0},
		"generates a trap with 99 turns":      {minTurns: 99, maxTurns: 99},
		"generates a trap with 0 to 10 turns": {minTurns: 0, maxTurns: 10},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			for i := 0; i < generateAilmentIterations; i++ {
				if got := generateTrap(tc.minTurns, tc.maxTurns, nil); got.state != trapAilment || got.turns < tc.minTurns || got.turns > tc.maxTurns {
					t.Fatalf("generateTrap(%d, %d, _) = {state: %s, turns: %d, _}, want {state: %s, turns: %d-%d, _}", tc.minTurns, tc.maxTurns, got.state, got.turns, trapAilment, tc.minTurns, tc.maxTurns)
				}
			}
		})
	}
}

func FunctionName(fn interface{}) string {
	full := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	full = path.Base(full)
	if i := strings.LastIndex(full, "."); i != -1 {
		return full[i+1:]
	}
	return full
}
