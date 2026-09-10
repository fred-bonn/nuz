package engine

import "testing"

func TestIsImmuneToPowderMovesDetectsPowderImmunity(t *testing.T) {
	tests := map[string]struct {
		t    pokemonType
		a    abilityState
		want bool
	}{
		"grass types are immune to powder moves":    {t: grassType, a: intimidateAbility, want: true},
		"grass types remain immune with overcoat":   {t: grassType, a: overcoatAbility, want: true},
		"non-grass types are not immune by default": {t: flyingType, a: intimidateAbility, want: false},
		"water types are not immune by default":     {t: waterType, a: intimidateAbility, want: false},
		"fire types are immune with overcoat":       {t: FireType, a: overcoatAbility, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Base: BasePokemon{
					Types: []pokemonType{
						tc.t,
					},
				},
				Ability: tc.a,
			}

			if got := mon.isImmuneToPowderMoves(); got != tc.want {
				t.Errorf("mon.isImmuneToPowderMoves() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestEffectiveStatAppliesStagesAndCritRules(t *testing.T) {
	tests := map[string]struct {
		stat  statState
		crit  bool
		stage int
		base  int
		want  int
	}{
		"increases a stat by one stage":                 {stat: attack, stage: 1, base: 100, want: 150},
		"decreases a stat by one stage":                 {stat: attack, stage: -1, base: 100, want: 66},
		"ignores positive stages on a crit for defense": {stat: defense, crit: true, stage: 1, base: 100, want: 100},
		"ignores negative stages on a crit for attack":  {stat: attack, crit: true, stage: -1, base: 100, want: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			}
			mon.Stats[tc.stat] = tc.base
			mon.Stages[tc.stat] = tc.stage

			if got := mon.effectiveStat(tc.stat, tc.crit); got != tc.want {
				t.Errorf("mon.effectiveStat(%s, %t) = %d, want %d", tc.stat, tc.crit, got, tc.want)
			}
		})
	}
}

func TestEffectiveSpeedAppliesSpeedModifiersCorrectly(t *testing.T) {
	tests := map[string]struct {
		stage     int
		base      int
		item      itemState
		ability   abilityState
		weather   WeatherState
		paralyzed bool
		unburden  bool
		want      int
	}{
		"boosts speed with a positive stage":                                   {stage: 1, base: 100, item: noneItem, ability: NoneAbility, weather: NoneWeather, paralyzed: false, unburden: false, want: 150},
		"reduces speed with a negative stage":                                  {stage: -1, base: 100, item: noneItem, ability: NoneAbility, weather: NoneWeather, paralyzed: false, unburden: false, want: 66},
		"halves speed when paralyzed":                                          {stage: 0, base: 100, item: noneItem, ability: NoneAbility, weather: NoneWeather, paralyzed: true, unburden: false, want: 25},
		"combines paralysis with a positive stage":                             {stage: 1, base: 100, item: noneItem, ability: NoneAbility, weather: NoneWeather, paralyzed: true, unburden: false, want: 37},
		"uses iron ball to halve speed":                                        {stage: 0, base: 100, item: ironBall, ability: NoneAbility, weather: NoneWeather, paralyzed: false, unburden: false, want: 50},
		"uses iron ball with a negative stage":                                 {stage: -1, base: 100, item: ironBall, ability: NoneAbility, weather: NoneWeather, paralyzed: false, unburden: false, want: 33},
		"does not boost speed with swift swim outside rain":                    {stage: 0, base: 100, item: noneItem, ability: swiftSwimAbility, weather: NoneWeather, paralyzed: false, unburden: false, want: 100},
		"boosts speed with swift swim in rain":                                 {stage: 0, base: 100, item: noneItem, ability: swiftSwimAbility, weather: rainWeather, paralyzed: false, unburden: false, want: 200},
		"does not boost speed with sand rush in rain":                          {stage: 0, base: 100, item: noneItem, ability: sandRushAbility, weather: rainWeather, paralyzed: false, unburden: false, want: 100},
		"boosts speed with sand rush in sandstorm":                             {stage: 0, base: 100, item: noneItem, ability: sandRushAbility, weather: sandstormWeather, paralyzed: false, unburden: false, want: 200},
		"does not boost speed with slush rush in rain":                         {stage: 0, base: 100, item: noneItem, ability: slushRushAbility, weather: rainWeather, paralyzed: false, unburden: false, want: 100},
		"does not boost speed with slush rush in hail while holding iron ball": {stage: 0, base: 100, item: ironBall, ability: slushRushAbility, weather: hailWeather, paralyzed: false, unburden: false, want: 100},
		"does not boost speed with chlorophyll outside sun":                    {stage: 0, base: 100, item: noneItem, ability: chlorophyllAbility, weather: rainWeather, paralyzed: false, unburden: false, want: 100},
		"boosts speed with chlorophyll in sun":                                 {stage: 1, base: 100, item: noneItem, ability: chlorophyllAbility, weather: sunWeather, paralyzed: false, unburden: false, want: 300},
		"does not boost speed for unburden without the ability":                {stage: 0, base: 100, item: noneItem, ability: NoneAbility, weather: NoneWeather, paralyzed: false, unburden: true, want: 100},
		"boosts speed for unburden with the ability":                           {stage: 0, base: 100, item: noneItem, ability: unburdenAbility, weather: NoneWeather, paralyzed: false, unburden: true, want: 200},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages:   []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Ability:  tc.ability,
				Ailments: make(map[ailmentState]*ailment),
				unburden: tc.unburden,
			}
			mon.Stats[Speed] = tc.base
			mon.Stages[Speed] = tc.stage
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item
			bbs := initBenchBattleState(tc.weather)
			if tc.paralyzed {
				mon.applyAilment(paralysisAilment, nil, nil)
			}

			if got := mon.effectiveSpeed(bbs); got != tc.want {
				t.Errorf("mon.effectiveSpeed(...) = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestEvasionFractionCalculatesTheCorrectFraction(t *testing.T) {
	tests := map[string]struct {
		stage   int
		keenEye bool
		want    struct {
			num int
			dem int
		}
	}{
		"uses a positive stage for evasion": {stage: 1, keenEye: false, want: struct {
			num int
			dem int
		}{num: 3, dem: 4}},
		"uses a larger positive stage for evasion": {stage: 5, keenEye: false, want: struct {
			num int
			dem int
		}{num: 3, dem: 8}},
		"uses a negative stage for evasion": {stage: -1, keenEye: false, want: struct {
			num int
			dem int
		}{num: 4, dem: 3}},
		"ignores evasion stages with keen eye": {stage: 5, keenEye: true, want: struct {
			num int
			dem int
		}{num: 1, dem: 1}},
		"ignores negative evasion stages with keen eye": {stage: -3, keenEye: true, want: struct {
			num int
			dem int
		}{num: 1, dem: 1}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			}
			mon.Stages[evasion] = tc.stage

			if num, dem := mon.evasionFraction(tc.keenEye); num != tc.want.num || dem != tc.want.dem {
				t.Errorf("mon.evasionFraction(%t) = (%d, %d), want (%d, %d)", tc.keenEye, num, dem, tc.want.num, tc.want.dem)
			}
		})
	}
}

func TestAccuracyFractionCalculatesTheCorrectFraction(t *testing.T) {
	tests := map[string]struct {
		stage int
		want  struct {
			num int
			dem int
		}
	}{
		"uses a positive stage for accuracy": {stage: 1, want: struct {
			num int
			dem int
		}{num: 4, dem: 3}},
		"uses a negative stage for accuracy": {stage: -1, want: struct {
			num int
			dem int
		}{num: 3, dem: 4}},
		"uses a larger negative stage for accuracy": {stage: -5, want: struct {
			num int
			dem int
		}{num: 3, dem: 8}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
			}
			mon.Stages[accuracy] = tc.stage

			if num, dem := mon.accuracyFraction(); num != tc.want.num || dem != tc.want.dem {
				t.Errorf("mon.evasionFraction() = (%d, %d), want (%d, %d)", num, dem, tc.want.num, tc.want.dem)
			}
		})
	}
}

