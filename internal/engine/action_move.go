package engine

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

func (ma *moveAction) prio(bs BattleState) int {
	bonus := 0
	if ma.userSlot.mon.Ability == pranksterAbility && ma.move.Class == statusClass {
		bonus++
	}

	return ma.move.Priority + bonus
}

func (ma *moveAction) speed(bs BattleState) int {
	return ma.userSlot.mon.effectiveSpeed(bs)
}

func (ma *moveAction) invoke(bs BattleState) {
	if ma.userSlot.mon.fainted {
		return
	}

	if (ma.move.Name == "fake out" || ma.move.Name == "first impression") && !ma.userSlot.firstTurn {
		vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s", ma.userSlot.mon.Base.Name, ma.move.Name)
		vprintf("but it failed")
		return
	}

	ma.userSlot.firstTurn = false

	ma.userSlot.suckerPunch = ma.move.Name == "sucker punch"

	if _, ok := ma.userSlot.mon.Ailments[freezeAilment]; ok {
		if isSelfThawingMove(ma.move.Name) || roll(1, 5) {
			vprintf("%s thawed out", ma.userSlot.mon.Base.Name)
			delete(ma.userSlot.mon.Ailments, freezeAilment)
		} else {
			ma.userSlot.invulnerableAction = nil
			vprintMove(ma.prio(bs), ma.speed(bs), "%s is frozen", ma.userSlot.mon.Base.Name)
			return
		}
	}

	if sleep := ma.userSlot.mon.hasAilment(sleepAilment); sleep != nil {
		if sleep.Turns <= 0 {
			vprintf("%s woke up", ma.userSlot.mon.Base.Name)
			delete(ma.userSlot.mon.Ailments, sleepAilment)
		} else {
			if ma.userSlot.mon.Ability == earlyBirdAbility {
				sleep.Turns -= 2
			} else {
				sleep.Turns--
			}
			ma.userSlot.invulnerableAction = nil
			vprintMove(ma.prio(bs), ma.speed(bs), "%s is asleep", ma.userSlot.mon.Base.Name)
			return
		}
	}

	if confusion := ma.userSlot.mon.hasAilment(confusionAilment); confusion != nil {
		if confusion.Turns > 0 {
			confusion.Turns -= 1
			vprintf("%s is confused", ma.userSlot.mon.Base.Name)
			if roll(1, 3) {
				damage := calculateDamage(ma.userSlot.mon, ma.userSlot.mon, &confusionMove, new(false), bs.getWeather(), false, false, false)
				ma.userSlot.invulnerableAction = nil
				vprintMove(ma.prio(bs), ma.speed(bs), "%s hit itself in confusion for %d damage", ma.userSlot.mon.Base.Name, damage)
				ma.userSlot.mon.HP -= int(damage)
				if ma.userSlot.mon.HP <= 0 {
					monFainted(bs, ma.userSlot, false)
				}
				return
			}
		} else {
			delete(ma.userSlot.mon.Ailments, confusionAilment)
			vprintf("%s snapped out of confusion", ma.userSlot.mon.Base.Name)
		}
	}

	if paralysis := ma.userSlot.mon.hasAilment(paralysisAilment); paralysis != nil {
		if roll(1, 4) {
			ma.userSlot.invulnerableAction = nil
			vprintf("%s is paralysed", ma.userSlot.mon.Base.Name)
			return
		}
	}

	if infatuation := ma.userSlot.mon.hasAilment(infatuationAilment); infatuation != nil {
		if roll(1, 2) {
			ma.userSlot.invulnerableAction = nil
			vprintf("%s is infatuated with %s", ma.userSlot.mon.Base.Name, infatuation.afflictedBy.mon.Base.Name)
			return
		}
	}

	if ma.flinch {
		vprintf("%s flinched", ma.userSlot.mon.Base.Name)
		return
	}

	ma.move.PP--

	if isMultipleTurnMove(ma.move.Name) {
		if ma.userSlot.invulnerableAction == nil {
			ma.move.PP++
			ma.userSlot.invulnerableAction = ma
			vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s and became invulnerable", ma.userSlot.mon.Base.Name, ma.move.Name)
			return
		}
		ma.userSlot.invulnerableAction = nil
	}

	if ma.userSlot.suckerPunch {
		targetMove := bs.getActions().getMoveActionBy(ma.targetSlot.mon)
		if targetMove == nil || targetMove.move.Class == statusClass {
			vprintMove(ma.prio(bs), ma.speed(bs), "%s used sucker punch but it failed", ma.userSlot.mon.Base.Name)
			return
		}
	}

	target := ma.targetSlot.mon
	if ma.move.Accuracy > 0 && !ma.pursuit && !accuracyRoll(bs, ma.userSlot.mon, target, ma.move) {
		vprintMove(ma.prio(bs), ma.speed(bs), "%s's move %s missed", ma.userSlot.mon.Base.Name, ma.move.Name)
		return
	}

	vprintMove(ma.prio(bs), ma.speed(bs), "%s used %s", ma.userSlot.mon.Base.Name, ma.move.Name)

	if ma.move.Name == "struggle" {
		ma.userSlot.mon.ChangeHpBy(-(ma.userSlot.mon.MaxHP() / 4))
	}

	if ma.targetSlot.protected {
		vprintln("but it failed")
		return
	}

	if ma.targetSlot.invulnerableAction != nil && ma.userSlot.mon.Ability != noGuardAbility && ma.targetSlot.mon.Ability != noGuardAbility {
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

	if ma.userSlot.mon.HP <= 0 {
		monFainted(bs, ma.userSlot, false)
		return
	}

	ma.userSlot.mon.CheckItemTrigger(true, makeLeppaBerryEvent(ma.move))

	if ma.move.Type == FireType {
		delete(ma.targetSlot.mon.Ailments, freezeAilment)
	}
}

func (ma *moveAction) applyStatusMove(bs BattleState, target *Pokemon, offensive bool) {
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
		if target.HP*2 <= target.MaxHP() {
			vprintf("but it failed")
			return
		}

		vprintf("%s took damage from belly drum", target.Base.Name)
		target.ChangeHpBy(-(target.MaxHP() / 2))
		target.changeStatStageBy(attack, 6, false)
	}

	if ma.move.Heal > 0 {
		change := target.MaxHP() * ma.move.Heal / 100
		target.ChangeHpBy(change)
		vprintf("%s healed for %d", target.Base.Name, change)
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

func (ma *moveAction) applyDamageMove(bs BattleState) {
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

func (ma *moveAction) resolveDamage(bs BattleState) bool {
	user := ma.userSlot.mon
	target := ma.targetSlot.mon

	crit := determineCrit(user, ma.move)

	damage := calculateDamage(user, target, ma.move, crit, bs.getWeather(), false, false, ma.pursuit)
	if damage == 0 {
		vprintf("it does not affect %s", target.Base.Name)
		return false
	}

	target.CheckItemTrigger(true, makeResistBerryEvent(ma.move.Type, nil))
	target.CheckItemTrigger(true, makeFocusSashEvent(&damage))
	if target.Ability == sturdyAbility && target.HP == target.MaxHP() {
		damage = min(damage, target.HP-1)
	}
	user.CheckItemTrigger(true, makeGemEvent(ma.move.Type, nil))

	damage = min(damage, target.HP)
	vprintf("%s took %d damage", target.Base.Name, int(damage))
	if *crit {
		vprintf("it was a critical hit!")
	}
	if ma.move.Name == "bug bite" && target.Item.State.isBerry() && !target.Item.Consumed {
		vprintf("%s's %s was consumed by bug bite", target.Base.Name, target.Item.String())
		item, _ := RegisterItem(target.Item.State, user)
		item.activate()
		target.Item, _ = RegisterItem(noneItem, target)

	} else if ma.move.Name == "wake up slap" {
		if a := target.hasAilment(sleepAilment); a != nil {
			vprintf("%s woke up", target.Base.Name)
			delete(target.Ailments, sleepAilment)
		}
	} else if ma.move.Name == "knock off" && !target.Item.Consumed && target.Ability != stickyHoldAbility {
		vprintf("%s had its %s knocked off", target.Base.Name, target.Item.String())
		target.Item = &item{
			Consumed: true,
		}
	}
	target.ChangeHpBy(-damage)
	if target.HP <= 0 {
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
		if target.Ability == liquidOozeAbility {
			if change > 0 {
				change = -change
			}
		}

		user.ChangeHpBy(change)
		if change >= 0 {
			vprintf("%s healed for %d", user.Base.Name, change)
		} else {
			vprintf("%s took recoil for %d", user.Base.Name, -change)
			if user.HP <= 0 {
				monFainted(bs, ma.userSlot, false)
			}
		}
	}

	if f, ok := contactDefensiveAbilities[target.Ability]; ok && ma.move.Contact {
		f(ma.userSlot, ma.targetSlot)
		if user.HP <= 0 {
			monFainted(bs, ma.userSlot, false)
		}
	} else if target.Ability == cottenDownAbility {
		for _, slot := range bs.getOtherSlots(ma.targetSlot) {
			slot.mon.changeStatStageBy(Speed, -1, true)
		}
	} else if target.Ability == waterCompactionAbility && ma.move.Type == waterType {
		target.changeStatStageBy(defense, 2, false)
	}
	if f, ok := contactOffensiveAbilities[user.Ability]; ok && ma.move.Contact {
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

	if target.Ability == shieldDustAbility {
		return true
	}

	if ma.move.AilmentChance > 0 && !target.fainted && roll(ma.move.AilmentChance*sg, 100) {
		target.applyAilment(ma.move.Ailment, ma.move, ma.userSlot)
	}

	if ma.move.FlinchChance > 0 && !target.fainted && target.Ability != innerFocusAbility && roll(ma.move.FlinchChance*sg, 100) {
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
