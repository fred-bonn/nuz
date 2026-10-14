package nuzengine

import "testing"

func TestCalculateDamageMatchesGameRollRange(t *testing.T) {
	newHorsea := func() *pokemon {
		return &pokemon{
			base:     BasePokemon{Name: "horsea", Types: []pokemonType{waterType}},
			level:    17,
			stats:    []int{0, 0, 34, 34, 18, 0},
			stages:   make([]int, 8),
			hp:       100,
			ailments: make(map[ailmentState]*ailment),
			item:     &item{State: noneItem},
		}
	}
	newDwebble := func() *pokemon {
		return &pokemon{
			base:     BasePokemon{Name: "dwebble", Types: []pokemonType{bugType, rockType}},
			level:    17,
			stats:    []int{0, 35, 0, 0, 22, 0},
			stages:   make([]int, 8),
			hp:       100,
			ailments: make(map[ailmentState]*ailment),
			item:     &item{State: noneItem},
		}
	}

	tests := map[string]struct {
		horseaAttacks bool
		move          Move
		wantMin       int
		wantMax       int
	}{
		"bubble beam": {horseaAttacks: true, move: Move{Move: "bubble beam", Type: waterType, Power: 65, Class: specialClass}, wantMin: 44, wantMax: 54},
		"twister":     {horseaAttacks: true, move: Move{Move: "twister", Type: flyingType, Power: 40, Class: specialClass}, wantMin: 9, wantMax: 11},
		"bug bite":    {move: Move{Move: "bug bite", Type: bugType, Power: 60, Class: physicalClass}, wantMin: 13, wantMax: 16},
		"rock blast":  {move: Move{Move: "rock blast", Type: rockType, Power: 25, Class: physicalClass}, wantMin: 7, wantMax: 9},
		"knock off":   {move: Move{Move: "knock off", Type: darkType, Power: 65, Class: physicalClass}, wantMin: 10, wantMax: 12},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			user, target := newDwebble(), newHorsea()
			if tc.horseaAttacks {
				user, target = target, user
			}

			if got := calculateDamage(user, target, &tc.move, new(false), noneWeather, rollMin, false, false); got != tc.wantMin {
				t.Errorf("min roll = %d, want %d", got, tc.wantMin)
			}
			if got := calculateDamage(user, target, &tc.move, new(false), noneWeather, rollMax, false, false); got != tc.wantMax {
				t.Errorf("max roll = %d, want %d", got, tc.wantMax)
			}
		})
	}
}
