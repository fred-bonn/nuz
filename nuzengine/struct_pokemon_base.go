package nuzengine

import (
	"fmt"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type basePokemon struct {
	Id     int
	Name   string
	Height int
	Weight int
	Types  []pokemonType
	Stats  map[string]int
}

func ToPokemon(pj pokeapi.PokemonJSON) (basePokemon, error) {
	types := make([]pokemonType, len(pj.Types))
	for i, t := range pj.Types {
		types[i] = stringToPokemonType(t.Type.Name)
		if types[i] == noType {
			return basePokemon{}, fmt.Errorf("%s is not a valid type for %s", t.Type.Name, pj.Name)
		}
	}

	stats := make(map[string]int, 6)
	for _, s := range pj.Stats {
		stats[s.Stat.Name] = s.BaseStat
	}

	return basePokemon{
		Id:     pj.Id,
		Name:   pj.Name,
		Height: pj.Height,
		Weight: pj.Weight,
		Types:  types,
		Stats:  stats,
	}, nil
}
