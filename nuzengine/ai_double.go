package nuzengine

const (
	noActionKey        = "none"
	compositeSeparator = ","
)

// doublePlanner picks the actions of both player slots as one composite
// action, ordered first slot then second slot, when the first slot is asked.
// The second slot's part is held until the second slot is asked.
type doublePlanner struct {
	second *actionCandidate
}

type compositeCandidate struct {
	first, second actionCandidate
}

func (c compositeCandidate) key() string {
	return c.first.key + compositeSeparator + c.second.key
}

// plan returns this slot's part of the composite action. When a new composite
// was chosen, recorded is true and state/key describe the trajectory step.
func (dp *doublePlanner) plan(dbs *doubleBattleState, s *slot, actions []*moveAction, pick func(state string, candidates []actionCandidate) actionCandidate) (mine actionCandidate, state, key string, recorded bool) {
	if s == dbs.activePlayerSlotTwo && dp.second != nil {
		mine = *dp.second
		dp.second = nil
		return mine, "", "", false
	}
	dp.second = nil

	none := []actionCandidate{{key: noActionKey}}
	ownCandidates := buildCandidates(dbs, s, actions)
	firstCandidates, secondCandidates := ownCandidates, none
	if s == dbs.activePlayerSlotTwo {
		firstCandidates, secondCandidates = none, ownCandidates
	} else if partner := dbs.activePlayerSlotTwo; !partner.mon.fainted && partner.invulnerableAction == nil {
		secondCandidates = buildCandidates(dbs, partner, possibleMoveActions(dbs, partner))
	}

	var composites []compositeCandidate
	var candidates []actionCandidate
	for _, first := range firstCandidates {
		for _, second := range secondCandidates {
			if first.target != nil && first.target == second.target {
				continue
			}
			composite := compositeCandidate{first: first, second: second}
			composites = append(composites, composite)
			candidates = append(candidates, actionCandidate{key: composite.key()})
		}
	}

	state = dbs.key()
	key = pick(state, candidates).key
	var chosen compositeCandidate
	for _, composite := range composites {
		if composite.key() == key {
			chosen = composite
			break
		}
	}

	if s == dbs.activePlayerSlotTwo {
		return chosen.second, state, key, true
	}
	if chosen.second.key != noActionKey {
		second := chosen.second
		dp.second = &second
	}
	return chosen.first, state, key, true
}

// resolveCandidate maps a planned candidate onto the slot's real actions.
func resolveCandidate(actions []*moveAction, c actionCandidate) (*moveAction, int) {
	if c.move == nil {
		return actions[0], -1
	}
	for _, a := range actions {
		if a.move == c.move.move && a.targetSlot == c.move.targetSlot {
			return a, 0
		}
	}
	return actions[0], 0
}
