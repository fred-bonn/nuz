package nuzengine

type replaceAction struct {
	oldSlot *slot
	Trainer *trainer
	midTurn bool
}

func (ra *replaceAction) prio(bs battleState) int {
	if ra.midTurn {
		return 10
	}
	return -10
}

func (ra *replaceAction) speed(bs battleState) int {
	return ra.oldSlot.mon.effectiveSpeed(bs)
}

func (ra *replaceAction) invoke(bs battleState) {
	if ra.midTurn {
		if a, ok := bs.getActions().queue.fetchBy(fetchPursuitMiddleware(ra.oldSlot.mon.base.Name)); ok {
			p, _ := a.(*moveAction)
			p.pursuit = true
			p.invoke(bs)
			if ra.oldSlot.mon.fainted {
				injectReplaceAction(bs, ra.oldSlot, false)
				return
			}
		}
	}

	mon := chooseSwitchIn(bs, ra.oldSlot, ra.Trainer.pokemonParty, ra.Trainer.ai)
	if mon == nil {
		ra.Trainer.lost = true
		return
	}

	for _, slot := range bs.getOtherSlots(ra.oldSlot) {
		if ailment := slot.mon.hasAilment(infatuationAilment); ailment != nil && ailment.afflictedBy == ra.oldSlot {
			delete(slot.mon.ailments, infatuationAilment)
		}
		if ailment := slot.mon.hasAilment(trapAilment); ailment != nil && ailment.afflictedBy == ra.oldSlot {
			delete(slot.mon.ailments, infatuationAilment)
		}
	}
	if f, ok := onSwitchAbilities[ra.oldSlot.mon.ability]; ok {
		f(ra.oldSlot, bs, false)
	}

	vprintReplace("%s was sent out", mon.base.Name)
	ra.oldSlot.setMon(bs, mon)
	if f, ok := onSwitchAbilities[ra.oldSlot.mon.ability]; ok {
		f(ra.oldSlot, bs, true)
	}
}
