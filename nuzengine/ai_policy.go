package nuzengine

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

type policyAi struct {
	q       qMap
	pending actionCandidate
}

type policyData struct {
	PlayerParty   string `json:"player_party"`
	OpponentParty string `json:"opponent_party"`
	Weather       int    `json:"weather"`
	Policy        qMap   `json:"policy"`
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

func (pa *policyAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if pa.pending.target != nil && slices.Contains(mons, pa.pending.target) {
		return pa.pending.target
	}

	return argmaxCandidate(pa.q, bs.key(), buildSwitchCandidates(mons)).target
}

func (pa *policyAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return pa.pending.move == nil
}

func savePolicy(q qMap, playerPartyStr, opponentPartyStr string, weatherInt int) error {
	data := policyData{
		PlayerParty:   playerPartyStr,
		OpponentParty: opponentPartyStr,
		Weather:       weatherInt,
		Policy:        q,
	}

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed marshaling policy: %w", err)
	}

	if err := os.WriteFile(policyFilePath, bytes, 0644); err != nil {
		return fmt.Errorf("failed writing policy file: %w", err)
	}
	return nil
}

func loadPolicy(path string) (policyData, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return policyData{}, fmt.Errorf("failed reading policy file: %w", err)
	}

	var data policyData
	if err := json.Unmarshal(bytes, &data); err != nil {
		return policyData{}, fmt.Errorf("failed parsing policy file: %w", err)
	}
	return data, nil
}
