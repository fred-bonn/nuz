package engine

import (
	"testing"
)

func TestBerryItemsCureAilments(t *testing.T) {
	tests := map[string]struct {
		ailment      ailmentState
		item         itemState
		unnerved     bool
		wantHp       bool
		wantConsumed bool
	}{
		"cheri cures paralysis":       {ailment: paralysisAilment, item: cheriBerry, wantHp: true, wantConsumed: true},
		"cheri blocked by unnerve":    {ailment: paralysisAilment, item: cheriBerry, unnerved: true, wantHp: false},
		"chesto cures sleep":          {ailment: sleepAilment, item: chestoBerry, wantHp: true, wantConsumed: true},
		"chesto does not cure poison": {ailment: poisonAilment, item: chestoBerry, wantHp: false},
		"pecha cures poison":          {ailment: poisonAilment, item: pechaBerry, wantHp: true, wantConsumed: true},
		"rawst cures burn":            {ailment: burnAilment, item: rawstBerry, wantHp: true, wantConsumed: true},
		"aspear cures freeze":         {ailment: freezeAilment, item: aspearBerry, wantHp: true, wantConsumed: true},
		"persim cures confusion":      {ailment: confusionAilment, item: persimBerry, wantHp: true, wantConsumed: true},
		"persim does not cure freeze": {ailment: freezeAilment, item: persimBerry, wantHp: false},
		"lum cures paralysis":         {ailment: paralysisAilment, item: lumBerry, wantHp: true, wantConsumed: true},
		"lum cures sleep":             {ailment: sleepAilment, item: lumBerry, wantHp: true, wantConsumed: true},
		"lum blocked by unnerve":      {ailment: paralysisAilment, item: lumBerry, unnerved: true, wantHp: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Base: BasePokemon{
					Types: []pokemonType{normalType},
				},
				unnerved: tc.unnerved,
				Ailments: make(map[ailmentState]*ailment),
			}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			if got := mon.applyAilment(tc.ailment, nil, nil); got != true {
				t.Fatalf("mon.applyAilment(%s, nil, nil) = %t, want true", tc.ailment.String(), got)
			}
			if got := mon.hasAilment(tc.ailment) == nil; got != tc.wantHp {
				t.Fatalf("mon.hasAilment(%s) = %t, want %t", tc.ailment.String(), got, tc.wantHp)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestBerryItemsHeal(t *testing.T) {
	tests := map[string]struct {
		initialHp    int
		maxHp        int
		item         itemState
		unnerved     bool
		wantHp       int
		wantConsumed bool
	}{
		"oran heals below threshold hp":        {initialHp: 50, maxHp: 100, item: oranBerry, wantHp: 60, wantConsumed: true},
		"oran does not heal below threshold":   {initialHp: 51, maxHp: 100, item: oranBerry, wantHp: 51},
		"sitrus heals below threshold hp":      {initialHp: 50, maxHp: 100, item: sitrusBerry, wantHp: 75, wantConsumed: true},
		"sitrus does not heal below threshold": {initialHp: 51, maxHp: 100, item: oranBerry, unnerved: true, wantHp: 51},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				HP:       tc.initialHp,
				unnerved: tc.unnerved,
			}
			mon.Stats[hitPoints] = tc.maxHp
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item
			mon.CheckItemTrigger(true, nil)

			if got := mon.HP; got != tc.wantHp {
				t.Errorf("mon.hp = %d, want %d", got, tc.wantHp)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestLeppaBerryRestoresPP(t *testing.T) {
	tests := map[string]struct {
		initialPP    int
		maxPP        int
		unnerved     bool
		wantPP       int
		wantConsumed bool
	}{
		"restores PP from 0 to 10":       {initialPP: 0, maxPP: 15, wantPP: 10, wantConsumed: true},
		"does not restore above max PP":  {initialPP: 0, maxPP: 3, wantPP: 3, wantConsumed: true},
		"does not restore when unnerved": {initialPP: 0, maxPP: 15, unnerved: true, wantPP: 0, wantConsumed: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{unnerved: tc.unnerved}
			item, _ := RegisterItem(leppaBerry, &mon)
			mon.Item = item

			move := Move{PP: tc.initialPP, MaxPP: tc.maxPP}
			mon.CheckItemTrigger(true, makeLeppaBerryEvent(&move))

			if got := move.PP; got != tc.wantPP {
				t.Errorf("move.PP = %d, want %d", got, tc.wantPP)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", leppaBerry.String())
				} else {
					t.Errorf("%s was not consumed", leppaBerry.String())
				}
			}
		})
	}
}

func TestLeftoversHealsAtEndOfTurn(t *testing.T) {
	mon := &Pokemon{
		Stats: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
		HP:    100,
		Item:  &item{State: leftovers},
	}
	mon.Stats[hitPoints] = 200

	slotVar := &slot{mon: mon}
	bs := &dummyBattleState{slots: []*slot{slotVar}}
	resolveEndOfTurn(bs)

	wantHP := 112
	if got := mon.HP; got != wantHP {
		t.Fatalf("mon.hp = %d, want %d", got, wantHP)
	}
}

func TestIsBerry(t *testing.T) {
	tests := map[string]struct {
		item itemState
		want bool
	}{
		"oran berry":   {item: oranBerry, want: true},
		"no item":      {item: noneItem, want: false},
		"yache berry":  {item: yacheBerry, want: true},
		"iron ball":    {item: ironBall, want: false},
		"liechi berry": {item: liechiBerry, want: true},
		"mystic water": {item: mysticWater, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.item.isBerry(); got != tc.want {
				t.Errorf("%s.isBerry() = %t, want %t", tc.item.String(), got, tc.want)
			}
		})
	}
}

func TestIsChoice(t *testing.T) {
	tests := map[string]struct {
		item itemState
		want bool
	}{
		"choice scarf":  {item: choiceScarf, want: true},
		"choice specs":  {item: choiceSpecs, want: true},
		"focus sash:":   {item: focusSash, want: false},
		"no item":       {item: noneItem, want: false},
		"oran berry":    {item: oranBerry, want: false},
		"silver powder": {item: silverPowder, want: false},
		"choice band":   {item: choiceBand, want: true},
		"assault vest":  {item: assaultVest, want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := tc.item.isChoice(); got != tc.want {
				t.Errorf("%s.isChoice() = %t, want %t", tc.item.String(), got, tc.want)
			}
		})
	}
}

func TestPinchHealBerries(t *testing.T) {
	tests := map[string]struct {
		item         itemState
		nature       string
		initialHp    int
		maxHp        int
		gluttony     bool
		unnerved     bool
		wantHp       int
		wantConfused bool
		wantConsumed bool
	}{
		"figy with lonely nature":                          {item: figyBerry, nature: "lonely", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConsumed: true},
		"figy with bold nature confuses":                   {item: figyBerry, nature: "bold", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConfused: true, wantConsumed: true},
		"figy heals at threshold":                          {item: figyBerry, nature: "lonely", initialHp: 50, maxHp: 100, gluttony: true, wantHp: 100, wantConsumed: true},
		"figy does not heal above threshold":               {item: figyBerry, nature: "bold", initialHp: 51, maxHp: 100, gluttony: true, wantHp: 51},
		"iapapa with adamant nature":                       {item: iapapaBerry, nature: "adamant", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConsumed: true},
		"iapapa with hasty nasture confuses":               {item: iapapaBerry, nature: "hasty", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConfused: true, wantConsumed: true},
		"wiki with rash nature":                            {item: wikiBerry, nature: "rash", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConsumed: true},
		"wiki with impish nature confuses":                 {item: wikiBerry, nature: "impish", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConfused: true, wantConsumed: true},
		"aguav with relaxed nature":                        {item: aguavBerry, nature: "relaxed", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConsumed: true},
		"aguav with rash nature confuses":                  {item: aguavBerry, nature: "rash", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConfused: true, wantConsumed: true},
		"mago wtih modest nature":                          {item: magoBerry, nature: "modest", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConsumed: true},
		"mago with sassy nature confuses":                  {item: magoBerry, nature: "sassy", initialHp: 40, maxHp: 100, gluttony: true, wantHp: 90, wantConfused: true, wantConsumed: true},
		"iapapa does not heal dead mons":                   {item: iapapaBerry, nature: "adamant", initialHp: 0, maxHp: 100, gluttony: true, wantHp: 0},
		"aguav blocked by unnerve":                         {item: aguavBerry, nature: "rash", initialHp: 40, maxHp: 100, gluttony: true, unnerved: true, wantHp: 40},
		"figy without gluttony does not heal":              {item: figyBerry, nature: "lonely", initialHp: 40, maxHp: 100, wantHp: 40},
		"figy without gluttony heals at threshold":         {item: figyBerry, nature: "lonely", initialHp: 25, maxHp: 100, wantHp: 75, wantConsumed: true},
		"figy without gluttony does not heal at threshold": {item: figyBerry, nature: "lonely", initialHp: 26, maxHp: 100, wantHp: 26},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Ailments: make(map[ailmentState]*ailment),
				HP:       tc.initialHp,
				unnerved: tc.unnerved,
			}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item
			nat, _ := getNature(tc.nature)
			mon.nat = nat
			mon.Stats[hitPoints] = tc.maxHp
			if tc.gluttony {
				mon.Ability = GluttonyAbility
			}

			mon.CheckItemTrigger(true, nil)
			if got := mon.HP; got != tc.wantHp {
				t.Errorf("mon.hp = %d, want %d", got, tc.wantHp)
			}

			ailment := mon.hasAilment(confusionAilment)
			if got := ailment != nil; got != tc.wantConfused {
				if got {
					t.Errorf("mon was confused by %s with a %s nature", tc.item.String(), tc.nature)
				} else {
					t.Errorf("mon was not confused by %s with a %s nature", tc.item.String(), tc.nature)
				}
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestStatBoostBerries(t *testing.T) {
	tests := map[string]struct {
		item         itemState
		stat         statState
		initialHp    int
		maxHp        int
		gluttony     bool
		unnerved     bool
		wantStage    int
		wantConsumed bool
	}{
		"liechi boosts attack at threshold":                   {item: liechiBerry, stat: attack, initialHp: 25, maxHp: 100, wantStage: 1, wantConsumed: true},
		"liechi does not boost above threshold":               {item: liechiBerry, stat: attack, initialHp: 26, maxHp: 100, wantStage: 0},
		"liechi blocked by unnerve":                           {item: liechiBerry, stat: attack, initialHp: 25, maxHp: 100, unnerved: true, wantStage: 0},
		"liechi uses gluttony threshold":                      {item: liechiBerry, stat: attack, initialHp: 50, maxHp: 100, gluttony: true, wantStage: 1, wantConsumed: true},
		"liechi does not boost with gluttony above threshold": {item: liechiBerry, stat: attack, initialHp: 51, maxHp: 100, gluttony: true, wantStage: 0},
		"salac uses gluttony threshold":                       {item: salacBerry, stat: Speed, initialHp: 50, maxHp: 100, gluttony: true, wantStage: 1, wantConsumed: true},
		"salac does not boost with gluttony above threshold":  {item: salacBerry, stat: Speed, initialHp: 51, maxHp: 100, gluttony: true, wantStage: 0},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:    []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages:   []int{0, 0, 0, 0, 0, 0, 0, 0},
				HP:       tc.initialHp,
				unnerved: tc.unnerved,
			}
			mon.Stats[hitPoints] = tc.maxHp
			if tc.gluttony {
				mon.Ability = GluttonyAbility
			}

			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			mon.CheckItemTrigger(true, nil)
			if got := mon.Stages[tc.stat]; got != tc.wantStage {
				t.Errorf("mon.stages[%s] = %d, want %d", tc.stat, got, tc.wantStage)
			}

			if tc.wantStage == 1 && !item.Consumed {
				t.Errorf("%s was not consumed", tc.item.String())
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestResistBerries(t *testing.T) {
	tests := map[string]struct {
		item          itemState
		Pokemon       pokemonType
		unnerved      bool
		event         bool
		initialDamage int
		wantDamage    int
		wantConsumed  bool
	}{
		"chople reduces fighting damage":          {item: chopleBerry, Pokemon: fightingType, event: true, initialDamage: 100, wantDamage: 50, wantConsumed: true},
		"chople ignores non-matching type":        {item: chopleBerry, Pokemon: normalType, event: true, initialDamage: 100, wantDamage: 100},
		"chople blocked by unnerve":               {item: chopleBerry, Pokemon: fightingType, event: true, initialDamage: 100, unnerved: true, wantDamage: 100},
		"berry wont activate if event is missing": {item: chopleBerry, Pokemon: fightingType, initialDamage: 100, wantDamage: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{unnerved: tc.unnerved}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			damage := tc.initialDamage
			if tc.event {
				mon.CheckItemTrigger(true, makeResistBerryEvent(tc.Pokemon, &damage))
			} else {
				mon.CheckItemTrigger(true, nil)
			}

			if got := damage; got != tc.wantDamage {
				t.Errorf("damage = %d, want %d", got, tc.wantDamage)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestTypeGems(t *testing.T) {
	tests := map[string]struct {
		item         itemState
		move         pokemonType
		initialPower int
		event        bool
		wantPower    int
		wantConsumed bool
	}{
		"normal gem increases normal type":              {item: normalGem, move: normalType, initialPower: 100, event: true, wantPower: 150, wantConsumed: true},
		"normal gem does not increase normal fire type": {item: normalGem, move: FireType, initialPower: 100, event: true, wantPower: 100},
		"gem wont activate if event is missing":         {item: normalGem, move: normalType, initialPower: 100, wantPower: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			power := tc.initialPower
			if tc.event {
				mon.CheckItemTrigger(true, makeGemEvent(tc.move, &power))
			} else {
				mon.CheckItemTrigger(true, nil)
			}

			if got := power; got != tc.wantPower {
				t.Errorf("power = %d, want %d", got, tc.wantPower)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("%s was consumed", tc.item.String())
				} else {
					t.Errorf("%s was not consumed", tc.item.String())
				}
			}
		})
	}
}

func TestChoiceScarf(t *testing.T) {
	tests := map[string]struct {
		speed int
		want  int
	}{
		"100 base speed": {speed: 100, want: 150},
		"66 base speed":  {speed: 66, want: 99},
		"1 base speed":   {speed: 1, want: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0},
			}
			mon.Stats[Speed] = tc.speed
			item, _ := RegisterItem(choiceScarf, &mon)
			mon.Item = item
			bs := initBenchBattleState(NoneWeather)

			if got := mon.effectiveSpeed(bs); got != tc.want {
				t.Errorf("mon.effectiveSpeed(bs) = %d, want %d", got, tc.want)
			}

			if got := mon.Stats[Speed]; got != tc.speed {
				t.Errorf("choice scarf should not modify the base stats directly")
			}
		})
	}
}

func TestAssaultVest(t *testing.T) {
	tests := map[string]struct {
		spdef int
		crit  bool
		want  int
	}{
		"100 base special defense":   {spdef: 100, want: 150},
		"crit should have no effect": {spdef: 100, crit: true, want: 150},
		"66 base special defense":    {spdef: 66, want: 99},
		"1 base special defense":     {spdef: 1, want: 1},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0},
			}
			mon.Stats[specialDefense] = tc.spdef
			item, _ := RegisterItem(assaultVest, &mon)
			mon.Item = item

			if got := mon.effectiveStat(specialDefense, tc.crit); got != tc.want {
				t.Errorf("mon.effectiveSpeed(bs) = %d, want %d", got, tc.want)
			}

			if got := mon.Stats[specialDefense]; got != tc.spdef {
				t.Errorf("assault vest should not modify the base stats directly")
			}
		})
	}
}

