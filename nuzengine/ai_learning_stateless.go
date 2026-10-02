package nuzengine

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type learningAiStateless struct {
	epsilon       float64
	sMap          sMap
	bestKey       string
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

const statelessEpsilon = 0.1

func LearnStateless(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
	Verbose = false
	cfg := &config{client: pokeapi.NewClient()}
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

	la := newLearningAIStateless()
	bs := initSingleBattleState(
		trainer{ai: la, player: true, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		playerParty,
		opponentParty,
		weatherState(weatherInt),
	)

	for episode := range iterations {
		if episode > 0 {
			bs.reset()
		}
		la.beginEpisode()
		if err := bs.execute(); err != nil {
			elogf("error: failed executing battle state: %s", err)
			return
		}
		la.endEpisode(monteCarloReward(bs))
	}

	bestKey := ""
	bestValue := -100.0
	bestCount := 0
	for key, entry := range la.sMap {
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
		epsilon: statelessEpsilon,
		sMap:    make(sMap),
	}
}

func (la *learningAiStateless) beginEpisode() {
	la.sequence = la.sequence[:0]
	la.replay = nil
	la.replayIndex = 0
	la.replaying = false
	if rand.Float64() < la.epsilon && la.bestKey != "" {
		la.replay = strings.Split(la.bestKey, ";")
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
	previousValue := entry.value
	entry.count++
	entry.value += (reward - entry.value) / float64(entry.count)
	la.sMap[key] = entry
	la.updateBestKey(key, previousValue)
}

func (la *learningAiStateless) updateBestKey(key string, previousValue float64) {
	candidate := la.sMap[key]
	if la.bestKey == "" {
		la.bestKey = key
		return
	}

	best := la.sMap[la.bestKey]
	if key == la.bestKey {
		if candidate.value <= previousValue {
			la.recomputeBestKey()
		}
		return
	}
	if candidate.value > best.value || (candidate.value == best.value && candidate.count < best.count) {
		la.bestKey = key
	}
}

func (la *learningAiStateless) recomputeBestKey() {
	la.bestKey = ""
	var best struct {
		count int
		value float64
	}
	for key, entry := range la.sMap {
		if la.bestKey == "" || entry.value > best.value || (entry.value == best.value && entry.count < best.count) {
			la.bestKey = key
			best = entry
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
