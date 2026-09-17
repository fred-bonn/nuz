package nuzengine

type trainer struct {
	pokemonParty []*pokemon
	player       bool
	ai           ai
	fieldEffects map[fieldEffect]int
	lost         bool
}

func (t *trainer) canReplace(bs battleState) bool {
	count := 0
	for _, mon := range t.pokemonParty {
		if !mon.fainted {
			count++
		}
		if count > 1 {
			return true
		}
	}
	return false
}
