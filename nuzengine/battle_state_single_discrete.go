package nuzengine

type discreteSingleBattleState struct {
	playerMon              *pokemon
	playerMonHasKill       bool
	playerHasFastKill      bool
	playerMonIsTrapped     bool
	alivePlayerMons        []bool
	opponentMon            *pokemon
	opponentMonHasKill     bool
	opponentMonHasFastKill bool
	weather                weatherState
}

func (sbs *singleBattleState) discretize() discreteBattleState {
	alive := make([]bool, len(sbs.player.pokemonParty))
	for i, mon := range sbs.player.pokemonParty {
		alive[i] = !mon.fainted
	}
	playerMinDamage := calculateMaxDamage(sbs, sbs.activePlayerSlot.mon, sbs.activeOpponentSlot.mon, true) * 85 / 100
	opponentMaxDamage := calculateMaxDamage(sbs, sbs.activeOpponentSlot.mon, sbs.activePlayerSlot.mon, true)

	return discreteSingleBattleState{
		playerMon:          sbs.activePlayerSlot.mon,
		playerMonHasKill:   playerMinDamage >= sbs.activeOpponentSlot.mon.hp,
		playerMonIsTrapped: sbs.activePlayerSlot.isTrapped(),
		alivePlayerMons:    alive,
		opponentMon:        sbs.activeOpponentSlot.mon,
		opponentMonHasKill: opponentMaxDamage >= sbs.activePlayerSlot.mon.hp,
		weather:            sbs.weather,
	}
}
