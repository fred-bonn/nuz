package nuzengine

type doubleBattleState struct {
	activePlayerSlots   []*slot
	activeOpponentSlots []*slot
	player              *trainer
	opponent            *trainer
	actions             actionQueue
	weather             weatherState
	fieldEffects        map[fieldEffect]int
	err                 error
	initialPlayerMons   []*pokemon
	initialOpponentMons []*pokemon
	initialWeather      weatherState
}

func (dbs *doubleBattleState) execute() error                       { return nil }
func (dbs *doubleBattleState) reset()                               {}
func (dbs *doubleBattleState) setError(error)                       {}
func (dbs *doubleBattleState) gatherActions()                       {}
func (dbs *doubleBattleState) getAllSlots() []*slot                 { return nil }
func (dbs *doubleBattleState) getOtherSlots(*slot) []*slot          { return nil }
func (dbs *doubleBattleState) getOpponentSlot(*slot) *slot          { return nil }
func (dbs *doubleBattleState) getPlayerTrainer() *trainer           { return nil }
func (dbs *doubleBattleState) getPokemonSlot(*pokemon) *slot        { return nil }
func (dbs *doubleBattleState) getActions() *actionQueue             { return nil }
func (dbs *doubleBattleState) getWeather() weatherState             { return 0 }
func (dbs *doubleBattleState) setWeather(weatherState)              {}
func (dbs *doubleBattleState) getFieldEffects() map[fieldEffect]int { return nil }
func (dbs *doubleBattleState) key() string                          { return "" }

var _ battleState = (*doubleBattleState)(nil)
