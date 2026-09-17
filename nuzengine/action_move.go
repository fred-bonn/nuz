package nuzengine

import (
	"fmt"
	"strings"
)

type moveAction struct {
	userSlot   *slot
	targetSlot *slot
	move       *Move
	flinch     bool
	pursuit    bool
}

func (ma *moveAction) prio(bs battleState) int {
	bonus := 0
	if ma.userSlot.mon.ability == pranksterAbility && ma.move.Class == statusClass {
		bonus++
	}

	return ma.move.Priority + bonus
}

func (ma *moveAction) speed(bs battleState) int {
	return ma.userSlot.mon.effectiveSpeed(bs)
}

func (ma *moveAction) invoke(bs battleState) {
	if ma.userSlot.mon.fainted {
		return
	}

	if (ma.move.Name == "fake out" || ma.move.Name == "first impression") && !ma.userSlot.firstTurn {
		vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s", ma.userSlot.mon.base.Name, ma.move.Name)
		vprintf("but it failed")
		return
	}

	ma.userSlot.firstTurn = false

	ma.userSlot.suckerPunch = ma.move.Name == "sucker punch"

	if _, ok := ma.userSlot.mon.ailments[freezeAilment]; ok {
		if isSelfThawingMove(ma.move.Name) || roll(1, 5) {
			vprintf("%s thawed out", ma.userSlot.mon.base.Name)
			delete(ma.userSlot.mon.ailments, freezeAilment)
		} else {
			ma.userSlot.invulnerableAction = nil
			vprintMove(ma.prio(bs), ma.speed(bs), "%s is frozen", ma.userSlot.mon.base.Name)
			return
		}
	}

	if sleep := ma.userSlot.mon.hasAilment(sleepAilment); sleep != nil {
		if sleep.Turns <= 0 {
			vprintf("%s woke up", ma.userSlot.mon.base.Name)
			delete(ma.userSlot.mon.ailments, sleepAilment)
		} else {
			if ma.userSlot.mon.ability == earlyBirdAbility {
				sleep.Turns -= 2
			} else {
				sleep.Turns--
			}
			ma.userSlot.invulnerableAction = nil
			vprintMove(ma.prio(bs), ma.speed(bs), "%s is asleep", ma.userSlot.mon.base.Name)
			return
		}
	}

	if confusion := ma.userSlot.mon.hasAilment(confusionAilment); confusion != nil {
		if confusion.Turns > 0 {
			confusion.Turns -= 1
			vprintf("%s is confused", ma.userSlot.mon.base.Name)
			if roll(1, 3) {
				damage := calculateDamage(ma.userSlot.mon, ma.userSlot.mon, getConfusionMove(), new(false), bs.getWeather(), false, false, false)
				ma.userSlot.invulnerableAction = nil
				vprintMove(ma.prio(bs), ma.speed(bs), "%s hit itself in confusion for %d damage", ma.userSlot.mon.base.Name, damage)
				ma.userSlot.mon.hp -= int(damage)
				if ma.userSlot.mon.hp <= 0 {
					monFainted(bs, ma.userSlot, false)
				}
				return
			}
		} else {
			delete(ma.userSlot.mon.ailments, confusionAilment)
			vprintf("%s snapped out of confusion", ma.userSlot.mon.base.Name)
		}
	}

	if paralysis := ma.userSlot.mon.hasAilment(paralysisAilment); paralysis != nil {
		if roll(1, 4) {
			ma.userSlot.invulnerableAction = nil
			vprintf("%s is paralysed", ma.userSlot.mon.base.Name)
			return
		}
	}

	if infatuation := ma.userSlot.mon.hasAilment(infatuationAilment); infatuation != nil {
		if roll(1, 2) {
			ma.userSlot.invulnerableAction = nil
			vprintf("%s is infatuated with %s", ma.userSlot.mon.base.Name, infatuation.afflictedBy.mon.base.Name)
			return
		}
	}

	if ma.flinch {
		vprintf("%s flinched", ma.userSlot.mon.base.Name)
		return
	}

	ma.move.PP--

	if isMultipleTurnMove(ma.move.Name) {
		if ma.userSlot.invulnerableAction == nil {
			ma.move.PP++
			ma.userSlot.invulnerableAction = ma
			vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s and became invulnerable", ma.userSlot.mon.base.Name, ma.move.Name)
			return
		}
		ma.userSlot.invulnerableAction = nil
	}

	if ma.userSlot.suckerPunch {
		targetMove := bs.getActions().getMoveActionBy(ma.targetSlot.mon)
		if targetMove == nil || targetMove.move.Class == statusClass {
			vprintMove(ma.prio(bs), ma.speed(bs), "%s used sucker punch but it failed", ma.userSlot.mon.base.Name)
			return
		}
	}

	target := ma.targetSlot.mon
	if ma.move.Accuracy > 0 && !ma.pursuit && !accuracyRoll(bs, ma.userSlot.mon, target, ma.move) {
		vprintMove(ma.prio(bs), ma.speed(bs), "%s's move %s missed", ma.userSlot.mon.base.Name, ma.move.Name)
		return
	}

	vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s", ma.userSlot.mon.base.Name, ma.move.Name)

	if ma.move.Name == "struggle" {
		ma.userSlot.mon.ChangeHpBy(-(ma.userSlot.mon.MaxHP() / 4))
	}

	if ma.targetSlot.protected {
		vprintln("but it failed")
		return
	}

	if ma.targetSlot.invulnerableAction != nil && ma.userSlot.mon.ability != noGuardAbility && ma.targetSlot.mon.ability != noGuardAbility {
		// special cases for surf against dive, thunder against fly, etc
		vprintln("but it failed")
		return
	}

	if isPowderMove(ma.move.Name) && ma.targetSlot.mon.isImmuneToPowderMoves() {
		vprintln("but it failed")
		return
	}

	if ma.move.Class == statusClass {
		if strings.HasPrefix(ma.move.Target, "user") {
			ma.applyStatusMove(bs, ma.userSlot.mon, false)
		} else {
			ma.applyStatusMove(bs, target, true)
		}
	} else {
		ma.applyDamageMove(bs)
	}

	if ma.userSlot.mon.hp <= 0 {
		monFainted(bs, ma.userSlot, false)
		return
	}

	ma.userSlot.mon.checkItemTrigger(true, makeLeppaBerryEvent(ma.move))

	if ma.move.Type == fireType {
		delete(ma.targetSlot.mon.ailments, freezeAilment)
	}
}

