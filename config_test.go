package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fred-bonn/nuz/internal/parser"
	"github.com/fred-bonn/nuz/internal/pokeapi"
)

func TestConfigValidateInputLoadsShowdownPartyFromLocalData(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}

	party, err := cfg.validateInput("data/player.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}
	if len(party) != 5 {
		t.Fatalf("unexpected party size: got %d want 5", len(party))
	}
	if party[0].base.Name != "horsea" {
		t.Fatalf("unexpected first pokemon name: %q", party[0].base.Name)
	}
	if len(party[0].moves) != 2 {
		t.Fatalf("unexpected move count for horsea: got %d want 2", len(party[0].moves))
	}
	if party[0].moves[0].Name != "bubble beam" {
		t.Fatalf("unexpected first move name: %q", party[0].moves[0].Name)
	}
}

func TestConfigLoadsMunchlaxGluttonySalacBerry(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}

	party, err := cfg.validateInput("data/rnb_trainer_1.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}

	var munchlax *pokemon
	for _, mon := range party {
		if mon.base.Name == "munchlax" {
			munchlax = mon
			break
		}
	}
	if munchlax == nil {
		t.Fatal("munchlax was not loaded")
	}
	if munchlax.ability != gluttonyAbility {
		t.Fatalf("munchlax ability = %s, want gluttony", munchlax.ability)
	}
	if munchlax.item == nil || munchlax.item.state != salacBerry {
		t.Fatalf("munchlax item = %v, want salac berry", munchlax.item)
	}

	munchlax.hp = munchlax.maxHP() / 2
	munchlax.checkItemTrigger(true, nil)
	if !munchlax.item.consumed {
		t.Fatal("munchlax did not consume Salac Berry at 50% HP")
	}
	if munchlax.stages[speed] != 1 {
		t.Fatalf("munchlax speed stage = %d, want 1", munchlax.stages[speed])
	}
}

func TestConfigLoadedDwebbleActivatesSalacAfterDamage(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}

	party, err := cfg.validateInput("data/rnb_trainer_2.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}
	if len(party) != 1 {
		t.Fatalf("loaded party size = %d, want 1", len(party))
	}

	dwebble := party[0]
	if dwebble.item == nil || dwebble.item.state != salacBerry {
		t.Fatalf("dwebble item = %v, want salac berry", dwebble.item)
	}
	dwebble.hp = dwebble.maxHP()
	dwebble.changeHpBy(-(dwebble.maxHP() - 1))

	if dwebble.hp != 1 {
		t.Fatalf("dwebble HP = %d, want 1", dwebble.hp)
	}
	if !dwebble.item.consumed {
		t.Fatal("dwebble did not consume Salac Berry after reaching 1 HP")
	}
	if dwebble.stages[speed] != 1 {
		t.Fatalf("dwebble speed stage = %d, want 1", dwebble.stages[speed])
	}
}

func TestClonedDwebbleActivatesSalacAfterDamage(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}
	party, err := cfg.validateInput("data/rnb_trainer_2.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}

	dwebble := clonePokemon(party[0])
	dwebble.hp = dwebble.maxHP()
	dwebble.changeHpBy(-(dwebble.maxHP() - 1))

	if !dwebble.item.consumed {
		t.Fatal("cloned dwebble did not consume Salac Berry after reaching 1 HP")
	}
	if dwebble.stages[speed] != 1 {
		t.Fatalf("cloned dwebble speed stage = %d, want 1", dwebble.stages[speed])
	}
}

func TestClonePokemonDeepCopiesOwnedState(t *testing.T) {
	cfg := &config{client: pokeapi.NewClient()}
	party, err := cfg.validateInput("data/rnb_trainer_1.txt")
	if err != nil {
		t.Fatalf("validateInput returned unexpected error: %v", err)
	}

	original := party[0]
	original.lockedMove = original.moves[0]
	clone := clonePokemon(original)

	clone.base.Types[0] = noType
	clone.base.Stats["hp"]++
	clone.ivs[0]--
	clone.stats[0]--
	clone.stages[0]++
	clone.moves[0].PP--
	clone.moves[0].StatChanges["attack"]++
	clone.ailments[paralysisAilment] = &ailment{state: paralysisAilment, turns: 2}
	clone.item.consumed = true
	clone.lockedMove.PP--

	if original.base.Types[0] == noType {
		t.Fatal("clone shares BasePokemon.Types with original")
	}
	if original.base.Stats["hp"] == clone.base.Stats["hp"] {
		t.Fatal("clone shares BasePokemon.Stats with original")
	}
	if original.ivs[0] == clone.ivs[0] || original.stats[0] == clone.stats[0] || original.stages[0] == clone.stages[0] {
		t.Fatal("clone shares a Pokemon slice with original")
	}
	if original.moves[0] == clone.moves[0] {
		t.Fatal("clone shares Move pointers with original")
	}
	if original.moves[0].PP == clone.moves[0].PP || original.moves[0].StatChanges["attack"] == clone.moves[0].StatChanges["attack"] {
		t.Fatal("clone shares mutable Move state with original")
	}
	if original.ailments[paralysisAilment] != nil {
		t.Fatal("clone shares ailments map with original")
	}
	if original.item.consumed {
		t.Fatal("clone shares item state with original")
	}
	if original.lockedMove != original.moves[0] || clone.lockedMove != clone.moves[0] {
		t.Fatal("clone did not preserve locked-move aliasing")
	}
}

func TestConfigNameHelpersAndHiddenPower(t *testing.T) {
	if got := apiName("Mr. Mime"); got != "mr-mime" {
		t.Fatalf("apiName returned %q, want %q", got, "mr-mime")
	}
	if got := cleanName("Mr. Mime"); got != "mr. mime" {
		t.Fatalf("cleanName returned %q, want %q", got, "mr. mime")
	}

	move, err := generateHiddenPower("hidden-power-fire")
	if err != nil {
		t.Fatalf("generateHiddenPower returned unexpected error: %v", err)
	}
	if move.Name != "hidden power" || move.Type != fireType || move.Power != 60 || move.Accuracy != 100 {
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
