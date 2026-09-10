package engine

import "testing"

func TestGetEffectivenessReturnsTheCorrectTypeMatchup(t *testing.T) {
	tests := map[string]struct {
		attacking pokemonType
		defending pokemonType
		want      effectiveness
	}{
		"normal against ghost is immune":              {attacking: normalType, defending: ghostType, want: immuneEffectivensss},
		"ghost against normal is immune":              {attacking: ghostType, defending: normalType, want: immuneEffectivensss},
		"fighting against normal is super effective":  {attacking: fightingType, defending: normalType, want: superEffectiveness},
		"psychic against fighting is super effective": {attacking: psychicType, defending: fightingType, want: superEffectiveness},
		"steel against psychic is neutral":            {attacking: steelType, defending: psychicType, want: normalEffectivenss},
		"bug against steel is resisted":               {attacking: bugType, defending: steelType, want: resistedEffectiveness},
		"ground against bug is resisted":              {attacking: groundType, defending: bugType, want: resistedEffectiveness},
		"water against ground is super effective":     {attacking: waterType, defending: groundType, want: superEffectiveness},
		"ice against water is resisted":               {attacking: iceType, defending: waterType, want: resistedEffectiveness},
		"fire against ice is super effective":         {attacking: FireType, defending: iceType, want: superEffectiveness},
		"fairy against fire is resisted":              {attacking: fairyType, defending: FireType, want: resistedEffectiveness},
		"dragon against fairy is immune":              {attacking: dragonType, defending: fairyType, want: immuneEffectivensss},
		"electric against dragon is resisted":         {attacking: electricType, defending: dragonType, want: resistedEffectiveness},
		"grass against electric is neutral":           {attacking: grassType, defending: electricType, want: normalEffectivenss},
		"rock against grass is neutral":               {attacking: rockType, defending: grassType, want: normalEffectivenss},
		"flying against rock is resisted":             {attacking: flyingType, defending: rockType, want: resistedEffectiveness},
		"dark against flying is neutral":              {attacking: darkType, defending: flyingType, want: normalEffectivenss},
		"poison against dark is neutral":              {attacking: poisonType, defending: darkType, want: normalEffectivenss},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := getEffectiveness(tc.attacking, tc.defending); got != tc.want {
				t.Errorf("%s: getEffectiveness(%s, %s) = %.1f, want %.1f", name, tc.attacking.String(), tc.defending.String(), got, tc.want)
			}
		})
	}
}
