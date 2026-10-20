package nuzengine

import (
	"math/rand"
)

type rnbAi struct{}

func (rnb rnbAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	opponentSlot := bs.getOpponentSlot(slot)
	for _, mon := range party {
		if mon == slot.mon || mon.fainted {
			continue
		}

		opponentDamage := calculateMaxDamageAmongMoves(bs, opponentSlot.mon, mon, calculateMaxDamageFlags{
			isFirstTurn: opponentSlot.firstTurn,
			rollType:    rollMax,
		})
		oneHitKill := opponentDamage >= mon.hp
		twoHitKillWhileSlower := opponentDamage*2 >= mon.hp && !opponentSlot.mon.isFasterThan(bs, mon)
		if !oneHitKill && !twoHitKillWhileSlower {
			return score <= 0 && !roll(1, 2) && slot.mon.hp > slot.mon.MaxHP()/2
		}
	}

	return false
}

func (rnb rnbAi) evaluateActions(bs battleState, userSlot *slot, actions []*moveAction) (*moveAction, int) {
	scores := make([]int, len(actions))
	damage := make([]int, len(actions))
	kills := make([]bool, len(actions))
	highestDamageIndex := make(map[*slot]int)
	canHighestKill := make(map[*slot]bool)

	for i, a := range actions {
		if a.move.PP <= 0 {
			damage[i] = -1
			scores[i] = -64
			continue
		}

		// TODO: add exceptions for moves that are useful against allies
		if a.targetSlot != userSlot && a.targetSlot.Trainer == userSlot.Trainer {
			damage[i] = -1
			scores[i] = -64
			continue
		}

		if a.move.Class == statusClass {
			damage[i] = -1
			scores[i], _ = a.scoreActionMove(bs)
			continue
		}

		if a.move.Move == "nuzzle" {
			damage[i] = -1
			scores[i] = a.scoreParalysisMove(bs)
			continue
		} else if a.move.Move == "rollout" {
			damage[i] = -1
			scores[i] = 7
			continue
		} else if a.move.Move == "fake out" {
			if !a.userSlot.firstTurn || a.targetSlot.mon.ability == innerFocusAbility || a.targetSlot.mon.ability == shieldDustAbility {
				damage[i] = -1
				scores[i] = -64
				continue
			}
			scores[i] = 9

		} else if a.move.Move == "first impression" && !a.userSlot.firstTurn {
			damage[i] = -1
			scores[i] = -64
			continue
		} else if a.move.Move == "belch" && (!a.userSlot.mon.item.State.isBerry() || !a.userSlot.mon.item.Consumed) {
			damage[i] = -1
			scores[i] = -64
			continue
		} else if a.move.Move == "sucker punch" && a.userSlot.suckerPunch && roll(1, 2) {
			scores[i] = -20
		}

		if a.move.Ailment == trapAilment {
			if _, ok := a.targetSlot.mon.ailments[trapAilment]; !ok {
				scores[i] = 6 + 2*rollInt(1, 5)
			}
		}

		damage[i], kills[i] = a.scoreActionMove(bs)
		if damage[i] == 0 {
			damage[i] = -1
			scores[i] = -64
		}

		if kills[i] {
			canHighestKill[a.targetSlot] = true
			continue
		}
		highestIndex, ok := highestDamageIndex[a.targetSlot]
		if !ok || damage[highestIndex] < damage[i] {
			highestDamageIndex[a.targetSlot] = i
		}
	}

	for i, a := range actions {
		if damage[i] == -1 {
			continue
		}

		// add additional scoring to damaging moves
		if a.move.Move == "fell stinger" && kills[i] {
			if a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
				scores[i] = 21 + 2*rollInt(1, 5)
			} else {
				scores[i] = 15 + 2*rollInt(1, 5)
			}
		} else if a.move.Move == "acid spray" {
			scores[i] = 6
		} else if a.move.Move == "future sight" {
			scores[i] = 6
			if a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) && calculateMaxDamageAmongMoves(bs, a.targetSlot.mon, a.userSlot.mon, calculateMaxDamageFlags{
				isFirstTurn: a.targetSlot.firstTurn,
				rollType:    rollMax,
			}) >= a.userSlot.mon.hp {
				scores[i] = 8
			}
		} else if a.move.Move == "pursuit" {
			if kills[i] {
				scores[i] = 10
			} else {
				if a.targetSlot.mon.hp < a.targetSlot.mon.MaxHP()/5 {
					scores[i] = 10
				} else if a.targetSlot.mon.hp < a.targetSlot.mon.MaxHP()*2/5 {
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
			for _, move := range a.targetSlot.mon.moves {
				if move.PP <= 0 || move.Class == statusClass {
					continue
				}
				if a.targetSlot.mon.lockedMove != nil && a.targetSlot.mon.lockedMove != move {
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
					dmg += calculateDamage(a.targetSlot.mon, a.userSlot.mon, move, new(critRate >= 3), bs.getWeather(), rollRandom, true, false)
				}
				if a.userSlot.mon.hp <= dmg {
					scores[i] += 11
					break
				}
			}
		}

		// If a move can kill this target, only consider moves that can kill it.
		if canHighestKill[a.targetSlot] {
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

		// Award the top-damage bonus independently for each target slot.
		if highestDamageIndex[a.targetSlot] == i {
			scores[i] += 6 + 2*rollInt(1, 5)
			continue
		}

		// moves from this point that gets a base score if and only if it neither kills or is highest damage
		if isSpeedControlMove(a.move.Move) && !a.userSlot.mon.isFasterThan(bs, a.targetSlot.mon) {
			scores[i] = 6
			continue
		}

		if c, ok := isOffenseControlMove(a.move.Move); ok {
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

	// vprintln(scores)

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

func (rnb rnbAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if len(mons) == 1 {
		return mons[0]
	}

	scores := make([]int, len(mons))
	opponent := opponentSlot.mon

	for i, mon := range mons {
		if mon.base.Name == "ditto" || mon.base.Name == "wobbufet" {
			scores[i] = 2
			continue
		}

		outspeeds := mon.isFasterThan(bs, opponent)

		monDamage := calculateMaxDamageAmongMoves(bs, mon, opponent, calculateMaxDamageFlags{
			isSwitchIn: true,
			rollType:   rollMax,
		})
		opponentDamage := calculateMaxDamageAmongMoves(bs, opponent, mon, calculateMaxDamageFlags{
			isFirstTurn: opponentSlot.firstTurn,
			rollType:    rollMax,
		})

		killsOpponent := monDamage >= opponent.hp
		monKilled := opponentDamage >= mon.hp

		monDamagePercent := monDamage * 100 / max(1, opponent.hp)
		opponentDamagePercent := opponentDamage * 100 / max(1, mon.hp)

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

type calculateMaxDamageFlags struct {
	rollType          damageRollType
	checkChoiceItem   bool
	checkOnlyPriority bool
	isCrit            bool
	isSwitchIn        bool
	isFirstTurn       bool
}

func calculateMaxDamageAmongMoves(bs battleState, user, target *pokemon, flags calculateMaxDamageFlags) int {
	var maxDmg, dmg int
	rolls := 1
	for _, move := range user.moves {
		if flags.checkOnlyPriority && move.Priority == 0 {
			continue
		}
		if (move.Move == "fake out" || move.Move == "first impression") && !(flags.isSwitchIn || flags.isFirstTurn) {
			continue
		}
		if move.PP <= 0 || move.Class == statusClass {
			continue
		}
		if flags.checkChoiceItem && user.lockedMove != nil && user.lockedMove != move {
			continue
		}

		rolls = 1
		flags.isCrit = flags.isCrit || determineCritRate(user, move) > 3
		if move.MaxHits == 5 {
			rolls = 3
		} else if move.MaxHits > 0 {
			rolls = move.MaxHits
		}
		for i := 0; i < rolls; i++ {
			dmg += calculateDamage(user, target, move, &flags.isCrit, bs.getWeather(), flags.rollType, true, false)
		}

		target.checkItemTrigger(false, makeFocusSashEvent(&dmg))
		if target.ability == sturdyAbility && target.hp == target.MaxHP() {
			dmg = min(dmg, target.hp-1)
		}

		if dmg > maxDmg {
			maxDmg = dmg
		}
		dmg = 0
	}
	return maxDmg
}