func (ma *moveAction) applyStatusMove(bs battleState, target *pokemon, offensive bool) {
	if isProtectMove(ma.move.Name) {
		ma.userSlot.resolveProtect()
		return
	}

	if ma.move.Category == "field-effect" {
		err := ma.targetSlot.applyFieldEffect(ma.move.Name)
		if err != nil {
			bs.setError(err)
		}

		return
	}

	switch ma.move.Name {
	case "swagger":
		target.changeStatStageBy(attack, 2, false)
		target.applyAilment(confusionAilment, ma.move, ma.userSlot)
		return
	case "focus energy":
		target.focusEnergy = true
		return
	case "laser focus":
		target.laserFocus = true
		return
	case "belly drum":
		if target.hp*2 <= target.MaxHP() {
			vprintf("but it failed")
			return
		}

		vprintf("%s took damage from belly drum", target.base.Name)
		target.ChangeHpBy(-(target.MaxHP() / 2))
		target.changeStatStageBy(attack, 6, false)
	}

	if ma.move.Heal > 0 {
		change := target.MaxHP() * ma.move.Heal / 100
		target.ChangeHpBy(change)
		vprintf("%s healed for %d", target.base.Name, change)
	}

	if ma.move.Ailment != noneAilment {
		target.applyAilment(ma.move.Ailment, ma.move, ma.userSlot)
	}

	for stat, change := range ma.move.StatChanges {
		s := stringToStat(stat)
		if s == noStat {
			bs.setError(fmt.Errorf("%s is not a valid stat for %s", stat, ma.move.Name))
			return
		}
		target.changeStatStageBy(s, change, offensive)
	}
}

func (ma *moveAction) applyDamageMove(bs battleState) {
	target := ma.targetSlot.mon
	if target.fainted {
		return
	}

	hits := 1
	if ma.move.MinHits > 0 {
		hits = determineHits(ma.move)
	}

	if ok := ma.resolveDamage(bs); !ok {
		return
	}

	for i := 1; i < hits; i++ {
		if ok := ma.resolveDamage(bs); !ok {
			return
		}
	}
}

