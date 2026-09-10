package engine

type itemState int

const (
	noneItem itemState = iota
	berryJuice
	oranBerry
	sitrusBerry
	lumBerry
	leppaBerry
	cheriBerry
	chestoBerry
	pechaBerry
	rawstBerry
	aspearBerry
	persimBerry
	liechiBerry
	ganlonBerry
	salacBerry
	petayaBerry
	apicotBerry
	figyBerry
	iapapaBerry
	wikiBerry
	aguavBerry
	magoBerry
	babiriBerry
	chilanBerry
	chartiBerry
	chopleBerry
	cobaBerry
	colburBerry
	habanBerry
	kasibBerry
	kebiaBerry
	occaBerry
	passhoBerry
	payapaBerry
	rindoBerry
	roseliBerry
	shucaBerry
	tangaBerry
	wacanBerry
	yacheBerry
	ironBall
	scopeLens
	leftovers
	safetyGoggles
	mysticWater
	dragonFang
	silverPowder
	magnet
	blackBelt
	charcoal
	sharpBeak
	spellTag
	miracleSeed
	softSand
	neverMeltIce
	silkScarf
	poisonBarb
	twistedSpoon
	hardStone
	metalCoat
	pixiePlate
	blackGlasses
	normalGem
	fireGem
	fightingGem
	waterGem
	flyingGem
	grassGem
	poisonGem
	electricGem
	groundGem
	psychicGem
	rockGem
	iceGem
	bugGem
	dragonGem
	ghostGem
	darkGem
	steelGem
	fairyGem
	assaultVest
	choiceScarf
	choiceBand
	choiceSpecs
	focusSash
)

var itemStateMap = map[string]itemState{
	"berry juice":    berryJuice,
	"oran berry":     oranBerry,
	"sitrus berry":   sitrusBerry,
	"lum berry":      lumBerry,
	"leppa berry":    leppaBerry,
	"cheri berry":    cheriBerry,
	"chesto berry":   chestoBerry,
	"pecha berry":    pechaBerry,
	"rawst berry":    rawstBerry,
	"aspear berry":   aspearBerry,
	"persim berry":   persimBerry,
	"liechi berry":   liechiBerry,
	"ganlon berry":   ganlonBerry,
	"salac berry":    salacBerry,
	"figy berry":     figyBerry,
	"iapapa berry":   iapapaBerry,
	"wiki berry":     wikiBerry,
	"aguav berry":    aguavBerry,
	"mago berry":     magoBerry,
	"petaya berry":   petayaBerry,
	"apicot berry":   apicotBerry,
	"babiri berry":   babiriBerry,
	"chilan berry":   chilanBerry,
	"charti berry":   chartiBerry,
	"chople berry":   chopleBerry,
	"coba berry":     cobaBerry,
	"colbur berry":   colburBerry,
	"haban berry":    habanBerry,
	"kasib berry":    kasibBerry,
	"kebia berry":    kebiaBerry,
	"occa berry":     occaBerry,
	"passho berry":   passhoBerry,
	"payapa berry":   payapaBerry,
	"rindo berry":    rindoBerry,
	"roseli berry":   roseliBerry,
	"shuca berry":    shucaBerry,
	"tanga berry":    tangaBerry,
	"wacan berry":    wacanBerry,
	"yache berry":    yacheBerry,
	"iron ball":      ironBall,
	"scope lens":     scopeLens,
	"leftovers":      leftovers,
	"safety goggles": safetyGoggles,
	"mystic water":   mysticWater,
	"dragon fang":    dragonFang,
	"silver powder":  silverPowder,
	"magnet":         magnet,
	"black belt":     blackBelt,
	"charcoal":       charcoal,
	"sharp beak":     sharpBeak,
	"spell tag":      spellTag,
	"miracle seed":   miracleSeed,
	"soft sand":      softSand,
	"never melt ice": neverMeltIce,
	"silk scarf":     silkScarf,
	"poison barb":    poisonBarb,
	"twisted spoon":  twistedSpoon,
	"hard stone":     hardStone,
	"metal coat":     metalCoat,
	"pixie plate":    pixiePlate,
	"black glasses":  blackGlasses,
	"normal gem":     normalGem,
	"fire gem":       fireGem,
	"fighting gem":   fightingGem,
	"water gem":      waterGem,
	"flying gem":     flyingGem,
	"grass gem":      grassGem,
	"poison gem":     poisonGem,
	"electric gem":   electricGem,
	"ground gem":     groundGem,
	"psychic gem":    psychicGem,
	"rock gem":       rockGem,
	"ice gem":        iceGem,
	"bug gem":        bugGem,
	"dragon gem":     dragonGem,
	"ghost gem":      ghostGem,
	"dark gem":       darkGem,
	"steel gem":      steelGem,
	"fairy gem":      fairyGem,
	"assault vest":   assaultVest,
	"choice scarf":   choiceScarf,
	"choice band":    choiceBand,
	"choice specs":   choiceSpecs,
	"focus sash":     focusSash,
}

