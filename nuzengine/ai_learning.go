package nuzengine

type learningAi struct {
}

const ()

func newLearningAI() *randomAi {
	return &randomAi{}
}

func loadPolicyFromDisk(path string) (any, error) {
	return nil, nil
}

func (la *learningAi) savePolicyToDisk() error {
	return nil
}

func (la *learningAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	return nil, 0
}

func (la *learningAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	return nil
}

func (la *learningAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return false
}
