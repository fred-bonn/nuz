package nuzengine

import "math/rand"

func calculateDamage(user, target *Pokemon, move *Move, crit *bool, weather weatherState, maxRoll, forScoring, pursuit bool) int {
	if f, ok := typeImmunityAbilities[target.Ability]; ok && user.Ability != moldBreakerAbility && f(target, move.Type, forScoring) {
		return 0
	}

	numerator := 1
	denominator := 1
	moveType := move.Type
	power := move.Power
	var offensiveStat, defensiveStat int
	if move.Class == physicalClass {
		offensiveStat = user.effectiveStat(attack, *crit)
		defensiveStat = target.effectiveStat(defense, *crit)
	} else {
		offensiveStat = user.effectiveStat(specialAttack, *crit)
		defensiveStat = target.effectiveStat(specialDefense, *crit)
		if weather == sandstormWeather && target.hasType(rockType) {
			defensiveStat = defensiveStat * 3 / 2
		}
	}

	if f, ok := typeConvertingAbilities[user.Ability]; ok {
		f(&moveType, &power)
	}
	numerator, denominator = target.applyMoveType(numerator, denominator, moveType)
	if weather != noneWeather {
		if f, ok := weatherFuncs[weather]; ok {
			f(&numerator, &denominator, moveType)
		}
		switch weather {
		case sunWeather:
			if user.Ability == solarPowerAbility && move.Class == SpecialClass {
				offensiveStat = offensiveStat * 3 / 2
			}
		case sandstormWeather:
			if user.Ability == sandForceAbility && (moveType == rockType || moveType == groundType || moveType == steelType) {
				power = power * 13 / 10
			}
		}
	}
	if numerator == 0 {
		return 0
	}

	switch move.Name {
	case "psywave":
		if maxRoll {
			return user.level
		}
		*crit = false
		return (user.level * (rand.Intn(100) + 51)) / 100
	case "seismic toss", "night shade":
		*crit = false
		return user.level
	case "sonic boom":
		*crit = false
		return 20
	case "dragon rage":
		*crit = false
		return 40
	case "endeavor":
		*crit = false
		return target.HP - user.HP
	case "super fang":
		*crit = false
		return max(1, target.HP/2)
	}

	if move.Name == "acrobatics" && (user.Item.Consumed || user.Item.State == flyingGem) {
		power *= 2
	} else if move.Name == "wake up slap" && target.hasAilment(sleepAilment) != nil {
		power *= 2
	} else if move.Name == "venoshock" && (target.hasAilment(poisonAilment) != nil || target.hasAilment(toxicAilment) != nil) {
		power *= 2
	} else if move.Name == "hex" && target.hasNonVolatileAilment() {
		power *= 2
	} else if move.Name == "flail" || move.Name == "reversal" {
		res := int(48 * (float64(user.HP) / float64(user.MaxHP())))
		if res <= 1 {
			power = 200
		} else if res <= 4 {
			power = 150
		} else if res <= 9 {
			power = 100
		} else if res <= 16 {
			power = 80
		} else if res <= 32 {
			power = 40
		} else {
			power = 20
		}
	} else if move.Name == "pursuit" && pursuit {
		power *= 2
	} else if move.Name == "knock off" && !target.Item.Consumed {
		power *= 2
	}

	if user.Ability == technicianAbility && move.Power <= 60 {
		power = power * 3 / 2
	} else if t, ok := pinchAbilities[user.Ability]; ok && t == moveType && user.HP*3 <= user.MaxHP() {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.flashFire && moveType == fireType {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.Ability == hustleAbility && move.Class == physicalClass {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.Ability == mercilessAbility {
		if a := target.hasAilment(poisonAilment); a != nil {
			*crit = true
		} else if a := target.hasAilment(toxicAilment); a != nil {
			*crit = true
		}
	}

	if target.Ability.blocksCrits() && (forScoring || user.Ability != moldBreakerAbility) {
		*crit = false
	}

	if user.hasType(moveType) {
		numerator *= 3
		denominator *= 2
	}

	if target.Ability == drySkinAbility && moveType == fireType {
		power = power * 5 / 4
	}

	if *crit {
		if user.Ability == sniperAbility {
			numerator *= 3
			denominator *= 2
		}
		numerator *= 3
		denominator *= 2
	}

	if move.Class == physicalClass && user.hasAilment(burnAilment) != nil {
		denominator *= 2
	}

	user.checkItemTrigger(false, makeGemEvent(moveType, &power))

	user.checkItemTrigger(false, makeChoiceItemEvent(move, noStat, &offensiveStat))

	user.checkItemTrigger(false, makeMoveBoostingEvent(moveType, &power))

	if !maxRoll {
		numerator *= rand.Intn(16) + 85
		denominator *= 100
	}

	damage := ((((2*user.level)/5)+2)*power*offensiveStat)/defensiveStat/50 + 2
	damage = damage * numerator / denominator

	target.checkItemTrigger(false, makeResistBerryEvent(moveType, &damage))

	damage = max(1, damage)

	return damage
}
