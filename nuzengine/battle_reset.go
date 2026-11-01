package nuzengine

import (
	"maps"
	"slices"
)

// A reset point is the state bs.reset() returns a battle to. It is stored in place: battle-level
// values in battleResetState and per-pokemon values in pokemon.resetPoint. Pokémon and slots are
// referenced by index, so the reset point of one battle can be taken from another battle
// with the same parties and slot layout.
type battleResetState struct {
	weather          weatherState
	fieldEffects     map[fieldEffect]int
	player, opponent trainerResetState
	slots            []slotResetState
	// midBattle is set when the reset point was taken from a battle in progress,
	// whose switch-in abilities have already taken effect.
	midBattle bool
}

type trainerResetState struct {
	fieldEffects map[fieldEffect]int
	lost         bool
}

type slotResetState struct {
	mon          int
	firstTurn    bool
	suckerPunch  bool
	protected    bool
	protectTurns int
	unnerved     bool
	invulnerable *invulnerableResetState
}

type invulnerableResetState struct {
	target int
	move   int
}

type pokemonResetState struct {
	hp           int
	fainted      bool
	stages       []int
	ailments     []ailmentResetState
	pp           []int
	lockedMove   int
	item         itemState
	itemConsumed bool
	ability      abilityState
	unnerved     bool
	flashFire    bool
	unburden     bool
	trace        bool
	focusEnergy  bool
	laserFocus   bool
}

type ailmentResetState struct {
	state       ailmentState
	turns       int
	afflictedBy int
}

// setResetPoint makes the current state of src the reset point of dst. It reads src and writes dst only.
func setResetPoint(dst, src battleState, midBattle bool) {
	srcSlots := src.getAllSlots()
	srcPlayer, srcOpponent := src.getPlayerTrainer(), opponentTrainer(src)
	dstPlayer, dstOpponent := dst.getPlayerTrainer(), opponentTrainer(dst)

	rp := dst.getResetPoint()
	rp.weather = src.getWeather()
	rp.fieldEffects = maps.Clone(src.getFieldEffects())
	rp.player = trainerResetState{fieldEffects: maps.Clone(srcPlayer.fieldEffects), lost: srcPlayer.lost}
	rp.opponent = trainerResetState{fieldEffects: maps.Clone(srcOpponent.fieldEffects), lost: srcOpponent.lost}
	rp.midBattle = midBattle

	for i, mon := range srcPlayer.pokemonParty {
		dstPlayer.pokemonParty[i].resetPoint.capture(mon, srcSlots)
	}
	for i, mon := range srcOpponent.pokemonParty {
		dstOpponent.pokemonParty[i].resetPoint.capture(mon, srcSlots)
	}

	rp.slots = rp.slots[:0]
	for _, s := range srcSlots {
		sr := slotResetState{
			mon:          slices.Index(s.Trainer.pokemonParty, s.mon),
			firstTurn:    s.firstTurn,
			suckerPunch:  s.suckerPunch,
			protected:    s.protected,
			protectTurns: s.protectTurns,
			unnerved:     s.unnerved,
		}
		if ia := s.invulnerableAction; ia != nil {
			sr.invulnerable = &invulnerableResetState{
				target: slices.Index(srcSlots, ia.targetSlot),
				move:   slices.Index(s.mon.moves, ia.move),
			}
		}
		rp.slots = append(rp.slots, sr)
	}
}

func (r *pokemonResetState) capture(mon *pokemon, slots []*slot) {
	*r = pokemonResetState{
		hp:           mon.hp,
		fainted:      mon.fainted,
		stages:       append(r.stages[:0], mon.stages...),
		ailments:     r.ailments[:0],
		pp:           r.pp[:0],
		lockedMove:   slices.Index(mon.moves, mon.lockedMove),
		item:         mon.item.State,
		itemConsumed: mon.item.Consumed,
		ability:      mon.ability,
		unnerved:     mon.unnerved,
		flashFire:    mon.flashFire,
		unburden:     mon.unburden,
		trace:        mon.trace,
		focusEnergy:  mon.focusEnergy,
		laserFocus:   mon.laserFocus,
	}
	for _, move := range mon.moves {
		r.pp = append(r.pp, move.PP)
	}
	for _, a := range mon.ailments {
		r.ailments = append(r.ailments, ailmentResetState{state: a.State, turns: a.Turns, afflictedBy: slices.Index(slots, a.afflictedBy)})
	}
}

// resetBattle returns bs to its reset point and clears any in-flight turn state.
func resetBattle(bs battleState) {
	rp := bs.getResetPoint()
	slots := bs.getAllSlots()
	player, opponent := bs.getPlayerTrainer(), opponentTrainer(bs)

	for _, t := range []struct {
		trainer *trainer
		state   trainerResetState
	}{{player, rp.player}, {opponent, rp.opponent}} {
		t.trainer.lost = t.state.lost
		t.trainer.fieldEffects = maps.Clone(t.state.fieldEffects)
		if t.trainer.fieldEffects == nil {
			t.trainer.fieldEffects = make(map[fieldEffect]int)
		}
		for _, mon := range t.trainer.pokemonParty {
			mon.reset(slots)
		}
	}

	for i, sr := range rp.slots {
		s := slots[i]
		s.mon = s.Trainer.pokemonParty[sr.mon]
		s.firstTurn = sr.firstTurn
		s.suckerPunch = sr.suckerPunch
		s.protected = sr.protected
		s.protectTurns = sr.protectTurns
		s.unnerved = sr.unnerved
		s.invulnerableAction = nil
		if inv := sr.invulnerable; inv != nil && inv.move >= 0 && inv.target >= 0 {
			s.invulnerableAction = &moveAction{userSlot: s, targetSlot: slots[inv.target], move: s.mon.moves[inv.move]}
		}
	}

	if fx := bs.getFieldEffects(); fx != nil {
		clear(fx)
		maps.Copy(fx, rp.fieldEffects)
	}
	bs.setWeather(rp.weather)

	actions := bs.getActions()
	actions.queue = actions.queue[:0]
	bs.setError(nil)

	if !rp.midBattle {
		resolveOnEntry(bs)
	}
}

func opponentTrainer(bs battleState) *trainer {
	for _, s := range bs.getAllSlots() {
		if s.Trainer != bs.getPlayerTrainer() {
			return s.Trainer
		}
	}
	return nil
}
