package nuzengine

type discreteSingleBattleState struct {
	playerMon              *Pokemon
	playerMonHasKill       bool
	playerHasFastKill      bool
	playerMonIsTrapped     bool
	alivePlayerMons        []bool
	opponentMon            *Pokemon
	opponentMonHasKill     bool
	opponentMonHasFastKill bool
	weather                weatherState
}

func (sbs *singleBattleState) discretize() discreteBattleState {
	alive := make([]bool, len(sbs.Player.PokemonParty))
	for i, mon := range sbs.Player.PokemonParty {
		alive[i] = !mon.fainted
	}
	playerMinDamage := calculateMaxDamage(sbs, sbs.activePlayerSlot.mon, sbs.activeOpponentSlot.mon, true) * 85 / 100
	opponentMaxDamage := calculateMaxDamage(sbs, sbs.activeOpponentSlot.mon, sbs.activePlayerSlot.mon, true)

	return discreteSingleBattleState{
		playerMon:          sbs.activePlayerSlot.mon,
		playerMonHasKill:   playerMinDamage >= sbs.activeOpponentSlot.mon.HP,
		playerMonIsTrapped: sbs.activePlayerSlot.isTrapped(),
		alivePlayerMons:    alive,
		opponentMon:        sbs.activeOpponentSlot.mon,
		opponentMonHasKill: opponentMaxDamage >= sbs.activePlayerSlot.mon.HP,
		weather:            sbs.weather,
	}
}
