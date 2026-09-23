package nuzengine

import "slices"

type policyAi struct {
	q       qMap
	pending candidate
}

func newPolicyAI(q qMap) *policyAi {
	return &policyAi{q: q}
}

func (pa *policyAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	chosen := argmaxCandidate(pa.q, bs.key(), buildCandidates(bs, slot, actions))
	pa.pending = chosen

	if chosen.move != nil {
		return chosen.move, 0
	}
	return actions[0], -1
}

func (pa *policyAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return pa.pending.move == nil
}

func (pa *policyAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if pa.pending.target != nil && slices.Contains(mons, pa.pending.target) {
		return pa.pending.target
	}

	return argmaxCandidate(pa.q, bs.key(), buildSwitchCandidates(mons)).target
}
