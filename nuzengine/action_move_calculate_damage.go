package nuzengine

import "math/rand"

type damageRollType int

const (
	rollRandom = iota
	rollMin
	rollMax
)

func calculateDamage(user, target *pokemon, move *Move, crit *bool, weather weatherState, roll damageRollType, forScoring, pursuit bool) int {
	if f, ok := typeImmunityAbilities[target.ability]; ok && user.ability != moldBreakerAbility && f(target, move.Type, forScoring) {
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

	if f, ok := typeConvertingAbilities[user.ability]; ok {
		f(&moveType, &power)
	}
	numerator, denominator = target.applyMoveType(numerator, denominator, moveType)
	if weather != noneWeather {
		if f, ok := weatherFuncs[weather]; ok {
			f(&numerator, &denominator, moveType)
		}
		switch weather {
		case sunWeather:
			if user.ability == solarPowerAbility && move.Class == specialClass {
				offensiveStat = offensiveStat * 3 / 2
			}
		case sandstormWeather:
			if user.ability == sandForceAbility && (moveType == rockType || moveType == groundType || moveType == steelType) {
				power = power * 13 / 10
			}
		}
	}
	if numerator == 0 {
		return 0
	}

	switch move.Move {
	case "psywave":
		if roll == rollMax {
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
		return target.hp - user.hp
	case "super fang":
		*crit = false
		return max(1, target.hp/2)
	}

	if move.Move == "acrobatics" && (user.item.Consumed || user.item.State == flyingGem) {
		power *= 2
	} else if move.Move == "wake up slap" && target.hasAilment(sleepAilment) != nil {
		power *= 2
	} else if move.Move == "venoshock" && (target.hasAilment(poisonAilment) != nil || target.hasAilment(toxicAilment) != nil) {
		power *= 2
	} else if move.Move == "hex" && target.hasNonVolatileAilment() {
		power *= 2
	} else if move.Move == "flail" || move.Move == "reversal" {
		res := int(48 * (float64(user.hp) / float64(user.MaxHP())))
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
	} else if move.Move == "pursuit" && pursuit {
		power *= 2
	} else if move.Move == "knock off" && !target.item.Consumed {
		power *= 2
	}

	if user.ability == technicianAbility && move.Power <= 60 {
		power = power * 3 / 2
	} else if t, ok := pinchAbilities[user.ability]; ok && t == moveType && user.hp*3 <= user.MaxHP() {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.flashFire && moveType == fireType {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.ability == hustleAbility && move.Class == physicalClass {
		offensiveStat = offensiveStat * 3 / 2
	} else if user.ability == mercilessAbility {
		if a := target.hasAilment(poisonAilment); a != nil {
			*crit = true
		} else if a := target.hasAilment(toxicAilment); a != nil {
			*crit = true
		}
	}

	if target.ability.blocksCrits() && (forScoring || user.ability != moldBreakerAbility) {
		*crit = false
	}

	if user.hasType(moveType) {
		numerator *= 3
		denominator *= 2
	}

	if target.ability == drySkinAbility && moveType == fireType {
		power = power * 5 / 4
	}

	if *crit {
		if user.ability == sniperAbility {
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

	switch roll {
	case rollRandom:
		numerator *= rand.Intn(16) + 85
		denominator *= 100
	case rollMin:
		numerator *= 85
		denominator *= 100
	}

	damage := ((((2*user.level)/5)+2)*power*offensiveStat)/defensiveStat/50 + 2
	damage = damage * numerator / denominator

	target.checkItemTrigger(false, makeResistBerryEvent(moveType, &damage))

	damage = max(1, damage)

	return damage
}
