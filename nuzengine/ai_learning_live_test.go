package nuzengine

import (
	"errors"
	"testing"
	"time"
)

func testRootBattle(t *testing.T, ai ai) *singleBattleState {
	t.Helper()
	move := func(name string) *Move {
		return &Move{Move: name, Type: normalType, Power: 40, Accuracy: 100, PP: 10, MaxPP: 10, Class: physicalClass}
	}
	party := func(prefix string) []*pokemon {
		a := testSwitchPokemon(prefix+" a", 100, 50, 100, 100, move("tackle"))
		a.moves = append(a.moves, move("scratch"))
		b := testSwitchPokemon(prefix+" b", 100, 40, 100, 100, move("tackle"))
		return []*pokemon{a, b}
	}
	bs, err := initSingleBattleState(
		trainer{ai: ai, player: true, fieldEffects: make(map[fieldEffect]int)},
		trainer{ai: rnbAi{}, fieldEffects: make(map[fieldEffect]int)},
		party("player"), party("opponent"), 0,
	)
	if err != nil {
		t.Fatal(err)
	}
	return bs.(*singleBattleState)
}

func TestBattleRootApply(t *testing.T) {
	tests := map[string]struct {
		mutate func(bs *singleBattleState)
		check  func(t *testing.T, bs *singleBattleState)
	}{
		"hp and faint": {
			mutate: func(bs *singleBattleState) {
				bs.player.pokemonParty[1].hp = 0
				bs.player.pokemonParty[1].fainted = true
				bs.activeOpponentSlot.mon.hp = 7
			},
			check: func(t *testing.T, bs *singleBattleState) {
				if !bs.player.pokemonParty[1].fainted || bs.activeOpponentSlot.mon.hp != 7 {
					t.Errorf("hp/faint not restored")
				}
			},
		},
		"active mon, stages and pp": {
			mutate: func(bs *singleBattleState) {
				bs.activePlayerSlot.mon = bs.player.pokemonParty[1]
				bs.player.pokemonParty[0].stages[attack] = 2
				bs.player.pokemonParty[0].moves[1].PP = 3
				bs.player.pokemonParty[0].lockedMove = bs.player.pokemonParty[0].moves[1]
			},
			check: func(t *testing.T, bs *singleBattleState) {
				first := bs.player.pokemonParty[0]
				if bs.activePlayerSlot.mon != bs.player.pokemonParty[1] {
					t.Errorf("active mon = %s", bs.activePlayerSlot.mon.base.Name)
				}
				if first.stages[attack] != 2 || first.moves[1].PP != 3 || first.lockedMove != first.moves[1] {
					t.Errorf("stages=%v pp=%d locked=%v", first.stages, first.moves[1].PP, first.lockedMove)
				}
			},
		},
		"ailments, weather and field effects": {
			mutate: func(bs *singleBattleState) {
				bs.activeOpponentSlot.mon.ailments[burnAilment] = generateAilment(burnAilment, bs.activePlayerSlot)
				bs.weather = rainWeather
				bs.player.fieldEffects[spikesEffect] = 0
				bs.activePlayerSlot.protected = true
				bs.activePlayerSlot.protectTurns = 2
			},
			check: func(t *testing.T, bs *singleBattleState) {
				a, ok := bs.activeOpponentSlot.mon.ailments[burnAilment]
				if !ok || a.afflictedBy != bs.activePlayerSlot {
					t.Errorf("ailment = %+v", a)
				}
				if bs.weather != rainWeather {
					t.Errorf("weather = %v", bs.weather)
				}
				if _, ok := bs.player.fieldEffects[spikesEffect]; !ok {
					t.Errorf("field effect missing")
				}
				if !bs.activePlayerSlot.protected || bs.activePlayerSlot.protectTurns != 2 {
					t.Errorf("protect state not restored")
				}
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			bs := testRootBattle(t, rnbAi{})
			tc.mutate(bs)
			root := captureRoot(bs)
			wantKey := bs.key()

			// Diverge completely, then restore.
			bs.reset()
			bs.player.pokemonParty[0].hp = 1
			bs.activeOpponentSlot.mon.fainted = true
			bs.player.lost = true
			bs.setError(errors.New("boom"))
			root.apply(bs)

			if bs.getError() != nil || bs.player.lost {
				t.Errorf("error=%v lost=%v, want cleared", bs.getError(), bs.player.lost)
			}
			if got := bs.key(); got != wantKey {
				t.Errorf("key = %q, want %q", got, wantKey)
			}
			tc.check(t, bs)
		})
	}
}

