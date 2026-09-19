package nuzengine

import "fmt"

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
	discreteState := discreteSingleBattleState{
		weather:         sbs.weather,
		alivePlayerMons: make([]bool, len(sbs.player.pokemonParty)),
	}

	discreteState.update(sbs)

	return &discreteState
}

func (dsbs *discreteSingleBattleState) update(bs battleState) bool {
	sbs, ok := bs.(*singleBattleState)
	if !ok {
		return ok
	}

	playerMon := sbs.activePlayerSlot.mon
	opponentMon := sbs.activeOpponentSlot.mon

	dsbs.playerMon = playerMon
	dsbs.opponentMon = opponentMon

	for i, mon := range sbs.player.pokemonParty {
		dsbs.alivePlayerMons[i] = !mon.fainted
	}

	playerMonMax := calculateMaxDamageAmongMoves(sbs, playerMon, opponentMon, true, false, rollMin)
	playerMonMaxPrio := calculateMaxDamageAmongMoves(sbs, playerMon, opponentMon, true, true, rollMin)
	opponentMonMax := calculateMaxDamageAmongMoves(sbs, opponentMon, playerMon, true, false, rollMax)
	opponentMonMaxPrio := calculateMaxDamageAmongMoves(sbs, opponentMon, playerMon, true, true, rollMax)

	dsbs.playerMonHasKill = playerMonMax >= opponentMon.hp
	dsbs.playerMonHasFastKill = playerMonMaxPrio >= opponentMon.hp || (dsbs.playerMonHasKill && playerMon.isFasterThan(sbs, opponentMon))
	dsbs.opponentMonHasKill = opponentMonMax >= playerMon.hp
	dsbs.opponentMonHasFastKill = opponentMonMaxPrio >= playerMon.hp || (dsbs.opponentMonHasKill && opponentMon.isFasterThan(sbs, playerMon))

	dsbs.playerMonIsTrapped = sbs.activePlayerSlot.isTrapped()

	return true
}

func (dsbs *discreteSingleBattleState) String() string {
	return fmt.Sprintf(
		"PlayerMon: %v, PlayerMonHasKill: %v, PlayerMonHasFastKill: %v, PlayerMonIsTrapped: %v, AlivePlayerMons: %v, OpponentMon: %v, OpponentMonHasKill: %v, OpponentMonHasFastKill: %v, Weather: %v",
		dsbs.playerMon.base.Name,
		dsbs.playerMonHasKill,
		dsbs.playerMonHasFastKill,
		dsbs.playerMonIsTrapped,
		dsbs.alivePlayerMons,
		dsbs.opponentMon.base.Name,
		dsbs.opponentMonHasKill,
		dsbs.opponentMonHasFastKill,
		dsbs.weather,
	)
}
