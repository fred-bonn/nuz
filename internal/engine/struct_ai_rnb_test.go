package engine

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
			opponent := testSwitchPokemon("opponent", 100, 50, 100, 100, &Move{Name: "sonic boom", Power: 1, PP: 1, Class: SpecialClass})
			bs := testSwitchBattleState(current, replacement, opponent)

			if got := (RnbAi{}).shouldSwitch(bs, bs.activePlayerSlot, -1, []*Pokemon{current, replacement}); got {
				t.Fatal("shouldSwitch returned true for an unsafe replacement")
			}
		})
	}
}

func TestRnbShouldSwitchCanChooseSafeReplacement(t *testing.T) {
	current := testSwitchPokemon("current", 100, 100, 100, 100, nil)
	replacement := testSwitchPokemon("replacement", 50, 100, 100, 100, nil)
	opponent := testSwitchPokemon("opponent", 100, 50, 100, 100, &Move{Name: "sonic boom", Power: 1, PP: 1, Class: SpecialClass})
	bs := testSwitchBattleState(current, replacement, opponent)

	for i := 0; i < 100; i++ {
		if (RnbAi{}).shouldSwitch(bs, bs.activePlayerSlot, -1, []*Pokemon{current, replacement}) {
			return
		}
	}
	t.Fatal("shouldSwitch never selected a safe replacement")
}

func testSwitchPokemon(name string, hp, speed, specialAttack, specialDefense int, move *Move) *Pokemon {
	moves := []*Move{}
	if move != nil {
		moves = append(moves, move)
	}
	return &Pokemon{
		Base:     BasePokemon{Name: name, Types: []pokemonType{normalType}},
		level:    50,
		Moves:    moves,
		Stats:    []int{hp, 100, 100, specialAttack, specialDefense, speed},
		Stages:   make([]int, 8),
		HP:       hp,
		Ailments: make(map[ailmentState]*ailment),
		Item:     &item{State: noneItem},
	}
}

func testSwitchBattleState(current, replacement, opponent *Pokemon) *SingleBattleState {
	return InitSingleBattleState(
		Trainer{AI: RnbAi{}, FieldEffects: make(map[FieldEffect]int)},
		Trainer{AI: RnbAi{}, FieldEffects: make(map[FieldEffect]int)},
		[]*Pokemon{current, replacement},
		[]*Pokemon{opponent},
		NoneWeather,
	)
}
