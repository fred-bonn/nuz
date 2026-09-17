package nuzengine

type singleBattleState struct {
	activePlayerSlot   *slot
	activeOpponentSlot *slot
	player             *trainer
	opponent           *trainer
	actions            actionQueue
	weather            weatherState
	fieldEffects       map[fieldEffect]int
	err                error
	initialPlayerMon   *pokemon
	initialOpponentMon *pokemon
	initialWeather     weatherState
}

func (sbs *singleBattleState) execute() error {
	vprintln("\nStarting battle...")

	for k := 0; !sbs.player.lost && !sbs.opponent.lost; k++ {
		vprintln("=====")
		vprintf("Turn %d:", k+1)
		vprintf("%s %d/%d - %s %d/%d", sbs.activePlayerSlot.mon.base.Name, sbs.activePlayerSlot.mon.hp, sbs.activePlayerSlot.mon.MaxHP(), sbs.activeOpponentSlot.mon.base.Name, sbs.activeOpponentSlot.mon.hp, sbs.activeOpponentSlot.mon.MaxHP())

		sbs.gatherActions()
		sbs.actions.sort(sbs)
		for len(sbs.actions.queue) > 0 {
			action, _ := sbs.actions.queue.pop()
			action.invoke(sbs)
			if sbs.err != nil {
				return sbs.err
			}
		}
		resolveEndOfTurn(sbs)
		// if the end of turn causes mons to faint, empty the queue for replace actions
		for len(sbs.actions.queue) > 0 {
			action, _ := sbs.actions.queue.pop()
			action.invoke(sbs)
			if sbs.err != nil {
				return sbs.err
			}
		}
	}
	vprintln("=====")
	vprintln("Ending battle...")

	return nil
}

func (sbs *singleBattleState) setError(err error) {
	sbs.err = err
}

func (sbs *singleBattleState) gatherActions() {
	sbs.actions.queue.push(chooseNextAction(sbs, sbs.activePlayerSlot, sbs.player.PokemonParty, sbs.player.AI))
	sbs.actions.queue.push(chooseNextAction(sbs, sbs.activeOpponentSlot, sbs.opponent.PokemonParty, sbs.opponent.AI))
}

func (sbs *singleBattleState) getAllSlots() []*slot {
	return []*slot{
		sbs.activePlayerSlot,
		sbs.activeOpponentSlot,
	}
}

func (sbs *singleBattleState) getOtherSlots(s *slot) []*slot {
	if s == sbs.activePlayerSlot {
		return []*slot{sbs.activeOpponentSlot}
	}
	return []*slot{sbs.activePlayerSlot}
}

func (sbs *singleBattleState) getOpponentSlot(s *slot) *slot {
	if s == sbs.activePlayerSlot {
		return sbs.activeOpponentSlot
	}
	return sbs.activePlayerSlot
}

func (sbs *singleBattleState) getPlayerTrainer() *trainer {
	return sbs.player
}

func (sbs *singleBattleState) getActions() *actionQueue {
	return &sbs.actions
}

func (sbs *singleBattleState) getWeather() weatherState {
	return sbs.weather
}

func (sbs *singleBattleState) setWeather(w weatherState) {
	sbs.weather = w
	w.onset()
}

func (sbs *singleBattleState) getFieldEffects() map[fieldEffect]int {
	return sbs.fieldEffects
}

func (sbs *singleBattleState) reset() error {
	sbs.activePlayerSlot.mon.switchReset()
	sbs.activePlayerSlot.mon = sbs.initialPlayerMon
	sbs.activeOpponentSlot.mon.switchReset()
	sbs.activeOpponentSlot.mon = sbs.initialOpponentMon

	resetPokemonParty(sbs.player.PokemonParty)
	resetPokemonParty(sbs.opponent.PokemonParty)

	sbs.player.FieldEffects = make(map[fieldEffect]int)
	sbs.opponent.FieldEffects = make(map[fieldEffect]int)

	sbs.player.lost = false
	sbs.opponent.lost = false

	sbs.weather = sbs.initialWeather

	sbs.err = nil

	resolveOnEntry(sbs)

	return nil
}

func initSingleBattleState(player, opponent trainer, playerParty, opponentParty []*pokemon, weather weatherState) *singleBattleState {
	player.PokemonParty = playerParty
	opponent.PokemonParty = opponentParty

	res := singleBattleState{
		activePlayerSlot: &slot{
			mon:       playerParty[0],
			Trainer:   &player,
			firstTurn: true,
		},
		activeOpponentSlot: &slot{
			mon:       opponentParty[0],
			Trainer:   &opponent,
			firstTurn: true,
		},
		player:   &player,
		opponent: &opponent,
		actions: actionQueue{
			queue: make(priorityQueue[action], 0, 3),
		},
		initialPlayerMon:   playerParty[0],
		initialOpponentMon: opponentParty[0],
		initialWeather:     weather,
	}

	res.setWeather(weather)

	resolveOnEntry(&res)

	return &res
}
