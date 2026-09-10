package engine

import (
	"fmt"

	"github.com/fred-bonn/nuz/internal/pokeapi"
)

type BasePokemon struct {
	Id     int
	Name   string
	Height int
	Weight int
	Types  []pokemonType
	Stats  map[string]int
}

func ToPokemon(pj pokeapi.PokemonJSON) (BasePokemon, error) {
	types := make([]pokemonType, len(pj.Types))
	for i, t := range pj.Types {
		types[i] = StringToPokemonType(t.Type.Name)
		if types[i] == NoType {
			return BasePokemon{}, fmt.Errorf("%s is not a valid type for %s", t.Type.Name, pj.Name)
		}
	}

	stats := make(map[string]int, 6)
	for _, s := range pj.Stats {
		stats[s.Stat.Name] = s.BaseStat
	}

	return BasePokemon{
		Id:     pj.Id,
		Name:   pj.Name,
		Height: pj.Height,
		Weight: pj.Weight,
		Types:  types,
		Stats:  stats,
	}, nil
}