func StringToItemState(s string) itemState {
	if state, ok := itemStateMap[s]; ok {
		return state
	}
	if s != "" {
		elogf("warning: %s is not a valid item and was made into noneItem", s)
	}
	return noneItem
}

func (i itemState) String() string {
	switch i {
	case berryJuice:
		return "berry juice"
	case oranBerry:
		return "oran berry"
	case sitrusBerry:
		return "sitrus berry"
	case lumBerry:
		return "lum berry"
	case leppaBerry:
		return "leppa berry"
	case cheriBerry:
		return "cheri berry"
	case chestoBerry:
		return "chesto berry"
	case pechaBerry:
		return "pecha berry"
	case rawstBerry:
		return "rawst berry"
	case aspearBerry:
		return "aspear berry"
	case persimBerry:
		return "persim berry"
	case liechiBerry:
		return "liechi berry"
	case ganlonBerry:
		return "ganlon berry"
	case salacBerry:
		return "salac berry"
	case petayaBerry:
		return "petaya berry"
	case apicotBerry:
		return "apicot berry"
	case figyBerry:
		return "figy berry"
	case iapapaBerry:
		return "iapapa berry"
	case wikiBerry:
		return "wiki berry"
	case aguavBerry:
		return "aguav berry"
	case magoBerry:
		return "mago berry"
	case babiriBerry:
		return "babiri berry"
	case chilanBerry:
		return "chilan berry"
	case chartiBerry:
		return "charti berry"
	case chopleBerry:
		return "chople berry"
	case cobaBerry:
		return "coba berry"
	case colburBerry:
		return "colbur berry"
	case habanBerry:
		return "haban berry"
	case kasibBerry:
		return "kasib berry"
	case kebiaBerry:
		return "kebia berry"
	case occaBerry:
		return "occa berry"
	case passhoBerry:
		return "passho berry"
	case payapaBerry:
		return "payapa berry"
	case rindoBerry:
		return "rindo berry"
	case roseliBerry:
		return "roseli berry"
	case shucaBerry:
		return "shuca berry"
	case tangaBerry:
		return "tanga berry"
	case wacanBerry:
		return "wacan berry"
	case yacheBerry:
		return "yache berry"
	case ironBall:
		return "iron ball"
	case scopeLens:
		return "scope lens"
	case leftovers:
		return "leftovers"
	case safetyGoggles:
		return "safety goggles"
	case mysticWater:
		return "mystic water"
	case dragonFang:
		return "dragon fang"
	case silverPowder:
		return "silver powder"
	case magnet:
		return "magnet"
	case blackBelt:
		return "black belt"
	case charcoal:
		return "charcoal"
	case sharpBeak:
		return "sharp beak"
	case spellTag:
		return "spell tag"
	case miracleSeed:
		return "miracle seed"
	case softSand:
		return "soft sand"
	case neverMeltIce:
		return "never-melt ice"
	case silkScarf:
		return "silk scarf"
	case poisonBarb:
		return "poison barb"
	case twistedSpoon:
		return "twisted spoon"
	case hardStone:
		return "hard stone"
	case metalCoat:
		return "metal coat"
	case pixiePlate:
		return "pixie plate"
	case blackGlasses:
		return "black glasses"
	case normalGem:
		return "normal gem"
	case fireGem:
		return "fire gem"
	case fightingGem:
		return "fighting gem"
	case waterGem:
		return "water gem"
	case flyingGem:
		return "flying gem"
	case grassGem:
		return "grass gem"
	case poisonGem:
		return "poison gem"
	case electricGem:
		return "electric gem"
	case groundGem:
		return "ground gem"
	case psychicGem:
		return "psychic gem"
	case rockGem:
		return "rock gem"
	case iceGem:
		return "ice gem"
	case bugGem:
		return "bug gem"
	case dragonGem:
		return "dragon gem"
	case ghostGem:
		return "ghost gem"
	case darkGem:
		return "dark gem"
	case steelGem:
		return "steel gem"
	case fairyGem:
		return "fairy gem"
	case assaultVest:
		return "assault vest"
	case choiceScarf:
		return "choice scarf"
	case choiceBand:
		return "choice band"
	case choiceSpecs:
		return "choice specs"
	case focusSash:
		return "focus sash"
	default:
		elogf("warning: itemState %d does not have a string version", i)
		return ""
	}
}

