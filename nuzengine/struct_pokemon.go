package nuzengine

import (
	"fmt"
	"slices"
	"strings"

	"github.com/fred-bonn/nuz/nuzengine/internal/parser"
)

type pokemon struct {
	base           basePokemon
	level          int
	ivs            []int
	nat            nature
	moves          []*Move
	lockedMove     *Move
	stats          []int
	stages         []int
	hp             int
	fainted        bool
	ailments       map[ailmentState]*ailment
	item           *item
	ability        abilityState
	unnerved       bool
	flashFire      bool
	unburden       bool
	trace          bool
	focusEnergy    bool
	laserFocus     bool
	initialHp      int
	initialAilment ailmentState
	initialItem    itemState
}

func initPokemon(baseMon basePokemon, moves []*Move, parsedMon parser.ParsedPokemon) (pokemon, error) {
	if parsedMon.Level < 1 || parsedMon.Level > 100 {
		return pokemon{}, fmt.Errorf("invalid level: %d", parsedMon.Level)
	}

	nat, err := getNature(parsedMon.Nature)
	if err != nil {
		return pokemon{}, err
	}

	res := pokemon{
		base:     baseMon,
		level:    parsedMon.Level,
		ivs:      []int{31, 31, 31, 31, 31, 31},
		nat:      nat,
		moves:    moves,
		stats:    []int{0, 0, 0, 0, 0, 0},
		stages:   []int{0, 0, 0, 0, 0, 0, 0, 0},
		hp:       0,
		fainted:  false,
		ailments: make(map[ailmentState]*ailment),
	}

	err = setIVs(&res, parsedMon.IVs)
	if err != nil {
		return pokemon{}, err
	}

	err = calculateStats(&res)
	if err != nil {
		return pokemon{}, err
	}

	if parsedMon.HP == -1 {
		parsedMon.HP = res.MaxHP()
	}

	res.hp = max(1, min(res.MaxHP(), parsedMon.HP))
	res.initialHp = res.hp

	res.initialAilment = noneAilment
	status := stringToAilmentState(parsedMon.Status)
	if status.isNonVolatileStatus() {
		res.ailments[status] = generateAilment(status, nil)
		res.initialAilment = status
	}

	item, err := registerItem(stringToItemState(strings.ToLower(parsedMon.Item)), &res)
	if err != nil {
		return pokemon{}, err
	}
	res.item = item
	res.initialItem = item.State

	res.ability = stringToAbility(strings.ToLower(parsedMon.Ability))
	if res.ability == NoneAbility {
		return pokemon{}, fmt.Errorf("none is not a valid ability")
	}

	return res, nil
}

func setIVs(Pokemon *pokemon, ivs map[string]int) error {
	for key, val := range ivs {
		stat := stringToStat(key)
		if stat == noStat {
			return fmt.Errorf("no stat is not a valid stat")
		}
		Pokemon.ivs[stat] = max(0, min(31, val))
	}

	return nil
}

func calculateStats(Pokemon *pokemon) error {
	for key, val := range Pokemon.base.Stats {
		stat := stringToStat(key)
		if stat == noStat {
			return fmt.Errorf("no stat is not a valid stat")
		}
		Pokemon.stats[stat] = ((val*2+Pokemon.ivs[stat])*Pokemon.level)/100 + 5
	}
	// Shedinja case: if HP is 1, it stays 1 regardless of level or IVs
	if Pokemon.stats[hitPoints] == 1 {
		Pokemon.stats[hitPoints] = 1
	} else {
		Pokemon.stats[hitPoints] += Pokemon.level + 5
	}

	// Apply nature modifiers
	posNat := Pokemon.nat.positive
	negNat := Pokemon.nat.negative

	if posNat != negNat {
		Pokemon.stats[posNat] = (Pokemon.stats[posNat] * 110) / 100
		Pokemon.stats[negNat] = (Pokemon.stats[negNat] * 90) / 100
	}

	return nil
}

func (p *pokemon) reset() {
	for _, move := range p.moves {
		move.PP = move.MaxPP
	}
	p.hp = p.initialHp
	for ailment := range p.ailments {
		delete(p.ailments, ailment)
	}
	if p.initialAilment != noneAilment {
		p.ailments[p.initialAilment] = generateAilment(p.initialAilment, nil)
	}
	p.item, _ = registerItem(p.initialItem, p)
	p.fainted = false
}

func (p *pokemon) switchReset() {
	for a := range volatileStatuses {
		delete(p.ailments, a)
	}

	for stat := range p.stages {
		p.stages[stat] = 0
	}

	if toxic, ok := p.ailments[toxicAilment]; ok {
		toxic.Turns = 0
	}

	if p.trace {
		p.trace = false
		p.ability = traceAbility
	}

	p.lockedMove = nil
	p.flashFire = false
	p.unburden = false
	p.focusEnergy = false
	p.laserFocus = false
}