func TestChangeStatStageByUpdatesStagesWithinTheAllowedRange(t *testing.T) {
	tests := map[string]struct {
		initial   int
		stat      statState
		change    int
		offensive bool
		ability   abilityState
		want      int
	}{
		"increases the stage by one":                          {initial: 0, stat: attack, change: 1, want: 1},
		"decreases the stage by one":                          {initial: 0, stat: defense, change: -1, want: -1},
		"caps a positive stage at the maximum":                {initial: 5, stat: Speed, change: 2, want: 6},
		"caps a negative stage at the minimum":                {initial: -5, stat: Speed, change: -2, want: -6},
		"blocks offensive stat drops with clear body":         {initial: 0, stat: attack, change: -1, offensive: true, ability: clearBodyAbility, want: 0},
		"does not block defensive stat drops with clear body": {initial: 0, stat: attack, change: -1, offensive: false, ability: clearBodyAbility, want: -1},
		"blocks accuracy drops with keen eye":                 {initial: 0, stat: accuracy, change: -1, ability: keenEyeAbility, want: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stages:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Ability: tc.ability,
			}
			mon.Stages[tc.stat] = tc.initial

			mon.changeStatStageBy(tc.stat, tc.change, tc.offensive)

			if got := mon.Stages[tc.stat]; got != tc.want {
				t.Errorf("mon.changeStatStageBy(%s, %d, %t) = %d, want %d", tc.stat, tc.change, tc.offensive, got, tc.want)
			}
		})
	}
}

