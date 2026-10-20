package nuzengine

import "testing"

type testAction struct {
	name     string
	slot     *slot
	invoked  *[]string
	invokeFn func()
}

func (a *testAction) invoke(battleState) {
	*a.invoked = append(*a.invoked, a.name)
	if a.invokeFn != nil {
		a.invokeFn()
	}
}

func (a *testAction) prio(battleState) int {
	return 0
}

func (a *testAction) speed(bs battleState) int {
	return a.slot.mon.effectiveSpeed(bs)
}

func TestExecuteActionsRefreshesOrderWhenRequired(t *testing.T) {
	tests := map[string]struct {
		boosted   bool
		wantOrder []string
	}{
		"keeps the existing queue order when speed is unchanged": {
			wantOrder: []string{"fast", "slower", "boosted"},
		},
		"salac berry lets the boosted pokemon move next": {
			boosted:   true,
			wantOrder: []string{"fast", "boosted", "slower"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fast := testSwitchPokemon("fast", 100, 40, 100, 100, nil)
			boosted := testSwitchPokemon("boosted", 100, 23, 100, 100, nil)
			if tc.boosted {
				boosted.hp = 25
			}
			boosted.item, _ = registerItem(salacBerry, boosted)
			slower := testSwitchPokemon("slower", 100, 30, 100, 100, nil)

			invoked := make([]string, 0, 3)
			var bs *singleBattleState
			bs = &singleBattleState{
				actions: actionQueue{
					queue: priorityQueue[action]{
						&testAction{name: "fast", slot: &slot{mon: fast}, invoked: &invoked, invokeFn: func() {
							if tc.boosted {
								boosted.checkItemTrigger(bs, true, nil)
							}
						}},
						&testAction{name: "boosted", slot: &slot{mon: boosted}, invoked: &invoked},
						&testAction{name: "slower", slot: &slot{mon: slower}, invoked: &invoked},
					},
				},
			}
			bs.actions.sort(bs)

			if err := executeActions(bs); err != nil {
				t.Fatalf("executeActions() error = %v", err)
			}
			for i, want := range tc.wantOrder {
				if invoked[i] != want {
					t.Errorf("action %d = %q, want %q", i, invoked[i], want)
				}
			}
		})
	}
}