func TestChoiceBand(t *testing.T) {
	tests := map[string]struct {
		initialAttack int
		class         moveClass
		event         bool
		want          int
	}{
		"100 base attack": {initialAttack: 100, class: physicalClass, event: true, want: 150},
		"66 base attack":  {initialAttack: 66, class: physicalClass, event: true, want: 99},
		"1 base attack":   {initialAttack: 1, class: physicalClass, event: true, want: 1},
		"choice band wont activate with special move class":   {initialAttack: 100, class: SpecialClass, event: true, want: 100},
		"choice band wont activate without choice item event": {initialAttack: 100, class: physicalClass, event: false, want: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0},
			}
			move := Move{
				Class: tc.class,
			}
			item, _ := RegisterItem(choiceBand, &mon)
			mon.Item = item

			newAttack := tc.initialAttack
			if tc.event {
				mon.CheckItemTrigger(false, makeChoiceItemEvent(&move, noStat, &newAttack))
			} else {
				mon.CheckItemTrigger(false, nil)
			}

			if got := newAttack; got != tc.want {
				t.Errorf("attack = %d, want %d", got, tc.want)
			}

			if newAttack == mon.effectiveStat(attack, false) {
				t.Errorf("choice specs should not modify the base stats directly")
			}
		})
	}
}

