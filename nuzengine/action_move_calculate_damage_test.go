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

			if got := calculateDamage(initBenchBattleState(noneWeather), user, target, &tc.move, new(false), noneWeather, rollMin, false, false); got != tc.wantMin {
				t.Errorf("min roll = %d, want %d", got, tc.wantMin)
			}
			if got := calculateDamage(initBenchBattleState(noneWeather), user, target, &tc.move, new(false), noneWeather, rollMax, false, false); got != tc.wantMax {
				t.Errorf("max roll = %d, want %d", got, tc.wantMax)
			}
		})
	}
}

func TestSpreadMoveTargets(t *testing.T) {
	playerTrainer := &trainer{}
	opponentTrainer := &trainer{}
	user := &slot{Trainer: playerTrainer, mon: &pokemon{}}
	ally := &slot{Trainer: playerTrainer, mon: &pokemon{}}
	opponentOne := &slot{Trainer: opponentTrainer, mon: &pokemon{}}
	opponentTwo := &slot{Trainer: opponentTrainer, mon: &pokemon{}}
	state := &doubleBattleState{
		activePlayerSlotOne:   user,
		activePlayerSlotTwo:   ally,
		activeOpponentSlotOne: opponentOne,
		activeOpponentSlotTwo: opponentTwo,
	}

	tests := []struct {
		name      string
		target    string
		wantSlots []*slot
	}{
		{name: "all opponents", target: "all-opponents", wantSlots: []*slot{opponentOne, opponentTwo}},
		{name: "all other pokemon", target: "all-other-pokemon", wantSlots: []*slot{ally, opponentOne, opponentTwo}},
		{name: "single target", target: "selected-pokemon", wantSlots: []*slot{opponentTwo}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			move := &Move{Target: test.target}
			action := &moveAction{userSlot: user, targetSlot: opponentTwo, move: move}
			got := action.targets(state)
			if len(got) != len(test.wantSlots) {
				t.Fatalf("got %d targets, want %d", len(got), len(test.wantSlots))
			}
			for i, slot := range test.wantSlots {
				if got[i] != slot {
					t.Errorf("target %d = %p, want %p", i, got[i], slot)
				}
			}
		})
	}

	opponentTwo.mon.fainted = true
	got := (&moveAction{
		userSlot:   user,
		targetSlot: opponentOne,
		move:       &Move{Target: "all-opponents"},
	}).targets(state)
	if len(got) != 1 || got[0] != opponentOne {
		t.Errorf("all-opponents with a fainted slot = %v, want only opponent one", got)
	}
}

func TestApplySpreadDamageModifier(t *testing.T) {
	tests := []struct {
		name   string
		damage int
		spread bool
		want   int
	}{
		{name: "spread damage reduced", damage: 20, spread: true, want: 15},
		{name: "minimum positive damage", damage: 1, spread: true, want: 1},
		{name: "immunity remains zero", damage: 0, spread: true, want: 0},
		{name: "single target unchanged", damage: 20, spread: false, want: 20},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := applySpreadDamageModifier(test.damage, test.spread); got != test.want {
				t.Errorf("applySpreadDamageModifier(%d, %t) = %d, want %d", test.damage, test.spread, got, test.want)
			}
		})
	}
}
