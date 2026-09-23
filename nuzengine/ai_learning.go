package nuzengine

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

const (
	monteCarloWorkers = 10
	monteCarloEpsilon = 0.5

	faintedMonPenalty = 10

	policyFilePath = "data/policy.json"
)

func Learn(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	episodesPerWorker := iterations / monteCarloWorkers

	results := make(chan qMap, monteCarloWorkers)
	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			q := runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker)
			if q != nil {
				results <- q
			}
		})
	}
	wg.Wait()
	close(results)

	var workerMaps []qMap
	for q := range results {
		workerMaps = append(workerMaps, q)
	}

	if err := savePolicy(mergeQMaps(workerMaps), playerPartyStr, opponentPartyStr, weatherInt); err != nil {
		elogf("error: failed saving policy: %s", err)
	}
}

func runWorker(cfg *config, playerPartyStr, opponentPartyStr string, weatherInt, episodeCount int) qMap {
	playerParty, err := cfg.validateInput(playerPartyStr)
	if err != nil {
		elogf("error: failed validating player party: %s", err)
		return nil
	}
	opponentParty, err := cfg.validateInput(opponentPartyStr)
	if err != nil {
		elogf("error: failed validating opponent party: %s", err)
		return nil
	}
	la := newLearningAI()

	bs := initSingleBattleState(
		trainer{ai: la, player: true, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		playerParty,
		opponentParty,
		weatherState(weatherInt),
	)

	for episode := range episodeCount {
		if episode > 0 {
			bs.reset()
		}

		la.beginEpisode()
		if err := bs.execute(); err != nil {
			elogf("error: failed executing battle state: %s", err)
			return la.q
		}
		la.endEpisode(monteCarloReward(bs))
	}

	return la.q
}

type policyData struct {
	PlayerParty   string `json:"player_party"`
	OpponentParty string `json:"opponent_party"`
	Weather       int    `json:"weather"`
	Policy        qMap   `json:"policy"`
}

func monteCarloReward(bs battleState) float64 {
	trainer := bs.getPlayerTrainer()

	reward := 0.0
	for _, mon := range trainer.pokemonParty {
		if mon.fainted {
			reward -= faintedMonPenalty
		}
	}
	return reward
}

type qEntry struct {
	Value float64 `json:"value"`
	Count int     `json:"count"`
}

type trajectoryStep struct {
	state  string
	action string
}

type qMap map[string]map[string]*qEntry

func newQMap() qMap {
	return make(qMap)
}

func (q qMap) update(state, action string, reward float64) {
	actions, ok := q[state]
	if !ok {
		actions = make(map[string]*qEntry)
		q[state] = actions
	}
	entry, ok := actions[action]
	if !ok {
		entry = &qEntry{}
		actions[action] = entry
	}
	entry.Count++
	entry.Value += (reward - entry.Value) / float64(entry.Count)
}

func (q qMap) valueFor(state, action string) float64 {
	actions, ok := q[state]
	if !ok {
		return 0
	}
	entry, ok := actions[action]
	if !ok {
		return 0
	}
	return entry.Value
}

func mergeQMaps(maps []qMap) qMap {
	merged := newQMap()
	for _, m := range maps {
		for state, actions := range m {
			dst, ok := merged[state]
			if !ok {
				dst = make(map[string]*qEntry)
				merged[state] = dst
			}
			for action, entry := range actions {
				existing, ok := dst[action]
				if !ok {
					dst[action] = &qEntry{Value: entry.Value, Count: entry.Count}
					continue
				}
				totalCount := existing.Count + entry.Count
				existing.Value = (existing.Value*float64(existing.Count) + entry.Value*float64(entry.Count)) / float64(totalCount)
				existing.Count = totalCount
			}
		}
	}
	return merged
}

type learningAi struct {
	epsilon    float64
	q          qMap
	trajectory []trajectoryStep
	pending    candidate
}

func newLearningAI() *learningAi {
	return &learningAi{
		epsilon: monteCarloEpsilon,
		q:       newQMap(),
	}
}

func (la *learningAi) beginEpisode() {
	la.trajectory = la.trajectory[:0]
}

func (la *learningAi) endEpisode(reward float64) {
	visited := make(map[trajectoryStep]bool, len(la.trajectory))
	for _, step := range la.trajectory {
		if visited[step] {
			continue
		}
		visited[step] = true

		la.q.update(step.state, step.action, reward)
	}
}

type candidate struct {
	key    string
	move   *moveAction
	target *pokemon
}

func buildCandidates(bs battleState, slot *slot, actions []*moveAction) []candidate {
	candidates := make([]candidate, 0, len(actions))
	for _, a := range actions {
		candidates = append(candidates, candidate{key: "move:" + strings.ToLower(a.move.Move), move: a})
	}

	if canReplace(slot.Trainer.pokemonParty) && !slot.isTrapped() {
		for _, mon := range slot.Trainer.pokemonParty {
			if mon == slot.mon || mon.fainted || bs.getActions().containstSwitchTo(mon) {
				continue
			}
			candidates = append(candidates, candidate{key: "switch:" + strings.ToLower(mon.base.Name), target: mon})
		}
	}

	return candidates
}

func buildSwitchCandidates(mons []*pokemon) []candidate {
	candidates := make([]candidate, 0, len(mons))
	for _, mon := range mons {
		candidates = append(candidates, candidate{key: "switch:" + strings.ToLower(mon.base.Name), target: mon})
	}
	return candidates
}

func argmaxCandidate(q qMap, state string, candidates []candidate) candidate {
	best := candidates[0]
	bestValue := q.valueFor(state, best.key)
	bestCount := 1
	for _, c := range candidates[1:] {
		value := q.valueFor(state, c.key)
		if value > bestValue {
			best, bestValue, bestCount = c, value, 1
			continue
		}
		if value == bestValue {
			bestCount++
			if rand.Intn(bestCount) == 0 {
				best = c
			}
		}
	}
	return best
}

func (la *learningAi) choose(state string, candidates []candidate) candidate {
	if rand.Float64() < la.epsilon {
		return candidates[rand.Intn(len(candidates))]
	}
	return argmaxCandidate(la.q, state, candidates)
}

func (la *learningAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	state := bs.key()
	chosen := la.choose(state, buildCandidates(bs, slot, actions))

	la.trajectory = append(la.trajectory, trajectoryStep{state: state, action: chosen.key})
	la.pending = chosen

	if chosen.move != nil {
		return chosen.move, 0
	}
	return actions[0], -1
}

func (la *learningAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return la.pending.move == nil
}

func (la *learningAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if la.pending.target != nil && slices.Contains(mons, la.pending.target) {
		return la.pending.target
	}

	state := bs.key()
	chosen := la.choose(state, buildSwitchCandidates(mons))
	la.trajectory = append(la.trajectory, trajectoryStep{state: state, action: chosen.key})
	return chosen.target
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
