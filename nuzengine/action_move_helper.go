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

func accuracyRoll(bs battleState, user *pokemon, target *pokemon, move *Move) bool {
	if user.ability == noGuardAbility || target.ability == noGuardAbility {
		return true
	} else if move.Move == "toxic" && user.hasType(poisonType) {
		return true
	} else if move.Move == "thunder wave" && user.hasType(electricType) {
		return true
	}

	moveAccuracy := move.Accuracy
	if user.ability == hustleAbility && move.Class == physicalClass {
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

func determineCrit(user *pokemon, move *Move) *bool {
	rate := determineCritRate(user, move)

	return new(roll(1, critRateMap[rate]))
}

func determineCritRate(user *pokemon, move *Move) int {
	if user.laserFocus {
		return 3
	}

	rate := move.CritRate
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
		if ma.move.Move != "pursuit" {
			return false
		}
		if ma.targetSlot.mon.base.Name != name {
			return false
		}
		return true
	}
}

func getStruggleMove() *Move {
	return &Move{
		Move:  "struggle",
		Type:  noType,
		Power: 50,
		Class: physicalClass,
	}
}

func getConfusionMove() *Move {
	return &Move{
		Move:  "confusion",
		Type:  noType,
		Power: 40,
		Class: physicalClass,
	}
}
