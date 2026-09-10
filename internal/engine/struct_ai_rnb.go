package engine

import (
	"math/rand"
)

type RnbAi struct{}

func (rnb RnbAi) shouldSwitch(bs BattleState, slot *slot, score int, party []*Pokemon) bool {
	opponent := bs.getOpponentSlot(slot).mon
	for _, mon := range party {
		if mon == slot.mon || mon.fainted {
			continue
		}

		opponentDamage := calculateMaxDamage(bs, opponent, mon, false)
		oneHitKill := opponentDamage >= mon.HP
		twoHitKillWhileSlower := opponentDamage*2 >= mon.HP && !opponent.isFasterThan(bs, mon)
		if !oneHitKill && !twoHitKillWhileSlower {
			return score <= 0 && !roll(1, 2) && slot.mon.HP > slot.mon.MaxHP()/2
		}
	}

	return false
}

func (rnb RnbAi) evaluateActions(bs BattleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	scores := make([]int, len(actions))
	damage := make([]int, len(actions))
	kills := make([]bool, len(actions))
	highestDamageIndex := -1
	canHighestKill := false

	for i, a := range actions {
		if a.move.PP <= 0 {
			damage[i] = -1
			scores[i] = -64
			continue
		}

		if a.move.Class == statusClass {
			damage[i] = -1
			scores[i], _ = a.scoreActionMove(bs)
			continue
		}

		if a.move.Name == "nuzzle" {
			damage[i] = -1
			scores[i] = a.scoreParalysisMove(bs)
			continue
		} else if a.move.Name == "rollout" {
			damage[i] = -1
			scores[i] = 7
			continue
		} else if a.move.Name == "fake out" {
			if !a.userSlot.firstTurn || a.targetSlot.mon.Ability == innerFocusAbility || a.targetSlot.mon.Ability == shieldDustAbility {
				damage[i] = -1
				scores[i] = -64
				continue
			}
			scores[i] = 9

		} else if a.move.Name == "first impression" && !a.userSlot.firstTurn {
			damage[i] = -1
			scores[i] = -64
			continue
		} else if a.move.Name == "belch" && (!a.userSlot.mon.Item.State.isBerry() || !a.userSlot.mon.Item.Consumed) {
			damage[i] = -1
			scores[i] = -64
			continue
		} else if a.move.Name == "sucker punch" && a.userSlot.suckerPunch && roll(1, 2) {
			scores[i] = -20
		}

		if a.move.Ailment == trapAilment {
			if _, ok := a.targetSlot.mon.Ailments[trapAilment]; !ok {
				scores[i] = 6 + 2*rollInt(1, 5)
			}
		}

		damage[i], kills[i] = a.scoreActionMove(bs)
		if damage[i] == 0 {
			damage[i] = -1
			scores[i] = -64
		}
		// vlogln(a.move.Name, damage[i])

		canHighestKill = canHighestKill || kills[i]
		if !canHighestKill && (highestDamageIndex == -1 || damage[highestDamageIndex] < damage[i]) {
			highestDamageIndex = i
		}
	}

	for i, a := range actions {
		if damage[i] == -1 {
			continue
		}

		// add additional scoring to damaging moves
		if a.move.Name == "fell stinger" && kills[i] {
			if a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
				scores[i] = 21 + 2*rollInt(1, 5)
			} else {
				scores[i] = 15 + 2*rollInt(1, 5)
			}
		} else if a.move.Name == "acid spray" {
			scores[i] = 6
		} else if a.move.Name == "future sight" {
			// needs 8 score instead if faster and dead to target
			scores[i] = 6
		} else if a.move.Name == "pursuit" {
			if kills[i] {
				scores[i] = 10
			} else {
				if a.targetSlot.mon.HP < a.targetSlot.mon.MaxHP()/5 {
					scores[i] = 10
				} else if a.targetSlot.mon.HP < a.targetSlot.mon.MaxHP()*2/5 {
					scores[i] = 8 * rollInt(1, 2)
				}
			}
			if a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
				vprintln(3)
				scores[i] += 3
			}
		}

		// add score if fast dead and the move has priority
		if a.move.Priority > 0 && !a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
			for _, move := range a.targetSlot.mon.Moves {
				if move.PP <= 0 || move.Class == statusClass {
					continue
				}
				if a.targetSlot.mon.LockedMove != nil && a.targetSlot.mon.LockedMove != move {
					continue
				}

				dmg := 0
				critRate := determineCritRate(a.userSlot.mon, move)
				rolls := 1
				if move.MaxHits == 5 {
					rolls = 3
				} else if move.MaxHits > 0 {
					rolls = move.MaxHits
				}
				for i := 0; i < rolls; i++ {
					dmg += calculateDamage(a.targetSlot.mon, a.userSlot.mon, move, new(critRate >= 3), bs.getWeather(), false, true, false)
				}
				if a.userSlot.mon.HP <= dmg {
					scores[i] += 11
					break
				}
			}
		}

		// if the highest damaging move kills then we only have to consider moves that can kill
		if canHighestKill {
			if !kills[i] {
				continue
			}
			if a.move.Priority > 0 || a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
				scores[i] += 12 + 2*rollInt(1, 5)
			} else {
				scores[i] += 9 + 2*rollInt(1, 5)
			}
			continue
		}

		// if no moves kills then the highest damaging moves gets additional score
		if highestDamageIndex == i {
			scores[i] += 6 + 2*rollInt(1, 5)
			continue
		}

		// moves from this point that gets a base score if and only if it neither kills or is highest damage
		if isSpeedControlMove(a.move.Name) && !a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
			scores[i] = 6
			continue
		}

		if c, ok := isOffenseControlMove(a.move.Name); ok {
			if a.targetSlot.mon.hasMovePredicate(func(m *Move) bool {
				return m.Class == c
			}) {
				scores[i] = 6
			} else {
				scores[i] = 5
			}
			continue
		}
	}

	vprintln(scores)

	maxScore := scores[0]
	for _, score := range scores {
		if score > maxScore {
			maxScore = score
		}
	}

	var bestIndices []int
	for i, score := range scores {
		if score == maxScore {
			bestIndices = append(bestIndices, i)
		}
	}

	resultIndex := rand.Intn(len(bestIndices))

	return actions[bestIndices[resultIndex]], scores[bestIndices[resultIndex]]
}

