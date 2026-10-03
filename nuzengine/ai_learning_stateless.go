package nuzengine

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"sync"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type learningAiStateless struct {
	epsilon       float64
	sMap          sMap
	bestKeys      []string
	bestIndex     map[string]int
	bestValue     float64
	pending       actionCandidate
	pendingSwitch bool
	sequence      []string
	replay        []string
	replayIndex   int
	replaying     bool
}

type sMap map[string]struct {
	count int
	value float64
}

const (
	statelessWorkers = 10
	statelessEpsilon = 0.1
)

func LearnStateless(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{client: pokeapi.NewClient()}

	episodesPerWorker := iterations / statelessWorkers

	results := make(chan sMap, statelessWorkers)
	var wg sync.WaitGroup
	for range monteCarloWorkers {
		wg.Go(func() {
			if s := runStatelessWorker(cfg, playerPartyStr, opponentPartyStr, weatherInt, episodesPerWorker); s != nil {
				results <- s
			}
		})
	}
	wg.Wait()
	close(results)

	merged := make(sMap)
	for s := range results {
		mergeSMaps(merged, s)
	}
	printStatelessResults(merged)
}

func runStatelessWorker(cfg *config, playerPartyStr, opponentPartyStr string, weatherInt, episodeCount int) sMap {
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

	la := newLearningAIStateless()
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
			return la.sMap
		}
		la.endEpisode(monteCarloReward(bs))
	}

	return la.sMap
}

func mergeSMaps(dst, src sMap) {
	for key, entry := range src {
		existing := dst[key]
		total := existing.count + entry.count
		existing.value = (existing.value*float64(existing.count) + entry.value*float64(entry.count)) / float64(total)
		existing.count = total
		dst[key] = existing
	}
}

func printStatelessResults(results sMap) {
	bestKey := ""
	bestValue := -100.0
	bestCount := 0
	for key, entry := range results {
		if bestKey == "" || entry.value > bestValue || (entry.value == bestValue && entry.count > bestCount) {
			bestKey = key
			bestValue, bestCount = entry.value, entry.count
		}
		if entry.value == 0.0 {
			fmt.Printf("%s\n - score: %.2f\n - runs: %d\n", key, entry.value, entry.count)
		}
	}
	fmt.Println("=== Best Sequence ===")
	fmt.Printf("%s\n - score: %.2f\n - runs: %d\n", bestKey, bestValue, bestCount)
}

func newLearningAIStateless() *learningAiStateless {
	return &learningAiStateless{
		epsilon:   statelessEpsilon,
		sMap:      make(sMap),
		bestIndex: make(map[string]int),
	}
}

func (la *learningAiStateless) beginEpisode() {
	la.sequence = la.sequence[:0]
	la.replay = nil
	la.replayIndex = 0
	la.replaying = false
	if rand.Float64() < la.epsilon && len(la.bestKeys) > 0 {
		la.replay = strings.Split(la.bestKeys[rand.Intn(len(la.bestKeys))], ";")
		la.replaying = len(la.replay) > 0
	}
	la.pendingSwitch = false
}

func (la *learningAiStateless) endEpisode(reward float64) {
	if len(la.sequence) == 0 {
		return
	}

	key := strings.Join(la.sequence, ";")
	entry := la.sMap[key]
	entry.count++
	entry.value += (reward - entry.value) / float64(entry.count)
	la.sMap[key] = entry
	la.updateBestKeys(key)
}

func (la *learningAiStateless) updateBestKeys(key string) {
	value := la.sMap[key].value
	_, isBest := la.bestIndex[key]
	switch {
	case len(la.bestKeys) == 0 || value > la.bestValue:
		la.resetBest(key, value)
	case value == la.bestValue:
		if !isBest {
			la.addBest(key)
		}
	case isBest:
		la.removeBest(key)
		if len(la.bestKeys) == 0 {
			la.recomputeBestKeys()
		}
	}
}

