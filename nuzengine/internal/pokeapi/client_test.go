package pokeapi

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestClientFetchHandlesAPIResponses(t *testing.T) {
	const baseURL = "https://pokeapi.co/api/v2"
	tests := map[string]struct {
		resource        string
		name            string
		statusCode      int
		status          string
		body            string
		transportErr    error
		wantError       bool
		wantPokemonName string
		wantPokemonType string
		wantMoveName    string
		wantMovePower   int
	}{
		"fetches pokemon data": {
			resource:        "pokemon",
			name:            "charizard",
			statusCode:      http.StatusOK,
			status:          "200 OK",
			body:            `{"id":6,"name":"charizard","types":[{"type":{"name":"fire"}}],"stats":[{"base_stat":84,"stat":{"name":"hp"}}]}`,
			wantPokemonName: "charizard",
			wantPokemonType: "fire",
		},
		"fetches move data": {
			resource:      "move",
			name:          "flamethrower",
			statusCode:    http.StatusOK,
			status:        "200 OK",
			body:          `{"name":"flamethrower","power":90,"pp":15,"accuracy":100,"type":{"name":"fire"},"damage_class":{"name":"special"},"meta":{"crit_rate":0,"category":{"name":"damage"}}}`,
			wantMoveName:  "flamethrower",
			wantMovePower: 90,
		},
		"reports transport errors": {
			resource:     "pokemon",
			name:         "missingno",
			transportErr: fmt.Errorf("boom"),
			wantError:    true,
		},
		"reports HTTP failures": {
			resource:   "move",
			name:       "missing-move",
			statusCode: http.StatusNotFound,
			status:     "404 Not Found",
			body:       `{"detail":"Not found"}`,
			wantError:  true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			client := Client{
				httpClient: &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
					wantURL := baseURL + "/" + tc.resource + "/" + tc.name
					if req.URL.String() != wantURL {
						t.Errorf("request URL = %q, want %q", req.URL.String(), wantURL)
					}
					if tc.transportErr != nil {
						return nil, tc.transportErr
					}
					return &http.Response{
						StatusCode: tc.statusCode,
						Status:     tc.status,
						Body:       io.NopCloser(strings.NewReader(tc.body)),
						Header:     make(http.Header),
					}, nil
				})},
				baseURL: baseURL,
			}

			if tc.resource == "pokemon" {
				pokemon, err := client.FetchPokemon(tc.name)
				if tc.wantError {
					if err == nil {
						t.Fatal("FetchPokemon returned no error")
					}
					return
				}
				if err != nil {
					t.Fatalf("FetchPokemon returned unexpected error: %v", err)
				}
				if pokemon.Name != tc.wantPokemonName {
					t.Errorf("pokemon name = %q, want %q", pokemon.Name, tc.wantPokemonName)
				}
				if got := pokemon.Types[0].Type.Name; got != tc.wantPokemonType {
					t.Errorf("primary type = %q, want %q", got, tc.wantPokemonType)
				}
				return
			}

			move, err := client.FetchMove(tc.name)
			if tc.wantError {
				if err == nil {
					t.Fatal("FetchMove returned no error")
				}
				return
			}
			if err != nil {
				t.Fatalf("FetchMove returned unexpected error: %v", err)
			}
			if move.Name != tc.wantMoveName {
				t.Errorf("move name = %q, want %q", move.Name, tc.wantMoveName)
			}
			if move.Power != tc.wantMovePower {
				t.Errorf("move power = %d, want %d", move.Power, tc.wantMovePower)
			}
		})
	}
}

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
			if got := cleanName(tc.input); got != tc.want {
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
			if got := cleanName(tc.input); got != tc.want {
				t.Errorf("%s: cleanMoveName(%q) = %q, want %q", name, tc.input, got, tc.want)
			}
		})
	}
}
