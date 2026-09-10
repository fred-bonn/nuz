package engine

func cloneFieldEffects(src map[FieldEffect]int) map[FieldEffect]int {
	if src == nil {
		return make(map[FieldEffect]int)
	}
	clone := make(map[FieldEffect]int, len(src))
	for k, v := range src {
		clone[k] = v
	}
	return clone
}

func ClonePokemonParty(src []*Pokemon) []*Pokemon {
	clone := make([]*Pokemon, len(src))
	for i, mon := range src {
		clone[i] = ClonePokemon(mon)
	}
	return clone
}

func ClonePokemon(p *Pokemon) *Pokemon {
	if p == nil {
		return nil
	}

	copyP := *p
	copyP.Base.Types = append([]pokemonType(nil), p.Base.Types...)
	copyP.Base.Stats = make(map[string]int, len(p.Base.Stats))
	for stat, value := range p.Base.Stats {
		copyP.Base.Stats[stat] = value
	}
	copyP.IVs = append([]int(nil), p.IVs...)
	copyP.Moves = make([]*Move, len(p.Moves))
	moveCopies := make(map[*Move]*Move, len(p.Moves))
	for i, move := range p.Moves {
		copyP.Moves[i] = cloneMove(move)
		if move != nil {
			moveCopies[move] = copyP.Moves[i]
		}
	}
	copyP.Stats = append([]int(nil), p.Stats...)
	copyP.Stages = append([]int(nil), p.Stages...)
	if p.LockedMove != nil {
		if move, ok := moveCopies[p.LockedMove]; ok {
			copyP.LockedMove = move
		} else {
			copyP.LockedMove = cloneMove(p.LockedMove)
		}
	}

	copyP.Ailments = make(map[ailmentState]*ailment, len(p.Ailments))
	for state, ailment := range p.Ailments {
		if ailment == nil {
			continue
		}
		copyA := *ailment
		copyP.Ailments[state] = &copyA
	}

	if p.Item != nil {
		copyP.Item, _ = RegisterItem(p.Item.State, &copyP)
		copyP.Item.Consumed = p.Item.Consumed
	}

	return &copyP
}

func cloneMove(move *Move) *Move {
	if move == nil {
		return nil
	}

	copyMove := *move
	copyMove.StatChanges = make(map[string]int, len(move.StatChanges))
	for stat, change := range move.StatChanges {
		copyMove.StatChanges[stat] = change
	}
	return &copyMove
}