func (la *learningAiStateless) resetBest(key string, value float64) {
	la.bestKeys = la.bestKeys[:0]
	clear(la.bestIndex)
	la.bestValue = value
	la.addBest(key)
}

func (la *learningAiStateless) addBest(key string) {
	la.bestIndex[key] = len(la.bestKeys)
	la.bestKeys = append(la.bestKeys, key)
}

func (la *learningAiStateless) removeBest(key string) {
	i := la.bestIndex[key]
	last := len(la.bestKeys) - 1
	la.bestKeys[i] = la.bestKeys[last]
	la.bestIndex[la.bestKeys[i]] = i
	la.bestKeys = la.bestKeys[:last]
	delete(la.bestIndex, key)
}

func (la *learningAiStateless) recomputeBestKeys() {
	la.bestKeys = la.bestKeys[:0]
	clear(la.bestIndex)
	for key, entry := range la.sMap {
		switch {
		case len(la.bestKeys) == 0 || entry.value > la.bestValue:
			la.resetBest(key, entry.value)
		case entry.value == la.bestValue:
			la.addBest(key)
		}
	}
}

func (la *learningAiStateless) choose(candidates []actionCandidate) actionCandidate {
	if la.replaying {
		if la.replayIndex < len(la.replay) {
			expected := la.replay[la.replayIndex]
			for _, candidate := range candidates {
				if candidate.key == expected {
					la.replayIndex++
					la.sequence = append(la.sequence, candidate.key)
					return candidate
				}
			}
			available := make([]string, 0, len(candidates))
			for _, candidate := range candidates {
				available = append(available, candidate.key)
			}
		}
		la.replaying = false
	}

	chosen := candidates[rand.Intn(len(candidates))]
	la.sequence = append(la.sequence, chosen.key)
	return chosen
}

func (la *learningAiStateless) replayCandidate(slot *slot, actions []*moveAction) (actionCandidate, bool) {
	if !la.replaying {
		return actionCandidate{}, false
	}
	if la.replayIndex >= len(la.replay) {
		la.replaying = false
		return actionCandidate{}, false
	}

	expected := la.replay[la.replayIndex]
	if strings.HasPrefix(expected, "move:") {
		for _, action := range actions {
			candidate := actionCandidate{key: "move:" + strings.ToLower(action.move.Move), move: action}
			if candidate.key == expected {
				la.replayIndex++
				la.sequence = append(la.sequence, candidate.key)
				return candidate, true
			}
		}
	}
	if strings.HasPrefix(expected, "switch:") && canReplace(slot.Trainer.pokemonParty) && !slot.isTrapped() {
		for _, mon := range slot.Trainer.pokemonParty {
			if mon == slot.mon || mon.fainted {
				continue
			}
			candidate := actionCandidate{key: "switch:" + strings.ToLower(mon.base.Name), target: mon}
			if candidate.key == expected {
				la.replayIndex++
				la.sequence = append(la.sequence, candidate.key)
				return candidate, true
			}
		}
	}
	return actionCandidate{}, false
}

func (la *learningAiStateless) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	chosen, replayed := la.replayCandidate(slot, actions)
	if !replayed {
		chosen = la.choose(buildCandidates(bs, slot, actions))
	}
	la.pending = chosen
	la.pendingSwitch = chosen.target != nil

	if chosen.move != nil {
		return chosen.move, 0
	}
	return actions[0], -1
}

func (la *learningAiStateless) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if la.pendingSwitch && la.pending.target != nil && slices.Contains(mons, la.pending.target) {
		la.pendingSwitch = false
		return la.pending.target
	}

	candidates := buildSwitchCandidates(mons)
	if la.pendingSwitch {
		la.sequence = la.sequence[:len(la.sequence)-1]
		la.pendingSwitch = false
		la.replaying = false
	} else {
		for i := range candidates {
			candidates[i].key = strings.Replace(candidates[i].key, "switch:", "replace:", 1)
		}
	}
	chosen := la.choose(candidates)
	return chosen.target
}

func (la *learningAiStateless) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return la.pending.move == nil
}
