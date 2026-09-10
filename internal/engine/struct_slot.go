package engine

import "fmt"

type slot struct {
	mon                *Pokemon
	Trainer            *Trainer
	firstTurn          bool
	suckerPunch        bool
	protected          bool
	protectTurns       int
	invulnerableAction *moveAction
	unnerved           bool
}

func (s *slot) setMon(bs BattleState, new *Pokemon) {
	s.mon.switchReset()
	s.firstTurn = true
	s.suckerPunch = false

	new.unnerved = s.mon.unnerved
	s.mon = new
	for effect := range s.Trainer.FieldEffects {
		if effect == stealthRockEffect {
			num, dem := s.mon.applyMoveType(1, 8, rockType)
			takeResidualDamage(bs, s, "stealth rock", num, dem)
			continue
		}
		if !s.mon.isGrounded() {
			continue
		}
		switch effect {
		case spikesEffect:
			takeResidualDamage(bs, s, "spikes", 1, 8)
		case toxicSpikesEffect:
			s.mon.applyAilment(poisonAilment, nil, nil)
		case stickyWebEffect:
			s.mon.changeStatStageBy(Speed, -1, true)
		}
	}
}

func (s *slot) isTrapped() bool {
	return s.mon.hasAilment(trapAilment) != nil || s.mon.hasAilment(boundAilment) != nil
}

func (s *slot) resolveProtect() {
	denominator := 1
	for i := 0; i < s.protectTurns; i++ {
		denominator *= 3
	}
	if roll(1, denominator) {
		s.protected = true
		s.protectTurns++
	} else {
		vprintln("but it failed")
	}
}

func (s *slot) hasFieldEffect(effect string) bool {
	if _, ok := s.Trainer.FieldEffects[stringToFieldEffect(effect)]; ok {
		return true
	}
	return false
}

func (s *slot) applyFieldEffect(effect string) error {
	e := stringToFieldEffect(effect)
	if e == noneEffect {
		return fmt.Errorf("%s is not a valid effect", effect)
	}

	turns := 0
	// code here to assign turn number based on which field effect it is
	s.Trainer.FieldEffects[e] = turns

	return nil
}
