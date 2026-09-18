package nuzengine

type discreteSingleBattleState struct {
	playerMon              *pokemon
	playerMonHasKill       bool
	playerMonHasFastKill   bool
	playerMonIsTrapped     bool
	alivePlayerMons        []bool
	opponentMon            *pokemon
	opponentMonHasKill     bool
	opponentMonHasFastKill bool
	weather                weatherState
}

func (sbs *singleBattleState) discretize() discreteBattleState {
	playerMon := sbs.activePlayerSlot.mon
	opponentMon := sbs.activeOpponentSlot.mon

	alive := make([]bool, len(sbs.player.pokemonParty))
	for i, mon := range sbs.player.pokemonParty {
		alive[i] = !mon.fainted
	}

	playerMinDamage := calculateMaxDamage(sbs, playerMon, opponentMon, true, false) * 85 / 100
	playerMonHasKill := playerMinDamage >= opponentMon.hp

	playerMinPriorityDamage := calculateMaxDamage(sbs, playerMon, opponentMon, true, true) * 85 / 100
	playerMonHasFastKill := playerMinPriorityDamage >= opponentMon.hp || (playerMonHasKill && playerMon.isFasterThan(sbs, opponentMon))

	opponentMaxDamage := calculateMaxDamage(sbs, opponentMon, playerMon, true, false)
	opponentMonHasKill := opponentMaxDamage >= playerMon.hp

	opponentMaxPriorityDamage := calculateMaxDamage(sbs, opponentMon, playerMon, true, true)
	opponentMonHasFastKill := opponentMaxPriorityDamage >= playerMon.hp || (opponentMaxDamage >= playerMon.hp && opponentMon.isFasterThan(sbs, playerMon))

	return discreteSingleBattleState{
		playerMon:              playerMon,
		playerMonHasKill:       playerMonHasKill,
		playerMonHasFastKill:   playerMonHasFastKill,
		playerMonIsTrapped:     sbs.activePlayerSlot.isTrapped(),
		alivePlayerMons:        alive,
		opponentMon:            sbs.activeOpponentSlot.mon,
		opponentMonHasKill:     opponentMonHasKill,
		opponentMonHasFastKill: opponentMonHasFastKill,
		weather:                sbs.weather,
	}
}
