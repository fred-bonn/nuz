package nuzengine

import "math/rand"

type ai interface {
	evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int)
	evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon
	shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool
}

type randomAi struct{}

func (ra randomAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	return actions[rand.Intn(len(actions))], 1
}

func (ra randomAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	return mons[rand.Intn(len(mons))]
}

func (ra randomAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return roll(1, 10)
}

func chooseNextAction(bs battleState, slot *slot, party []*pokemon, decisionAI ai) action {
	if slot.invulnerableAction != nil {
		return slot.invulnerableAction
	}

	possibleActions := make([]*moveAction, 0)
	for _, opponentSlot := range bs.getOtherSlots(slot) {
		if slot.mon.lockedMove != nil && slot.mon.lockedMove.pp > 0 {
			possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: slot.mon.lockedMove})
			continue
		}
		for _, move := range slot.mon.moves {
			if move.pp <= 0 || (slot.mon.item.State == assaultVest && move.class != statusClass) {
				continue
			}
			possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: move})
		}
	}

	if len(possibleActions) == 0 {
		for _, opponentSlot := range bs.getOtherSlots(slot) {
			if opponentSlot.Trainer != slot.Trainer {
				possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: getStruggleMove()})
			}
		}
	}

	chosenAction, score := decisionAI.evaluateActions(bs, slot, possibleActions)
	if slot.mon.item.State.isChoice() {
		slot.mon.lockedMove = chosenAction.move
	}
	if !canReplace(party) || slot.isTrapped() || !decisionAI.shouldSwitch(bs, slot, score, party) {
		return chosenAction
	}

	var possibleMons []*pokemon
	for _, mon := range party {
		if mon != slot.mon && !mon.fainted && !bs.getActions().containstSwitchTo(mon) {
			possibleMons = append(possibleMons, mon)
		}
	}
	if len(possibleMons) == 0 {
		return chosenAction
	}
	chosenMon := decisionAI.evaluteSwitchIns(bs, possibleMons, bs.getOpponentSlot(slot))
	return &switchAction{oldSlot: slot, new: chosenMon}
}

func chooseSwitchIn(bs battleState, slot *slot, party []*pokemon, decisionAI ai) *pokemon {
	var possibleMons []*pokemon
	for _, mon := range party {
		if mon != slot.mon && !mon.fainted {
			possibleMons = append(possibleMons, mon)
		}
	}
	if len(possibleMons) == 0 {
		return nil
	}

	return decisionAI.evaluteSwitchIns(bs, possibleMons, bs.getOpponentSlot(slot))
}

func canReplace(party []*pokemon) bool {
	count := 0
	for _, mon := range party {
		if !mon.fainted {
			count++
		}
		if count > 1 {
			return true
		}
	}
	return false
}
