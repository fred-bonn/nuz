package nuzengine

import "fmt"

type doubleBattleState struct {
	activePlayerSlotOne   *slot
	activePlayerSlotTwo   *slot
	activeOpponentSlotOne *slot
	activeOpponentSlotTwo *slot
	player                *trainer
	opponent              *trainer
	actions               actionQueue
	weather               weatherState
	fieldEffects          map[fieldEffect]int
	err                   error
	resetPoint            battleResetState
}

func (dbs *doubleBattleState) execute() error {
	vprintln("\nStarting battle...")

	for turn := 0; !dbs.player.lost && !dbs.opponent.lost; turn++ {
		vprintln("=====")
		vprintf("Turn %d:", turn+1)
		vprintf("%s %d/%d - %s %d/%d", dbs.activePlayerSlotOne.mon.base.Name, dbs.activePlayerSlotOne.mon.hp, dbs.activePlayerSlotOne.mon.MaxHP(), dbs.activeOpponentSlotOne.mon.base.Name, dbs.activeOpponentSlotOne.mon.hp, dbs.activeOpponentSlotOne.mon.MaxHP())
		vprintf("%s %d/%d - %s %d/%d", dbs.activePlayerSlotTwo.mon.base.Name, dbs.activePlayerSlotTwo.mon.hp, dbs.activePlayerSlotTwo.mon.MaxHP(), dbs.activeOpponentSlotTwo.mon.base.Name, dbs.activeOpponentSlotTwo.mon.hp, dbs.activeOpponentSlotTwo.mon.MaxHP())

		dbs.gatherActions()
		if dbs.err != nil {
			dbs.actions.queue = dbs.actions.queue[:0]
			return dbs.err
		}
		dbs.actions.sort(dbs)
		if err := executeActions(dbs); err != nil {
			return err
		}
		resolveEndOfTurn(dbs)
		// if the end of turn causes mons to faint, empty the queue for replace actions
		if err := executeActions(dbs); err != nil {
			return err
		}
	}
	vprintln("=====")
	vprintln("Ending battle...")

	return nil
}
func (dbs *doubleBattleState) reset() { resetBattle(dbs) }

func (dbs *doubleBattleState) getResetPoint() *battleResetState { return &dbs.resetPoint }

func (dbs *doubleBattleState) setError(err error) {
	dbs.err = err
}

func (dbs *doubleBattleState) getError() error {
	return dbs.err
}

func (dbs *doubleBattleState) gatherActions() {
	// slot one is chosen first so slot two can exclude mons slot one is switching to
	choose := func(s *slot, t *trainer) {
		if s.mon.fainted || dbs.err != nil {
			return
		}
		dbs.actions.queue.push(chooseNextAction(dbs, s, t.pokemonParty, t.ai))
	}
	choose(dbs.activePlayerSlotOne, dbs.player)
	choose(dbs.activePlayerSlotTwo, dbs.player)
	choose(dbs.activeOpponentSlotOne, dbs.opponent)
	choose(dbs.activeOpponentSlotTwo, dbs.opponent)
}

func (dbs *doubleBattleState) getAllSlots() []*slot {
	return []*slot{
		dbs.activePlayerSlotOne,
		dbs.activePlayerSlotTwo,
		dbs.activeOpponentSlotOne,
		dbs.activeOpponentSlotTwo,
	}
}

func (dbs *doubleBattleState) getOtherSlots(s *slot) []*slot {
	var otherSlots []*slot
	for _, slot := range dbs.getAllSlots() {
		if slot != s {
			otherSlots = append(otherSlots, slot)
		}
	}
	return otherSlots
}

func (dbs *doubleBattleState) getOpponentSlot(s *slot) *slot {
	switch s {
	case dbs.activePlayerSlotOne:
		return dbs.activeOpponentSlotOne
	case dbs.activePlayerSlotTwo:
		return dbs.activeOpponentSlotTwo
	case dbs.activeOpponentSlotOne:
		return dbs.activePlayerSlotOne
	case dbs.activeOpponentSlotTwo:
		return dbs.activePlayerSlotTwo
	default:
		return nil
	}
}

func (dbs *doubleBattleState) getPlayerTrainer() *trainer {
	return dbs.player
}

func (dbs *doubleBattleState) getPokemonSlot(p *pokemon) *slot {
	for _, slot := range dbs.getAllSlots() {
		if slot.mon == p {
			return slot
		}
	}
	return nil
}

func (dbs *doubleBattleState) getActions() *actionQueue {
	return &dbs.actions
}

func (dbs *doubleBattleState) getWeather() weatherState {
	return dbs.weather
}

func (dbs *doubleBattleState) setWeather(weather weatherState) {
	dbs.weather = weather
	dbs.actions.makeDirty()
	weather.onset()
}

func (dbs *doubleBattleState) getFieldEffects() map[fieldEffect]int {
	return dbs.fieldEffects
}

func (dbs *doubleBattleState) slotLabel(s *slot) string {
	switch s {
	case dbs.activePlayerSlotOne:
		return "p1"
	case dbs.activePlayerSlotTwo:
		return "p2"
	case dbs.activeOpponentSlotOne:
		return "o1"
	case dbs.activeOpponentSlotTwo:
		return "o2"
	default:
		return ""
	}
}

var _ battleState = (*doubleBattleState)(nil)

func initDoubleBattleState(player, opponent trainer, playerParty, opponentParty []*pokemon, weather weatherState) (battleState, error) {
	if len(playerParty) < 2 {
		return nil, fmt.Errorf("player party must contain at least 2 Pokemon")
	}
	if len(opponentParty) < 2 {
		return nil, fmt.Errorf("opponent party must contain at least 2 Pokemon")
	}

	player.pokemonParty = playerParty
	opponent.pokemonParty = opponentParty

	res := doubleBattleState{
		activePlayerSlotOne: &slot{
			mon:       playerParty[0],
			Trainer:   &player,
			firstTurn: true,
		},
		activePlayerSlotTwo: &slot{
			mon:       playerParty[1],
			Trainer:   &player,
			firstTurn: true,
		},
		activeOpponentSlotOne: &slot{
			mon:       opponentParty[0],
			Trainer:   &opponent,
			firstTurn: true,
		},
		activeOpponentSlotTwo: &slot{
			mon:       opponentParty[1],
			Trainer:   &opponent,
			firstTurn: true,
		},
		player:       &player,
		opponent:     &opponent,
		weather:      weather,
		fieldEffects: make(map[fieldEffect]int),
		actions: actionQueue{
			queue: make(priorityQueue[action], 0, 3),
		},
	}

	res.setWeather(weather)
	setResetPoint(&res, &res, false)

	resolveOnEntry(&res)

	return &res, nil
}
