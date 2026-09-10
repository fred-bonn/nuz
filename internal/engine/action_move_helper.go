package engine

import (
	"math/rand"
)

var critRateMap = map[int]int{
	0: 16,
	1: 8,
	2: 2,
	3: 1,
	4: 1,
	5: 1,
	6: 1,
}

var confusionMove = Move{
	Name:  "confusion",
	Type:  NoType,
	Power: 40,
	Class: physicalClass,
}

var struggleMove = Move{
	Name:  "struggle",
	Type:  NoType,
	Power: 50,
	Class: physicalClass,
}

func calculateDamage(user, target *Pokemon, move *Move, crit *bool, weather WeatherState, maxRoll, forScoring, pursuit bool) int {
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
	if weather != NoneWeather {
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
	} else if user.flashFire && moveType == FireType {
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

	if target.Ability == drySkinAbility && moveType == FireType {
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

	user.CheckItemTrigger(false, makeGemEvent(moveType, &power))

	user.CheckItemTrigger(false, makeChoiceItemEvent(move, noStat, &offensiveStat))

	user.CheckItemTrigger(false, makeMoveBoostingEvent(moveType, &power))

	if !maxRoll {
		numerator *= rand.Intn(16) + 85
		denominator *= 100
	}

	damage := ((((2*user.level)/5)+2)*power*offensiveStat)/defensiveStat/50 + 2
	damage = damage * numerator / denominator

	target.CheckItemTrigger(false, makeResistBerryEvent(moveType, &damage))

	damage = max(1, damage)

	return damage
}

func roll(numerator int, denominator int) bool {
	return rand.Intn(denominator) < numerator
}

func rollInt(numerator int, denominator int) int {
	if roll(numerator, denominator) {
		return 1
	}
	return 0
}

func accuracyRoll(bs BattleState, user *Pokemon, target *Pokemon, move *Move) bool {
	if user.Ability == noGuardAbility || target.Ability == noGuardAbility {
		return true
	} else if move.Name == "toxic" && user.hasType(poisonType) {
		return true
	} else if move.Name == "thunder wave" && user.hasType(electricType) {
		return true
	}

	moveAccuracy := move.Accuracy
	if user.Ability == hustleAbility && move.Class == physicalClass {
		moveAccuracy = moveAccuracy * 80 / 100
	}

	accNum, accDen := user.accuracyFraction()
	evNum, evDen := target.evasionFraction(user.Ability == keenEyeAbility)
	numerator := moveAccuracy * accNum * evNum
	denominator := 100 * accDen * evDen
	if user.Ability == compoundEyesAbility {
		numerator *= 13
		denominator *= 10
	}

	if bs.getWeather() != NoneWeather {
		switch bs.getWeather() {
		case hailWeather:
			if target.Ability == snowCloakAbility {
				numerator *= 4
				denominator *= 5
			}
		case sandstormWeather:
			if target.Ability == sandVeilAbility {
				numerator *= 4
				denominator *= 5
			}
		}
	}

	return roll(numerator, denominator)
}

func determineHits(move *Move) int {
	if move.MaxHits == 5 && move.MinHits == 2 {
		r := rand.Intn(100) + 1
		if r <= 35 {
			return 2
		} else if r <= 70 {
			return 3
		} else if r <= 85 {
			return 4
		} else {
			return 5
		}
	}
	return move.MaxHits
}

func determineCrit(user *Pokemon, move *Move) *bool {
	rate := determineCritRate(user, move)

	return new(roll(1, critRateMap[rate]))
}

func determineCritRate(user *Pokemon, move *Move) int {
	if user.laserFocus {
		return 3
	}

	rate := move.CritRate
	if user.Item.State == scopeLens {
		rate++
	}
	if user.Ability == superLuckAbility {
		rate++
	}
	if user.focusEnergy {
		rate += 2
	}

	return rate
}

func monFainted(bs BattleState, slot *slot, pursuit bool) {
	if slot.mon.fainted {
		return
	}

	slot.mon.fainted = true
	if !pursuit {
		injectReplaceAction(bs, slot, false)
	}
	vprintf("%s fainted!", slot.mon.Base.Name)
}

func fetchPursuitMiddleware(name string) func(a action) bool {
	return func(a action) bool {
		ma, ok := a.(*moveAction)
		if !ok {
			return false
		}
		if ma.move.Name != "pursuit" {
			return false
		}
		if ma.targetSlot.mon.Base.Name != name {
			return false
		}
		return true
	}
}
