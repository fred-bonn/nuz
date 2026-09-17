package nuzengine

type weatherState int

const (
	noneWeather weatherState = iota
	rainWeather
	sunWeather
	sandstormWeather
	hailWeather
)

var weatherFuncs = map[weatherState]func(*int, *int, pokemonType){
	rainWeather: func(num, den *int, t pokemonType) {
		switch t {
		case waterType:
			*num = *num * 3
			*den = *den * 2
		case fireType:
			*den = *den * 2
		}
	},
	sunWeather: func(num, den *int, t pokemonType) {
		switch t {
		case waterType:
			*den = *den * 2
		case fireType:
			*num = *num * 3
			*den = *den * 2
		}
	},
}

func (ws weatherState) affectsMon(mon *pokemon) bool {
	if mon.ability == overcoatAbility || mon.ability == magicGuardAbility || mon.item.State == safetyGoggles {
		return false
	}

	switch ws {
	case sandstormWeather:
		if !mon.hasType(rockType) && !mon.hasType(steelType) && !mon.hasType(groundType) && (mon.ability < sandVeilAbility || mon.ability > sandForceAbility) {
			return true
		}
	case hailWeather:
		if !mon.hasType(iceType) && (mon.ability < iceBodyAbility || mon.ability > snowCloakAbility) {
			return true
		}
	}

	return false
}

func (ws weatherState) activateMonAbility(bs battleState, slot *slot) {
	mon := slot.mon

	switch ws {
	case hailWeather:
		if mon.ability == iceBodyAbility {
			vprintf("%s healed due to ice body", mon.base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 16)
		}
	case rainWeather:
		switch mon.ability {
		case raindDishAbility:
			vprintf("%s healed due to rain dish", mon.base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 16)
		case drySkinAbility:
			takeResidualDamage(bs, slot, "dry skin", 1, 8)
			vprintf("%s healed due to dry skin", mon.base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 8)
		case hydrationAbility:
			if mon.hasNonVolatileAilment() {
				for ailment := range nonVolatileStatuses {
					if mon.hasAilment(ailment) != nil {
						delete(mon.ailments, ailment)
						vprintf("%s had its %s removed", mon.base.Name, ailment.String())
						return
					}
				}
			}
		}
	case sunWeather:
		switch mon.ability {
		case drySkinAbility:
			takeResidualDamage(bs, slot, "dry skin", 1, 8)
		case solarPowerAbility:
			takeResidualDamage(bs, slot, "solar power", 1, 8)
		}
	}
}

func (ws weatherState) String() string {
	switch ws {
	case rainWeather:
		return "rain"
	case sunWeather:
		return "sun"
	case hailWeather:
		return "hail"
	case sandstormWeather:
		return "sandstorm"
	default:
		return "no weather"
	}
}

func (ws weatherState) onset() {
	switch ws {
	case rainWeather:
		vprintln("it started to rain")
	case sunWeather:
		vprintln("the sunlight turned harsh")
	case sandstormWeather:
		vprintln("a sandstorm brewed")
	case hailWeather:
		vprintln("it started to hail")
	}
}
