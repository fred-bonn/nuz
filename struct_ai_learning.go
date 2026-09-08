package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fred-bonn/nuz/internal/parser"
)

const (
	learnRewardWinBonus            = 5000000.0
	learnRewardDeadMonPenalty      = 1000000.0
	learnRewardGuaranteedKillBonus = 750.0
	learnRewardCritRiskPenalty     = 3200.0
	learnRewardStateRiskPenalty    = 1500.0
	learnActionCountBoost          = 2.5
	learnRewardDiscountFactor      = 0.9
	learnRewardDecayFactor         = 0.98
	learnRewardLearningRate        = 0.25
	learnRewardStableThreshold     = 2
)

type discreteBattleState struct {
	PlayerPokemon            string `json:"player_pokemon"`
	OpponentPokemon          string `json:"opponent_pokemon"`
	PlayerMonIsFaster        bool   `json:"player_mon_is_faster"`
	OpponentHasMoveThatKills bool   `json:"opponent_has_move_that_kills"`
}

// discretizeBattleState returns a canonical structured representation suitable for a tabular policy.
func discretizeBattleState(bs battleState) discreteBattleState {
	state := discreteBattleState{}
	if bs == nil {
		return state
	}

	var playerSlot *slot
	for _, slot := range bs.getAllSlots() {
		if slot != nil && slot.trainer != nil && slot.trainer.player {
			playerSlot = slot
			break
		}
	}
	if playerSlot == nil {
		return state
	}
	opponentSlot := bs.getOpponentSlot(playerSlot)
	if playerSlot.mon == nil || opponentSlot == nil || opponentSlot.mon == nil {
		return state
	}

	state.PlayerPokemon = playerSlot.mon.base.Name
	state.OpponentPokemon = opponentSlot.mon.base.Name
	playerSpeed := playerSlot.mon.effectiveSpeed(bs)
	opponentSpeed := opponentSlot.mon.effectiveSpeed(bs)
	state.PlayerMonIsFaster = playerSpeed > opponentSpeed
	state.OpponentHasMoveThatKills = opponentHasMoveThatKills(bs, opponentSlot.mon, playerSlot.mon)

	return state
}

func (state discreteBattleState) key() string {
	return string(mustMarshalDiscreteState(state))
}

func mustMarshalDiscreteState(state discreteBattleState) []byte {
	encoded, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}
	return encoded
}

func opponentHasMoveThatKills(bs battleState, user, target *pokemon) bool {
	for _, move := range user.moves {
		if move == nil || move.PP <= 0 || move.Class == statusClass {
			continue
		}
		if moveCanKill(bs, user, target, move) {
			return true
		}
		if !target.ability.blocksCrits() && user.ability != moldBreakerAbility && moveCanCritKill(bs, user, target, move) {
			return true
		}
	}
	return false
}

func moveCanKill(bs battleState, user, target *pokemon, move *Move) bool {
	rolls := 1
	if move.MaxHits == 5 {
		rolls = 3
	} else if move.MaxHits > 0 {
		rolls = move.MaxHits
	}
	damage := 0
	critRate := determineCritRate(user, move)
	for i := 0; i < rolls; i++ {
		damage += calculateDamage(user, target, move, new(critRate >= 3), bs.getWeather(), true, true, false)
	}
	target.checkItemTrigger(false, makeFocusSashEvent(&damage))
	if target.ability == sturdyAbility && target.hp == target.maxHP() {
		damage = min(damage, target.hp-1)
	}
	return damage >= target.hp
}

func moveCanCritKill(bs battleState, user, target *pokemon, move *Move) bool {
	if move == nil || move.PP <= 0 || move.Class == statusClass {
		return false
	}
	critRate := determineCritRate(user, move)
	if critRate <= 0 {
		return false
	}
	critMult := 1
	if critRate >= 3 {
		critMult = 2
	}
	damage := 0
	rolls := 1
	if move.MaxHits == 5 {
		rolls = 3
	} else if move.MaxHits > 0 {
		rolls = move.MaxHits
	}
	for i := 0; i < rolls; i++ {
		damage += calculateDamage(user, target, move, new(true), bs.getWeather(), true, true, false)
	}
	if critMult > 1 {
		damage *= critMult
	}
	target.checkItemTrigger(false, makeFocusSashEvent(&damage))
	if target.ability == sturdyAbility && target.hp == target.maxHP() {
		damage = min(damage, target.hp-1)
	}
	return damage >= target.hp
}

