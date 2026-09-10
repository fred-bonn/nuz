package engine

type Trainer struct {
	PokemonParty []*Pokemon
	Player       bool
	AI           AI
	FieldEffects map[FieldEffect]int
	lost         bool
}

func (t *Trainer) canReplace(bs BattleState) bool {
	count := 0
	for _, mon := range t.PokemonParty {
		if !mon.fainted {
			count++
		}
		if count > 1 {
			return true
		}
	}
	return false
}
