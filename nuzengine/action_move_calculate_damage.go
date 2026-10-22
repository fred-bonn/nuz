package nuzengine

import "math/rand"

type damageRollType int

const (
	rollRandom = iota
	rollMin
	rollMax
)

func calculateDamage(bs battleState, user, target *pokemon, move *Move, crit *bool, weather weatherState, roll damageRollType, forScoring, pursuit bool) int {
	if f, ok := typeImmunityAbilities[target.ability]; ok && user.ability != moldBreakerAbility && f(bs, target, move.Type, forScoring) {
		return 0
	}

	typeNum, typeDen := 1, 1
	weatherNum, weatherDen := 1, 1
	moveType := move.Type
	power := move.Power
	var offensiveStat, defensiveStat int
	if move.Class == physicalClass {
		offensiveStat = user.effectiveStat(bs, attack, *crit)
		defensiveStat = target.effectiveStat(bs, defense, *crit)
	} else {
		offensiveStat = user.effectiveStat(bs, specialAttack, *crit)
		defensiveStat = target.effectiveStat(bs, specialDefense, *crit)
		if weather == sandstormWeather && target.hasType(rockType) {
			defensiveStat = defensiveStat * 3 / 2
		}
	}

	if f, ok := typeConvertingAbilities[user.ability]; ok {
		f(&moveType, &power)
	}
	typeNum, typeDen = target.applyMoveType(typeNum, typeDen, moveType)
	if weather != noneWeather {
		if f, ok := weatherFuncs[weather]; ok {
			f(&weatherNum, &weatherDen, moveType)
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
	if typeNum == 0 {
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
		return max(0, target.hp-user.hp)
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
	} else if move.Move == "knock off" && target.item.State != noneItem && !target.item.Consumed {
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

	if target.ability == drySkinAbility && moveType == fireType {
		power = power * 5 / 4
	}

	user.checkItemTrigger(bs, false, makeGemEvent(moveType, &power))

	user.checkItemTrigger(bs, false, makeChoiceItemEvent(move, noStat, &offensiveStat))

	user.checkItemTrigger(bs, false, makeMoveBoostingEvent(moveType, &power))

	damage := ((((2*user.level)/5)+2)*power*offensiveStat)/defensiveStat/50 + 2
	damage = damage * weatherNum / weatherDen

	if *crit {
		if user.ability == sniperAbility {
			damage = damage * 3 / 2
		}
		damage = damage * 3 / 2
	}

	switch roll {
	case rollRandom:
		damage = damage * (rand.Intn(16) + 85) / 100
	case rollMin:
		damage = damage * 85 / 100
	}

	if user.hasType(moveType) {
		damage = damage * 3 / 2
	}

	damage = damage * typeNum / typeDen

	if move.Class == physicalClass && user.hasAilment(burnAilment) != nil {
		damage /= 2
	}

	target.checkItemTrigger(bs, false, makeResistBerryEvent(moveType, &damage))

	damage = max(1, damage)

	return damage
}
