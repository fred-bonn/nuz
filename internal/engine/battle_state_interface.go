package engine

type BattleState interface {
	Execute() error
	Reset() error
	setError(error)
	gatherActions()
	getAllSlots() []*slot
	getOtherSlots(slot *slot) []*slot
	getOpponentSlot(slot *slot) *slot
	getActions() *actionQueue
	getWeather() WeatherState
	setWeather(WeatherState)
	getFieldEffects() map[FieldEffect]int
	GetStatistics() *battleStatistics
	RecordStatistics()
	PrintStatistics()
}

func injectReplaceAction(bs BattleState, slot *slot, midTurn bool) {
	bs.getActions().queue.push(&replaceAction{
		oldSlot: slot,
		Trainer: slot.Trainer,
		midTurn: midTurn,
	})
	bs.getActions().sort(bs)
}

func resolveEndOfTurn(bs BattleState) {
	for _, slot := range bs.getAllSlots() {
		// resolve end of return effects from ailments and statuses
		for _, ailment := range slot.mon.Ailments {
			switch ailment.State {
			case burnAilment:
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 16)
			case poisonAilment:
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
			case toxicAilment:
				ailment.Turns++
				takeResidualDamage(bs, slot, ailment.State.String(), ailment.Turns, 16)
			case trapAilment:
				ailment.Turns--
				takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
				if ailment.Turns <= 0 {
					vprintf("%s was freed", slot.mon.Base.Name)
					delete(slot.mon.Ailments, ailment.State)
				}
			case leechSeedAilment:
				vprintf("%s leeched health from %s", ailment.afflictedBy.mon.Base.Name, slot.mon.Base.Name)
				dmg := takeResidualDamage(bs, slot, ailment.State.String(), 1, 8)
				ailment.afflictedBy.mon.ChangeHpBy(dmg)
			case yawnAilment:
				ailment.Turns--
				if ailment.Turns == 0 {
					slot.mon.applyAilment(sleepAilment, nil, ailment.afflictedBy)
					delete(slot.mon.Ailments, ailment.State)
				}
			}
		}

		// resolve end of turn effects of weather
		if w := bs.getWeather(); w != NoneWeather {
			if w.affectsMon(slot.mon) {
				takeResidualDamage(bs, slot, w.String(), 1, 16)
			}
			w.activateMonAbility(bs, slot)
		}

		// reset protect counter if the slot was not protected this turn
		if !slot.protected {
			slot.protectTurns = 0
		} else {
			slot.protected = false
		}

		if slot.mon.Ability == harvestAbility && roll(1, 2) && slot.mon.Item.State.isBerry() {
			vprintf("%s harvested its %s", slot.mon.Base.Name, slot.mon.Item.String())
			slot.mon.Item.Consumed = false
			slot.mon.CheckItemTrigger(true, nil)
		} else if slot.mon.Ability == speedBoostAbility && !slot.firstTurn {
			slot.mon.changeStatStageBy(Speed, 1, false)
		}

		if slot.mon.Item.State == leftovers {
			change := slot.mon.MaxHP() / 16
			vprintItem("%s restored %d health from leftovers", slot.mon.Base.Name, change)
			slot.mon.ChangeHpBy(change)
		}

		slot.mon.laserFocus = false
	}
}

func takeResidualDamage(bs BattleState, slot *slot, effect string, num, den int) int {
	if slot.mon.fainted {
		return 0
	}

	change := slot.mon.MaxHP() * num / den
	vprintf("%s took %d damage from %s", slot.mon.Base.Name, change, effect)
	slot.mon.ChangeHpBy(-change)
	if slot.mon.HP <= 0 {
		slot.mon.fainted = true
		injectReplaceAction(bs, slot, false)
		vprintf("%s fainted!", slot.mon.Base.Name)
	}
	return change
}

func resolveOnEntry(bs BattleState) {
	for _, slot := range bs.getAllSlots() {
		if f, ok := onSwitchAbilities[slot.mon.Ability]; ok {
			f(slot, bs, true)
		}
	}
}