type learningAi struct {
	policy              map[string][]string
	scores              map[string]map[string]float64
	counts              map[string]map[string]int
	history             []stateActionEntry
	seen                map[string]map[string]bool
	decayFactor         float64
	learningRate        float64
	discountFactor      float64
	lastPolicySignature string
	stableRounds        int
}

type stateActionEntry struct {
	stateKey string
	action   string
}

// savedPolicy stores the entire contents of the player/opponent Showdown party files
// (as if read via `cat player.txt`) so the policy can rebuild the parties without the
// original files, by running player_party/opponent_party back through the parser.
type savedPolicy struct {
	PlayerParty   string                        `json:"player_party"`
	OpponentParty string                        `json:"opponent_party"`
	Policy        map[string][]string           `json:"policy"`
	Scores        map[string]map[string]float64 `json:"scores"`
	Counts        map[string]map[string]int     `json:"counts"`
	Version       int                           `json:"version"`
	Metadata      map[string]string             `json:"metadata,omitempty"`
}

func newLearningAi() *learningAi {
	return &learningAi{
		policy:         make(map[string][]string),
		scores:         make(map[string]map[string]float64),
		counts:         make(map[string]map[string]int),
		seen:           make(map[string]map[string]bool),
		decayFactor:    learnRewardDecayFactor,
		learningRate:   learnRewardLearningRate,
		discountFactor: learnRewardDiscountFactor,
	}
}

func policyPathForInputs(playerInput, opponentInput string) string {
	playerName := strings.TrimSuffix(filepath.Base(playerInput), filepath.Ext(playerInput))
	opponentName := strings.TrimSuffix(filepath.Base(opponentInput), filepath.Ext(opponentInput))
	return filepath.Join("policies", fmt.Sprintf("%s__vs__%s.json", playerName, opponentName))
}

func loadPolicyFromDisk(path string) (*savedPolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	policy := &savedPolicy{}
	if err := json.Unmarshal(data, policy); err != nil {
		return nil, err
	}
	if policy.Policy == nil {
		policy.Policy = make(map[string][]string)
	}
	if policy.Scores == nil {
		policy.Scores = make(map[string]map[string]float64)
	}
	if policy.Counts == nil {
		policy.Counts = make(map[string]map[string]int)
	}
	return policy, nil
}

