package nuzengine

import (
	"math/rand"
	"slices"
	"strings"
	"sync"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

const (
	monteCarloEpisodes         = 100000
	monteCarloWorkers          = 10
	monteCarloEpsilon          = 0.3
	monteCarloPolicyIterations = 10

	winReward         = 15
	faintedMonPenalty = 5
)

// Learn trains a Monte Carlo control policy by running monteCarloWorkers battle states
// concurrently, each with its own Q-map, then merges the maps and evaluates the resulting
// greedy policy for monteCarloPolicyIterations battles.
func Learn(playerPartyStr, opponentPartyStr string, weatherInt int) {
	cfg := &config{
		client: pokeapi.NewClient(),
	}

	episodesPerWorker := monteCarloEpisodes / monteCarloWorkers

	results := make(chan *workerResult, monteCarloWorkers)
	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			result := runWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker)
			if result != nil {
				results <- result
			}
		})
	}
	wg.Wait()
	close(results)

	var combinedStats *battleStatistics
	var workerMaps []qMap
	for result := range results {
		if combinedStats == nil {
			combinedStats = result.stats
		} else {
			combinedStats.battleCount += result.stats.battleCount
			combinedStats.winCount += result.stats.winCount
			for i, survived := range result.stats.monSurvivalCount {
				combinedStats.monSurvivalCount[i] += survived
			}
		}
		workerMaps = append(workerMaps, result.q)
	}
	if combinedStats != nil {
		combinedStats.print()
	}

	runPolicy(cfg, mergeQMaps(workerMaps), playerPartyStr, opponentPartyStr, weatherInt)
}

// workerResult is one worker's outcome: its battle statistics and its own, unshared Q-map.
type workerResult struct {
	stats *battleStatistics
	q     qMap
}

// runWorker runs one goroutine's battle state for episodeCount episodes, training its own
// Q-map; no locking is needed since each worker owns an independent map.
func runWorker(cfg *config, playerPartyStr, opponentPartyStr string, weatherInt, episodeCount int) *workerResult {
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

	statistics := newBattleStatistics(bs)

	for episode := range episodeCount {
		if episode > 0 {
			bs.reset()
		}

		la.beginEpisode()
		if err := bs.execute(); err != nil {
			elogf("error: failed executing battle state: %s", err)
			return &workerResult{stats: statistics, q: la.q}
		}
		statistics.record()
		la.endEpisode(monteCarloReward(bs))
	}

	return &workerResult{stats: statistics, q: la.q}
}

// runPolicy evaluates the merged, greedy policy for monteCarloPolicyIterations battles.
func runPolicy(cfg *config, q qMap, playerPartyStr, opponentPartyStr string, weatherInt int) {
	playerParty, err := cfg.validateInput(playerPartyStr)
	if err != nil {
		elogf("error: failed validating player party: %s", err)
		return
	}

	opponentParty, err := cfg.validateInput(opponentPartyStr)
	if err != nil {
		elogf("error: failed validating opponent party: %s", err)
		return
	}

	bs := initSingleBattleState(
		trainer{ai: newPolicyAI(q), player: true, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		playerParty,
		opponentParty,
		weatherState(weatherInt),
	)

	if err := Execute(bs, monteCarloPolicyIterations); err != nil {
		elogf("error: failed executing policy battle state: %s", err)
	}
}

// monteCarloReward is the terminal reward applied to every (state, action) pair visited in an episode.
func monteCarloReward(bs battleState) float64 {
	trainer := bs.getPlayerTrainer()

	reward := 0.0
	if !trainer.lost {
		reward += winReward
	}
	for _, mon := range trainer.pokemonParty {
		if mon.fainted {
			reward -= faintedMonPenalty
		}
	}
	return reward
}

// qEntry is the running sample average of returns observed for a (state, action) pair.
type qEntry struct {
	value float64
	count int
}

type trajectoryStep struct {
	state  string
	action string
}

// qMap is a state -> action -> running-average Q-value table. Each worker trains its own,
// unshared qMap, so no locking is required; maps are merged only after training completes.
type qMap map[string]map[string]*qEntry

func newQMap() qMap {
	return make(qMap)
}

// update applies one first-visit Monte Carlo sample-average update.
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
	entry.count++
	entry.value += (reward - entry.value) / float64(entry.count)
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
	return entry.value
}

// mergeQMaps combines independently-trained worker Q-maps into one, weighting each
// (state, action) value by how many samples contributed to it.
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
					dst[action] = &qEntry{value: entry.value, count: entry.count}
					continue
				}
				totalCount := existing.count + entry.count
				existing.value = (existing.value*float64(existing.count) + entry.value*float64(entry.count)) / float64(totalCount)
				existing.count = totalCount
			}
		}
	}
	return merged
}

// learningAi trains its own, unshared Q-map via epsilon-greedy Monte Carlo control.
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

// beginEpisode clears the recorded trajectory ahead of a new battle.
func (la *learningAi) beginEpisode() {
	la.trajectory = la.trajectory[:0]
}

// endEpisode applies a first-visit Monte Carlo update using the given terminal reward.
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

// candidate is one of the actions available to the active slot: either a move or a switch target.
type candidate struct {
	key    string
	move   *moveAction
	target *pokemon
}

// buildCandidates lists every legal move plus, when the slot may voluntarily switch, every
// alive, non-active, not-already-queued party member as a switch candidate.
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

// buildSwitchCandidates lists only the mons actually offered for a (possibly forced) switch-in.
func buildSwitchCandidates(mons []*pokemon) []candidate {
	candidates := make([]candidate, 0, len(mons))
	for _, mon := range mons {
		candidates = append(candidates, candidate{key: "switch:" + strings.ToLower(mon.base.Name), target: mon})
	}
	return candidates
}

// argmaxCandidate picks the candidate with the highest known Q-value, breaking ties uniformly at random.
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

// choose picks a candidate using epsilon-greedy selection over the current Q-map.
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
	// the voluntary-switch decision from evaluateActions is still valid if its target is still available
	if la.pending.target != nil && slices.Contains(mons, la.pending.target) {
		return la.pending.target
	}

	// forced replacement (a mon fainted): choose independently over the mons actually available now
	state := bs.key()
	chosen := la.choose(state, buildSwitchCandidates(mons))
	la.trajectory = append(la.trajectory, trajectoryStep{state: state, action: chosen.key})
	return chosen.target
}

// policyAi always exploits a fixed, already-trained Q-map: no exploration, no learning.
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
