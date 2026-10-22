package nuzengine

import "strconv"

type singleBattleStateKey struct {
	PlayerMon              string
	PlayerMonHasKill       bool
	PlayerMonHasFastKill   bool
	PlayerMonIsTrapped     bool
	PlayerItemUsed         bool
	PlayerOutOfPP          uint64
	PlayerAilment          ailmentState
	AlivePlayerMons        uint64
	OpponentMon            string
	OpponentMonHasKill     bool
	OpponentMonHasFastKill bool
	OpponentItemUsed       bool
	OpponentOutOfPP        uint64
	OpponentAilment        ailmentState
	AliveOpponentMons      uint64
	Weather                weatherState
}

// String renders the key as "mon.flags.ailment" for the player and opponent
// followed by the alive masks and weather, all numbers in base 36.
func (stateKey singleBattleStateKey) String() string {
	flags := func(bools ...bool) uint64 {
		var bits uint64
		for _, flag := range bools {
			bits <<= 1
			if flag {
				bits |= 1
			}
		}
		return bits
	}
	playerBits := flags(stateKey.PlayerMonHasKill, stateKey.PlayerMonHasFastKill, stateKey.PlayerMonIsTrapped, stateKey.PlayerItemUsed)<<8 | stateKey.PlayerOutOfPP&0xff
	opponentBits := flags(stateKey.OpponentMonHasKill, stateKey.OpponentMonHasFastKill, stateKey.OpponentItemUsed)<<8 | stateKey.OpponentOutOfPP&0xff

	buf := make([]byte, 0, 64)
	buf = append(buf, stateKey.PlayerMon...)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, playerBits, 36)
	buf = append(buf, '.')
	buf = strconv.AppendInt(buf, int64(stateKey.PlayerAilment), 36)
	buf = append(buf, '|')
	buf = append(buf, stateKey.OpponentMon...)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, opponentBits, 36)
	buf = append(buf, '.')
	buf = strconv.AppendInt(buf, int64(stateKey.OpponentAilment), 36)
	buf = append(buf, '|')
	buf = strconv.AppendUint(buf, stateKey.AlivePlayerMons, 36)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, stateKey.AliveOpponentMons, 36)
	buf = append(buf, '.')
	buf = strconv.AppendInt(buf, int64(stateKey.Weather), 36)
	return string(buf)
}

func (sbs *singleBattleState) key() string {
	playerMon := sbs.activePlayerSlot.mon
	opponentMon := sbs.activeOpponentSlot.mon

	stateKey := singleBattleStateKey{
		Weather:           sbs.weather,
		PlayerMon:         playerMon.base.Name,
		OpponentMon:       opponentMon.base.Name,
		PlayerAilment:     addNonVolatileAilmentState(playerMon),
		OpponentAilment:   addNonVolatileAilmentState(opponentMon),
		PlayerItemUsed:    isItemUsed(playerMon),
		OpponentItemUsed:  isItemUsed(opponentMon),
		PlayerOutOfPP:     outOfPPMask(playerMon),
		OpponentOutOfPP:   outOfPPMask(opponentMon),
		AlivePlayerMons:   aliveMask(sbs.player.pokemonParty),
		AliveOpponentMons: aliveMask(sbs.opponent.pokemonParty),
	}

	playerMonMax := calculateMaxDamageAmongMoves(sbs, playerMon, opponentMon, calculateMaxDamageFlags{
		checkChoiceItem: true,
		isFirstTurn:     sbs.activePlayerSlot.firstTurn,
		rollType:        rollMin,
	})
	playerMonMaxPrio := calculateMaxDamageAmongMoves(sbs, playerMon, opponentMon, calculateMaxDamageFlags{
		checkChoiceItem:   true,
		checkOnlyPriority: true,
		isFirstTurn:       sbs.activePlayerSlot.firstTurn,
		rollType:          rollMin,
	})
	opponentMonMax := calculateMaxDamageAmongMoves(sbs, opponentMon, playerMon, calculateMaxDamageFlags{
		checkChoiceItem: true,
		isCrit:          true,
		isFirstTurn:     sbs.activeOpponentSlot.firstTurn,
		rollType:        rollMax,
	})
	opponentMonMaxPrio := calculateMaxDamageAmongMoves(sbs, opponentMon, playerMon, calculateMaxDamageFlags{
		checkChoiceItem:   true,
		checkOnlyPriority: true,
		isCrit:            true,
		isFirstTurn:       sbs.activeOpponentSlot.firstTurn,
		rollType:          rollMax,
	})

	stateKey.PlayerMonHasKill = playerMonMax >= opponentMon.hp
	stateKey.PlayerMonHasFastKill = playerMonMaxPrio >= opponentMon.hp || (stateKey.PlayerMonHasKill && playerMon.isFasterThan(sbs, opponentMon))
	stateKey.OpponentMonHasKill = opponentMonMax >= playerMon.hp
	stateKey.OpponentMonHasFastKill = opponentMonMaxPrio >= playerMon.hp || (stateKey.OpponentMonHasKill && opponentMon.isFasterThan(sbs, playerMon))
	stateKey.PlayerMonIsTrapped = sbs.activePlayerSlot.isTrapped()

	return stateKey.String()
}

func isItemUsed(mon *pokemon) bool {
	return mon.item == nil || mon.item.State == noneItem || mon.item.Consumed
}

func outOfPPMask(mon *pokemon) uint64 {
	var mask uint64
	for i, move := range mon.moves {
		if i < 8 && move.PP <= 0 {
			mask |= 1 << i
		}
	}
	return mask
}

func addNonVolatileAilmentState(mon *pokemon) ailmentState {
	for ailment := range mon.ailments {
		if ailment.isNonVolatileStatus() {
			return ailment
		}
	}
	return noneAilment
}
