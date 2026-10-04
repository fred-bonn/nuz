package nuzengine

import (
	"math/rand"
	"slices"
	"strings"

	"github.com/fred-bonn/nuz/nuzengine/internal/pokeapi"
)

type learningAiStateless struct {
	sequences     sMap
	bestValue     float64
	pending       actionCandidate
	pendingSwitch bool
	sequence      []string
	replay        []string
	replayIndex   int
	replaying     bool
	verbose       bool
}

type sMap map[string]int

func LearnStateless(playerPartyStr, opponentPartyStr string, weatherInt, iterations int) {
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

	la := &learningAiStateless{
		sequences: make(sMap),
		bestValue: -1e9,
	}
	if Verbose {
		la.verbose = true
	}
	Verbose = false

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

	if la.verbose {
		Verbose = true
		la.printStatelessResults()
	}
}

func (la *learningAiStateless) printStatelessResults() {
	firstActionCounts := make(map[string]int)
	for sequence, count := range la.sequences {
		firstAction, _, _ := strings.Cut(sequence, ";")
		if firstAction != "" {
			firstActionCounts[firstAction] += count
		}
	}

	mostFrequentActions := make([]string, 0)
	mostFrequentCount := 0
	for action, count := range firstActionCounts {
		switch {
		case count > mostFrequentCount:
			mostFrequentActions = append(mostFrequentActions[:0], action)
			mostFrequentCount = count
		case count == mostFrequentCount:
			mostFrequentActions = append(mostFrequentActions, action)
		}
	}
	slices.Sort(mostFrequentActions)
	vprintln("Most frequent first action(s):")
	if len(mostFrequentActions) == 0 {
		vprintln("none")
		return
	}
	for _, action := range mostFrequentActions {
		vprintf("%s - %d occurrences", action, mostFrequentCount)
	}
}

func (la *learningAiStateless) beginEpisode() {
	la.sequence = la.sequence[:0]
	la.replay = nil
	la.replayIndex = 0
	la.replaying = false
	la.pendingSwitch = false
}

func (la *learningAiStateless) endEpisode(reward float64) {
	if reward < la.bestValue {
		return
	}
	if reward > la.bestValue {
		la.bestValue = reward
		la.sequences = make(sMap)
	}

	key := strings.Join(la.sequence, ";")
	la.sequences[key] = 1
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