func TestHasAilmentDetectsAppliedAilments(t *testing.T) {
	tests := map[string]struct {
		has   ailmentState
		check ailmentState
		want  bool
	}{
		"finds a matching paralysis ailment": {has: paralysisAilment, check: paralysisAilment, want: true},
		"does not find a different ailment":  {has: paralysisAilment, check: freezeAilment, want: false},
		"finds a matching poison ailment":    {has: poisonAilment, check: poisonAilment, want: true},
		"does not treat toxic as poison":     {has: poisonAilment, check: toxicAilment, want: false},
		"does not treat poison as toxic":     {has: toxicAilment, check: poisonAilment, want: false},
		"finds a matching yawn ailment":      {has: yawnAilment, check: yawnAilment, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Ailments: map[ailmentState]*ailment{
					tc.has: {
						State: tc.has,
					},
				},
			}

			if got := mon.hasAilment(tc.check) != nil; tc.want != got {
				t.Errorf("mon.hasAilment(%s) = %t, want %t", tc.check.String(), got, tc.want)
			}
		})
	}
}

func TestHasNonVolatileAilmentDetectsNonVolatileStatuses(t *testing.T) {
	tests := map[string]struct {
		has  ailmentState
		want bool
	}{
		"detects paralysis as non-volatile":          {has: paralysisAilment, want: true},
		"detects sleep as non-volatile":              {has: sleepAilment, want: true},
		"detects toxic as non-volatile":              {has: toxicAilment, want: true},
		"does not treat yawn as non-volatile":        {has: yawnAilment, want: false},
		"does not treat infatuation as non-volatile": {has: infatuationAilment, want: false},
		"does not treat confusion as non-volatile":   {has: confusionAilment, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Ailments: map[ailmentState]*ailment{
					tc.has: {
						State: tc.has,
					},
				},
			}

			if got := mon.hasNonVolatileAilment(); tc.want != got {
				t.Errorf("mon.hasAilment() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestApplyAilmentAppliesAilmentsWhenAllowed(t *testing.T) {
	tests := map[string]struct {
		ailment     ailmentState
		pokemonType pokemonType
		ability     abilityState
		want        bool
	}{
		"applies burn to a normal type":                {ailment: burnAilment, pokemonType: normalType, ability: intimidateAbility, want: true},
		"does not apply burn to a fire type":           {ailment: burnAilment, pokemonType: FireType, ability: intimidateAbility, want: false},
		"applies paralysis to a normal type":           {ailment: paralysisAilment, pokemonType: normalType, ability: intimidateAbility, want: true},
		"does not apply paralysis to an electric type": {ailment: paralysisAilment, pokemonType: electricType, ability: intimidateAbility, want: false},
		"applies freeze to a normal type":              {ailment: freezeAilment, pokemonType: normalType, ability: NoneAbility, want: true},
		"does not apply freeze to an ice type":         {ailment: freezeAilment, pokemonType: iceType, ability: intimidateAbility, want: false},
		"does not apply freeze with magma armor":       {ailment: freezeAilment, pokemonType: normalType, ability: magmaArmorAbility, want: false},
		"applies poison to a normal type":              {ailment: poisonAilment, pokemonType: normalType, ability: intimidateAbility, want: true},
		"does not apply poison to a steel type":        {ailment: poisonAilment, pokemonType: steelType, ability: intimidateAbility, want: false},
		"does not apply poison with immunity":          {ailment: poisonAilment, pokemonType: normalType, ability: immunityAbility, want: false},
		"applies sleep to a normal type":               {ailment: sleepAilment, pokemonType: normalType, ability: NoneAbility, want: true},
		"does not apply sleep with vital spirit":       {ailment: sleepAilment, pokemonType: normalType, ability: vitalSpiritAbility, want: false},
		"applies yawn to a normal type":                {ailment: yawnAilment, pokemonType: normalType, ability: NoneAbility, want: true},
		"does not apply yawn with vital spirit":        {ailment: yawnAilment, pokemonType: normalType, ability: vitalSpiritAbility, want: false},
		"applies confusion to a normal type":           {ailment: confusionAilment, pokemonType: normalType, ability: NoneAbility, want: true},
		"applies infatuation to a normal type":         {ailment: infatuationAilment, pokemonType: normalType, ability: NoneAbility, want: true},
		"does not apply infatuation with oblivious":    {ailment: infatuationAilment, pokemonType: normalType, ability: obliviousAbility, want: false},
		"does not apply burn with water veil":          {ailment: burnAilment, pokemonType: normalType, ability: waterVeilAbility, want: false},
		"does not apply paralysis with limber":         {ailment: paralysisAilment, pokemonType: normalType, ability: limberAbility, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Base: BasePokemon{
					Types: []pokemonType{tc.pokemonType},
				},
				Ability:  tc.ability,
				Ailments: make(map[ailmentState]*ailment),
			}
			item, _ := RegisterItem(noneItem, &mon)
			mon.Item = item

			got := mon.applyAilment(tc.ailment, nil, nil)
			if got != tc.want {
				t.Fatalf("mon.applyAilment(%s, nil, nil) = %t, want %t", tc.ailment.String(), got, tc.want)
			}
			if tc.want && mon.hasAilment(tc.ailment) == nil {
				t.Fatalf("mon.applyAilment(%s, nil, nil) did not leave %s applied", tc.ailment.String(), tc.ailment.String())
			}
			if !tc.want && mon.hasAilment(tc.ailment) != nil {
				t.Fatalf("mon.applyAilment(%s, nil, nil) applied %s despite a blocking condition", tc.ailment.String(), tc.ailment.String())
			}
		})
	}
}

func TestChangeHpByCapsHealingAndDamageAtMaxHP(t *testing.T) {
	tests := map[string]struct {
		initialHP int
		maxHP     int
		change    int
		want      int
	}{
		"increases HP by a normal amount":     {initialHP: 50, maxHP: 100, change: 20, want: 70},
		"caps HP at the maximum when healing": {initialHP: 90, maxHP: 100, change: 20, want: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				HP:    tc.initialHP,
			}
			mon.Stats[hitPoints] = tc.maxHP

			mon.ChangeHpBy(tc.change)

			if got := mon.HP; got != tc.want {
				t.Fatalf("mon.changeHpBy(%d) hp = %d, want %d", tc.change, got, tc.want)
			}
		})
	}
}