func savePolicyToDisk(la *learningAi, playerInput, opponentInput string) error {
	if la == nil {
		return fmt.Errorf("nil learning AI")
	}
	path := policyPathForInputs(playerInput, opponentInput)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	playerPartyFile, err := os.ReadFile(playerInput)
	if err != nil {
		return fmt.Errorf("failed reading player input '%s': %w", playerInput, err)
	}
	opponentPartyFile, err := os.ReadFile(opponentInput)
	if err != nil {
		return fmt.Errorf("failed reading opponent input '%s': %w", opponentInput, err)
	}
	policy := &savedPolicy{
		PlayerParty:   string(playerPartyFile),
		OpponentParty: string(opponentPartyFile),
		Policy:        cloneActionMap(la.policy),
		Scores:        cloneScoreMap(la.scores),
		Counts:        cloneCountMap(la.counts),
		Version:       1,
		Metadata:      map[string]string{"player_input": playerInput, "opponent_input": opponentInput},
	}
	data, err := json.MarshalIndent(policy, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func cloneActionMap(source map[string][]string) map[string][]string {
	if source == nil {
		return make(map[string][]string)
	}
	clone := make(map[string][]string, len(source))
	for key, values := range source {
		copied := make([]string, len(values))
		copy(copied, values)
		clone[key] = copied
	}
	return clone
}

func cloneScoreMap(source map[string]map[string]float64) map[string]map[string]float64 {
	if source == nil {
		return make(map[string]map[string]float64)
	}
	clone := make(map[string]map[string]float64, len(source))
	for key, values := range source {
		copied := make(map[string]float64, len(values))
		for action, score := range values {
			copied[action] = score
		}
		clone[key] = copied
	}
	return clone
}

func cloneCountMap(source map[string]map[string]int) map[string]map[string]int {
	if source == nil {
		return make(map[string]map[string]int)
	}
	clone := make(map[string]map[string]int, len(source))
	for key, values := range source {
		copied := make(map[string]int, len(values))
		for action, count := range values {
			copied[action] = count
		}
		clone[key] = copied
	}
	return clone
}

func countScoreEntries(scores map[string]map[string]float64) int {
	count := 0
	for _, actions := range scores {
		count += len(actions)
	}
	return count
}

func countCountEntries(counts map[string]map[string]int) int {
	count := 0
	for _, actions := range counts {
		count += len(actions)
	}
	return count
}

func loadLearningAIFromPolicy(policy *savedPolicy) *learningAi {
	if policy == nil {
		return newLearningAi()
	}
	la := &learningAi{
		policy:      cloneActionMap(policy.Policy),
		scores:      cloneScoreMap(policy.Scores),
		counts:      cloneCountMap(policy.Counts),
		seen:        make(map[string]map[string]bool),
		decayFactor: 0.98,
	}
	if la.policy == nil {
		la.policy = make(map[string][]string)
	}
	if la.scores == nil {
		la.scores = make(map[string]map[string]float64)
	}
	if la.counts == nil {
		la.counts = make(map[string]map[string]int)
	}
	for stateKey, actions := range la.policy {
		la.seen[stateKey] = make(map[string]bool)
		for _, action := range actions {
			la.seen[stateKey][action] = true
		}
	}
	return la
}

type staticPolicyAi struct {
	policy map[string][]string
	scores map[string]map[string]float64
	counts map[string]map[string]int
}

func newStaticPolicyAiFromPolicy(policy *savedPolicy) *staticPolicyAi {
	if policy == nil {
		return &staticPolicyAi{policy: map[string][]string{}, scores: map[string]map[string]float64{}, counts: map[string]map[string]int{}}
	}
	return &staticPolicyAi{
		policy: cloneActionMap(policy.Policy),
		scores: cloneScoreMap(policy.Scores),
		counts: cloneCountMap(policy.Counts),
	}
}

func (spa *staticPolicyAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	if len(actions) == 0 {
		return nil, 0
	}
	stateKey := discretizeBattleState(bs).key()
	fallbackAction, fallbackScore := rnbAi{}.evaluateActions(bs, slot, actions)
	hasRecordedEvidence := false
	for _, action := range actions {
		key := actionKey(action)
		if spa.counts[stateKey][key] > 0 || spa.scores[stateKey][key] != 0 {
			hasRecordedEvidence = true
			break
		}
	}
	if !hasRecordedEvidence {
		return actions[0], 0
	}

	// Try to find an action with a positive score from the policy
	bestAction := fallbackAction
	bestScore := float64(fallbackScore)
	riskPenalty := 0.0
	if stateHasCritKillRisk(stateKey) {
		riskPenalty = learnRewardStateRiskPenalty
	}
	positiveFound := false
	for _, action := range actions {
		key := actionKey(action)
		score := spa.scoreFor(stateKey, key)
		if count := spa.counts[stateKey][key]; count > 0 {
			score += float64(count) * learnActionCountBoost
		}
		if action != nil && action.move != nil && action.userSlot != nil && action.targetSlot != nil && action.userSlot.mon != nil && action.targetSlot.mon != nil {
			if moveCanKill(bs, action.userSlot.mon, action.targetSlot.mon, action.move) {
				score += learnRewardGuaranteedKillBonus
			}
		}
		score -= riskPenalty
		if score > 0 {
			positiveFound = true
			if score > bestScore {
				bestAction = action
				bestScore = score
			}
		}
	}
	if !positiveFound {
		return bestAction, int(bestScore)
	}
	return bestAction, int(bestScore)
}

func (spa *staticPolicyAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if len(mons) == 0 {
		return nil
	}
	stateKey := discretizeBattleState(bs).key()
	fallbackMon := rnbAi{}.evaluteSwitchIns(bs, mons, opponentSlot)
	for _, mon := range mons {
		key := actionKeyForSwitch(mon)
		if spa.counts[stateKey][key] > 0 || spa.scores[stateKey][key] != 0 {
			goto recordedSwitchEvidence
		}
	}
	return mons[0]

recordedSwitchEvidence:
	bestMon := fallbackMon
	bestScore := 0.0
	riskPenalty := 0.0
	if stateHasCritKillRisk(stateKey) {
		riskPenalty = learnRewardStateRiskPenalty
	}

	positiveFound := false
	for _, mon := range mons {
		key := actionKeyForSwitch(mon)
		score := spa.scoreFor(stateKey, key)
		if count := spa.counts[stateKey][key]; count > 0 {
			score += float64(count) * learnActionCountBoost
		}
		score -= riskPenalty
		if score > 0 {
			positiveFound = true
			if score > bestScore {
				bestMon = mon
				bestScore = score
			}
		}
	}
	if !positiveFound {
		return fallbackMon
	}
	return bestMon
}

func (spa *staticPolicyAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	if spa == nil || len(party) <= 1 {
		return false
	}
	stateKey := discretizeBattleState(bs).key()
	candidateFound := false
	for _, mon := range party {
		if mon == nil || mon == slot.mon || mon.fainted {
			continue
		}
		candidateFound = true
		key := actionKeyForSwitch(mon)
		if learned, ok := spa.scores[stateKey][key]; ok && learned > float64(score) {
			return true
		}
	}
	if !candidateFound {
		return false
	}
	if (rnbAi{}).shouldSwitch(bs, slot, score, party) {
		return true
	}
	if slot.mon.hp <= slot.mon.maxHP()/3 {
		return true
	}
	return false
}

func (spa *staticPolicyAi) scoreFor(stateKey, actionKey string) float64 {
	if spa == nil {
		return 0
	}
	if spa.scores == nil {
		return 0
	}
	if stateScores, ok := spa.scores[stateKey]; ok {
		if score, ok := stateScores[actionKey]; ok {
			return score
		}
	}
	if spa.policy != nil {
		for _, allowedKey := range spa.policy[stateKey] {
			if allowedKey == actionKey {
				return 0
			}
		}
	}
	return -1e18
}

func validatePolicyCompatibility(policy *savedPolicy, playerParty, opponentParty []*pokemon) error {
	if policy == nil {
		return fmt.Errorf("policy is nil")
	}
	if policy.PlayerParty == "" || policy.OpponentParty == "" {
		return fmt.Errorf("policy is missing embedded player/opponent party files")
	}
	if err := validatePartyMatchesShowdown(policy.PlayerParty, playerParty); err != nil {
		return fmt.Errorf("player party: %w", err)
	}
	if err := validatePartyMatchesShowdown(policy.OpponentParty, opponentParty); err != nil {
		return fmt.Errorf("opponent party: %w", err)
	}
	return nil
}

// validatePartyMatchesShowdown parses raw Showdown party text and checks that it names
// the same Pok\u00e9mon, in the same order, as the already-loaded party.
func validatePartyMatchesShowdown(content string, party []*pokemon) error {
	parsed, err := parser.ParseShowdown(content)
	if err != nil {
		return fmt.Errorf("failed parsing embedded party: %w", err)
	}
	if len(parsed) != len(party) {
		return fmt.Errorf("expected %d pokemon, got %d", len(parsed), len(party))
	}
	for i, mon := range parsed {
		if party[i] == nil || cleanName(mon.Name) != party[i].base.Name {
			return fmt.Errorf("pokemon at position %d does not match: policy=%q loaded=%q", i, mon.Name, party[i].base.Name)
		}
	}
	return nil
}

func (la *learningAi) ensureState(stateKey string) {
	if la == nil {
		return
	}
	if la.policy == nil {
		la.policy = make(map[string][]string)
	}
	if la.scores == nil {
		la.scores = make(map[string]map[string]float64)
	}
	if la.counts == nil {
		la.counts = make(map[string]map[string]int)
	}
	if la.seen == nil {
		la.seen = make(map[string]map[string]bool)
	}
	if _, ok := la.policy[stateKey]; !ok {
		la.policy[stateKey] = []string{}
	}
	if _, ok := la.scores[stateKey]; !ok {
		la.scores[stateKey] = make(map[string]float64)
	}
	if _, ok := la.counts[stateKey]; !ok {
		la.counts[stateKey] = make(map[string]int)
	}
	if _, ok := la.seen[stateKey]; !ok {
		la.seen[stateKey] = make(map[string]bool)
	}
}

func (la *learningAi) decayScores(factor float64) {
	if la == nil {
		return
	}
	if factor <= 0 {
		factor = la.decayFactor
	}
	if factor == 0 {
		factor = 0.98
	}
	for stateKey, actionScores := range la.scores {
		for action, score := range actionScores {
			actionScores[action] = score * factor
			if la.counts[stateKey] != nil && la.counts[stateKey][action] > 0 {
				actionScores[action] += float64(la.counts[stateKey][action]) * 0.5
			}
		}
	}
}

func (la *learningAi) policySignature() string {
	if la == nil {
		return ""
	}
	stateKeys := make([]string, 0, len(la.policy))
	for stateKey := range la.policy {
		stateKeys = append(stateKeys, stateKey)
	}
	sort.Strings(stateKeys)
	var b strings.Builder
	for _, stateKey := range stateKeys {
		actionKeys := make([]string, 0, len(la.scores[stateKey]))
		for actionKey := range la.scores[stateKey] {
			actionKeys = append(actionKeys, actionKey)
		}
		sort.Strings(actionKeys)
		for _, actionKey := range actionKeys {
			fmt.Fprintf(&b, "%s:%s=%.6f;", stateKey, actionKey, la.scores[stateKey][actionKey])
		}
	}
	return b.String()
}

func (la *learningAi) policySaturated() bool {
	if la == nil || len(la.policy) == 0 {
		return false
	}
	current := la.policySignature()
	if la.lastPolicySignature == "" {
		la.lastPolicySignature = current
		la.stableRounds = 0
		return false
	}
	if current == la.lastPolicySignature {
		la.stableRounds++
		if la.stableRounds >= 2 {
			return true
		}
		return false
	}
	la.lastPolicySignature = current
	la.stableRounds = 0
	return false
}

func (la *learningAi) recordStateAction(stateKey string, action string) {
	if la == nil || stateKey == "" || action == "" {
		return
	}
	la.ensureState(stateKey)
	la.policy[stateKey] = appendUnique(la.policy[stateKey], action)
	la.counts[stateKey][action]++
	if la.seen[stateKey][action] {
		return
	}
	la.seen[stateKey][action] = true
	la.history = append(la.history, stateActionEntry{stateKey: stateKey, action: action})
}

func (la *learningAi) recordBattleOutcome(stats *battleStatistics) {
	if la == nil || stats == nil {
		return
	}

	la.decayScores(la.decayFactor)
	reward := stats.outcomeScore()
	if stats.winCount == stats.battleCount {
		reward += learnRewardWinBonus
	}
	for _, survivors := range stats.pokemonSurvivors {
		if survivors < stats.battleCount {
			reward -= learnRewardDeadMonPenalty
		}
	}
	for _, entry := range la.history {
		la.ensureState(entry.stateKey)
		current := la.scores[entry.stateKey][entry.action]
		bestFuture := 0.0
		for _, candidateScore := range la.scores[entry.stateKey] {
			if candidateScore > bestFuture {
				bestFuture = candidateScore
			}
		}
		if stateHasCritKillRisk(entry.stateKey) {
			reward -= learnRewardCritRiskPenalty
		}
		target := reward + la.discountFactor*bestFuture
		la.scores[entry.stateKey][entry.action] = current + la.learningRate*(target-current)
		if la.counts[entry.stateKey] != nil {
			la.counts[entry.stateKey][entry.action]++
		}
	}
	la.history = nil
	la.seen = make(map[string]map[string]bool)
}

func (la *learningAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	if len(actions) == 0 {
		return nil, 0
	}
	if la == nil {
		return rnbAi{}.evaluateActions(bs, slot, actions)
	}
	stateKey := discretizeBattleState(bs).key()
	la.ensureState(stateKey)
	firstAction := actions[0]
	fallbackAction, fallbackScore := rnbAi{}.evaluateActions(bs, slot, actions)
	for _, action := range actions {
		key := actionKey(action)
		la.policy[stateKey] = appendUnique(la.policy[stateKey], key)
		if _, ok := la.scores[stateKey][key]; !ok {
			la.scores[stateKey][key] = 0
		}
		if _, ok := la.counts[stateKey][key]; !ok {
			la.counts[stateKey][key] = 0
		}
	}

	hasRecordedEvidence := false
	for _, action := range actions {
		key := actionKey(action)
		if la.counts[stateKey][key] > 0 || la.scores[stateKey][key] != 0 {
			hasRecordedEvidence = true
			break
		}
	}
	if !hasRecordedEvidence {
		la.recordStateAction(stateKey, actionKey(firstAction))
		return firstAction, 0
	}

	bestAction := fallbackAction
	bestScore := float64(fallbackScore)
	riskPenalty := 0.0
	if stateHasCritKillRisk(stateKey) {
		riskPenalty = learnRewardStateRiskPenalty
	}
	positiveFound := false
	for _, action := range actions {
		key := actionKey(action)
		score := la.scores[stateKey][key]
		if count := la.counts[stateKey][key]; count > 0 {
			score += float64(count) * learnActionCountBoost
		}
		if action != nil && action.move != nil && action.userSlot != nil && action.targetSlot != nil && action.userSlot.mon != nil && action.targetSlot.mon != nil {
			if moveCanKill(bs, action.userSlot.mon, action.targetSlot.mon, action.move) {
				score += learnRewardGuaranteedKillBonus
			}
		}
		score -= riskPenalty
		if score > 0 {
			positiveFound = true
			if score > bestScore {
				bestAction = action
				bestScore = score
			}
		}
	}
	if !positiveFound {
		la.recordStateAction(stateKey, actionKey(bestAction))
		return bestAction, int(bestScore)
	}
	la.recordStateAction(stateKey, actionKey(bestAction))
	return bestAction, int(bestScore)
}

func (la *learningAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if len(mons) == 0 {
		return nil
	}
	if la == nil {
		return rnbAi{}.evaluteSwitchIns(bs, mons, opponentSlot)
	}
	stateKey := discretizeBattleState(bs).key()
	la.ensureState(stateKey)
	firstMon := mons[0]
	fallbackMon := rnbAi{}.evaluteSwitchIns(bs, mons, opponentSlot)
	bestMon := fallbackMon
	bestScore := 0.0
	riskPenalty := 0.0
	if stateHasCritKillRisk(stateKey) {
		riskPenalty = learnRewardStateRiskPenalty
	}

	hasRecordedEvidence := false
	for _, mon := range mons {
		key := actionKeyForSwitch(mon)
		if la.counts[stateKey][key] > 0 || la.scores[stateKey][key] != 0 {
			hasRecordedEvidence = true
			break
		}
	}
	if !hasRecordedEvidence {
		la.recordStateAction(stateKey, actionKeyForSwitch(firstMon))
		return firstMon
	}

	positiveFound := false
	for _, mon := range mons {
		key := actionKeyForSwitch(mon)
		la.policy[stateKey] = appendUnique(la.policy[stateKey], key)
		if _, ok := la.scores[stateKey][key]; !ok {
			la.scores[stateKey][key] = 0
		}
		if _, ok := la.counts[stateKey][key]; !ok {
			la.counts[stateKey][key] = 0
		}
		score := la.scores[stateKey][key]
		if count := la.counts[stateKey][key]; count > 0 {
			score += float64(count) * 2.5
		}
		score -= riskPenalty
		if score > 0 {
			positiveFound = true
			if score > bestScore {
				bestMon = mon
				bestScore = score
			}
		}
	}
	if !positiveFound {
		la.recordStateAction(stateKey, actionKeyForSwitch(bestMon))
		return bestMon
	}
	la.recordStateAction(stateKey, actionKeyForSwitch(bestMon))
	return bestMon
}

func (la *learningAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	if la == nil || len(party) <= 1 {
		return false
	}
	stateKey := discretizeBattleState(bs).key()
	la.ensureState(stateKey)
	candidateFound := false
	for _, mon := range party {
		if mon == nil || mon == slot.mon || mon.fainted {
			continue
		}
		candidateFound = true
		key := actionKeyForSwitch(mon)
		if learned, ok := la.scores[stateKey][key]; ok && learned > float64(score) {
			return true
		}
	}
	if !candidateFound {
		return false
	}
	if (rnbAi{}).shouldSwitch(bs, slot, score, party) {
		return true
	}
	if slot.mon.hp <= slot.mon.maxHP()/3 {
		return true
	}
	return false
}

func stateHasCritKillRisk(stateKey string) bool {
	if stateKey == "" {
		return false
	}
	return strings.Contains(stateKey, "\"opponent_crit_kill_risk\":true") || strings.Contains(stateKey, "\"opponent_has_move_that_kills\":true")
}

func actionKey(action *moveAction) string {
	return "move:" + action.move.Name
}

func appendUnique(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func actionKeyForSwitch(mon *pokemon) string {
	return "switch:" + mon.base.Name
}