func (is itemState) isBerry() bool {
	return is >= oranBerry && is <= yacheBerry
}

func (is itemState) isChoice() bool {
	return is >= choiceScarf && is <= choiceSpecs
}

type item struct {
	State    itemState
	trigger  func(any) bool
	activate func()
	Consumed bool
}

func (i item) String() string {
	return i.State.String()
}

func (p *Pokemon) CheckItemTrigger(consume bool, event any) {
	if p.Item == nil {
		if p.Base.Name != "" {
			elogf("warning: p.checkItemTrigger: item is nil for %s", p.Base.Name)
		}
		return
	}
	if p.Item.trigger == nil || p.Item.Consumed {
		return
	}

	if p.Item.trigger(event) {
		if consume {
			p.Item.Consumed = true
			p.unburden = true
		}
		p.Item.activate()
	}
}

type ItemFactoryBuilder func(*Pokemon) *item

var itemBuilders = map[itemState]ItemFactoryBuilder{
	berryJuice:    makeBerryJuice,
	oranBerry:     makeOranBerry,
	sitrusBerry:   makeSitrusBerry,
	lumBerry:      makeLumBerry,
	leppaBerry:    makeLeppaBerry,
	cheriBerry:    makeCheriBerry,
	chestoBerry:   makeChestoBerry,
	pechaBerry:    makePechaBerry,
	rawstBerry:    makeRawstBerry,
	aspearBerry:   makeAspearBerry,
	persimBerry:   makePersimBerry,
	liechiBerry:   makeStatBoostBerryMiddleware(liechiBerry, attack),
	ganlonBerry:   makeStatBoostBerryMiddleware(ganlonBerry, defense),
	salacBerry:    makeStatBoostBerryMiddleware(salacBerry, Speed),
	petayaBerry:   makeStatBoostBerryMiddleware(petayaBerry, specialAttack),
	apicotBerry:   makeStatBoostBerryMiddleware(apicotBerry, specialDefense),
	figyBerry:     makePinchHealingBerryMiddleware(figyBerry, attack),
	iapapaBerry:   makePinchHealingBerryMiddleware(iapapaBerry, defense),
	wikiBerry:     makePinchHealingBerryMiddleware(wikiBerry, specialAttack),
	aguavBerry:    makePinchHealingBerryMiddleware(aguavBerry, specialDefense),
	magoBerry:     makePinchHealingBerryMiddleware(magoBerry, Speed),
	babiriBerry:   makeResistBerryMiddleware(babiriBerry, steelType),
	chilanBerry:   makeResistBerryMiddleware(chilanBerry, normalType),
	chartiBerry:   makeResistBerryMiddleware(chartiBerry, rockType),
	chopleBerry:   makeResistBerryMiddleware(chopleBerry, fightingType),
	cobaBerry:     makeResistBerryMiddleware(cobaBerry, flyingType),
	colburBerry:   makeResistBerryMiddleware(colburBerry, darkType),
	habanBerry:    makeResistBerryMiddleware(habanBerry, dragonType),
	kasibBerry:    makeResistBerryMiddleware(kasibBerry, ghostType),
	kebiaBerry:    makeResistBerryMiddleware(kebiaBerry, poisonType),
	occaBerry:     makeResistBerryMiddleware(occaBerry, FireType),
	passhoBerry:   makeResistBerryMiddleware(passhoBerry, waterType),
	payapaBerry:   makeResistBerryMiddleware(payapaBerry, psychicType),
	rindoBerry:    makeResistBerryMiddleware(rindoBerry, grassType),
	roseliBerry:   makeResistBerryMiddleware(roseliBerry, fairyType),
	shucaBerry:    makeResistBerryMiddleware(shucaBerry, groundType),
	tangaBerry:    makeResistBerryMiddleware(tangaBerry, bugType),
	wacanBerry:    makeResistBerryMiddleware(wacanBerry, electricType),
	yacheBerry:    makeResistBerryMiddleware(yacheBerry, iceType),
	ironBall:      makePassiveItemMiddleware(ironBall),
	scopeLens:     makePassiveItemMiddleware(scopeLens),
	leftovers:     makePassiveItemMiddleware(leftovers),
	safetyGoggles: makePassiveItemMiddleware(safetyGoggles),
	mysticWater:   makeTypeBoostingItemMiddleware(mysticWater, waterType),
	dragonFang:    makeTypeBoostingItemMiddleware(dragonFang, dragonType),
	silverPowder:  makeTypeBoostingItemMiddleware(silverPowder, bugType),
	magnet:        makeTypeBoostingItemMiddleware(magnet, electricType),
	blackBelt:     makeTypeBoostingItemMiddleware(blackBelt, fightingType),
	charcoal:      makeTypeBoostingItemMiddleware(charcoal, FireType),
	sharpBeak:     makeTypeBoostingItemMiddleware(sharpBeak, flyingType),
	spellTag:      makeTypeBoostingItemMiddleware(spellTag, ghostType),
	miracleSeed:   makeTypeBoostingItemMiddleware(miracleSeed, grassType),
	softSand:      makeTypeBoostingItemMiddleware(softSand, groundType),
	neverMeltIce:  makeTypeBoostingItemMiddleware(neverMeltIce, iceType),
	silkScarf:     makeTypeBoostingItemMiddleware(silkScarf, normalType),
	poisonBarb:    makeTypeBoostingItemMiddleware(poisonBarb, poisonType),
	twistedSpoon:  makeTypeBoostingItemMiddleware(twistedSpoon, psychicType),
	hardStone:     makeTypeBoostingItemMiddleware(hardStone, rockType),
	metalCoat:     makeTypeBoostingItemMiddleware(metalCoat, steelType),
	pixiePlate:    makeTypeBoostingItemMiddleware(pixiePlate, fairyType),
	blackGlasses:  makeTypeBoostingItemMiddleware(blackGlasses, darkType),
	normalGem:     makeGemMiddleware(normalGem, normalType),
	fireGem:       makeGemMiddleware(fireGem, FireType),
	fightingGem:   makeGemMiddleware(fightingGem, fightingType),
	waterGem:      makeGemMiddleware(waterGem, waterType),
	flyingGem:     makeGemMiddleware(flyingGem, flyingType),
	grassGem:      makeGemMiddleware(grassGem, grassType),
	poisonGem:     makeGemMiddleware(poisonGem, poisonType),
	electricGem:   makeGemMiddleware(electricGem, electricType),
	groundGem:     makeGemMiddleware(groundGem, groundType),
	psychicGem:    makeGemMiddleware(psychicGem, psychicType),
	rockGem:       makeGemMiddleware(rockGem, rockType),
	iceGem:        makeGemMiddleware(iceGem, iceType),
	bugGem:        makeGemMiddleware(bugGem, bugType),
	dragonGem:     makeGemMiddleware(dragonGem, dragonType),
	ghostGem:      makeGemMiddleware(ghostGem, ghostType),
	darkGem:       makeGemMiddleware(darkGem, darkType),
	steelGem:      makeGemMiddleware(steelGem, steelType),
	fairyGem:      makeGemMiddleware(fairyGem, fairyType),
	assaultVest:   makeAssaultVest,
	choiceScarf:   makeChoiceScarf,
	choiceBand:    makeChoiceBand,
	choiceSpecs:   makeChoiceSpecs,
	focusSash:     makeFocusSash,
}