func TestToxicConversionTurnsPoisonIntoToxicWhenAppropriate(t *testing.T) {
	tests := map[string]struct {
		move string
		want bool
	}{
		"converts poison fang into toxic":    {move: "poison fang", want: true},
		"converts toxic into toxic":          {move: "toxic", want: true},
		"does not convert tackle into toxic": {move: "tackle", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Ailments: make(map[ailmentState]*ailment),
			}
			move := Move{
				Name: tc.move,
			}

			mon.applyAilment(poisonAilment, &move, nil)
			if tc.want && mon.hasAilment(toxicAilment) == nil {
				t.Fatalf("mon.applyAilment(poison, %s, nil) did not convert the poison to toxic", tc.move)
			}
			if !tc.want && mon.hasAilment(toxicAilment) != nil {
				t.Fatalf("mon.applyAilment(poison, %s, nil) did converted the poison to toxic", tc.move)
			}
		})
	}
}

func TestIsGroundedAccountsForTypeAbilitiesAndItems(t *testing.T) {
	tests := map[string]struct {
		pokemonType pokemonType
		ability     abilityState
		item        itemState
		want        bool
	}{
		"flying types are not grounded by default":    {pokemonType: flyingType, ability: NoneAbility, item: noneItem, want: false},
		"flying types become grounded with iron ball": {pokemonType: flyingType, ability: NoneAbility, item: ironBall, want: true},
		"normal types are grounded by intimidate":     {pokemonType: normalType, ability: intimidateAbility, item: noneItem, want: true},
		"normal types remain grounded with iron ball": {pokemonType: normalType, ability: intimidateAbility, item: ironBall, want: true},
		"levitate prevents grounding by default":      {pokemonType: normalType, ability: levitateAbility, item: noneItem, want: false},
		"levitate is bypassed by iron ball":           {pokemonType: normalType, ability: levitateAbility, item: ironBall, want: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Base: BasePokemon{
					Types: []pokemonType{tc.pokemonType},
				},
				Ability: tc.ability,
			}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			if got := mon.isGrounded(); got != tc.want {
				t.Fatalf("mon.isGrounded() = %t, want %t", got, tc.want)
			}
		})
	}
}

