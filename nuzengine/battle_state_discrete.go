package nuzengine

type discreteBattleState interface {
	update(battleState) bool
}
