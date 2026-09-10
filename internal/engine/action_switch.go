package engine

type switchAction struct {
	oldSlot *slot
	new     *Pokemon
}

func (sa *switchAction) invoke(bs BattleState) {
	for _, slot := range bs.getOtherSlots(sa.oldSlot) {
		if ailment := slot.mon.hasAilment(infatuationAilment); ailment != nil && ailment.afflictedBy == sa.oldSlot {
			delete(slot.mon.Ailments, infatuationAilment)
		}
		if ailment := slot.mon.hasAilment(trapAilment); ailment != nil && ailment.afflictedBy == sa.oldSlot {
			delete(slot.mon.Ailments, infatuationAilment)
		}
	}
	if f, ok := onSwitchAbilities[sa.oldSlot.mon.Ability]; ok {
		f(sa.oldSlot, bs, false)
	}

	if a, ok := bs.getActions().queue.fetchBy(fetchPursuitMiddleware(sa.oldSlot.mon.Base.Name)); ok {
		p, _ := a.(*moveAction)
		p.pursuit = true
		p.invoke(bs)
	}

	vprintSwitch("switched %s for %s", sa.oldSlot.mon.Base.Name, sa.new.Base.Name)
	sa.oldSlot.setMon(bs, sa.new)
	if f, ok := onSwitchAbilities[sa.oldSlot.mon.Ability]; ok {
		f(sa.oldSlot, bs, true)
	}
}

func (sa *switchAction) prio(bs BattleState) int {
	return 10
}

func (sa *switchAction) speed(bs BattleState) int {
	return sa.oldSlot.mon.effectiveSpeed(bs)
}