func (p *pokemon) effectiveStat(stat statState, crit bool) int {
	stage := p.stages[stat]
	base := p.stats[stat]
	p.checkItemTrigger(false, makeChoiceItemEvent(nil, stat, &base))

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

func (p *pokemon) effectiveSpeed(bs battleState) int {
	stage := p.stages[Speed]
	base := p.stats[Speed]
	p.checkItemTrigger(false, makeChoiceItemEvent(nil, Speed, &base))
	numerator := 1
	denominator := 1

	if p.item.State == ironBall {
		denominator *= 2
	} else if p.unburden && p.ability == unburdenAbility {
		numerator *= 2
	}
	if _, ok := p.ailments[paralysisAilment]; ok {
		denominator *= 4
	}
	switch bs.getWeather() {
	case rainWeather:
		if p.ability == swiftSwimAbility {
			numerator *= 2
		}
	case sunWeather:
		if p.ability == chlorophyllAbility {
			numerator *= 2
		}
	case hailWeather:
		if p.ability == slushRushAbility {
			numerator *= 2
		}
	case sandstormWeather:
		if p.ability == sandRushAbility {
			numerator *= 2
		}
	}

	base = base * numerator / denominator

	if stage >= 0 {
		return base * (2 + stage) / 2
	}
	return base * 2 / (2 - stage)
}

func (p *pokemon) isFasterThan(bs battleState, mon *pokemon) bool {
	return p.effectiveSpeed(bs) >= mon.effectiveSpeed(bs)
}

func (p *pokemon) evasionFraction(keenEye bool) (int, int) {
	if keenEye {
		return 1, 1
	}

	stage := p.stages[evasion]
	if stage == 0 {
		return 3, 3
	} else if stage > 0 {
		return 3, 3 + stage
	}
	return 3 - stage, 3
}

func (p *pokemon) accuracyFraction() (int, int) {
	stage := p.stages[accuracy]
	if stage == 0 {
		return 3, 3
	} else if stage > 0 {
		return 3 + stage, 3
	}
	return 3, 3 - stage
}

func (p *pokemon) hasType(pokemonType pokemonType) bool {
	return slices.Contains(p.base.Types, pokemonType)
}

func (p *pokemon) applyAilment(ailment ailmentState, move *Move, afflictedBy *slot) bool {
	if ailment == noneAilment {
		elogf("warning: %s applies an ailment but is none", ailment.String())
		return false
	}

	if _, ok := p.ailments[ailment]; ok {
		return false
	}
	if ailment.isNonVolatileStatus() && p.hasNonVolatileAilment() {
		return false
	}

	switch ailment {
	case burnAilment:
		if p.hasType(fireType) || p.ability == waterVeilAbility {
			return false
		}
	case paralysisAilment:
		if p.hasType(electricType) || p.ability == limberAbility {
			return false
		}
	case poisonAilment, toxicAilment:
		if p.ability == immunityAbility {
			return false
		}
		if (p.hasType(poisonType) || p.hasType(steelType)) && (afflictedBy == nil || afflictedBy.mon.ability != corrosionAbility) {
			return false
		}
	case freezeAilment:
		if p.hasType(iceType) || p.ability == magmaArmorAbility {
			return false
		}
	case sleepAilment, yawnAilment:
		if p.ability.blocksSleep() || p.hasNonVolatileAilment() {
			return false
		}
	case trapAilment:
		p.ailments[ailment] = generateTrap(move.MinTurns, move.MaxTurns, afflictedBy)
		return true
	case infatuationAilment:
		if p.ability == obliviousAbility {
			return false
		}
	}

	if ailment == poisonAilment {
		if move != nil && (move.Name == "toxic" || move.Name == "poison fang") {
			ailment = toxicAilment
		}
	}

	p.ailments[ailment] = generateAilment(ailment, afflictedBy)
	vprintf("%s became afflicted with %s", p.base.Name, ailment.String())
	if ailment.isNonVolatileStatus() && p.ability == synchronizeAbility {
		afflictedBy.mon.applyAilment(ailment, nil, nil)
	}
	p.checkItemTrigger(true, nil)

	return true
}

func (p *pokemon) hasAilment(ailment ailmentState) *ailment {
	if a, ok := p.ailments[ailment]; ok {
		return a
	}
	return nil
}

func (p *pokemon) hasNonVolatileAilment() bool {
	for ailment := range p.ailments {
		if ailment <= sleepAilment {
			return true
		}
	}
	return false
}

func (p *pokemon) isGrounded() bool {
	if p.item.State == ironBall {
		return true
	}
	if p.hasType(flyingType) || p.ability == levitateAbility {
		return false
	}
	return true
}

func (p *pokemon) ChangeHpBy(change int) {
	p.hp = min(p.hp+change, p.MaxHP())
	p.checkItemTrigger(true, nil)
}

func (p *pokemon) hasMovePredicate(f func(*Move) bool) bool {
	return slices.ContainsFunc(p.moves, f)
}

func (p *pokemon) changeStatStageBy(stat statState, change int, offensive bool) {
	if offensive && (p.ability == clearBodyAbility || p.ability == clearSmokeAbility) {
		vprintf("blocked by clear body")
		return
	}
	if p.ability == keenEyeAbility && stat == accuracy && change < 0 {
		return
	}

	p.stages[stat] = max(-6, min(6, p.stages[stat]+change))
	vprintf("%s's %s changed by %d stages (%d)", p.base.Name, stat, change, p.stages[stat])
}

func (p *pokemon) MaxHP() int {
	return p.stats[hitPoints]
}

func (p *pokemon) serenceGraceBonus() int {
	if p.ability == serenceGraceAbility {
		return 2
	}
	return 1
}

func (p *pokemon) applyMoveType(num, dem int, moveType pokemonType) (int, int) {
	for _, t := range p.base.Types {
		if t == flyingType && moveType == groundType && p.isGrounded() {
			continue
		}
		if p.ability == levitateAbility && moveType == groundType && !p.isGrounded() {
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

func (p *pokemon) isImmuneToPowderMoves() bool {
	return p.hasType(grassType) || p.ability == overcoatAbility
}
