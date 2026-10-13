package nuzengine

import "testing"

func TestRnbShouldSwitchReturnsExpectedDecision(t *testing.T) {
	tests := map[string]struct {
		hp         int
		wantSwitch bool
	}{
		"opponent can one hit KO":        {hp: 20},
		"slower opponent can two hit KO": {hp: 30},
		"can choose a safe replacement":  {hp: 50, wantSwitch: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			current := testSwitchPokemon("current", 100, 100, 100, 100, nil)
			replacement := testSwitchPokemon("replacement", tc.hp, 100, 100, 100, nil)
			opponent := testSwitchPokemon("opponent", 100, 50, 100, 100, &Move{Move: "sonic boom", Power: 1, PP: 1, Class: specialClass})
			bs := testSwitchBattleState(current, replacement, opponent)

			attempts := 1
			if tc.wantSwitch {
				attempts = 100
			}
			for i := 0; i < attempts; i++ {
				got := (rnbAi{}).shouldSwitch(bs, bs.activePlayerSlot, -1, []*pokemon{current, replacement})
				if got == tc.wantSwitch {
					return
				}
			}
			if tc.wantSwitch {
				t.Fatal("shouldSwitch never selected the safe replacement")
			}
			t.Fatal("shouldSwitch returned true for an unsafe replacement")
		})
	}
}

func TestBattleStateInitializersRequireAtLeastTwoPokemon(t *testing.T) {
	validParty := []*pokemon{
		testSwitchPokemon("one", 100, 100, 100, 100, nil),
		testSwitchPokemon("two", 100, 100, 100, 100, nil),
		testSwitchPokemon("three", 100, 100, 100, 100, nil),
	}
	initializers := []struct {
		name          string
		minimumLength int
		init          func(trainer, trainer, []*pokemon, []*pokemon, weatherState) (battleState, error)
	}{
		{name: "single", minimumLength: 1, init: initSingleBattleState},
		{name: "double", minimumLength: 2, init: initDoubleBattleState},
	}

	for _, initializer := range initializers {
		t.Run(initializer.name, func(t *testing.T) {
			minimumParty := validParty[:initializer.minimumLength]
			shortParty := validParty[:initializer.minimumLength-1]
			for _, parties := range []struct {
				name          string
				playerParty   []*pokemon
				opponentParty []*pokemon
			}{
				{name: "player", playerParty: shortParty, opponentParty: validParty},
				{name: "opponent", playerParty: validParty, opponentParty: shortParty},
			} {
				t.Run(parties.name, func(t *testing.T) {
					if _, err := initializer.init(trainer{}, trainer{}, parties.playerParty, parties.opponentParty, 0); err == nil {
						t.Fatalf("initializer succeeded with a party shorter than %d Pokemon", initializer.minimumLength)
					}
				})
			}

			if _, err := initializer.init(trainer{}, trainer{}, minimumParty, minimumParty, 0); err != nil {
				t.Fatalf("initializer rejected parties of %d Pokemon: %v", initializer.minimumLength, err)
			}
		})
	}
}

func testSwitchPokemon(name string, hp, speed, specialAttack, specialDefense int, mv *Move) *pokemon {
	moves := make([]*Move, 0)
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
	bs, err := initSingleBattleState(
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		[]*pokemon{current, replacement, testSwitchPokemon("reserve", 100, 100, 100, 100, nil)},
		[]*pokemon{opponent, testSwitchPokemon("reserve one", 100, 100, 100, 100, nil), testSwitchPokemon("reserve two", 100, 100, 100, 100, nil)},
		0,
	)
	if err != nil {
		panic(err)
	}
	singleBattleState, ok := bs.(*singleBattleState)
	if !ok {
		panic("expected singleBattleState")
	}
	return singleBattleState
}
