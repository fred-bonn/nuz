package nuzengine

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type moveClass int

const (
	noneClass moveClass = iota
	physicalClass
	specialClass
	statusClass
)

func stringToMoveClass(s string) moveClass {
	switch s {
	case "physical":
		return physicalClass
	case "special":
		return specialClass
	case "status":
		return statusClass
	default:
		return noneClass
	}
}

type move struct {
	name          string
	moveType      pokemonType
	power         int
	accuracy      int
	pp            int
	maxPP         int
	class         moveClass
	priority      int
	critRate      int
	drain         int
	heal          int
	flinchChance  int
	isContact     bool
	ailment       ailmentState
	ailmentChance int
	maxHits       int
	minHits       int
	maxTurns      int
	minTurns      int
	statChance    int
	statChanges   map[string]int
	target        string
	category      string
}

var contactMoves map[string]any

func toMove(mj pokeapi.MoveJSON) (move, error) {
	isContact := false
	statChanges := make(map[string]int)
	for _, sc := range mj.StatChanges {
		statChanges[sc.Stat.Name] = sc.Change
	}
	var statChance int
	var ailmentChance int
	if mj.DamageClass.Name == "status" {
		statChance = 100
		ailmentChance = 100
	} else {
		statChance = mj.Meta.StatChance
		ailmentChance = mj.Meta.AilmentChance
	}

	if contactMoves == nil {
		initContactMoves()
	}

	_, isContact = contactMoves[mj.Name]

	class := stringToMoveClass(mj.DamageClass.Name)
	if class == noneClass {
		return move{}, fmt.Errorf("%s is not a valid move class for %s", mj.DamageClass.Name, mj.Name)
	}

	moveType := stringToPokemonType(mj.Type.Name)
	if moveType == noType {
		return move{}, fmt.Errorf("%s is not a valid type for %s", mj.Type.Name, mj.Name)
	}

	ailment := stringToAilmentState(mj.Meta.Ailment.Name)
	if mj.Meta.Ailment.Name != "" && mj.Meta.Ailment.Name != "none" && ailment == noneAilment {
		return move{}, fmt.Errorf("%s is not a valid ailment for %s", mj.Meta.Ailment.Name, mj.Name)
	}

	return move{
		name:          mj.Name,
		moveType:      moveType,
		power:         mj.Power,
		accuracy:      mj.Accuracy,
		pp:            mj.PP,
		maxPP:         mj.PP,
		class:         class,
		priority:      mj.Priority,
		critRate:      mj.Meta.CritRate,
		drain:         mj.Meta.Drain,
		heal:          mj.Meta.Heal,
		flinchChance:  mj.Meta.FlinchChance,
		isContact:     isContact,
		ailment:       ailment,
		ailmentChance: ailmentChance,
		maxHits:       mj.Meta.MaxHits,
		minHits:       mj.Meta.MinHits,
		maxTurns:      mj.Meta.MaxTurns,
		minTurns:      mj.Meta.MinTurns,
		statChance:    statChance,
		statChanges:   statChanges,
		target:        mj.Target.Name,
		category:      mj.Meta.Category.Name,
	}, nil
}

func initContactMoves() error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("unable to determine source file location")
	}
	dir := filepath.Dir(filename)
	filePath := filepath.Join(dir, "contact_moves.json")

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var moves []string
	err = json.Unmarshal(data, &moves)
	if err != nil {
		return err
	}

	contactMoves = make(map[string]any, len(moves))
	for _, move := range moves {
		contactMoves[move] = struct{}{}
	}

	return nil
}
