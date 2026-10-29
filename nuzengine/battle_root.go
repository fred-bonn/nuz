package nuzengine

import (
	"maps"
	"slices"
)

// battleRoot is a copy of a battle's mutable state. Applying it to a structurally
// identical battle (same parties and slot layout) makes that battle continue from
// the captured state. Pokémon and slots are referenced by index, never by pointer.
type battleRoot struct {
	player, opponent partyRoot
	slots            []slotRoot
	weather          weatherState
	fieldEffects     map[fieldEffect]int
}

type partyRoot struct {
	mons         []pokemonRoot
	fieldEffects map[fieldEffect]int
	lost         bool
}

type pokemonRoot struct {
	hp           int
	fainted      bool
	stages       []int
	ailments     []ailmentRoot
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

type ailmentRoot struct {
	state       ailmentState
	turns       int
	afflictedBy int
}

type slotRoot struct {
	mon          int
	firstTurn    bool
	suckerPunch  bool
	protected    bool
	protectTurns int
	unnerved     bool
	invulnerable *invulnerableRoot
}

type invulnerableRoot struct {
	target int
	move   int
}

func captureRoot(bs battleState) battleRoot {
	slots := bs.getAllSlots()
	root := battleRoot{
		player:       capturePartyRoot(bs.getPlayerTrainer(), slots),
		opponent:     capturePartyRoot(opponentTrainer(bs), slots),
		weather:      bs.getWeather(),
		fieldEffects: maps.Clone(bs.getFieldEffects()),
	}
	for _, s := range slots {
		sr := slotRoot{
			mon:          slices.Index(s.Trainer.pokemonParty, s.mon),
			firstTurn:    s.firstTurn,
			suckerPunch:  s.suckerPunch,
			protected:    s.protected,
			protectTurns: s.protectTurns,
			unnerved:     s.unnerved,
		}
		if ia := s.invulnerableAction; ia != nil {
			sr.invulnerable = &invulnerableRoot{
				target: slices.Index(slots, ia.targetSlot),
				move:   slices.Index(s.mon.moves, ia.move),
			}
		}
		root.slots = append(root.slots, sr)
	}
	return root
}

func opponentTrainer(bs battleState) *trainer {
	for _, s := range bs.getAllSlots() {
		if s.Trainer != bs.getPlayerTrainer() {
			return s.Trainer
		}
	}
	return nil
}

func capturePartyRoot(t *trainer, slots []*slot) partyRoot {
	pr := partyRoot{fieldEffects: maps.Clone(t.fieldEffects), lost: t.lost}
	for _, mon := range t.pokemonParty {
		pr.mons = append(pr.mons, capturePokemonRoot(mon, slots))
	}
	return pr
}

func capturePokemonRoot(mon *pokemon, slots []*slot) pokemonRoot {
	pr := pokemonRoot{
		hp:           mon.hp,
		fainted:      mon.fainted,
		stages:       slices.Clone(mon.stages),
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
		pr.pp = append(pr.pp, move.PP)
	}
	for _, a := range mon.ailments {
		pr.ailments = append(pr.ailments, ailmentRoot{state: a.State, turns: a.Turns, afflictedBy: slices.Index(slots, a.afflictedBy)})
	}
	return pr
}

// apply overwrites bs with the captured state and clears any in-flight turn state.
func (r battleRoot) apply(bs battleState) {
	slots := bs.getAllSlots()
	r.player.apply(bs.getPlayerTrainer(), slots)
	r.opponent.apply(opponentTrainer(bs), slots)

	for i, sr := range r.slots {
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
		maps.Copy(fx, r.fieldEffects)
	}
	bs.setWeather(r.weather)

	actions := bs.getActions()
	actions.queue = actions.queue[:0]
	bs.setError(nil)
}

func (pr partyRoot) apply(t *trainer, slots []*slot) {
	t.lost = pr.lost
	t.fieldEffects = maps.Clone(pr.fieldEffects)
	if t.fieldEffects == nil {
		t.fieldEffects = make(map[fieldEffect]int)
	}
	for i, mr := range pr.mons {
		mr.apply(t.pokemonParty[i], slots)
	}
}

func (pr pokemonRoot) apply(mon *pokemon, slots []*slot) {
	mon.hp = pr.hp
	mon.fainted = pr.fainted
	copy(mon.stages, pr.stages)
	for i, pp := range pr.pp {
		mon.moves[i].PP = pp
	}
	mon.lockedMove = nil
	if pr.lockedMove >= 0 {
		mon.lockedMove = mon.moves[pr.lockedMove]
	}
	mon.item, _ = registerItem(pr.item, mon)
	mon.item.Consumed = pr.itemConsumed
	mon.ability = pr.ability
	mon.unnerved = pr.unnerved
	mon.flashFire = pr.flashFire
	mon.unburden = pr.unburden
	mon.trace = pr.trace
	mon.focusEnergy = pr.focusEnergy
	mon.laserFocus = pr.laserFocus

	clear(mon.ailments)
	for _, ar := range pr.ailments {
		var by *slot
		if ar.afflictedBy >= 0 {
			by = slots[ar.afflictedBy]
		}
		a := generateAilment(ar.state, by)
		a.Turns = ar.turns
		mon.ailments[ar.state] = a
	}
}