func createItemFactory(builder ItemFactoryBuilder, mon *Pokemon) func() *item {
	return func() *item {
		return builder(mon)
	}
}

func RegisterItem(is itemState, mon *Pokemon) (*item, error) {
	if is == noneItem {
		return &item{
			Consumed: true,
		}, nil
	}

	builder := itemBuilders[is]
	factory := createItemFactory(builder, mon)

	return factory(), nil
}

func makePassiveItemMiddleware(is itemState) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		return &item{
			State: is,
			trigger: func(e any) bool {
				return false
			},
		}
	}
}

func makeTypeBoostingItemMiddleware(is itemState, pokemonType pokemonType) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		var p *int
		return &item{
			State: is,
			trigger: func(e any) bool {
				event, ok := e.(moveBoostingEvent)
				if !ok {
					return false
				}
				if pokemonType != event.pokemonType {
					return false
				}
				p = event.power
				return true
			},
			activate: func() {
				*p = *p * 6 / 5
			},
		}
	}
}

func makeBerryJuice(mon *Pokemon) *item {
	return &item{
		State: berryJuice,
		trigger: func(any) bool {
			return mon.HP > 0 && mon.HP*2 <= mon.MaxHP()
		},
		activate: func() {
			mon.ChangeHpBy(20)
			vprintItem("%s drank its berry juice and restored 20 hp", mon.Base.Name)
			cheekPouch(mon)
		},
	}
}

