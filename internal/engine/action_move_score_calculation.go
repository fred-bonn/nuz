package engine

func (ma *moveAction) scoreActionMove(bs BattleState) (int, bool) {
	if ma.move.Class == statusClass {
		return ma.scoreStatusMove(bs), false
	}

	damageRoll := 0
	critRate := determineCritRate(ma.userSlot.mon, ma.move)
	rolls := 1
	if ma.move.MaxHits == 5 {
		rolls = 3
	} else if ma.move.MaxHits > 0 {
		rolls = ma.move.MaxHits
	}
	for i := 0; i < rolls; i++ {
		damageRoll += calculateDamage(ma.userSlot.mon, ma.targetSlot.mon, ma.move, new(critRate >= 3), bs.getWeather(), false, true, false)
	}

	ma.targetSlot.mon.CheckItemTrigger(false, makeFocusSashEvent(&damageRoll))
	if ma.targetSlot.mon.Ability == sturdyAbility && ma.targetSlot.mon.HP == ma.targetSlot.mon.MaxHP() {
		damageRoll = min(damageRoll, ma.targetSlot.mon.HP-1)
	}

	return damageRoll, damageRoll >= ma.targetSlot.mon.HP
}

func (ma *moveAction) scoreStatusMove(bs BattleState) int {
	if ma.move.Category == "heal" {
		if ma.userSlot.mon.HP > ma.userSlot.mon.MaxHP()*85/100 {
			return -64
		}
		if ma.shouldMonHeal(bs) {
			return 7
		}
		return 5
	}

	if isPowderMove(ma.move.Name) && ma.targetSlot.mon.isImmuneToPowderMoves() {
		return -64
	}
	if isParalysisMove(ma.move.Name) {
		return ma.scoreParalysisMove(bs)
	}
	if isSleepMove(ma.move.Name) {
		return ma.scoreSleepMove(bs)
	}
	if isProtectMove(ma.move.Name) {
		return ma.scoreProtectMove(bs)
	}

	switch ma.move.Name {
	case "sticky web":
		if ma.targetSlot.hasFieldEffect(ma.move.Name) {
			return -64
		}
		if ma.userSlot.firstTurn {
			return 9 + 3*rollInt(3, 4)
		}
		return 6 + 3*rollInt(3, 4)
	case "stealth rock", "spikes", "toxic spikes":
		if ma.targetSlot.hasFieldEffect(ma.move.Name) {
			return -64
		}
		if ma.userSlot.firstTurn {
			return 8 + rollInt(3, 4)
		}
		return 6 + rollInt(3, 4)
	case "attract":
		if ma.targetSlot.mon.hasAilment(infatuationAilment) != nil || ma.targetSlot.mon.Ability == obliviousAbility {
			return -64
		}
	case "leech seed":
		if ma.targetSlot.mon.hasAilment(leechSeedAilment) != nil || ma.targetSlot.mon.hasType(grassType) {
			return -64
		}
	case "toxic":
		return ma.scoreToxic(bs)
	case "focus energy", "laser focus":
		return ma.scoreCritStatus()
	case "belly drum":
		return ma.scoreBellyDrum(bs)
	}

	return 6
}

func (ma *moveAction) shouldMonHeal(bs BattleState) bool {
	if ma.userSlot.mon.hasAilment(toxicAilment) != nil {
		return false
	}

	maxDmg := calculateMaxDamage(bs, ma.targetSlot.mon, ma.userSlot.mon, true)
	if maxDmg >= ma.userSlot.mon.MaxHP()*ma.move.Heal/100 {
		return false
	}

	if ma.userSlot.mon.isFasterThan(bs, ma.targetSlot.mon) {
		if maxDmg < min(ma.userSlot.mon.MaxHP(), ma.userSlot.mon.HP+ma.userSlot.mon.MaxHP()*ma.move.Heal/100) {
			return true
		} else {
			if ma.userSlot.mon.HP < ma.userSlot.mon.MaxHP()*40/100 {
				return true
			} else if ma.userSlot.mon.HP <= ma.userSlot.mon.MaxHP()*66/100 {
				return roll(1, 2)
			}
		}
	} else {
		if ma.userSlot.mon.HP < ma.userSlot.mon.MaxHP()*50/100 {
			return true
		} else if ma.userSlot.mon.HP <= ma.userSlot.mon.MaxHP()*70/100 {
			return roll(3, 4)
		}
	}

	return false
}

func (ma *moveAction) scoreParalysisMove(bs BattleState) int {
	target := ma.targetSlot.mon
	user := ma.userSlot.mon

	if target.hasNonVolatileAilment() || target.hasType(electricType) || target.Ability == limberAbility {
		return -64
	}

	score := 6
	if target.isFasterThan(bs, user) && user.effectiveSpeed(bs) > target.effectiveSpeed(bs)/4 {
		score++
	} else if user.hasMovePredicate(func(m *Move) bool {
		return m.Name == "hex" || m.FlinchChance > 0
	}) {
		score++
	} else if target.hasAilment(confusionAilment) != nil {
		score++
	} else if target.hasAilment(infatuationAilment) != nil {
		score++
	}

	return score + rollInt(1, 2)
}