func (rnb RnbAi) evaluteSwitchIns(bs BattleState, mons []*Pokemon, opponentSlot *slot) *Pokemon {
	if len(mons) == 1 {
		return mons[0]
	}

	scores := make([]int, len(mons))
	opponent := opponentSlot.mon

	for i, mon := range mons {
		if mon.Base.Name == "ditto" || mon.Base.Name == "wobbufet" {
			scores[i] = 2
			continue
		}

		outspeeds := mon.isFasterThan(bs, opponent)

		monDamage := calculateMaxDamage(bs, mon, opponent, false)
		opponentDamage := calculateMaxDamage(bs, opponent, mon, false)

		killsOpponent := monDamage >= opponent.HP
		monKilled := opponentDamage >= mon.HP

		monDamagePercent := monDamage * 100 / max(1, opponent.HP)
		opponentDamagePercent := opponentDamage * 100 / max(1, mon.HP)

		if outspeeds && killsOpponent {
			scores[i] = 5
		} else if !outspeeds && killsOpponent && !monKilled {
			scores[i] = 4
		} else if outspeeds && monDamagePercent > opponentDamagePercent {
			scores[i] = 3
		} else if !outspeeds && monDamagePercent > opponentDamagePercent && !monKilled {
			scores[i] = 2
		} else if outspeeds {
			scores[i] = 1
		} else if !outspeeds && monKilled {
			scores[i] = -1
		}
		// vlogln("%s switch in: outspeeds=%t, monDamage=%d, opponentDamage=%d, killsOpponent=%t, monKilled=%t, monDamagePercent=%d, opponentDamagePercent=%d, score=%d\n", mon.Base.Name, outspeeds, monDamage, opponentDamage, killsOpponent, monKilled, monDamagePercent, opponentDamagePercent, scores[i])
	}

	maxScore := scores[0]
	bestIndex := 0
	for i := 1; i < len(scores); i++ {
		if scores[i] > maxScore {
			maxScore = scores[i]
			bestIndex = i
		}
	}

	// vlogln(scores[bestIndex])

	return mons[bestIndex]
}

func calculateMaxDamage(bs BattleState, user, target *Pokemon, checkChoice bool) int {
	var maxDmg, dmg int
	rolls := 1
	for _, move := range user.Moves {
		if move.PP <= 0 || move.Class == statusClass {
			continue
		}
		if checkChoice && user.LockedMove != nil && user.LockedMove != move {
			continue
		}

		rolls = 1
		critRate := determineCritRate(user, move)
		if move.MaxHits == 5 {
			rolls = 3
		} else if move.MaxHits > 0 {
			rolls = move.MaxHits
		}
		for i := 0; i < rolls; i++ {
			dmg += calculateDamage(user, target, move, new(critRate >= 3), bs.getWeather(), true, true, false)
		}

		target.CheckItemTrigger(false, makeFocusSashEvent(&dmg))
		if target.Ability == sturdyAbility && target.HP == target.MaxHP() {
			dmg = min(dmg, target.HP-1)
		}

		if dmg > maxDmg {
			maxDmg = dmg
		}
		dmg = 0
	}
	return maxDmg
}