func TestChoiceSpecs(t *testing.T) {
	tests := map[string]struct {
		initialSpecialAttack int
		class                moveClass
		event                bool
		want                 int
	}{
		"100 base attack": {initialSpecialAttack: 100, class: SpecialClass, event: true, want: 150},
		"66 base attack":  {initialSpecialAttack: 66, class: SpecialClass, event: true, want: 99},
		"1 base attack":   {initialSpecialAttack: 1, class: SpecialClass, event: true, want: 1},
		"choice specs wont activate with physical move class":  {initialSpecialAttack: 100, class: physicalClass, event: true, want: 100},
		"choice specs wont activate without choice item event": {initialSpecialAttack: 100, class: SpecialClass, event: false, want: 100},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats:  []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				Stages: []int{0, 0, 0, 0, 0, 0, 0, 0},
			}
			move := Move{
				Class: tc.class,
			}
			item, _ := RegisterItem(choiceSpecs, &mon)
			mon.Item = item

			newSpecialAttack := tc.initialSpecialAttack
			if tc.event {
				mon.CheckItemTrigger(false, makeChoiceItemEvent(&move, noStat, &newSpecialAttack))
			} else {
				mon.CheckItemTrigger(false, nil)
			}

			if got := newSpecialAttack; got != tc.want {
				t.Errorf("attack = %d, want %d", got, tc.want)
			}

			if newSpecialAttack == mon.effectiveStat(specialAttack, false) {
				t.Errorf("choice specs should not modify the base stats directly")
			}
		})
	}
}

