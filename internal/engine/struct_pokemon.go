package engine

import (
	"fmt"
	"slices"
)

type Pokemon struct {
	Base        BasePokemon
	level       int
	IVs         []int
	nat         nature
	Moves       []*Move
	LockedMove  *Move
	Stats       []int
	Stages      []int
	HP          int
	fainted     bool
	Ailments    map[ailmentState]*ailment
	Item        *item
	Ability     abilityState
	unnerved    bool
	flashFire   bool
	unburden    bool
	trace       bool
	focusEnergy bool
	laserFocus  bool
}

func getNature(nat string) (nature, error) {
	res, ok := natureChart[nat]
	if !ok {
		return nature{}, fmt.Errorf("invalid nature: %s", nat)
	}

	return res, nil

}

func InitPokemon(base BasePokemon, level int, ivs map[string]int, nature string, moves []*Move, hp int, status ailmentState) (Pokemon, error) {
	if level < 1 || level > 100 {
		return Pokemon{}, fmt.Errorf("invalid level: %d", level)
	}

	nat, err := getNature(nature)
	if err != nil {
		return Pokemon{}, err
	}

	res := Pokemon{
		Base:     base,
		level:    level,
		IVs:      []int{31, 31, 31, 31, 31, 31},
		nat:      nat,
		Moves:    moves,
		Stats:    []int{0, 0, 0, 0, 0, 0},
		Stages:   []int{0, 0, 0, 0, 0, 0, 0, 0},
		HP:       0,
		fainted:  false,
		Ailments: make(map[ailmentState]*ailment),
	}

	err = setIVs(&res, ivs)
	if err != nil {
		return Pokemon{}, err
	}

	err = calculateStats(&res)
	if err != nil {
		return Pokemon{}, err
	}

	if hp == -1 {
		hp = res.MaxHP()
	}

	res.HP = max(1, min(res.MaxHP(), hp))

	if status.isNonVolatileStatus() {
		res.Ailments[status] = GenerateAilment(status, nil)
	}

	return res, nil
}

func setIVs(Pokemon *Pokemon, ivs map[string]int) error {
	for key, val := range ivs {
		stat := stringToStat(key)
		if stat == noStat {
			return fmt.Errorf("%s is not a valid stat for %s", key, Pokemon.Base.Name)
		}
		Pokemon.IVs[stat] = max(0, min(31, val))
	}

	return nil
}

func calculateStats(Pokemon *Pokemon) error {
	for key, val := range Pokemon.Base.Stats {
		stat := stringToStat(key)
		if stat == noStat {
			return fmt.Errorf("%s is not a valid stat for %s", key, Pokemon.Base.Name)
		}
		Pokemon.Stats[stat] = ((val*2+Pokemon.IVs[stat])*Pokemon.level)/100 + 5
	}
	// Shedinja case: if HP is 1, it stays 1 regardless of level or IVs
	if Pokemon.Stats[hitPoints] == 1 {
		Pokemon.Stats[hitPoints] = 1
	} else {
		Pokemon.Stats[hitPoints] += Pokemon.level + 5
	}

	// Apply nature modifiers
	posNat := Pokemon.nat.positive
	negNat := Pokemon.nat.negative

	if posNat != negNat {
		Pokemon.Stats[posNat] = (Pokemon.Stats[posNat] * 110) / 100
		Pokemon.Stats[negNat] = (Pokemon.Stats[negNat] * 90) / 100
	}

	return nil
}

func (p *Pokemon) switchReset() {
	for a := range volatileStatuses {
		delete(p.Ailments, a)
	}

	for stat := range p.Stages {
		p.Stages[stat] = 0
	}

	if toxic, ok := p.Ailments[toxicAilment]; ok {
		toxic.Turns = 0
	}

	if p.trace {
		p.trace = false
		p.Ability = traceAbility
	}

	p.LockedMove = nil
	p.flashFire = false
	p.unburden = false
	p.focusEnergy = false
	p.laserFocus = false
}

func (p *Pokemon) resetMovePPs() {
	if p == nil {
		return
	}
	for _, move := range p.Moves {
		if move == nil {
			continue
		}
		move.PP = move.MaxPP
	}
	if p.LockedMove != nil {
		p.LockedMove.PP = p.LockedMove.MaxPP
	}
	p.LockedMove = nil
}

func resetPokemonPartyPPs(party []*Pokemon) {
	for _, mon := range party {
		if mon != nil {
			mon.resetMovePPs()
		}
	}
}

func (p *Pokemon) effectiveStat(stat statState, crit bool) int {
	stage := p.Stages[stat]
	base := p.Stats[stat]
	p.CheckItemTrigger(false, makeChoiceItemEvent(nil, stat, &base))

	if crit {
		switch stat {
		case defense, specialDefense:
			stage = min(0, stage)
		case attack, specialAttack:
			stage = max(0, stage)
		}
	}

	if stage >= 0 {
		return base * (2 + stage) / 2
	}
	return base * 2 / (2 - stage)
}

func (p *Pokemon) effectiveSpeed(bs BattleState) int {
	stage := p.Stages[Speed]
	base := p.Stats[Speed]
	p.CheckItemTrigger(false, makeChoiceItemEvent(nil, Speed, &base))
	numerator := 1
	denominator := 1

	if p.Item.State == ironBall {
		denominator *= 2
	} else if p.unburden && p.Ability == unburdenAbility {
		numerator *= 2
	}
	if _, ok := p.Ailments[paralysisAilment]; ok {
		denominator *= 4
	}
	switch bs.getWeather() {
	case rainWeather:
		if p.Ability == swiftSwimAbility {
			numerator *= 2
		}
	case sunWeather:
		if p.Ability == chlorophyllAbility {
			numerator *= 2
		}
	case hailWeather:
		if p.Ability == slushRushAbility {
			numerator *= 2
		}
	case sandstormWeather:
		if p.Ability == sandRushAbility {
			numerator *= 2
		}
	}

	base = base * numerator / denominator

	if stage >= 0 {
		return base * (2 + stage) / 2
	}
	return base * 2 / (2 - stage)
}

