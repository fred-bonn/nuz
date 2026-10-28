package nuzengine

import (
	"reflect"
	"testing"
)

func TestSnapshotSlots(t *testing.T) {
	newTrainer := func(player bool) trainer {
		return trainer{player: player, ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)}
	}
	party := func(prefix string, n int) []*pokemon {
		var res []*pokemon
		for i := 0; i < n; i++ {
			res = append(res, testSwitchPokemon(prefix+string(rune('a'+i)), 40+i, 100, 100, 100, nil))
		}
		return res
	}
	names := func(mons []PokemonSnapshot) []string {
		res := []string{}
		for _, m := range mons {
			res = append(res, m.Name)
		}
		return res
	}

	tests := []struct {
		name         string
		double       bool
		prepare      func(player, opponent []*pokemon)
		wantPActive  []string
		wantPBench   []string
		wantOActive  []string
		wantOBench   []string
		wantFirstHP  PokemonSnapshot
		wantOppFirst PokemonSnapshot
	}{
		{
			name:        "single battle",
			prepare:     func(p, o []*pokemon) { p[0].hp = 20; p[0].ailments[burnAilment] = generateAilment(burnAilment, nil) },
			wantPActive: []string{"pa"}, wantPBench: []string{"pb", "pc"},
			wantOActive: []string{"oa"}, wantOBench: []string{"ob", "oc"},
			wantFirstHP:  PokemonSnapshot{Name: "pa", HP: 20, MaxHP: 40, Ailments: []string{"burn"}},
			wantOppFirst: PokemonSnapshot{Name: "oa", HP: 40, MaxHP: 40, Ailments: []string{}},
		},
		{
			name:   "double battle with fainted bench mon and multiple ailments",
			double: true,
			prepare: func(p, o []*pokemon) {
				p[3].hp, p[3].fainted = 0, true
				o[0].ailments[poisonAilment] = generateAilment(poisonAilment, nil)
				o[0].ailments[confusionAilment] = generateAilment(confusionAilment, nil)
			},
			wantPActive: []string{"pa", "pb"}, wantPBench: []string{"pc", "pd"},
			wantOActive: []string{"oa", "ob"}, wantOBench: []string{"oc", "od"},
			wantFirstHP:  PokemonSnapshot{Name: "pa", HP: 40, MaxHP: 40, Ailments: []string{}},
			wantOppFirst: PokemonSnapshot{Name: "oa", HP: 40, MaxHP: 40, Ailments: []string{"poison", "confusion"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := 3
			if tt.double {
				n = 4
			}
			p, o := party("p", n), party("o", n)
			tt.prepare(p, o)
			var bs battleState
			var err error
			if tt.double {
				bs, err = initDoubleBattleState(newTrainer(true), newTrainer(false), p, o, 0)
			} else {
				bs, err = initSingleBattleState(newTrainer(true), newTrainer(false), p, o, 0)
			}
			if err != nil {
				t.Fatal(err)
			}
			snap := snapshotSlots(bs.getAllSlots(), bs.getPlayerTrainer())

			for _, c := range []struct {
				label string
				got   []PokemonSnapshot
				want  []string
			}{
				{"player active", snap.Player.Active, tt.wantPActive},
				{"player bench", snap.Player.Bench, tt.wantPBench},
				{"opponent active", snap.Opponent.Active, tt.wantOActive},
				{"opponent bench", snap.Opponent.Bench, tt.wantOBench},
			} {
				if got := names(c.got); !reflect.DeepEqual(got, c.want) {
					t.Errorf("%s = %v, want %v", c.label, got, c.want)
				}
			}
			if got := snap.Player.Active[0]; !reflect.DeepEqual(got, tt.wantFirstHP) {
				t.Errorf("player first = %+v, want %+v", got, tt.wantFirstHP)
			}
			if got := snap.Opponent.Active[0]; !reflect.DeepEqual(got, tt.wantOppFirst) {
				t.Errorf("opponent first = %+v, want %+v", got, tt.wantOppFirst)
			}
			if tt.double {
				if last := snap.Player.Bench[1]; !last.Fainted || last.HP != 0 {
					t.Errorf("expected fainted bench mon, got %+v", last)
				}
			}
		})
	}
}
