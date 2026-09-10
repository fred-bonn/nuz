package engine

type WeatherState int

const (
	NoneWeather WeatherState = iota
	rainWeather
	sunWeather
	sandstormWeather
	hailWeather
)

var weatherFuncs = map[WeatherState]func(*int, *int, pokemonType){
	rainWeather: func(num, den *int, t pokemonType) {
		switch t {
		case waterType:
			*num = *num * 3
			*den = *den * 2
		case FireType:
			*den = *den * 2
		}
	},
	sunWeather: func(num, den *int, t pokemonType) {
		switch t {
		case waterType:
			*den = *den * 2
		case FireType:
			*num = *num * 3
			*den = *den * 2
		}
	},
}

func (ws WeatherState) affectsMon(mon *Pokemon) bool {
	if mon.Ability == overcoatAbility || mon.Ability == magicGuardAbility || mon.Item.State == safetyGoggles {
		return false
	}

	switch ws {
	case sandstormWeather:
		if !mon.hasType(rockType) && !mon.hasType(steelType) && !mon.hasType(groundType) && (mon.Ability < sandVeilAbility || mon.Ability > sandForceAbility) {
			return true
		}
	case hailWeather:
		if !mon.hasType(iceType) && (mon.Ability < iceBodyAbility || mon.Ability > snowCloakAbility) {
			return true
		}
	}

	return false
}

func (ws WeatherState) activateMonAbility(bs BattleState, slot *slot) {
	mon := slot.mon

	switch ws {
	case hailWeather:
		if mon.Ability == iceBodyAbility {
			vprintf("%s healed due to ice body", mon.Base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 16)
		}
	case rainWeather:
		switch mon.Ability {
		case raindDishAbility:
			vprintf("%s healed due to rain dish", mon.Base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 16)
		case drySkinAbility:
			takeResidualDamage(bs, slot, "dry skin", 1, 8)
			vprintf("%s healed due to dry skin", mon.Base.Name)
			mon.ChangeHpBy(mon.MaxHP() / 8)
		case hydrationAbility:
			if mon.hasNonVolatileAilment() {
				for ailment := range nonVolatileStatuses {
					if mon.hasAilment(ailment) != nil {
						delete(mon.Ailments, ailment)
						vprintf("%s had its %s removed", mon.Base.Name, ailment.String())
						return
					}
				}
			}
		}
	case sunWeather:
		switch mon.Ability {
		case drySkinAbility:
			takeResidualDamage(bs, slot, "dry skin", 1, 8)
		case solarPowerAbility:
			takeResidualDamage(bs, slot, "solar power", 1, 8)
		}
	}
}

func (ws WeatherState) String() string {
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

func (ws WeatherState) onset() {
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
