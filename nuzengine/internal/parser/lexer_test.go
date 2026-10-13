package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNextTokenReadsTheExpectedTokenType(t *testing.T) {
	tests := map[string][]struct {
		input string
		want  token
	}{
		"skips leading whitespace": {
			{
				input: "\t\t   \r\n\t",
				want: token{
					Type:    t_NEWLINE,
					Literal: "\\n",
				},
			},
		},
		"parses an item token": {
			{
				input: "@ item-name",
				want: token{
					Type:    t_ITEM,
					Literal: "item-name",
				},
			},
		},
		"returns EOF at the end of input": {
			{
				input: "",
				want: token{
					Type:    t_EOF,
					Literal: "",
				},
			},
		},
		"parses a level token": {
			{
				input: ": 69",
				want: token{
					Type:    t_IDENT,
					Literal: "69",
				},
			},
		},
		"parses a nature token": {
			{
				input: "Jolly Nature\n",
				want: token{
					Type:    t_IDENT,
					Literal: "Jolly Nature",
				},
			},
		},
		"recognizes a level token": {
			{
				input: "Level: 30",
				want: token{
					Type:    t_LEVEL,
					Literal: "",
				},
			},
		},
		"recognizes a status token": {
			{
				input: "   Status>",
				want: token{
					Type:    t_STATUS,
					Literal: "",
				},
			},
		},
		"parses a move token": {
			{
				input: "\t - Tackle",
				want: token{
					Type:    t_MOVE,
					Literal: "-",
				},
			},
		},
	}

	for name, tcs := range tests {
		for _, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				l := newLexer(string(tc.input))

				if got := l.nextToken(); got != tc.want {
					t.Errorf("%s: l.NextToken() = %q, want %q", name, got, tc.want)
				}
			})
		}
	}
}

func TestReadIdentifierReadsPokemonNamesCorrectly(t *testing.T) {
	tests := map[string][]struct {
		input string
		want  token
	}{
		"parses a name with a period": {
			{
				input: "Mr. Mime   \n",
				want: token{
					Type:    t_IDENT,
					Literal: "Mr Mime",
				},
			},
		},
		"parses a hyphenated name": {
			{
				input: "Nidoran-F\n",
				want: token{
					Type:    t_IDENT,
					Literal: "Nidoran-F",
				},
			},
		},
		"parses a name with an apostrophe": {
			{
				input: "   Farfetc'h\n",
				want: token{
					Type:    t_IDENT,
					Literal: "Farfetch",
				},
			},
		},
		"parses a name with a numeral": {
			{
				input: "Porygon2\n",
				want: token{
					Type:    t_IDENT,
					Literal: "Porygon2",
				},
			},
		},
	}

	for name, tcs := range tests {
		for _, tc := range tcs {
			t.Run(name, func(t *testing.T) {
				l := newLexer(string(tc.input))

				if got := l.readIdent(); got != tc.want {
					t.Errorf("%s: parsePokemonLine(%q) = %q, want %q", name, tc.input, got, tc.want)
				}
				if l.ch != '\n' {
					t.Errorf("%s: l.ch = '%c', want '\n'", name, l.ch)
				}
			})
		}
	}
}

func TestReadShowdownFileCases(t *testing.T) {
	tests := map[string]struct {
		input             string
		wantError         bool
		wantErrorContains string
		wantName          string
		wantHP            int
		checkHP           bool
		wantStatus        string
		checkStatus       bool
		wantMoves         int
	}{
		"rejects malformed nature line": {
			input:             "Pikachu\nLevel: 25\nJolly\nAbility: Static\n- Thunderbolt\n",
			wantError:         true,
			wantErrorContains: "nature line not enough fields",
		},
		"rejects unknown status": {
			input:             "Pikachu\nLevel: 25\nJolly Nature\nAbility: Static\nStatus: Confused\n- Thunderbolt\n",
			wantError:         true,
			wantErrorContains: "unrecognized status",
		},
		"allows minimal pokemon without status or hp": {
			input:       "Pikachu\nLevel: 25\nJolly Nature\nAbility: Static\n- Thunderbolt\n- Quick Attack\n- Tail Whip\n- Slam\n",
			wantName:    "Pikachu",
			wantHP:      -1,
			checkHP:     true,
			wantStatus:  "",
			checkStatus: true,
			wantMoves:   4,
		},
		"rejects missing level line": {
			input:     "Horsea\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n",
			wantError: true,
		},
		"rejects invalid nature": {
			input:     "Horsea\nLevel: 17\nBad Nature\nAbility: Swift Swim\n- Bubble Beam\n",
			wantError: true,
		},
		"rejects invalid status": {
			input:     "Horsea\nLevel: 17\nModest Nature\nAbility: Swift Swim\nStatus> bad\n- Bubble Beam\n",
			wantError: true,
		},
		"accepts valid multi-move party": {
			input:     "Horsea\nLevel: 17\nModest Nature\nAbility: Swift Swim\n- Bubble Beam\n- Twister\n",
			wantName:  "Horsea",
			wantMoves: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "party.txt")
			if err := os.WriteFile(path, []byte(tc.input), 0o644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			parsed, err := ReadShowdownFile(path)
			if tc.wantError {
				if err == nil {
					t.Fatal("ReadShowdownFile() returned nil error")
				}
				if tc.wantErrorContains != "" && !strings.Contains(err.Error(), tc.wantErrorContains) {
					t.Fatalf("ReadShowdownFile() error = %v, want it to contain %q", err, tc.wantErrorContains)
				}
				return
			}
			if err == nil {
				if len(parsed) == 0 {
					t.Fatal("expected at least one parsed pokemon")
				}
			} else {
				t.Fatalf("ReadShowdownFile() returned unexpected error: %v", err)
			}
			if len(parsed) != 1 {
				t.Fatalf("ReadShowdownFile() returned %d entries, want 1", len(parsed))
			}
			if parsed[0].Name != tc.wantName {
				t.Errorf("pokemon name = %q, want %q", parsed[0].Name, tc.wantName)
			}
			if tc.checkHP && parsed[0].HP != tc.wantHP {
				t.Errorf("pokemon HP = %d, want %d", parsed[0].HP, tc.wantHP)
			}
			if tc.checkStatus && parsed[0].Status != tc.wantStatus {
				t.Errorf("pokemon status = %q, want %q", parsed[0].Status, tc.wantStatus)
			}
			if len(parsed[0].Moves) != tc.wantMoves {
				t.Errorf("pokemon has %d moves, want %d", len(parsed[0].Moves), tc.wantMoves)
			}
		})
	}
}
