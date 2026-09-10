package engine

type SingleBattleState struct {
	activePlayerSlot   *slot
	activeOpponentSlot *slot
	Player             *Trainer
	opponent           *Trainer
	actions            actionQueue
	weather            WeatherState
	fieldEffects       map[FieldEffect]int
	err                error
	initialPlayer      Trainer
	initialOpponent    Trainer
	initialWeather     WeatherState
	statistics         battleStatistics
}

func (sbs *SingleBattleState) Execute() error {
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

func (sbs *SingleBattleState) setError(err error) {
	sbs.err = err
}

func (sbs *SingleBattleState) gatherActions() {
	sbs.actions.queue.push(chooseNextAction(sbs, sbs.activePlayerSlot, sbs.Player.PokemonParty, sbs.Player.AI))
	sbs.actions.queue.push(chooseNextAction(sbs, sbs.activeOpponentSlot, sbs.opponent.PokemonParty, sbs.opponent.AI))
}

func (sbs *SingleBattleState) getAllSlots() []*slot {
	return []*slot{
		sbs.activePlayerSlot,
		sbs.activeOpponentSlot,
	}
}

func (sbs *SingleBattleState) getOtherSlots(s *slot) []*slot {
	if s == sbs.activePlayerSlot {
		return []*slot{sbs.activeOpponentSlot}
	}
	return []*slot{sbs.activePlayerSlot}
}

func (sbs *SingleBattleState) getOpponentSlot(s *slot) *slot {
	if s == sbs.activePlayerSlot {
		return sbs.activeOpponentSlot
	}
	return sbs.activePlayerSlot
}

func (sbs *SingleBattleState) getActions() *actionQueue {
	return &sbs.actions
}

func (sbs *SingleBattleState) getWeather() WeatherState {
	return sbs.weather
}

func (sbs *SingleBattleState) setWeather(w WeatherState) {
	sbs.weather = w
	w.onset()
}

func (sbs *SingleBattleState) getFieldEffects() map[FieldEffect]int {
	return sbs.fieldEffects
}

func (sbs *SingleBattleState) GetStatistics() *battleStatistics {
	return &sbs.statistics
}

func (sbs *SingleBattleState) RecordStatistics() {
	sbs.statistics.record(sbs.Player)
}

func (sbs *SingleBattleState) PrintStatistics() {
	sbs.statistics.print(sbs.initialPlayer.PokemonParty)
}

func (sbs *SingleBattleState) Reset() error {
	playerParty := ClonePokemonParty(sbs.initialPlayer.PokemonParty)
	opponentParty := ClonePokemonParty(sbs.initialOpponent.PokemonParty)
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

func InitSingleBattleState(player, opponent Trainer, playerParty, opponentParty []*Pokemon, weather WeatherState) *SingleBattleState {
	player.PokemonParty = playerParty
	opponent.PokemonParty = opponentParty

	res := SingleBattleState{
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

	res.initialPlayer.PokemonParty = ClonePokemonParty(playerParty)
	res.initialOpponent.PokemonParty = ClonePokemonParty(opponentParty)
	res.initialPlayer.FieldEffects = cloneFieldEffects(player.FieldEffects)
	res.initialOpponent.FieldEffects = cloneFieldEffects(opponent.FieldEffects)
	res.Player.PokemonParty = playerParty
	res.opponent.PokemonParty = opponentParty

	res.setWeather(weather)
	resolveOnEntry(&res)

	return &res
}
