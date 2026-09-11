package engine

import "math/rand"

type AI interface {
	evaluateActions(bs BattleState, slot *slot, actions []*moveAction) (*moveAction, int)
	evaluteSwitchIns(bs BattleState, mons []*Pokemon, opponentSlot *slot) *Pokemon
	shouldSwitch(bs BattleState, slot *slot, score int, party []*Pokemon) bool
}

type RandomAi struct{}

func (ra RandomAi) evaluateActions(bs BattleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	return actions[rand.Intn(len(actions))], 1
}

func (ra RandomAi) evaluteSwitchIns(bs BattleState, mons []*Pokemon, opponentSlot *slot) *Pokemon {
	return mons[rand.Intn(len(mons))]
}

func (ra RandomAi) shouldSwitch(bs BattleState, slot *slot, score int, party []*Pokemon) bool {
	return roll(1, 10)
}

func chooseNextAction(bs BattleState, slot *slot, party []*Pokemon, decisionAI AI) action {
	if slot.invulnerableAction != nil {
		return slot.invulnerableAction
	}

	possibleActions := make([]*moveAction, 0)
	for _, opponentSlot := range bs.getOtherSlots(slot) {
		if slot.mon.LockedMove != nil && slot.mon.LockedMove.PP > 0 {
			possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: slot.mon.LockedMove})
			continue
		}
		for _, move := range slot.mon.Moves {
			if move.PP <= 0 || (slot.mon.Item.State == assaultVest && move.Class != statusClass) {
				continue
			}
			possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: move})
		}
	}

	if len(possibleActions) == 0 {
		for _, opponentSlot := range bs.getOtherSlots(slot) {
			if opponentSlot.Trainer != slot.Trainer {
				possibleActions = append(possibleActions, &moveAction{userSlot: slot, targetSlot: opponentSlot, move: &struggleMove})
			}
		}
	}

	chosenAction, score := decisionAI.evaluateActions(bs, slot, possibleActions)
	if slot.mon.Item.State.isChoice() {
		slot.mon.LockedMove = chosenAction.move
	}
	if !canReplace(party) || slot.isTrapped() || !decisionAI.shouldSwitch(bs, slot, score, party) {
		return chosenAction
	}

	var possibleMons []*Pokemon
	for _, mon := range party {
		if mon != slot.mon && !mon.fainted && !bs.getActions().containstSwitchTo(mon) {
			possibleMons = append(possibleMons, mon)
		}
	}
	if len(possibleMons) == 0 {
		return chosenAction
	}
	chosenMon := decisionAI.evaluteSwitchIns(bs, possibleMons, bs.getOpponentSlot(slot))
	if la, ok := decisionAI.(*LearningAI); ok {
		la.recordStateAction(discretizeBattleState(bs).key(), actionKeyForSwitch(chosenMon))
	}
	return &switchAction{oldSlot: slot, new: chosenMon}
}

func chooseSwitchIn(bs BattleState, slot *slot, party []*Pokemon, decisionAI AI) *Pokemon {
	var possibleMons []*Pokemon
	for _, mon := range party {
		if mon != slot.mon && !mon.fainted {
			possibleMons = append(possibleMons, mon)
		}
	}
	if len(possibleMons) == 0 {
		return nil
	}
	chosenMon := decisionAI.evaluteSwitchIns(bs, possibleMons, bs.getOpponentSlot(slot))
	if la, ok := decisionAI.(*LearningAI); ok {
		la.recordStateAction(discretizeBattleState(bs).key(), actionKeyForSwitch(chosenMon))
	}

	return chosenMon
}

func canReplace(party []*Pokemon) bool {
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
