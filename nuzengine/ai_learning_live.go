package nuzengine

import (
	"context"
	"slices"
	"strings"
	"sync"
)

// liveLearner runs one Monte Carlo worker without an iteration limit beside a live
// battle. The worker plays its own copy of the battle; rebase moves the reset point of
// that copy to the live state while keeping everything learned so far.
type liveLearner struct {
	la *learningAiMonteCarlo
	bs battleState

	// mu is held while an episode runs and while the reset point is replaced.
	mu      sync.Mutex
	cancel  context.CancelFunc
	done    chan struct{}
	started bool
}

// optionHint describes the prompt currently shown to the player.
type optionHint struct {
	state stateKey
	// actionKeys are the keys of the options, in display order.
	actionKeys []string
	// composite marks double battle action prompts, whose learned actions cover both player slots.
	composite bool
	// second marks the prompt of the second player slot; first is the key chosen for the first slot.
	second bool
	first  string
}

func newLiveLearner(ctx context.Context, battleStateInt int, playerPartyStr, opponentPartyStr string, weatherInt int) (*liveLearner, error) {
	la := newLearningAIMonteCarlo()
	bs, err := initBattleState(ctx, battleStateInt, playerPartyStr, opponentPartyStr, 0, weatherInt, "", la)
	if err != nil {
		return nil, err
	}
	return &liveLearner{
		la:   la,
		bs:   bs,
		done: make(chan struct{}),
	}, nil
}

// start launches the episode loop on a context derived from parent.
func (l *liveLearner) start(parent context.Context) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.started {
		return
	}
	l.started = true
	ctx, cancel := context.WithCancel(parent)
	l.cancel = cancel
	go l.run(ctx)
}

func (l *liveLearner) run(ctx context.Context) {
	defer close(l.done)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if err := l.episode(); err != nil {
			elogf("error: live learner stopped: %s", err)
			return
		}
	}
}

func (l *liveLearner) episode() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.bs.reset()
	l.la.pending = actionCandidate{}
	l.la.planner = doublePlanner{}
	l.la.beginEpisode()
	if err := l.bs.execute(); err != nil {
		return err
	}
	l.la.endEpisode(monteCarloReward(l.bs))
	return nil
}

// rebase makes the current state of src the reset point of all further episodes.
// It must be called from the goroutine that owns src.
func (l *liveLearner) rebase(src battleState) {
	l.mu.Lock()
	defer l.mu.Unlock()
	setResetPoint(l.bs, src, true)
}

// stop cancels the episode loop; it is safe to call more than once. Called before
// start, it stops nothing because no loop is running yet; a loop started afterwards
// is stopped by close, which Run always defers.
func (l *liveLearner) stop() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.cancel != nil {
		l.cancel()
	}
}

func (l *liveLearner) close() {
	l.stop()
	l.mu.Lock()
	started := l.started
	l.mu.Unlock()
	if started {
		<-l.done
	}
}

// bestOption returns the index of the option with the best learned value.
func (l *liveLearner) bestOption(h optionHint) (int, bool) {
	l.la.mu.RLock()
	defer l.la.mu.RUnlock()
	actions := l.la.q[h.state]

	best, bestValue := -1, 0.0
	consider := func(idx int, entry *qEntry) {
		if idx < 0 || entry.Count == 0 {
			return
		}
		if best < 0 || entry.Value > bestValue || (entry.Value == bestValue && idx < best) {
			best, bestValue = idx, entry.Value
		}
	}
	for action, entry := range actions {
		if !h.composite {
			consider(slices.Index(h.actionKeys, action), entry)
			continue
		}
		first, second, _ := strings.Cut(action, compositeSeparator)
		if !h.second {
			consider(slices.Index(h.actionKeys, first), entry)
		} else if first == h.first {
			consider(slices.Index(h.actionKeys, second), entry)
		}
	}
	return best, best >= 0
}