func TestApplyMoveTypeCalculatesTheCorrectDamageMultiplier(t *testing.T) {
	tests := map[string]struct {
		input struct {
			num int
			dem int
		}
		pokemonTypes []pokemonType
		ability      abilityState
		item         itemState
		moveType     pokemonType
		want         struct {
			num int
			dem int
		}
	}{
		"calculates doubled damage against normal types": {input: struct {
			num int
			dem int
		}{
			num: 6, dem: 3,
		}, pokemonTypes: []pokemonType{normalType}, moveType: fightingType, want: struct {
			num int
			dem int
		}{
			num: 12, dem: 3,
		}},
		"calculates doubled damage against a dual-type target": {input: struct {
			num int
			dem int
		}{
			num: 6, dem: 3,
		}, pokemonTypes: []pokemonType{normalType, flyingType}, moveType: fightingType, want: struct {
			num int
			dem int
		}{
			num: 12, dem: 6,
		}},
		"returns zero damage against an immune target": {input: struct {
			num int
			dem int
		}{
			num: 99, dem: 1,
		}, pokemonTypes: []pokemonType{normalType}, moveType: ghostType, want: struct {
			num int
			dem int
		}{
			num: 0, dem: 1,
		}},
		"returns zero damage against a levitating target": {input: struct {
			num int
			dem int
		}{
			num: 2, dem: 1,
		}, pokemonTypes: []pokemonType{normalType}, ability: levitateAbility, moveType: groundType, want: struct {
			num int
			dem int
		}{
			num: 0, dem: 1,
		}},
		"bypasses levitate with iron ball": {input: struct {
			num int
			dem int
		}{
			num: 2, dem: 1,
		}, pokemonTypes: []pokemonType{normalType}, ability: levitateAbility, item: ironBall, moveType: groundType, want: struct {
			num int
			dem int
		}{
			num: 2, dem: 1,
		}},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Base: BasePokemon{
					Types: tc.pokemonTypes,
				},
				Ability: tc.ability,
			}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			if num, dem := mon.applyMoveType(tc.input.num, tc.input.dem, tc.moveType); num != tc.want.num || dem != tc.want.dem {
				t.Fatalf("mon.applyMoveType(%d, %d, %s) = %d, %d; want %d, %d", tc.input.num, tc.input.dem, tc.moveType.String(), num, dem, tc.want.num, tc.want.dem)
			}
		})
	}
}
