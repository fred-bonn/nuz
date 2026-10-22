package nuzengine

import (
	"encoding/json"
	"strings"
)

// doubleSlotKey describes one active slot. The kill flags are indexed by the
// opponent's slot order (one, two) and are false for fainted targets.
type doubleSlotKey struct {
	Mon         string       `json:"mon"`
	Fainted     bool         `json:"fainted"`
	HasKill     [2]bool      `json:"has_kill"`
	HasFastKill [2]bool      `json:"has_fast_kill"`
	IsTrapped   bool         `json:"is_trapped"`
	Ailment     ailmentState `json:"ailment"`
}

type doubleBattleStateKey struct {
	PlayerSlots       [2]doubleSlotKey `json:"player_slots"`
	OpponentSlots     [2]doubleSlotKey `json:"opponent_slots"`
	AlivePlayerMons   []bool           `json:"alive_player_mons"`
	AliveOpponentMons []bool           `json:"alive_opponent_mons"`
	MovesOutOfPP      map[string]bool  `json:"moves_out_of_pp"`
	ItemsConsumed     map[string]bool  `json:"items_consumed"`
	Weather           weatherState     `json:"weather"`
}

func (dbs *doubleBattleState) key() string {
	stateKey := doubleBattleStateKey{
		Weather:           dbs.weather,
		AlivePlayerMons:   make([]bool, len(dbs.player.pokemonParty)),
		AliveOpponentMons: make([]bool, len(dbs.opponent.pokemonParty)),
		MovesOutOfPP:      make(map[string]bool),
		ItemsConsumed:     make(map[string]bool),
	}

	playerSlots := [2]*slot{dbs.activePlayerSlotOne, dbs.activePlayerSlotTwo}
	opponentSlots := [2]*slot{dbs.activeOpponentSlotOne, dbs.activeOpponentSlotTwo}

	for i := range 2 {
		stateKey.PlayerSlots[i] = dbs.slotKey(playerSlots[i], opponentSlots, true)
		stateKey.OpponentSlots[i] = dbs.slotKey(opponentSlots[i], playerSlots, false)
		stateKey.addMoveAndItemState(playerSlots[i].mon)
		stateKey.addMoveAndItemState(opponentSlots[i].mon)
	}

	for i, mon := range dbs.player.pokemonParty {
		stateKey.AlivePlayerMons[i] = !mon.fainted
	}
	for i, mon := range dbs.opponent.pokemonParty {
		stateKey.AliveOpponentMons[i] = !mon.fainted
	}

	bytes, err := json.Marshal(stateKey)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (dbs *doubleBattleState) slotKey(user *slot, targets [2]*slot, isPlayer bool) doubleSlotKey {
	key := doubleSlotKey{
		Mon:       user.mon.base.Name,
		Fainted:   user.mon.fainted,
		IsTrapped: user.isTrapped(),
		Ailment:   addNonVolatileAilmentState(user.mon),
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

func (stateKey *doubleBattleStateKey) addMoveAndItemState(mon *pokemon) {
	monName := strings.ToLower(mon.base.Name)
	for _, move := range mon.moves {
		stateKey.MovesOutOfPP[monName+": "+move.Move] = move.PP <= 0
	}
	itemName := "#"
	consumed := true
	if mon.item != nil && mon.item.State != noneItem {
		itemName = mon.item.State.String()
		consumed = mon.item.Consumed
	}
	stateKey.ItemsConsumed[monName+": "+itemName] = consumed
}
