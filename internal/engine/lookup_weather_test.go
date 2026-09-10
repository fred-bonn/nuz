package engine

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestWeatherAffectsMonAccordingToTypeAndAbilities(t *testing.T) {
	tests := map[string]struct {
		weather     WeatherState
		ability     abilityState
		goggles     bool
		pokemonType pokemonType
		want        bool
	}{
		"overcoat blocks sandstorm":                            {weather: sandstormWeather, ability: overcoatAbility, goggles: false, pokemonType: normalType, want: false},
		"overcoat blocks hail":                                 {weather: hailWeather, ability: overcoatAbility, goggles: false, pokemonType: normalType, want: false},
		"magic guard blocks sandstorm":                         {weather: sandstormWeather, ability: magicGuardAbility, goggles: false, pokemonType: normalType, want: false},
		"magic guard blocks hail":                              {weather: hailWeather, ability: magicGuardAbility, goggles: false, pokemonType: normalType, want: false},
		"intimidate is affected by sandstorm":                  {weather: sandstormWeather, ability: intimidateAbility, goggles: false, pokemonType: normalType, want: true},
		"intimidate is protected by safety goggles":            {weather: sandstormWeather, ability: intimidateAbility, goggles: true, pokemonType: normalType, want: false},
		"sand rush is unaffected by sandstorm":                 {weather: sandstormWeather, ability: sandRushAbility, goggles: false, pokemonType: normalType, want: false},
		"sand rush is affected by hail":                        {weather: hailWeather, ability: sandRushAbility, goggles: false, pokemonType: normalType, want: true},
		"sand rush is protected by safety goggles in hail":     {weather: hailWeather, ability: sandRushAbility, goggles: true, pokemonType: normalType, want: false},
		"ice body is unaffected by hail":                       {weather: hailWeather, ability: iceBodyAbility, goggles: false, pokemonType: normalType, want: false},
		"ice body is affected by sandstorm":                    {weather: sandstormWeather, ability: iceBodyAbility, goggles: false, pokemonType: normalType, want: true},
		"ice body is protected by safety goggles in sandstorm": {weather: sandstormWeather, ability: iceBodyAbility, goggles: true, pokemonType: normalType, want: false},
		"ground types are immune to sandstorm":                 {weather: sandstormWeather, ability: intimidateAbility, goggles: false, pokemonType: groundType, want: false},
		"steel types are immune to sandstorm":                  {weather: sandstormWeather, ability: intimidateAbility, goggles: false, pokemonType: steelType, want: false},
		"rock types are immune to sandstorm":                   {weather: sandstormWeather, ability: intimidateAbility, goggles: false, pokemonType: rockType, want: false},
		"ice types are affected by sandstorm":                  {weather: sandstormWeather, ability: intimidateAbility, goggles: false, pokemonType: iceType, want: true},
		"ice types are unaffected by hail":                     {weather: hailWeather, ability: intimidateAbility, goggles: false, pokemonType: iceType, want: false},
		"ground types are affected by hail":                    {weather: hailWeather, ability: intimidateAbility, goggles: false, pokemonType: groundType, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			dummy := &Pokemon{
				Base: BasePokemon{
					Types: []pokemonType{
						tc.pokemonType,
					},
				},
				Ability: tc.ability,
			}

			if tc.goggles {
				dummy.Item = &item{
					State: safetyGoggles,
				}
			} else {
				dummy.Item = &item{
					State:    noneItem,
					Consumed: true,
				}
			}

			if got := tc.weather.affectsMon(dummy); got != tc.want {
				t.Errorf("%s: tc.weather.affectsMon(%q) = %t, want %t", name, tc.ability, got, tc.want)
			}
		})
	}
}

func TestWeatherOnsetReportsTheCorrectWeatherMessage(t *testing.T) {
	tests := map[string]struct {
		weather  WeatherState
		contains string
	}{
		"reports rain onset":      {weather: rainWeather, contains: "to rain"},
		"reports sun onset":       {weather: sunWeather, contains: "turned harsh"},
		"reports sandstorm onset": {weather: sandstormWeather, contains: "brewed"},
		"reports hail onset":      {weather: hailWeather, contains: "to hail"},
	}

	oldVerbose := Verbose
	Verbose = true
	t.Cleanup(func() {
		Verbose = oldVerbose
	})

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var bf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			t.Cleanup(func() {
				os.Stdout = oldStdout
			})

			tc.weather.onset()
			w.Close()
			io.Copy(&bf, r)

			if !strings.Contains(bf.String(), tc.contains) {
				t.Errorf("%s: tc.weather.onset() did not contain \"%s\" in standard output: %s", name, tc.contains, bf.String())
			}
		})
	}
}
