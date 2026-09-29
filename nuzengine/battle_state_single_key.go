package nuzengine

import (
	"encoding/json"
	"strings"
)

type singleBattleStateKey struct {
	PlayerMon              string          `json:"player_mon"`
	PlayerMonHasKill       bool            `json:"player_mon_has_kill"`
	PlayerMonHasFastKill   bool            `json:"player_mon_has_fast_kill"`
	PlayerMonIsTrapped     bool            `json:"player_mon_is_trapped"`
	AlivePlayerMons        []bool          `json:"alive_player_mons"`
	PlayerAilment          ailmentState    `json:"player_ailment"`
	OpponentMon            string          `json:"opponent_mon"`
	OpponentMonHasKill     bool            `json:"opponent_mon_has_kill"`
	OpponentMonHasFastKill bool            `json:"opponent_mon_has_fast_kill"`
	AliveOpponentMons      []bool          `json:"alive_opponent_mons"`
	OpponentAilment        ailmentState    `json:"opponent_ailment"`
	MovesOutOfPP           map[string]bool `json:"moves_out_of_pp"`
	ItemsConsumed          map[string]bool `json:"items_consumed"`
	Weather                weatherState    `json:"weather"`
}

func (sbs *singleBattleState) key() string {
	stateKey := singleBattleStateKey{
		Weather:           sbs.weather,
		AlivePlayerMons:   make([]bool, len(sbs.player.pokemonParty)),
		AliveOpponentMons: make([]bool, len(sbs.opponent.pokemonParty)),
		MovesOutOfPP:      make(map[string]bool),
		ItemsConsumed:     make(map[string]bool),
	}

	playerMon := sbs.activePlayerSlot.mon
	opponentMon := sbs.activeOpponentSlot.mon

	stateKey.PlayerMon = playerMon.base.Name
	stateKey.OpponentMon = opponentMon.base.Name
	stateKey.PlayerAilment = addNonVolatileAilmentState(playerMon)
	stateKey.OpponentAilment = addNonVolatileAilmentState(opponentMon)
	stateKey.addMoveAndItemState(playerMon)
	stateKey.addMoveAndItemState(opponentMon)

	for i, mon := range sbs.player.pokemonParty {
		stateKey.AlivePlayerMons[i] = !mon.fainted
	}
	for i, mon := range sbs.opponent.pokemonParty {
		stateKey.AliveOpponentMons[i] = !mon.fainted
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

	bytes, err := json.Marshal(stateKey)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (stateKey *singleBattleStateKey) addMoveAndItemState(mon *pokemon) {
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

func addNonVolatileAilmentState(mon *pokemon) ailmentState {
	for ailment := range mon.ailments {
		if ailment.isNonVolatileStatus() {
			return ailment
		}
	}
	return noneAilment
}
