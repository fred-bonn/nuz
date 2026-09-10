package engine

type dummyAction struct {
	priority int
	spd      int
}

func (ba *dummyAction) invoke(bs BattleState)    {}
func (ba *dummyAction) prio(bs BattleState) int  { return ba.priority }
func (ba *dummyAction) speed(bs BattleState) int { return ba.spd }

type dummyBattleState struct {
	actions *actionQueue
	weather WeatherState
	slots   []*slot
}

func initBenchBattleState(w WeatherState) *dummyBattleState {
	return &dummyBattleState{
		weather: w,
	}
}

func (bs *dummyBattleState) Execute() error                       { return nil }
func (bs *dummyBattleState) setError(error)                       {}
func (bs *dummyBattleState) gatherActions()                       {}
func (bs *dummyBattleState) getAllSlots() []*slot                 { return bs.slots }
func (bs *dummyBattleState) getOtherSlots(s *slot) []*slot        { return nil }
func (bs *dummyBattleState) getOpponentSlot(s *slot) *slot        { return nil }
func (bs *dummyBattleState) getActions() *actionQueue             { return bs.actions }
func (bs *dummyBattleState) getWeather() WeatherState             { return bs.weather }
func (bs *dummyBattleState) setWeather(WeatherState)              {}
func (bs *dummyBattleState) getFieldEffects() map[FieldEffect]int { return nil }
func (bs *dummyBattleState) Reset() error                         { return nil }
func (bs *dummyBattleState) GetStatistics() *battleStatistics     { return nil }
func (bs *dummyBattleState) RecordStatistics()                    {}
func (bs *dummyBattleState) PrintStatistics()                     {}
