package nuzengine

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

func roll(numerator int, denominator int) bool {
	return rand.Intn(denominator) < numerator
}

func rollInt(numerator int, denominator int) int {
	if roll(numerator, denominator) {
		return 1
	}
	return 0
}

func accuracyRoll(bs battleState, user *pokemon, target *pokemon, move *move) bool {
	if user.ability == noGuardAbility || target.ability == noGuardAbility {
		return true
	} else if move.name == "toxic" && user.hasType(poisonType) {
		return true
	} else if move.name == "thunder wave" && user.hasType(electricType) {
		return true
	}

	moveAccuracy := move.accuracy
	if user.ability == hustleAbility && move.class == physicalClass {
		moveAccuracy = moveAccuracy * 80 / 100
	}

	accNum, accDen := user.accuracyFraction()
	evNum, evDen := target.evasionFraction(user.ability == keenEyeAbility)
	numerator := moveAccuracy * accNum * evNum
	denominator := 100 * accDen * evDen
	if user.ability == compoundEyesAbility {
		numerator *= 13
		denominator *= 10
	}

	if bs.getWeather() != noneWeather {
		switch bs.getWeather() {
		case hailWeather:
			if target.ability == snowCloakAbility {
				numerator *= 4
				denominator *= 5
			}
		case sandstormWeather:
			if target.ability == sandVeilAbility {
				numerator *= 4
				denominator *= 5
			}
		}
	}

	return roll(numerator, denominator)
}

func determineHits(move *move) int {
	if move.maxHits == 5 && move.minHits == 2 {
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
	return move.maxHits
}

func determineCrit(user *pokemon, move *move) *bool {
	rate := determineCritRate(user, move)

	return new(roll(1, critRateMap[rate]))
}

func determineCritRate(user *pokemon, move *move) int {
	if user.laserFocus {
		return 3
	}

	rate := move.critRate
	if user.item.State == scopeLens {
		rate++
	}
	if user.ability == superLuckAbility {
		rate++
	}
	if user.focusEnergy {
		rate += 2
	}

	return rate
}

func monFainted(bs battleState, slot *slot, pursuit bool) {
	if slot.mon.fainted {
		return
	}

	slot.mon.fainted = true
	if !pursuit {
		injectReplaceAction(bs, slot, false)
	}
	vprintf("%s fainted!", slot.mon.base.Name)
}

func fetchPursuitMiddleware(name string) func(a action) bool {
	return func(a action) bool {
		ma, ok := a.(*moveAction)
		if !ok {
			return false
		}
		if ma.move.name != "pursuit" {
			return false
		}
		if ma.targetSlot.mon.base.Name != name {
			return false
		}
		return true
	}
}

func getStruggleMove() *move {
	return &move{
		name:     "struggle",
		moveType: noType,
		power:    50,
		class:    physicalClass,
	}
}

func getConfusionMove() *move {
	return &move{
		name:     "confusion",
		moveType: noType,
		power:    40,
		class:    physicalClass,
	}
}
