package nuzengine

import (
	"slices"
)

// PokemonSnapshot is a read-only view of a Pokémon for display.
type PokemonSnapshot struct {
	Name     string   `json:"name"`
	HP       int      `json:"hp"`
	MaxHP    int      `json:"maxHp"`
	Fainted  bool     `json:"fainted"`
	Ailments []string `json:"ailments"`
}

// SideSnapshot is one trainer's Pokémon: those on the field and the rest of the party.
type SideSnapshot struct {
	Active []PokemonSnapshot `json:"active"`
	Bench  []PokemonSnapshot `json:"bench"`
}

// BattleSnapshot is a read-only view of the battle state.
type BattleSnapshot struct {
	Player   SideSnapshot `json:"player"`
	Opponent SideSnapshot `json:"opponent"`
}

// Snapshot returns the current state of the battle.
func (ib *InteractiveBattle) Snapshot() BattleSnapshot {
	return snapshotSlots(ib.bs.getAllSlots(), ib.bs.getPlayerTrainer())
}

func snapshotSlots(slots []*slot, player *trainer) BattleSnapshot {
	snap := BattleSnapshot{
		Player:   SideSnapshot{Active: []PokemonSnapshot{}, Bench: []PokemonSnapshot{}},
		Opponent: SideSnapshot{Active: []PokemonSnapshot{}, Bench: []PokemonSnapshot{}},
	}
	done := map[*trainer]bool{}
	for _, s := range slots {
		side := &snap.Opponent
		if s.Trainer == player {
			side = &snap.Player
		}
		side.Active = append(side.Active, snapshotPokemon(s.mon))
		if done[s.Trainer] {
			continue
		}
		done[s.Trainer] = true
		for _, mon := range s.Trainer.pokemonParty {
			if !isActive(slots, s.Trainer, mon) {
				side.Bench = append(side.Bench, snapshotPokemon(mon))
			}
		}
	}
	return snap
}

func isActive(slots []*slot, t *trainer, mon *pokemon) bool {
	return slices.ContainsFunc(slots, func(s *slot) bool { return s.Trainer == t && s.mon == mon })
}

func snapshotPokemon(mon *pokemon) PokemonSnapshot {
	ailments := []string{}
	states := make([]ailmentState, 0, len(mon.ailments))
	for state := range mon.ailments {
		states = append(states, state)
	}
	slices.Sort(states)
	for _, state := range states {
		ailments = append(ailments, state.String())
	}
	return PokemonSnapshot{
		Name:     mon.base.Name,
		HP:       mon.hp,
		MaxHP:    mon.MaxHP(),
		Fainted:  mon.fainted,
		Ailments: ailments,
	}
}