func makeOranBerry(mon *Pokemon) *item {
	return &item{
		State: oranBerry,
		trigger: func(any) bool {
			return mon.HP > 0 && !mon.unnerved && mon.HP*2 <= mon.MaxHP()
		},
		activate: func() {
			mon.ChangeHpBy(10)
			vprintItem("%s ate its oran berry and restored 10 hp", mon.Base.Name)
			cheekPouch(mon)
		},
	}
}

func makeSitrusBerry(mon *Pokemon) *item {
	return &item{
		State: sitrusBerry,
		trigger: func(any) bool {
			return mon.HP > 0 && !mon.unnerved && mon.HP*2 <= mon.MaxHP()
		},
		activate: func() {
			restore := mon.MaxHP() / 4
			mon.ChangeHpBy(restore)
			vprintItem("%s ate its sitrus berry and restored %d hp", mon.Base.Name, restore)
			cheekPouch(mon)
		},
	}
}

func makeCheriBerry(mon *Pokemon) *item {
	return &item{
		State: cheriBerry,
		trigger: func(any) bool {
			return !mon.unnerved && mon.hasAilment(paralysisAilment) != nil
		},
		activate: func() {
			vprintItem("%s ate its cheri berry", mon.Base.Name)
			delete(mon.Ailments, paralysisAilment)
			cheekPouch(mon)
		},
	}
}

func makeChestoBerry(mon *Pokemon) *item {
	return &item{
		State: chestoBerry,
		trigger: func(any) bool {
			return !mon.unnerved && mon.hasAilment(sleepAilment) != nil
		},
		activate: func() {
			vprintItem("%s ate its chesto berry", mon.Base.Name)
			delete(mon.Ailments, sleepAilment)
			cheekPouch(mon)
		},
	}
}

func makePechaBerry(mon *Pokemon) *item {
	return &item{
		State: pechaBerry,
		trigger: func(any) bool {
			return !mon.unnerved && (mon.hasAilment(poisonAilment) != nil || mon.hasAilment(toxicAilment) != nil)
		},
		activate: func() {
			vprintItem("%s ate its pecha berry", mon.Base.Name)
			delete(mon.Ailments, poisonAilment)
			delete(mon.Ailments, toxicAilment)
			cheekPouch(mon)
		},
	}
}

func makeRawstBerry(mon *Pokemon) *item {
	return &item{
		State: rawstBerry,
		trigger: func(any) bool {
			return !mon.unnerved && mon.hasAilment(burnAilment) != nil
		},
		activate: func() {
			vprintItem("%s ate its rawst berry", mon.Base.Name)
			delete(mon.Ailments, burnAilment)
			cheekPouch(mon)
		},
	}
}

