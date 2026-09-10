package engine

type replaceAction struct {
	oldSlot *slot
	Trainer *Trainer
	midTurn bool
}

func (ra *replaceAction) prio(bs BattleState) int {
	if ra.midTurn {
		return 10
	}
	return -10
}

func (ra *replaceAction) speed(bs BattleState) int {
	return ra.oldSlot.mon.effectiveSpeed(bs)
}

func (ra *replaceAction) invoke(bs BattleState) {
	if ra.midTurn {
		if a, ok := bs.getActions().queue.fetchBy(fetchPursuitMiddleware(ra.oldSlot.mon.Base.Name)); ok {
			p, _ := a.(*moveAction)
			p.pursuit = true
			p.invoke(bs)
			if ra.oldSlot.mon.fainted {
				injectReplaceAction(bs, ra.oldSlot, false)
				return
			}
		}
	}

	mon := chooseSwitchIn(bs, ra.oldSlot, ra.Trainer.PokemonParty, ra.Trainer.AI)
	if mon == nil {
		ra.Trainer.lost = true
		return
	}

	for _, slot := range bs.getOtherSlots(ra.oldSlot) {
		if ailment := slot.mon.hasAilment(infatuationAilment); ailment != nil && ailment.afflictedBy == ra.oldSlot {
			delete(slot.mon.Ailments, infatuationAilment)
		}
		if ailment := slot.mon.hasAilment(trapAilment); ailment != nil && ailment.afflictedBy == ra.oldSlot {
			delete(slot.mon.Ailments, infatuationAilment)
		}
	}
	if f, ok := onSwitchAbilities[ra.oldSlot.mon.Ability]; ok {
		f(ra.oldSlot, bs, false)
	}

	vprintReplace("%s was sent out", mon.Base.Name)
	ra.oldSlot.setMon(bs, mon)
	if f, ok := onSwitchAbilities[ra.oldSlot.mon.Ability]; ok {
		f(ra.oldSlot, bs, true)
	}
}