func (ma *moveAction) resolveDamage(bs battleState) bool {
	user := ma.userSlot.mon
	target := ma.targetSlot.mon

	crit := determineCrit(user, ma.move)

	damage := calculateDamage(user, target, ma.move, crit, bs.getWeather(), false, false, ma.pursuit)
	if damage == 0 {
		vprintf("it does not affect %s", target.base.Name)
		return false
	}

	target.checkItemTrigger(true, makeResistBerryEvent(ma.move.Type, nil))
	target.checkItemTrigger(true, makeFocusSashEvent(&damage))
	if target.ability == sturdyAbility && target.hp == target.MaxHP() {
		damage = min(damage, target.hp-1)
	}
	user.checkItemTrigger(true, makeGemEvent(ma.move.Type, nil))

	damage = min(damage, target.hp)
	vprintf("%s took %d damage", target.base.Name, int(damage))
	if *crit {
		vprintf("it was a critical hit!")
	}
	if ma.move.Name == "bug bite" && target.item.State.isBerry() && !target.item.Consumed {
		vprintf("%s's %s was consumed by bug bite", target.base.Name, target.item.String())
		item, _ := registerItem(target.item.State, user)
		item.activate()
		target.item, _ = registerItem(noneItem, target)

	} else if ma.move.Name == "wake up slap" {
		if a := target.hasAilment(sleepAilment); a != nil {
			vprintf("%s woke up", target.base.Name)
			delete(target.ailments, sleepAilment)
		}
	} else if ma.move.Name == "knock off" && !target.item.Consumed && target.ability != stickyHoldAbility {
		vprintf("%s had its %s knocked off", target.base.Name, target.item.String())
		target.item = &item{
			Consumed: true,
		}
	}
	target.ChangeHpBy(-damage)
	if target.hp <= 0 {
		monFainted(bs, ma.targetSlot, ma.pursuit)
	}

	if ma.move.Drain != 0 {
		change := damage * ma.move.Drain / 100
		if change == 0 {
			if ma.move.Drain > 0 {
				change = 1
			} else {
				change = -1
			}
		}
		if target.ability == liquidOozeAbility {
			if change > 0 {
				change = -change
			}
		}

		user.ChangeHpBy(change)
		if change >= 0 {
			vprintf("%s healed for %d", user.base.Name, change)
		} else {
			vprintf("%s took recoil for %d", user.base.Name, -change)
			if user.hp <= 0 {
				monFainted(bs, ma.userSlot, false)
			}
		}
	}

	if f, ok := contactDefensiveAbilities[target.ability]; ok && ma.move.Contact {
		f(ma.userSlot, ma.targetSlot)
		if user.hp <= 0 {
			monFainted(bs, ma.userSlot, false)
		}
	} else if target.ability == cottenDownAbility {
		for _, slot := range bs.getOtherSlots(ma.targetSlot) {
			slot.mon.changeStatStageBy(Speed, -1, true)
		}
	} else if target.ability == waterCompactionAbility && ma.move.Type == waterType {
		target.changeStatStageBy(defense, 2, false)
	}
	if f, ok := contactOffensiveAbilities[user.ability]; ok && ma.move.Contact {
		f(ma.userSlot, ma.targetSlot)
	}

	sg := user.serenceGraceBonus()

	if ma.move.StatChance > 0 && ma.move.Category == "damage-raise" && roll(ma.move.StatChance*sg, 100) {
		for stat, change := range ma.move.StatChanges {
			s := stringToStat(stat)
			if s == noStat {
				bs.setError(fmt.Errorf("%s is not a valid stat for %s", stat, ma.move.Name))
				return false
			}
			user.changeStatStageBy(s, change, false)
		}
	}

	if isPivotMove(ma.move.Name) {
		if ma.userSlot.Trainer.canReplace(bs) {
			injectReplaceAction(bs, ma.userSlot, true)
		}
	}

	if target.fainted {
		return false
	}

	if target.ability == shieldDustAbility {
		return true
	}

	if ma.move.AilmentChance > 0 && !target.fainted && roll(ma.move.AilmentChance*sg, 100) {
		target.applyAilment(ma.move.Ailment, ma.move, ma.userSlot)
	}

	if ma.move.FlinchChance > 0 && !target.fainted && target.ability != innerFocusAbility && roll(ma.move.FlinchChance*sg, 100) {
		if targetMove := bs.getActions().getMoveActionBy(target); targetMove != nil {
			targetMove.flinch = true
		}
	}

	if ma.move.StatChance > 0 && ma.move.Category == "damage-lower" && roll(ma.move.StatChance*sg, 100) {
		for stat, change := range ma.move.StatChanges {
			s := stringToStat(stat)
			if s == noStat {
				bs.setError(fmt.Errorf("%s is not a valid stat for %s", stat, ma.move.Name))
				return false
			}
			target.changeStatStageBy(s, change, true)
		}
	}

	return true
}