func TestFocusSash(t *testing.T) {
	tests := map[string]struct {
		initialHp     int
		maxHp         int
		initialDamage int
		event         bool
		want          int
		wantConsumed  bool
	}{
		"focus sash prevent 1-shot":                             {initialHp: 100, maxHp: 100, initialDamage: 150, event: true, want: 99, wantConsumed: true},
		"focus sash wont activate below max hp":                 {initialHp: 99, maxHp: 100, initialDamage: 150, event: true, want: 150, wantConsumed: false},
		"focus sash wont activate with damage less than max hp": {initialHp: 100, maxHp: 100, initialDamage: 99, event: true, want: 99, wantConsumed: false},
		"focus sash wont activate without choice item event":    {initialHp: 100, maxHp: 100, initialDamage: 150, event: false, want: 150, wantConsumed: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{
				Stats: []int{0, 0, 0, 0, 0, 0, 0, 0, 0},
				HP:    tc.initialHp,
			}
			mon.Stats[hitPoints] = tc.maxHp
			item, _ := RegisterItem(focusSash, &mon)
			mon.Item = item

			damage := tc.initialDamage
			if tc.event {
				mon.CheckItemTrigger(true, makeFocusSashEvent(&damage))
			} else {
				mon.CheckItemTrigger(true, nil)
			}

			if got := damage; got != tc.want {
				t.Errorf("damage = %d, want %d", got, tc.want)
			}

			if got := item.Consumed; got != tc.wantConsumed {
				if got {
					t.Errorf("focus sash was consumed at %d initial, %d max, and %d damage", tc.initialHp, tc.maxHp, tc.initialDamage)
				} else {
					t.Errorf("focus sash was not consumed at %d initial, %d max, and %d damage", tc.initialHp, tc.maxHp, tc.initialDamage)
				}
			}
		})
	}
}

