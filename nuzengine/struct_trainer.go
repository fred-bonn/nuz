package nuzengine

type trainer struct {
	PokemonParty []*Pokemon
	Player       bool
	AI           ai
	FieldEffects map[fieldEffect]int
	lost         bool
}

func (t *trainer) canReplace(bs battleState) bool {
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
