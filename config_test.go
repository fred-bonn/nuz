package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fred-bonn/nuz/internal/engine"
	"github.com/fred-bonn/nuz/internal/parser"
	"github.com/fred-bonn/nuz/internal/pokeapi"
)

func TestConfigValidateInputLoadsShowdownPartyFromLocalData(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}

	party, err := cfg.validateInput("showdown_demo_files/player.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}
	if len(party) != 5 {
		t.Fatalf("unexpected party size: got %d want 5", len(party))
	}
	if party[0].Base.Name != "horsea" {
		t.Fatalf("unexpected first pokemon name: %q", party[0].Base.Name)
	}
	if len(party[0].Moves) != 2 {
		t.Fatalf("unexpected move count for horsea: got %d want 2", len(party[0].Moves))
	}
	if party[0].Moves[0].Name != "bubble beam" {
		t.Fatalf("unexpected first move name: %q", party[0].Moves[0].Name)
	}
}

func TestConfigNameHelpersAndHiddenPower(t *testing.T) {
	if got := apiName("Mr. Mime"); got != "mr-mime" {
		t.Fatalf("apiName returned %q, want %q", got, "mr-mime")
	}
	if got := engine.CleanName("Mr. Mime"); got != "mr. mime" {
		t.Fatalf("cleanName returned %q, want %q", got, "mr. mime")
	}

	move, err := generateHiddenPower("hidden-power-fire")
	if err != nil {
		t.Fatalf("generateHiddenPower returned unexpected error: %v", err)
	}
	if move.Name != "hidden power" || move.Type != engine.FireType || move.Power != 60 || move.Accuracy != 100 {
		t.Fatalf("unexpected generated hidden power: %+v", move)
	}

	if _, err := generateHiddenPower("hidden-power-unknown"); err == nil {
		t.Fatal("generateHiddenPower should reject invalid hidden power types")
	}
}

func TestParserReadShowdownFileRejectsMalformedInput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.txt")
	content := "Pikachu\nJolly Nature\nAbility: Static\n- Tackle\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if _, err := parser.ReadShowdownFile(path); err == nil {
		t.Fatal("ReadShowdownFile should reject malformed party data")
	}
}
