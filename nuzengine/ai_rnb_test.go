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

func TestRnbChoosesTargetsWithIndependentHighestDamageScores(t *testing.T) {
	tests := map[string]struct {
		trials        int
		minSelections int
	}{
		"selects each target": {trials: 1000, minSelections: 400},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			player := &trainer{ai: rnbAi{}}
			opponent := &trainer{ai: rnbAi{}}
			staravia := testSwitchPokemon("staravia", 50, 45, 40, 40, &Move{
				Move:  "aerial ace",
				Type:  flyingType,
				Power: 60,
				PP:    20,
				Class: physicalClass,
			})
			staravia.moves = append(staravia.moves,
				&Move{Move: "endeavor", Type: normalType, Power: 1, PP: 5, Class: physicalClass},
				&Move{Move: "quick attack", Type: normalType, Power: 40, PP: 30, Class: physicalClass, Priority: 1},
			)
			staravia.base.Types = []pokemonType{normalType, flyingType}
			staravia.stats[attack] = 38
			dwebble := testSwitchPokemon("dwebble", 40, 24, 35, 35, nil)
			dwebble.base.Types = []pokemonType{bugType, rockType}
			dwebble.stats[defense] = 31
			sandygast := testSwitchPokemon("sandygast", 43, 11, 35, 35, nil)
			sandygast.base.Types = []pokemonType{groundType, ghostType}
			sandygast.stats[defense] = 27

			staraviaSlot := &slot{mon: staravia, Trainer: player, firstTurn: true}
			bs := &doubleBattleState{
				activePlayerSlotOne:   staraviaSlot,
				activePlayerSlotTwo:   &slot{mon: testSwitchPokemon("ally", 50, 20, 35, 35, nil), Trainer: player},
				activeOpponentSlotOne: &slot{mon: dwebble, Trainer: opponent, firstTurn: true},
				activeOpponentSlotTwo: &slot{mon: sandygast, Trainer: opponent, firstTurn: true},
			}

			targets := make(map[string]int)
			for range tc.trials {
				action := chooseNextAction(bs, staraviaSlot, []*pokemon{staravia}, rnbAi{})
				move, ok := action.(*moveAction)
				if !ok {
					t.Fatalf("action = %T, want *moveAction", action)
				}
				targets[move.targetSlot.mon.base.Name]++
			}

			for _, target := range []string{"dwebble", "sandygast"} {
				if targets[target] < tc.minSelections {
					t.Errorf("%s selected %d times, want at least %d", target, targets[target], tc.minSelections)
				}
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
