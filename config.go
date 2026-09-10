package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fred-bonn/nuz/internal/engine"
	"github.com/fred-bonn/nuz/internal/parser"
	"github.com/fred-bonn/nuz/internal/pokeapi"
)

type config struct {
	client pokeapi.Client
}

func (cfg *config) validateInput(trainerPath string) ([]*engine.Pokemon, error) {
	trainerFullPath, err := filepath.Abs(trainerPath)
	if err != nil {
		return nil, fmt.Errorf("failed getting absolute path: %w", err)
	}

	trainerPokemon, err := parser.ReadShowdownFile(trainerFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed reading showdown file: %w", err)
	}
	if len(trainerPokemon) == 0 {
		return nil, err
	}

	trainerParty, err := cfg.loadShowdown(trainerPokemon)
	if err != nil {
		return nil, fmt.Errorf("failed loading showdown file: %w", err)
	}

	return trainerParty, nil
}

// validateInputContent parses Showdown-style party text embedded directly in a saved policy,
// rather than reading it from a file on disk.
func (cfg *config) validateInputContent(content string) ([]*engine.Pokemon, error) {
	trainerPokemon, err := parser.ParseShowdown(content)
	if err != nil {
		return nil, fmt.Errorf("failed parsing showdown content: %w", err)
	}
	if len(trainerPokemon) == 0 {
		return nil, fmt.Errorf("no pokemon parsed from showdown content")
	}

	trainerParty, err := cfg.loadShowdown(trainerPokemon)
	if err != nil {
		return nil, fmt.Errorf("failed loading showdown content: %w", err)
	}

	return trainerParty, nil
}

func (cfg *config) loadShowdown(mons []parser.ParsedPokemon) ([]*engine.Pokemon, error) {
	var res []*engine.Pokemon

	for _, mon := range mons {
		var moves []*engine.Move

		basePokemon, err := cfg.loadPokemon(apiName(mon.Name))
		if err != nil {
			return nil, err
		}

		basePokemon.Name = engine.CleanName(mon.Name)

		for _, moveName := range mon.Moves {
			baseMove, err := cfg.loadMove(apiName(moveName))
			if err != nil {
				return nil, err
			}

			if mb, ok := engine.MoveBalanceMap[baseMove.Name]; ok {
				mb.Apply(&baseMove)
			}

			baseMove.Name = engine.CleanName(moveName)

			moves = append(moves, &baseMove)
		}

		finalPokemon, err := engine.InitPokemon(basePokemon, mon.Level, mon.IVs, mon.Nature, moves, mon.HP, engine.StringToAilmentState(mon.Status))
		if err != nil {
			return nil, err
		}

		item, err := engine.RegisterItem(engine.StringToItemState(strings.ToLower(mon.Item)), &finalPokemon)
		if err != nil {
			return nil, err
		}
		finalPokemon.Item = item

		finalPokemon.Ability = engine.StringToAbility(strings.ToLower(mon.Ability))
		if finalPokemon.Ability == engine.NoneAbility {
			return nil, fmt.Errorf("%s is not a valid ability for %s", strings.ToLower(mon.Ability), mon.Name)
		}

		res = append(res, &finalPokemon)
	}

	return res, nil
}

func (cfg *config) loadPokemon(name string) (engine.BasePokemon, error) {
	var p engine.BasePokemon

	data, err := os.ReadFile(fmt.Sprintf("data/pokemon/%s.json", name))
	if err == nil {
		// If the file exists and is read successfully, unmarshal it into a Pokemon struct
		err = json.Unmarshal(data, &p)
		if err != nil {
			return engine.BasePokemon{}, fmt.Errorf("failed unmarshaling '%s' Pokemon data: %w", name, err)
		}

		return p, nil
	}

	// Otherwise, fetch the Pokemon data from the API
	pokemonJSON, err := cfg.client.FetchPokemon(name)
	if err != nil {
		return engine.BasePokemon{}, fmt.Errorf("failed fetching Pokemon '%s': %w", name, err)
	}
	fmt.Printf("Fetched '%s' from API\n", name)

	p, err = engine.ToPokemon(pokemonJSON)
	if err != nil {
		return engine.BasePokemon{}, err
	}

	// Save the fetched Pokemon data to a file for future use
	data, err = json.Marshal(p)
	if err != nil {
		return engine.BasePokemon{}, fmt.Errorf("failed marshaling Pokemon JSON data '%s' to file: %w", name, err)
	}
	writeToFile(fmt.Sprintf("data/pokemon/%s.json", name), data)

	return p, nil
}

func (cfg *config) loadMove(name string) (engine.Move, error) {
	var m engine.Move

	if strings.HasPrefix(name, "hidden-power") {
		// If the move is Hidden Power, generate it
		var err error
		m, err = generateHiddenPower(name)
		if err != nil {
			return engine.Move{}, err
		}
		return m, nil
	}

	data, err := os.ReadFile(fmt.Sprintf("data/moves/%s.json", name))
	if err == nil {
		// If the file exists and is read successfully, unmarshal it into a Move struct
		err = json.Unmarshal(data, &m)
		if err != nil {
			return engine.Move{}, fmt.Errorf("failed unmarshaling Move '%s' data: %w", name, err)
		}

		return m, nil
	}

	// Otherwise, fetch the Move data from the API
	moveJson, err := cfg.client.FetchMove(name)
	if err != nil {
		return engine.Move{}, fmt.Errorf("failed fetching Move '%s': %w", name, err)
	}
	fmt.Printf("Fetched '%s' from API\n", name)

	m, err = engine.ToMove(moveJson)
	if err != nil {
		return engine.Move{}, err
	}

	// Save the fetched Move data using the internal Move struct to a file for future use
	data, err = json.Marshal(m)
	if err != nil {
		return m, fmt.Errorf("failed marshaling Move JSON data '%s' to file: %w", name, err)
	}
	writeToFile(fmt.Sprintf("data/moves/%s.json", name), data)

	return m, nil
}

func generateHiddenPower(name string) (engine.Move, error) {
	parts := strings.Split(name, "-")
	if len(parts) != 3 {
		return engine.Move{}, fmt.Errorf("type not specified for hidden power")
	}

	moveType := engine.StringToPokemonType(parts[2])
	if moveType == engine.NoType {
		return engine.Move{}, fmt.Errorf("%s is not a valid type for %s", parts[2], name)
	}

	move := engine.Move{
		Name:     "hidden power",
		Type:     moveType,
		Power:    60,
		Accuracy: 100,
		Class:    engine.SpecialClass,
	}

	return move, nil
}

func writeToFile(filename string, data []byte) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}
	return os.WriteFile(filename, data, 0644)
}

func apiName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "’", "")
	return name
}