func (p *Pokemon) isFasterThan(bs BattleState, mon *Pokemon) bool {
	return p.effectiveSpeed(bs) >= mon.effectiveSpeed(bs)
}

func (p *Pokemon) evasionFraction(keenEye bool) (int, int) {
	if keenEye {
		return 1, 1
	}

	stage := p.Stages[evasion]
	if stage == 0 {
		return 3, 3
	} else if stage > 0 {
		return 3, 3 + stage
	}
	return 3 - stage, 3
}

func (p *Pokemon) accuracyFraction() (int, int) {
	stage := p.Stages[accuracy]
	if stage == 0 {
		return 3, 3
	} else if stage > 0 {
		return 3 + stage, 3
	}
	return 3, 3 - stage
}

func (p *Pokemon) hasType(pokemonType pokemonType) bool {
	return slices.Contains(p.Base.Types, pokemonType)
}

func (p *Pokemon) applyAilment(ailment ailmentState, move *Move, afflictedBy *slot) bool {
	if ailment == noneAilment {
		elogf("warning: %s applies an ailment but is none", ailment.String())
		return false
	}

	if _, ok := p.Ailments[ailment]; ok {
		return false
	}
	if ailment.isNonVolatileStatus() && p.hasNonVolatileAilment() {
		return false
	}

	switch ailment {
	case burnAilment:
		if p.hasType(FireType) || p.Ability == waterVeilAbility {
			return false
		}
	case paralysisAilment:
		if p.hasType(electricType) || p.Ability == limberAbility {
			return false
		}
	case poisonAilment, toxicAilment:
		if p.Ability == immunityAbility {
			return false
		}
		if (p.hasType(poisonType) || p.hasType(steelType)) && (afflictedBy == nil || afflictedBy.mon.Ability != corrosionAbility) {
			return false
		}
	case freezeAilment:
		if p.hasType(iceType) || p.Ability == magmaArmorAbility {
			return false
		}
	case sleepAilment, yawnAilment:
		if p.Ability.blocksSleep() || p.hasNonVolatileAilment() {
			return false
		}
	case trapAilment:
		p.Ailments[ailment] = generateTrap(move.MinTurns, move.MaxTurns, afflictedBy)
		return true
	case infatuationAilment:
		if p.Ability == obliviousAbility {
			return false
		}
	}

	if ailment == poisonAilment {
		if move != nil && (move.Name == "toxic" || move.Name == "poison fang") {
			ailment = toxicAilment
		}
	}

	p.Ailments[ailment] = GenerateAilment(ailment, afflictedBy)
	vprintf("%s became afflicted with %s", p.Base.Name, ailment.String())
	if ailment.isNonVolatileStatus() && p.Ability == synchronizeAbility {
		afflictedBy.mon.applyAilment(ailment, nil, nil)
	}
	p.CheckItemTrigger(true, nil)

	return true
}

func (p *Pokemon) hasAilment(ailment ailmentState) *ailment {
	if a, ok := p.Ailments[ailment]; ok {
		return a
	}
	return nil
}

func (p *Pokemon) hasNonVolatileAilment() bool {
	for ailment := range p.Ailments {
		if ailment <= sleepAilment {
			return true
		}
	}
	return false
}

func (p *Pokemon) isGrounded() bool {
	if p.Item.State == ironBall {
		return true
	}
	if p.hasType(flyingType) || p.Ability == levitateAbility {
		return false
	}
	return true
}

func (p *Pokemon) ChangeHpBy(change int) {
	p.HP = min(p.HP+change, p.MaxHP())
	p.CheckItemTrigger(true, nil)
}

func (p *Pokemon) hasMovePredicate(f func(*Move) bool) bool {
	return slices.ContainsFunc(p.Moves, f)
}

func (p *Pokemon) changeStatStageBy(stat statState, change int, offensive bool) {
	if offensive && (p.Ability == clearBodyAbility || p.Ability == clearSmokeAbility) {
		vprintf("blocked by clear body")
		return
	}
	if p.Ability == keenEyeAbility && stat == accuracy && change < 0 {
		return
	}

	p.Stages[stat] = max(-6, min(6, p.Stages[stat]+change))
	vprintf("%s's %s changed by %d stages (%d)", p.Base.Name, stat, change, p.Stages[stat])
}

func (p *Pokemon) MaxHP() int {
	return p.Stats[hitPoints]
}

func (p *Pokemon) serenceGraceBonus() int {
	if p.Ability == serenceGraceAbility {
		return 2
	}
	return 1
}

func (p *Pokemon) applyMoveType(num, dem int, moveType pokemonType) (int, int) {
	for _, t := range p.Base.Types {
		if t == flyingType && moveType == groundType && p.isGrounded() {
			continue
		}
		if p.Ability == levitateAbility && moveType == groundType && !p.isGrounded() {
			num = 0
			continue
		}

		switch getEffectiveness(moveType, t) {
		case immuneEffectivensss:
			num = 0
		case resistedEffectiveness:
			dem *= 2
		case superEffectiveness:
			num *= 2
		}
	}

	return num, dem
}

func (p *Pokemon) isImmuneToPowderMoves() bool {
	return p.hasType(grassType) || p.Ability == overcoatAbility
}
