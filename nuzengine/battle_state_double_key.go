package nuzengine

import "strconv"

// doubleSlotKey describes one active slot. The kill flags are indexed by the
// opponent's slot order (one, two) and are false for fainted targets.
type doubleSlotKey struct {
	Mon         string
	Fainted     bool
	HasKill     [2]bool
	HasFastKill [2]bool
	IsTrapped   bool
	Ailment     ailmentState
	ItemUsed    bool
	OutOfPP     uint64
}

// bits packs the slot's flags into one number, so the key stays short.
func (k doubleSlotKey) bits() uint64 {
	var bits uint64
	for _, flag := range []bool{k.Fainted, k.HasKill[0], k.HasKill[1], k.HasFastKill[0], k.HasFastKill[1], k.IsTrapped, k.ItemUsed} {
		bits <<= 1
		if flag {
			bits |= 1
		}
	}
	return bits<<8 | k.OutOfPP&0xff
}

type doubleBattleStateKey struct {
	PlayerSlots   [2]doubleSlotKey
	OpponentSlots [2]doubleSlotKey
	AlivePlayer   uint64
	AliveOpponent uint64
	Weather       weatherState
}

func (dbs *doubleBattleState) key() string {
	stateKey := doubleBattleStateKey{
		Weather:       dbs.weather,
		AlivePlayer:   aliveMask(dbs.player.pokemonParty),
		AliveOpponent: aliveMask(dbs.opponent.pokemonParty),
	}

	playerSlots := [2]*slot{dbs.activePlayerSlotOne, dbs.activePlayerSlotTwo}
	opponentSlots := [2]*slot{dbs.activeOpponentSlotOne, dbs.activeOpponentSlotTwo}
	for i := range 2 {
		stateKey.PlayerSlots[i] = dbs.slotKey(playerSlots[i], opponentSlots, true)
		stateKey.OpponentSlots[i] = dbs.slotKey(opponentSlots[i], playerSlots, false)
	}

	return stateKey.String()
}

// String renders the key as "mon.flags.ailment" per slot followed by the
// alive masks and weather, all numbers in base 36.
func (stateKey doubleBattleStateKey) String() string {
	buf := make([]byte, 0, 96)
	for _, slots := range [2][2]doubleSlotKey{stateKey.PlayerSlots, stateKey.OpponentSlots} {
		for _, k := range slots {
			buf = append(buf, k.Mon...)
			buf = append(buf, '.')
			buf = strconv.AppendUint(buf, k.bits(), 36)
			buf = append(buf, '.')
			buf = strconv.AppendInt(buf, int64(k.Ailment), 36)
			buf = append(buf, '|')
		}
	}
	buf = strconv.AppendUint(buf, stateKey.AlivePlayer, 36)
	buf = append(buf, '.')
	buf = strconv.AppendUint(buf, stateKey.AliveOpponent, 36)
	buf = append(buf, '.')
	buf = strconv.AppendInt(buf, int64(stateKey.Weather), 36)
	return string(buf)
}

func aliveMask(party []*pokemon) uint64 {
	var mask uint64
	for i, mon := range party {
		if !mon.fainted {
			mask |= 1 << i
		}
	}
	return mask
}

func (dbs *doubleBattleState) slotKey(user *slot, targets [2]*slot, isPlayer bool) doubleSlotKey {
	key := doubleSlotKey{
		Mon:       user.mon.base.Name,
		Fainted:   user.mon.fainted,
		IsTrapped: user.isTrapped(),
		Ailment:   addNonVolatileAilmentState(user.mon),
		ItemUsed:  isItemUsed(user.mon),
		OutOfPP:   outOfPPMask(user.mon),
	}
	if user.mon.fainted {
		return key
	}

	// The player is assumed to roll low and the opponent to roll high and crit.
	var rollType damageRollType = rollMin
	if !isPlayer {
		rollType = rollMax
	}
	for i, target := range targets {
		if target.mon.fainted {
			continue
		}
		maxDamage := calculateMaxDamageAmongMoves(dbs, user.mon, target.mon, calculateMaxDamageFlags{
			checkChoiceItem: true,
			isCrit:          !isPlayer,
			isFirstTurn:     user.firstTurn,
			rollType:        rollType,
		})
		maxPriorityDamage := calculateMaxDamageAmongMoves(dbs, user.mon, target.mon, calculateMaxDamageFlags{
			checkChoiceItem:   true,
			checkOnlyPriority: true,
			isCrit:            !isPlayer,
			isFirstTurn:       user.firstTurn,
			rollType:          rollType,
		})
		key.HasKill[i] = maxDamage >= target.mon.hp
		key.HasFastKill[i] = maxPriorityDamage >= target.mon.hp || (key.HasKill[i] && user.mon.isFasterThan(dbs, target.mon))
	}
	return key
}
