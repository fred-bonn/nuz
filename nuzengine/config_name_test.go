package nuzengine

import (
	"strings"
	"testing"
)

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Pikachu", false},
		{"Mr. Mime", false},
		{"Mime Jr.", false},
		{"Farfetch'd", false},
		{"Farfetch’d", false},
		{"Ho-Oh", false},
		{"Porygon2", false},
		{"Kommo-o", false},
		{"Tapu Koko", false},
		{"Will-O-Wisp", false},
		{"Hidden Power Fire", false},
		{"", true},
		{".Mime", true},
		{"-Mime", true},
		{"'Mime", true},
		{"2Mime", true},
		{" Mime", true},
		{"../etc/passwd", true},
		{"a/b", true},
		{"a\\b", true},
		{"~root", true},
		{"a~", true},
		{"a%2e", true},
		{"a?b", true},
		{"a#b", true},
		{"a:b", true},
		{"a\x00b", true},
		{"a\nb", true},
		{"Flabébé", true},
		{strings.Repeat("a", maxNameLength), false},
		{strings.Repeat("a", maxNameLength+1), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateName(tt.name); (err != nil) != tt.wantErr {
				t.Fatalf("validateName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateInputRejectsBadNamesBeforeLoading(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{"pokemon", "Pika~chu\n- Thunderbolt\n"},
		{"move", "Pikachu\n- Thunder~bolt\n"},
		{"non ascii pokemon", "Pikachü\n- Thunderbolt\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config{}
			// A zero-value client would hit the network if loading were reached.
			if _, err := cfg.validateInput(tt.content); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestAPISlugPattern(t *testing.T) {
	for name, want := range map[string]bool{
		"pikachu": true, "mr-mime": true, "porygon2": true, "ho-oh": true,
		"": false, "-a": false, "a/b": false, "a.b": false, "A": false, "a b": false, "../a": false,
	} {
		if got := apiSlugPattern.MatchString(name); got != want {
			t.Errorf("apiSlugPattern(%q) = %v, want %v", name, got, want)
		}
	}
	cfg := &config{}
	if _, err := cfg.loadPokemon("../x"); err == nil {
		t.Error("loadPokemon accepted a path")
	}
	if _, err := cfg.loadMove("a/b"); err == nil {
		t.Error("loadMove accepted a path")
	}
}