func TestTypeBoostingItem(t *testing.T) {
	tests := map[string]struct {
		initialPower int
		item         itemState
		move         pokemonType
		event        bool
		want         int
	}{
		"silk scarf increases power of normal moves by 20%":              {initialPower: 40, item: silkScarf, move: normalType, event: true, want: 48},
		"silver powder increases power of bug moves by 20%":              {initialPower: 80, item: silverPowder, move: bugType, event: true, want: 96},
		"miracle seed increases power of grass moves by 20%":             {initialPower: 10, item: miracleSeed, move: grassType, event: true, want: 12},
		"dark glasses increases power of dark moves by 20%":              {initialPower: 25, item: blackGlasses, move: darkType, event: true, want: 30},
		"silk scarf does not increase power of bug moves":                {initialPower: 40, item: silkScarf, move: bugType, event: true, want: 40},
		"silk scarf does not increase power without move boosting event": {initialPower: 40, item: silkScarf, move: normalType, want: 40},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mon := Pokemon{}
			item, _ := RegisterItem(tc.item, &mon)
			mon.Item = item

			power := tc.initialPower
			if tc.event {
				mon.CheckItemTrigger(false, makeMoveBoostingEvent(tc.move, &power))
			} else {
				mon.CheckItemTrigger(false, nil)
			}

			if got := power; got != tc.want {
				t.Errorf("power = %d, want %d", got, tc.want)
			}

			if item.Consumed {
				t.Errorf("%s should not be consumed", tc.item.String())
			}
		})

	}
}