func TestLiveLearnerBestOption(t *testing.T) {
	const state = "s"
	tests := map[string]struct {
		q      map[string]float64
		hint   optionHint
		want   int
		wantOk bool
	}{
		"no data": {
			hint: optionHint{state: state, keys: []string{"move:a", "move:b"}},
		},
		"picks highest value": {
			q:      map[string]float64{"move:a": -10, "move:b": -2},
			hint:   optionHint{state: state, keys: []string{"move:a", "move:b"}},
			want:   1,
			wantOk: true,
		},
		"ignores learned actions that are not options": {
			q:      map[string]float64{"move:a": -10, "move:z": 0},
			hint:   optionHint{state: state, keys: []string{"move:a", "move:b"}},
			want:   0,
			wantOk: true,
		},
		"ties pick the first option": {
			q:      map[string]float64{"move:a": -5, "move:b": -5},
			hint:   optionHint{state: state, keys: []string{"move:a", "move:b"}},
			want:   0,
			wantOk: true,
		},
		"composite first slot": {
			q:      map[string]float64{"move:a,move:x": -8, "move:b,move:x": -1, "move:c,move:x": 0},
			hint:   optionHint{state: state, keys: []string{"move:a", "move:b"}, composite: true},
			want:   1,
			wantOk: true,
		},
		"composite second slot follows the chosen first": {
			q:      map[string]float64{"move:a,move:x": -8, "move:a,move:y": -3, "move:b,move:x": 0},
			hint:   optionHint{state: state, keys: []string{"move:x", "move:y"}, composite: true, second: true, first: "move:a"},
			want:   1,
			wantOk: true,
		},
		"composite second slot after a skipped first slot": {
			q:      map[string]float64{"none,move:y": -3},
			hint:   optionHint{state: state, keys: []string{"move:x", "move:y"}, composite: true, second: true, first: noActionKey},
			want:   1,
			wantOk: true,
		},
		"composite second slot with no matching first": {
			q:    map[string]float64{"move:b,move:x": 0},
			hint: optionHint{state: state, keys: []string{"move:x"}, composite: true, second: true, first: "move:a"},
		},
		"other state": {
			q:    map[string]float64{"move:a": 0},
			hint: optionHint{state: "other", keys: []string{"move:a"}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			la := newLearningAIMonteCarlo()
			for action, value := range tc.q {
				la.q.update(state, action, value)
			}
			l := &liveLearner{la: la}
			got, ok := l.bestOption(tc.hint)
			if ok != tc.wantOk || (ok && got != tc.want) {
				t.Errorf("bestOption() = %d, %v; want %d, %v", got, ok, tc.want, tc.wantOk)
			}
		})
	}
}

func TestLiveLearnerKeepsLearningAcrossRebase(t *testing.T) {
	la := newLearningAIMonteCarlo()
	bs := testRootBattle(t, la)
	l := &liveLearner{la: la, bs: bs, root: captureRoot(bs), stop: make(chan struct{}), done: make(chan struct{})}
	rootKey := bs.key()

	total := func(state string) int {
		la.mu.RLock()
		defer la.mu.RUnlock()
		n := 0
		for _, e := range la.q[state] {
			n += e.Count
		}
		return n
	}
	waitFor := func(what string, cond func() bool) {
		t.Helper()
		for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); time.Sleep(time.Millisecond) {
			if cond() {
				return
			}
		}
		t.Fatalf("timed out waiting for %s", what)
	}

	l.start()
	defer l.close()
	waitFor("episodes from the initial root", func() bool { return total(rootKey) > 0 })

	live := testRootBattle(t, rnbAi{})
	live.activeOpponentSlot.mon.hp = 1
	l.rebase(live)

	before := total(rootKey)
	liveKey := live.key()
	waitFor("episodes from the new root", func() bool { return total(liveKey) > 0 })
	if got := total(rootKey); got < before {
		t.Errorf("learned count dropped from %d to %d", before, got)
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	if got := l.root.opponent.mons[0].hp; got != 1 {
		t.Errorf("root opponent hp = %d, want 1", got)
	}
}
