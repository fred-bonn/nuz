package engine

type pokemonType int

const (
	NoType pokemonType = iota
	normalType
	FireType
	waterType
	electricType
	grassType
	iceType
	fightingType
	poisonType
	groundType
	flyingType
	psychicType
	bugType
	rockType
	ghostType
	dragonType
	darkType
	steelType
	fairyType
)

func StringToPokemonType(s string) pokemonType {
	switch s {
	case "normal":
		return normalType
	case "fire":
		return FireType
	case "water":
		return waterType
	case "electric":
		return electricType
	case "grass":
		return grassType
	case "ice":
		return iceType
	case "fighting":
		return fightingType
	case "poison":
		return poisonType
	case "ground":
		return groundType
	case "flying":
		return flyingType
	case "psychic":
		return psychicType
	case "bug":
		return bugType
	case "rock":
		return rockType
	case "ghost":
		return ghostType
	case "dragon":
		return dragonType
	case "dark":
		return darkType
	case "steel":
		return steelType
	case "fairy":
		return fairyType
	default:
		return NoType
	}
}

func (pt pokemonType) String() string {
	switch pt {
	case normalType:
		return "normal"
	case FireType:
		return "fire"
	case waterType:
		return "water"
	case electricType:
		return "electric"
	case grassType:
		return "grass"
	case iceType:
		return "ice"
	case fightingType:
		return "fighting"
	case poisonType:
		return "poison"
	case groundType:
		return "ground"
	case flyingType:
		return "flying"
	case psychicType:
		return "psychic"
	case bugType:
		return "bug"
	case rockType:
		return "rock"
	case ghostType:
		return "ghost"
	case dragonType:
		return "dragon"
	case darkType:
		return "dark"
	case steelType:
		return "steel"
	case fairyType:
		return "fairy"
	default:
		return "no type"
	}
}

type effectiveness float64

const (
	immuneEffectivensss   effectiveness = 0.0
	resistedEffectiveness effectiveness = 0.5
	normalEffectivenss    effectiveness = 1.0
	superEffectiveness    effectiveness = 2.0
)

var typeChart = map[pokemonType]map[pokemonType]effectiveness{
	normalType: {
		rockType: resistedEffectiveness, ghostType: immuneEffectivensss, steelType: resistedEffectiveness,
	},
	FireType: {
		FireType: resistedEffectiveness, waterType: resistedEffectiveness, grassType: superEffectiveness, iceType: superEffectiveness,
		bugType: superEffectiveness, rockType: resistedEffectiveness, dragonType: resistedEffectiveness, steelType: superEffectiveness,
	},
	waterType: {
		FireType: superEffectiveness, waterType: resistedEffectiveness, grassType: resistedEffectiveness, groundType: superEffectiveness,
		rockType: superEffectiveness, dragonType: resistedEffectiveness,
	},
	electricType: {
		waterType: superEffectiveness, electricType: resistedEffectiveness, grassType: resistedEffectiveness, groundType: immuneEffectivensss,
		flyingType: superEffectiveness, dragonType: resistedEffectiveness,
	},
	grassType: {
		FireType: resistedEffectiveness, waterType: superEffectiveness, grassType: resistedEffectiveness, poisonType: resistedEffectiveness,
		groundType: superEffectiveness, flyingType: resistedEffectiveness, bugType: resistedEffectiveness, rockType: superEffectiveness,
		dragonType: resistedEffectiveness, steelType: resistedEffectiveness,
	},
	iceType: {
		FireType: resistedEffectiveness, waterType: resistedEffectiveness, grassType: superEffectiveness, groundType: superEffectiveness,
		flyingType: superEffectiveness, dragonType: superEffectiveness, steelType: resistedEffectiveness, iceType: resistedEffectiveness,
	},
	fightingType: {
		normalType: superEffectiveness, iceType: superEffectiveness, rockType: superEffectiveness, darkType: superEffectiveness,
		steelType: superEffectiveness, poisonType: resistedEffectiveness, flyingType: resistedEffectiveness, psychicType: resistedEffectiveness,
		bugType: resistedEffectiveness, fairyType: resistedEffectiveness, ghostType: immuneEffectivensss,
	},
	poisonType: {
		grassType: superEffectiveness, fairyType: superEffectiveness, poisonType: resistedEffectiveness, groundType: resistedEffectiveness,
		rockType: resistedEffectiveness, ghostType: resistedEffectiveness, steelType: immuneEffectivensss,
	},
	groundType: {
		FireType: superEffectiveness, electricType: superEffectiveness, grassType: resistedEffectiveness, poisonType: superEffectiveness,
		flyingType: immuneEffectivensss, bugType: resistedEffectiveness, rockType: superEffectiveness, steelType: superEffectiveness,
	},
	flyingType: {
		electricType: resistedEffectiveness, grassType: superEffectiveness, fightingType: superEffectiveness,
		bugType: superEffectiveness, rockType: resistedEffectiveness, steelType: resistedEffectiveness,
	},
	psychicType: {
		fightingType: superEffectiveness, poisonType: superEffectiveness, psychicType: resistedEffectiveness,
		steelType: resistedEffectiveness, darkType: immuneEffectivensss,
	},
	bugType: {
		FireType: resistedEffectiveness, grassType: superEffectiveness, fightingType: resistedEffectiveness, poisonType: resistedEffectiveness,
		flyingType: resistedEffectiveness, psychicType: superEffectiveness, ghostType: resistedEffectiveness,
		darkType: superEffectiveness, steelType: resistedEffectiveness, fairyType: resistedEffectiveness,
	},
	rockType: {
		FireType: superEffectiveness, iceType: superEffectiveness, flyingType: superEffectiveness, bugType: superEffectiveness,
		fightingType: resistedEffectiveness, groundType: resistedEffectiveness, steelType: resistedEffectiveness,
	},
	ghostType: {
		normalType: immuneEffectivensss, psychicType: superEffectiveness, ghostType: superEffectiveness, darkType: resistedEffectiveness,
	},
	dragonType: {
		dragonType: superEffectiveness, steelType: resistedEffectiveness, fairyType: immuneEffectivensss,
	},
	darkType: {
		fightingType: resistedEffectiveness, psychicType: superEffectiveness, ghostType: superEffectiveness,
		darkType: resistedEffectiveness, fairyType: resistedEffectiveness,
	},
	steelType: {
		FireType: resistedEffectiveness, waterType: resistedEffectiveness, electricType: resistedEffectiveness, iceType: superEffectiveness,
		rockType: superEffectiveness, fairyType: superEffectiveness, steelType: resistedEffectiveness,
	},
	fairyType: {
		FireType: resistedEffectiveness, fightingType: superEffectiveness, poisonType: resistedEffectiveness,
		dragonType: superEffectiveness, darkType: superEffectiveness, steelType: resistedEffectiveness,
	},
}

func getEffectiveness(attacking, defending pokemonType) effectiveness {
	if row, ok := typeChart[attacking]; ok {
		if eff, ok := row[defending]; ok {
			return eff
		}
	}
	return normalEffectivenss
}