func (ma *moveAction) scoreSleepMove(bs BattleState) int {
	target := ma.targetSlot.mon
	user := ma.userSlot.mon

	if target.Ability.blocksSleep() {
		return -64
	}
	if a := target.hasAilment(yawnAilment); a != nil {
		return -64
	}
	if target.hasNonVolatileAilment() {
		return -64
	}

	isHex := func(m *Move) bool {
		return m.Name == "hex"
	}

	score := 6
	maxDmg := calculateMaxDamage(bs, target, ma.userSlot.mon, true)
	if maxDmg < user.HP && roll(1, 2) {
		if user.hasMovePredicate(isHex) {
			score += 1
		} else {
			for _, slot := range bs.getOtherSlots(ma.userSlot) {
				if slot.Trainer == ma.userSlot.Trainer && slot.mon.hasMovePredicate(isHex) {
					score += 1
				}
			}
		}

		if user.hasMovePredicate(func(m *Move) bool {
			return m.Name == "dream eater" || m.Name == "nightmare"
		}) && !target.hasMovePredicate(func(m *Move) bool {
			return m.Name == "snore" || m.Name == "sleep talk"
		}) {
			score += 1
		}
	}

	return score
}

func (ma *moveAction) scoreToxic(bs BattleState) int {
	target := ma.targetSlot.mon
	user := ma.userSlot.mon

	if target.hasNonVolatileAilment() {
		return -64
	}
	if target.Ability == immunityAbility {
		return -64
	}
	if (target.hasType(poisonType) || target.hasType(steelType)) && ma.userSlot.mon.Ability != corrosionAbility {
		return -64
	}

	score := 6
	maxDmg := calculateMaxDamage(bs, target, user, true)
	if maxDmg < user.HP && roll(19, 50) {
		if !target.hasMovePredicate(func(m *Move) bool {
			return m.Class == physicalClass || m.Class == SpecialClass
		}) {
			score += 1
		}

		if user.hasMovePredicate(func(m *Move) bool {
			return m.Name == "hex" || m.Name == "venoshock"
		}) || user.Ability == mercilessAbility {
			score += 2
		} else {
			score += 1
		}
	}

	return score
}

func (ma *moveAction) scoreProtectMove(bs BattleState) int {
	user := ma.userSlot.mon
	target := ma.targetSlot.mon

	if ma.userSlot.protectTurns == 2 || (ma.userSlot.protectTurns == 1 && roll(1, 2)) {
		return -64
	}
	if deadToSecondaryDamage(user, bs) {
		return -64
	}

	score := 6
	if ma.userSlot.firstTurn {
		score--
	}
	// still needs perish song and cursed
	if target.hasAilment(poisonAilment) != nil || target.hasAilment(toxicAilment) != nil || target.hasAilment(burnAilment) != nil || target.hasAilment(leechSeedAilment) != nil || target.hasAilment(yawnAilment) != nil || target.hasAilment(infatuationAilment) != nil {
		score++
	}

	if user.hasAilment(poisonAilment) != nil || user.hasAilment(toxicAilment) != nil || user.hasAilment(burnAilment) != nil || user.hasAilment(leechSeedAilment) != nil || user.hasAilment(yawnAilment) != nil || user.hasAilment(infatuationAilment) != nil {
		score -= 2
	}

	return score
}

func deadToSecondaryDamage(mon *Pokemon, bs BattleState) bool {
	if mon.Ability == magicGuardAbility {
		return false
	}

	dmg := 0
	if mon.hasAilment(burnAilment) != nil {
		dmg += mon.MaxHP() / 16
	} else if mon.hasAilment(poisonAilment) != nil {
		dmg += mon.MaxHP() / 8
	} else if a := mon.hasAilment(toxicAilment); a != nil {
		dmg += (mon.MaxHP() * (a.Turns + 1)) / 16
	}
	if mon.hasAilment(trapAilment) != nil {
		dmg += mon.MaxHP() / 8
	}
	if bs.getWeather().affectsMon(mon) {
		dmg += mon.MaxHP() / 16
	}

	return dmg >= mon.HP
}

func (ma *moveAction) scoreCritStatus() int {
	user := ma.userSlot.mon
	if ma.targetSlot.mon.Ability.blocksCrits() && user.Ability != moldBreakerAbility {
		return -64
	}
	if ma.move.Name == "focus energy" && user.focusEnergy {
		return -64
	}

	if user.hasMovePredicate(func(m *Move) bool {
		return m.CritRate > 0
	}) {
		return 7
	} else if user.Item.State == scopeLens {
		return 7
	} else if user.Ability == superLuckAbility || user.Ability == sniperAbility {
		return 7
	}

	return 6
}

func (ma *moveAction) scoreBellyDrum(bs BattleState) int {
	user := ma.userSlot.mon
	target := ma.targetSlot.mon

	if a := target.hasAilment(freezeAilment); a != nil && target.hasMovePredicate(func(m *Move) bool {
		return isSelfThawingMove(m.Name)
	}) {
		return 9
	}
	if a := target.hasAilment(sleepAilment); a != nil {
		return 9
	}

	dmg := calculateMaxDamage(bs, target, user, true)
	threshhold := user.HP - (user.MaxHP() / 2)
	if user.Item.State == sitrusBerry && !user.Item.Consumed {
		threshhold += user.MaxHP() / 4
	}
	if dmg < threshhold {
		return 8
	}

	return 4
}
