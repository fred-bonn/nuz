package engine

import "testing"

func TestCleanNameFormatsPokemonNamesConsistently(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"lowercases mixed-case names":    {input: "pIkAcHu", want: "pikachu"},
		"preserves punctuation in names": {input: "Mr. Mime", want: "mr. mime"},
		"preserves apostrophes in names": {input: "Farfetch’d", want: "farfetch’d"},
		"handles empty input":            {input: "", want: ""},
		"keeps numerals in names":        {input: "Porygon2", want: "porygon2"},
		"normalizes hyphenated names":    {input: "Ho-Oh", want: "ho-oh"},
		"normalizes regional forms":      {input: "Arcanine-Hisui", want: "arcanine-hisui"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CleanName(tc.input); got != tc.want {
				t.Errorf("%s: cleanPokemonName(%q) = %q, want %q", name, tc.input, got, tc.want)
			}
		})
	}
}

func TestHasHyphenDetectsHyphenatedNames(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"detects no hyphen in empty input":      {input: "", want: false},
		"detects a hyphen in a hyphenated name": {input: "ho-oh", want: true},
		"detects no hyphen in a simple name":    {input: "pikachu", want: false},
		"detects no hyphen in a regional form":  {input: "arcanine-hisui", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := hasHyphen(tc.input); got != tc.want {
				t.Errorf("%s: hasHyphen(%q) = %t, want %t", name, tc.input, got, tc.want)
			}
		})
	}
}

func TestIsRegionalPokemonDetectsRegionalForms(t *testing.T) {
	tests := map[string]struct {
		input string
		want  bool
	}{
		"detects no regional form in empty input":       {input: "", want: false},
		"detects no regional form in a hyphenated name": {input: "ho-oh", want: false},
		"detects no regional form in a simple name":     {input: "pikachu", want: false},
		"detects a regional form in a regional name":    {input: "arcanine-hisui", want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := isRegionalPokemon(tc.input); got != tc.want {
				t.Errorf("%s: isRegionalPokemon(%q) = %t, want %t", name, tc.input, got, tc.want)
			}
		})
	}
}

func TestCleanNameFormatsMoveNamesConsistently(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"lowercases mixed-case move names": {input: "ThUndEr SHoCk", want: "thunder shock"},
		"handles empty move input":         {input: "", want: ""},
		"normalizes hyphenated move names": {input: "Tri-Attack", want: "tri attack"},
		"preserves numerals in move names": {input: "conversion 2", want: "conversion 2"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := CleanName(tc.input); got != tc.want {
				t.Errorf("%s: cleanMoveName(%q) = %q, want %q", name, tc.input, got, tc.want)
			}
		})
	}
}
