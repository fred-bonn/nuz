package nuzengine

import "testing"

func TestRnbShouldSwitchRejectsUnsafeReplacements(t *testing.T) {
	tests := map[string]struct {
		hp int
	}{
		"opponent can one hit KO":        {hp: 20},
		"slower opponent can two hit KO": {hp: 30},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			current := testSwitchPokemon("current", 100, 100, 100, 100, nil)
			replacement := testSwitchPokemon("replacement", tc.hp, 100, 100, 100, nil)
			opponent := testSwitchPokemon("opponent", 100, 50, 100, 100, &move{name: "sonic boom", power: 1, pp: 1, class: specialClass})
			bs := testSwitchBattleState(current, replacement, opponent)

			if got := (rnbAi{}).shouldSwitch(bs, bs.activePlayerSlot, -1, []*pokemon{current, replacement}); got {
				t.Fatal("shouldSwitch returned true for an unsafe replacement")
			}
		})
	}
}

func TestRnbShouldSwitchCanChooseSafeReplacement(t *testing.T) {
	current := testSwitchPokemon("current", 100, 100, 100, 100, nil)
	replacement := testSwitchPokemon("replacement", 50, 100, 100, 100, nil)
	opponent := testSwitchPokemon("opponent", 100, 50, 100, 100, &move{name: "sonic boom", power: 1, pp: 1, class: specialClass})
	bs := testSwitchBattleState(current, replacement, opponent)

	for i := 0; i < 100; i++ {
		if (rnbAi{}).shouldSwitch(bs, bs.activePlayerSlot, -1, []*pokemon{current, replacement}) {
			return
		}
	}
	t.Fatal("shouldSwitch never selected a safe replacement")
}

func testSwitchPokemon(name string, hp, speed, specialAttack, specialDefense int, mv *move) *pokemon {
	moves := make([]*move, 0)
	if mv != nil {
		moves = append(moves, mv)
	}
	return &pokemon{
		base:     BasePokemon{Name: name, Types: []pokemonType{normalType}},
		level:    50,
		moves:    moves,
		stats:    []int{hp, 100, 100, specialAttack, specialDefense, speed},
		stages:   make([]int, 8),
		hp:       hp,
		ailments: make(map[ailmentState]*ailment),
		item:     &item{State: noneItem},
	}
}

func testSwitchBattleState(current, replacement, opponent *pokemon) *singleBattleState {
	return initSingleBattleState(
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		[]*pokemon{current, replacement},
		[]*pokemon{opponent},
		0,
	)
}
