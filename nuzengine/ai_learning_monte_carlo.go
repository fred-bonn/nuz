package nuzengine

import (
	"math/rand"
	"slices"
	"strings"
	"sync"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type learningAiMonteCarlo struct {
	epsilon    float64
	q          qMap
	trajectory []trajectoryStep
	pending    actionCandidate
}

type qMap map[string]map[string]*qEntry

type qEntry struct {
	Value float64 `json:"value"`
	Count int     `json:"count"`
}

type actionCandidate struct {
	key    string
	move   *moveAction
	target *pokemon
}

type trajectoryStep struct {
	state  string
	action string
}

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
	la := newLearningAIMonteCarlo()

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

func newLearningAIMonteCarlo() *learningAiMonteCarlo {
	return &learningAiMonteCarlo{
		epsilon: monteCarloEpsilon,
		q:       newQMap(),
	}
}

func (la *learningAiMonteCarlo) beginEpisode() {
	la.trajectory = la.trajectory[:0]
}

func (la *learningAiMonteCarlo) endEpisode(reward float64) {
	visited := make(map[trajectoryStep]bool, len(la.trajectory))
	for _, step := range la.trajectory {
		if visited[step] {
			continue
		}
		visited[step] = true

		la.q.update(step.state, step.action, reward)
	}
}

func buildCandidates(bs battleState, slot *slot, actions []*moveAction) []actionCandidate {
	candidates := make([]actionCandidate, 0, len(actions))
	for _, a := range actions {
		candidates = append(candidates, actionCandidate{key: "move:" + strings.ToLower(a.move.Move), move: a})
	}
	if canReplace(slot.Trainer.pokemonParty) && !slot.isTrapped() {
		var possibleMons []*pokemon
		for _, mon := range slot.Trainer.pokemonParty {
			if mon != slot.mon && !mon.fainted {
				possibleMons = append(possibleMons, mon)
			}
		}
		candidates = append(candidates, buildSwitchCandidates(possibleMons)...)
	}

	return candidates
}

func buildSwitchCandidates(mons []*pokemon) []actionCandidate {
	candidates := make([]actionCandidate, 0, len(mons))
	for _, mon := range mons {
		candidates = append(candidates, actionCandidate{key: "switch:" + strings.ToLower(mon.base.Name), target: mon})
	}
	return candidates
}

func argmaxCandidate(q qMap, state string, candidates []actionCandidate) actionCandidate {
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

func (la *learningAiMonteCarlo) choose(state string, candidates []actionCandidate) actionCandidate {
	if rand.Float64() < la.epsilon {
		return candidates[rand.Intn(len(candidates))]
	}
	return argmaxCandidate(la.q, state, candidates)
}

func (la *learningAiMonteCarlo) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	state := bs.key()
	chosen := la.choose(state, buildCandidates(bs, slot, actions))

	la.trajectory = append(la.trajectory, trajectoryStep{state: state, action: chosen.key})
	la.pending = chosen

	if chosen.move != nil {
		return chosen.move, 0
	}
	return actions[0], -1
}

func (la *learningAiMonteCarlo) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if la.pending.target != nil && slices.Contains(mons, la.pending.target) {
		return la.pending.target
	}

	state := bs.key()
	chosen := la.choose(state, buildSwitchCandidates(mons))
	la.trajectory = append(la.trajectory, trajectoryStep{state: state, action: chosen.key})
	return chosen.target
}

func (la *learningAiMonteCarlo) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return la.pending.move == nil
}