func makeAspearBerry(mon *Pokemon) *item {
	return &item{
		State: aspearBerry,
		trigger: func(any) bool {
			return !mon.unnerved && mon.hasAilment(freezeAilment) != nil
		},
		activate: func() {
			vprintItem("%s ate its aspear berry", mon.Base.Name)
			delete(mon.Ailments, freezeAilment)
			cheekPouch(mon)
		},
	}
}

func makePersimBerry(mon *Pokemon) *item {
	return &item{
		State: persimBerry,
		trigger: func(any) bool {
			return !mon.unnerved && mon.hasAilment(confusionAilment) != nil
		},
		activate: func() {
			vprintItem("%s ate its persim berry", mon.Base.Name)
			delete(mon.Ailments, confusionAilment)
			cheekPouch(mon)
		},
	}
}

func makeLumBerry(mon *Pokemon) *item {
	return &item{
		State: lumBerry,
		trigger: func(any) bool {
			return !mon.unnerved && (mon.hasNonVolatileAilment() || mon.hasAilment(confusionAilment) != nil)
		},
		activate: func() {
			vprintItem("%s ate its lum berry", mon.Base.Name)
			for ailment := range nonVolatileStatuses {
				if mon.hasAilment(ailment) != nil {
					delete(mon.Ailments, ailment)
					vprintItem("%s had its %s removed", mon.Base.Name, ailment.String())
					break
				}
			}
			if mon.hasAilment(confusionAilment) != nil {
				delete(mon.Ailments, confusionAilment)
				vprintItem("%s had its confusion removed", mon.Base.Name)
				cheekPouch(mon)
			}
		},
	}
}

func makeLeppaBerry(mon *Pokemon) *item {
	var m *Move
	return &item{
		trigger: func(e any) bool {
			event, ok := e.(leppaBerryEvent)
			if !ok || mon.unnerved {
				return false
			}
			m = event.move
			return event.move.PP <= 0
		},
		activate: func() {
			m.PP += min(10, m.MaxPP)
			cheekPouch(mon)
		},
	}
}

func makeStatBoostBerryMiddleware(is itemState, stat statState) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		return &item{
			State: is,
			trigger: func(any) bool {
				if mon.unnerved {
					return false
				}
				if mon.HP <= 0 {
					return false
				}
				if mon.Ability == GluttonyAbility {
					return mon.HP*2 <= mon.MaxHP()
				}
				return mon.HP*4 <= mon.MaxHP()
			},
			activate: func() {
				vprintItem("%s ate its %s", mon.Base.Name, is)
				mon.changeStatStageBy(stat, 1, false)
				cheekPouch(mon)
			},
		}
	}
}

func makePinchHealingBerryMiddleware(is itemState, stat statState) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		return &item{
			State: is,
			trigger: func(any) bool {
				if mon.unnerved {
					return false
				}
				if mon.HP <= 0 {
					return false
				}
				if mon.Ability == GluttonyAbility {
					return mon.HP*2 <= mon.MaxHP()
				}
				return mon.HP*4 <= mon.MaxHP()
			},
			activate: func() {
				restore := mon.MaxHP() / 2
				mon.ChangeHpBy(restore)
				vprintItem("%s ate its %s and restored %d hp", mon.Base.Name, is.String(), restore)
				cheekPouch(mon)
				if mon.nat.negative == stat {
					mon.applyAilment(confusionAilment, nil, nil)
				}
			},
		}
	}
}

func makeResistBerryMiddleware(is itemState, pokemonType pokemonType) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		var d *int
		return &item{
			State: is,
			trigger: func(e any) bool {
				event, ok := e.(resistBerryEvent)
				if !ok || mon.unnerved {
					return false
				}
				d = event.damage
				return event.pokemonType == pokemonType
			},
			activate: func() {
				if d == nil {
					vprintItem("%s ate its %s and reduced the damage", mon.Base.Name, is)
					cheekPouch(mon)
				} else {
					*d /= 2
				}
			},
		}
	}
}

