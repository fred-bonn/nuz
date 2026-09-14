package nuzengine

type singleBattleState struct {
	activePlayerSlot   *slot
	activeOpponentSlot *slot
	Player             *trainer
	opponent           *trainer
	actions            actionQueue
	weather            weatherState
	fieldEffects       map[fieldEffect]int
	err                error
	initialPlayer      trainer
	initialOpponent    trainer
	initialWeather     weatherState
	statistics         battleStatistics
}

func (sbs *singleBattleState) execute() error {
	vprintln("\nStarting battle...")

	for k := 0; !sbs.Player.lost && !sbs.opponent.lost; k++ {
		vprintln("=====")
		vprintf("Turn %d:", k+1)
		vprintf("%s %d/%d - %s %d/%d", sbs.activePlayerSlot.mon.Base.Name, sbs.activePlayerSlot.mon.HP, sbs.activePlayerSlot.mon.MaxHP(), sbs.activeOpponentSlot.mon.Base.Name, sbs.activeOpponentSlot.mon.HP, sbs.activeOpponentSlot.mon.MaxHP())

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
	sbs.actions.queue.push(chooseNextAction(sbs, sbs.activePlayerSlot, sbs.Player.PokemonParty, sbs.Player.AI))
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
	return sbs.Player
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

func (sbs *singleBattleState) getStatistics() *battleStatistics {
	return &sbs.statistics
}

func (sbs *singleBattleState) recordStatistics() {
	sbs.statistics.record(sbs.Player)
}

func (sbs *singleBattleState) printStatistics() {
	sbs.statistics.print(sbs.initialPlayer.PokemonParty)
}

func (sbs *singleBattleState) reset() error {
	playerParty := clonePokemonParty(sbs.initialPlayer.PokemonParty)
	opponentParty := clonePokemonParty(sbs.initialOpponent.PokemonParty)
	resetPokemonPartyPPs(playerParty)
	resetPokemonPartyPPs(opponentParty)

	player := sbs.initialPlayer
	player.PokemonParty = playerParty
	player.lost = false
	player.FieldEffects = cloneFieldEffects(sbs.initialPlayer.FieldEffects)

	opponent := sbs.initialOpponent
	opponent.PokemonParty = opponentParty
	opponent.lost = false
	opponent.FieldEffects = cloneFieldEffects(sbs.initialOpponent.FieldEffects)

	sbs.activePlayerSlot = &slot{
		mon:       playerParty[0],
		Trainer:   &player,
		firstTurn: true,
	}
	sbs.activeOpponentSlot = &slot{
		mon:       opponentParty[0],
		Trainer:   &opponent,
		firstTurn: true,
	}
	sbs.Player = &player
	sbs.opponent = &opponent
	sbs.actions = actionQueue{queue: make(priorityQueue[action], 0, 3)}
	sbs.weather = sbs.initialWeather
	sbs.err = nil

	sbs.setWeather(sbs.initialWeather)
	resolveOnEntry(sbs)
	return nil
}

func initSingleBattleState(player, opponent trainer, playerParty, opponentParty []*Pokemon, weather weatherState) *singleBattleState {
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
		Player:   &player,
		opponent: &opponent,
		actions: actionQueue{
			queue: make(priorityQueue[action], 0, 3),
		},
		initialPlayer:   player,
		initialOpponent: opponent,
		initialWeather:  weather,
		statistics:      newBattleStatistics(playerParty),
	}

	res.initialPlayer.PokemonParty = clonePokemonParty(playerParty)
	res.initialOpponent.PokemonParty = clonePokemonParty(opponentParty)
	res.initialPlayer.FieldEffects = cloneFieldEffects(player.FieldEffects)
	res.initialOpponent.FieldEffects = cloneFieldEffects(opponent.FieldEffects)
	res.Player.PokemonParty = playerParty
	res.opponent.PokemonParty = opponentParty

	res.setWeather(weather)
	resolveOnEntry(&res)

	return &res
}
