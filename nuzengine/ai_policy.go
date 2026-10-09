package nuzengine

import (
	"compress/gzip"
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
	PlayerParty   string
	OpponentParty string
	Weather       int
	Policy        qMap
}

// diskPolicy is the on-disk form; action keys are interned into Actions.
type diskPolicy struct {
	PlayerParty   string      `json:"player_party"`
	OpponentParty string      `json:"opponent_party"`
	Weather       int         `json:"weather"`
	Actions       []string    `json:"actions"`
	States        []diskState `json:"states"`
}

type diskState struct {
	Key     string      `json:"k"`
	Entries []diskEntry `json:"e"`
}

type diskEntry struct {
	Action int     `json:"a"`
	Value  float64 `json:"v"`
	Count  int     `json:"c"`
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
	disk := diskPolicy{
		PlayerParty:   playerPartyStr,
		OpponentParty: opponentPartyStr,
		Weather:       weatherInt,
		States:        make([]diskState, 0, len(q)),
	}

	actionIndex := make(map[string]int)
	for state, actions := range q {
		ds := diskState{Key: state, Entries: make([]diskEntry, 0, len(actions))}
		for action, entry := range actions {
			idx, ok := actionIndex[action]
			if !ok {
				idx = len(disk.Actions)
				actionIndex[action] = idx
				disk.Actions = append(disk.Actions, action)
			}
			ds.Entries = append(ds.Entries, diskEntry{Action: idx, Value: entry.Value, Count: entry.Count})
		}
		disk.States = append(disk.States, ds)
	}

	file, err := os.Create(policyFilePath)
	if err != nil {
		return fmt.Errorf("failed creating policy file: %w", err)
	}
	gz := gzip.NewWriter(file)
	if err := json.NewEncoder(gz).Encode(disk); err != nil {
		file.Close()
		return fmt.Errorf("failed writing policy: %w", err)
	}
	if err := gz.Close(); err != nil {
		file.Close()
		return fmt.Errorf("failed writing policy: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed writing policy file: %w", err)
	}
	return nil
}

func loadPolicy(path string) (policyData, error) {
	file, err := os.Open(path)
	if err != nil {
		return policyData{}, fmt.Errorf("failed reading policy file: %w", err)
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return policyData{}, fmt.Errorf("failed reading policy file: %w", err)
	}
	defer gz.Close()

	var disk diskPolicy
	if err := json.NewDecoder(gz).Decode(&disk); err != nil {
		return policyData{}, fmt.Errorf("failed parsing policy file: %w", err)
	}

	q := make(qMap, len(disk.States))
	for _, ds := range disk.States {
		actions := make(map[string]*qEntry, len(ds.Entries))
		for _, e := range ds.Entries {
			if e.Action < 0 || e.Action >= len(disk.Actions) {
				return policyData{}, fmt.Errorf("failed parsing policy file: invalid action index %d", e.Action)
			}
			actions[disk.Actions[e.Action]] = &qEntry{Value: e.Value, Count: e.Count}
		}
		q[ds.Key] = actions
	}

	return policyData{
		PlayerParty:   disk.PlayerParty,
		OpponentParty: disk.OpponentParty,
		Weather:       disk.Weather,
		Policy:        q,
	}, nil
}
