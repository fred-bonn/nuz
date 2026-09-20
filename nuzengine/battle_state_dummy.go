package nuzengine

type dummyAction struct {
	priority int
	spd      int
}

func (ba *dummyAction) invoke(bs battleState)    {}
func (ba *dummyAction) prio(bs battleState) int  { return ba.priority }
func (ba *dummyAction) speed(bs battleState) int { return ba.spd }

type dummyBattleState struct {
	actions *actionQueue
	weather weatherState
	slots   []*slot
}

func (bs *dummyBattleState) execute() error                       { return nil }
func (bs *dummyBattleState) setError(error)                       {}
func (bs *dummyBattleState) gatherActions()                       {}
func (bs *dummyBattleState) getAllSlots() []*slot                 { return bs.slots }
func (bs *dummyBattleState) getOtherSlots(s *slot) []*slot        { return nil }
func (bs *dummyBattleState) getOpponentSlot(s *slot) *slot        { return nil }
func (bs *dummyBattleState) getPlayerTrainer() *trainer           { return nil }
func (bs *dummyBattleState) getPokemonSlot(p *pokemon) *slot      { return nil }
func (bs *dummyBattleState) getActions() *actionQueue             { return bs.actions }
func (bs *dummyBattleState) getWeather() weatherState             { return bs.weather }
func (bs *dummyBattleState) setWeather(weatherState)              {}
func (bs *dummyBattleState) getFieldEffects() map[fieldEffect]int { return nil }
func (bs *dummyBattleState) reset()                               {}
func (bs *dummyBattleState) key() string                          { return "" }

func initBenchBattleState(w weatherState) *dummyBattleState {
	return &dummyBattleState{
		weather: w,
	}
}
