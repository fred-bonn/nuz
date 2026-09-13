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
	SpecialClass
	statusClass
)

func stringToMoveClass(s string) moveClass {
	switch s {
	case "physical":
		return physicalClass
	case "special":
		return SpecialClass
	case "status":
		return statusClass
	default:
		return noneClass
	}
}

type Move struct {
	Name          string
	Type          pokemonType
	Power         int
	Accuracy      int
	PP            int
	MaxPP         int
	Class         moveClass
	Priority      int
	CritRate      int
	Drain         int
	Heal          int
	FlinchChance  int
	Contact       bool
	Ailment       ailmentState
	AilmentChance int
	MaxHits       int
	MinHits       int
	MaxTurns      int
	MinTurns      int
	StatChance    int
	StatChanges   map[string]int
	Target        string
	Category      string
}

var contactMoves map[string]any

func ToMove(mj pokeapi.MoveJSON) (Move, error) {
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
		return Move{}, fmt.Errorf("%s is not a valid move class for %s", mj.DamageClass.Name, mj.Name)
	}

	moveType := stringToPokemonType(mj.Type.Name)
	if moveType == noType {
		return Move{}, fmt.Errorf("%s is not a valid type for %s", mj.Type.Name, mj.Name)
	}

	ailment := stringToAilmentState(mj.Meta.Ailment.Name)
	if mj.Meta.Ailment.Name != "" && mj.Meta.Ailment.Name != "none" && ailment == noneAilment {
		return Move{}, fmt.Errorf("%s is not a valid ailment for %s", mj.Meta.Ailment.Name, mj.Name)
	}

	return Move{
		Name:          mj.Name,
		Type:          moveType,
		Power:         mj.Power,
		Accuracy:      mj.Accuracy,
		PP:            mj.PP,
		MaxPP:         mj.PP,
		Class:         class,
		Priority:      mj.Priority,
		CritRate:      mj.Meta.CritRate,
		Drain:         mj.Meta.Drain,
		Heal:          mj.Meta.Heal,
		FlinchChance:  mj.Meta.FlinchChance,
		Contact:       isContact,
		Ailment:       ailment,
		AilmentChance: ailmentChance,
		MaxHits:       mj.Meta.MaxHits,
		MinHits:       mj.Meta.MinHits,
		MaxTurns:      mj.Meta.MaxTurns,
		MinTurns:      mj.Meta.MinTurns,
		StatChance:    statChance,
		StatChanges:   statChanges,
		Target:        mj.Target.Name,
		Category:      mj.Meta.Category.Name,
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