func makeGemMiddleware(is itemState, pokemonType pokemonType) func(mon *Pokemon) *item {
	return func(mon *Pokemon) *item {
		var p *int
		return &item{
			State: is,
			trigger: func(e any) bool {
				event, ok := e.(gemEvent)
				if !ok {
					return false
				}
				p = event.power
				return event.pokemonType == pokemonType
			},
			activate: func() {
				if p == nil {
					vprintItem("%s consumed its %s gem and boosted the damage", mon.Base.Name, pokemonType.String())
				} else {
					*p = *p * 3 / 2
				}
			},
		}
	}
}

func makeAssaultVest(mon *Pokemon) *item {
	var s *int
	return &item{
		State: assaultVest,
		trigger: func(e any) bool {
			event, ok := e.(choiceItemEvent)
			if !ok {
				return false
			}
			if event.statState != specialDefense {
				return false
			}
			s = event.stat
			return true
		},
		activate: func() {
			*s = *s * 3 / 2
		},
	}
}

func makeChoiceScarf(mon *Pokemon) *item {
	var s *int
	return &item{
		State: choiceScarf,
		trigger: func(e any) bool {
			event, ok := e.(choiceItemEvent)
			if !ok {
				return false
			}
			if event.statState != Speed {
				return false
			}
			s = event.stat
			return true
		},
		activate: func() {
			*s = *s * 3 / 2
		},
	}
}

func makeChoiceBand(mon *Pokemon) *item {
	var s *int
	return &item{
		State: choiceBand,
		trigger: func(e any) bool {
			event, ok := e.(choiceItemEvent)
			if !ok {
				return false
			}
			if event.move.Class != physicalClass {
				return false
			}
			s = event.stat
			return true
		},
		activate: func() {
			*s = *s * 3 / 2
		},
	}
}

func makeChoiceSpecs(mon *Pokemon) *item {
	var s *int
	return &item{
		State: choiceSpecs,
		trigger: func(e any) bool {
			event, ok := e.(choiceItemEvent)
			if !ok {
				return false
			}
			if event.move.Class != SpecialClass {
				return false
			}
			s = event.stat
			return true
		},
		activate: func() {
			*s = *s * 3 / 2
		},
	}
}

func makeFocusSash(mon *Pokemon) *item {
	var dmg *int
	return &item{
		trigger: func(e any) bool {
			event, ok := e.(focusSashEvent)
			if !ok {
				return false
			}
			dmg = event.damage
			return mon.HP == mon.MaxHP() && *event.damage >= mon.HP
		},
		activate: func() {
			*dmg = mon.HP - 1
		},
	}
}

type resistBerryEvent struct {
	pokemonType pokemonType
	damage      *int
}

func makeResistBerryEvent(pokemonType pokemonType, damage *int) resistBerryEvent {
	return resistBerryEvent{
		pokemonType: pokemonType,
		damage:      damage,
	}
}

type gemEvent struct {
	pokemonType pokemonType
	power       *int
}

func makeGemEvent(pokemonType pokemonType, power *int) gemEvent {
	return gemEvent{
		pokemonType: pokemonType,
		power:       power,
	}
}

type leppaBerryEvent struct {
	move *Move
}

func makeLeppaBerryEvent(move *Move) leppaBerryEvent {
	return leppaBerryEvent{
		move: move,
	}
}

type choiceItemEvent struct {
	move      *Move
	statState statState
	stat      *int
}

func makeChoiceItemEvent(move *Move, statState statState, stat *int) choiceItemEvent {
	res := choiceItemEvent{
		statState: statState,
		stat:      stat,
	}
	if move == nil {
		res.move = &Move{}
	} else {
		res.move = move
	}

	return res
}

type focusSashEvent struct {
	damage *int
}

func makeFocusSashEvent(damage *int) focusSashEvent {
	return focusSashEvent{
		damage: damage,
	}
}

type moveBoostingEvent struct {
	pokemonType pokemonType
	power       *int
}

func makeMoveBoostingEvent(pokemonType pokemonType, power *int) moveBoostingEvent {
	return moveBoostingEvent{
		pokemonType: pokemonType,
		power:       power,
	}
}
