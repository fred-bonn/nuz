package nuzengine

import (
	"slices"
	"strings"
	"sync"
)

// liveLearner runs one Monte Carlo worker without an iteration limit beside a live
// battle. The worker plays its own copy of the battle; rebase moves the root of
// that copy to the live state while keeping everything learned so far.
type liveLearner struct {
	la   *learningAiMonteCarlo
	bs   battleState
	root battleRoot

	// mu is held while an episode runs and while the root is replaced.
	mu       sync.Mutex
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
	started  bool
}

// optionHint describes the prompt currently shown to the player.
type optionHint struct {
	state string
	// keys are the action keys of the options, in display order.
	keys []string
	// composite marks double battle action prompts, whose learned actions cover both player slots.
	composite bool
	// second marks the prompt of the second player slot; first is the key chosen for the first slot.
	second bool
	first  string
}

func newLiveLearner(battleStateInt int, playerPartyStr, opponentPartyStr string, weatherInt int) (*liveLearner, error) {
	la := newLearningAIMonteCarlo()
	bs, err := initBattleState(battleStateInt, playerPartyStr, opponentPartyStr, 0, weatherInt, "", la)
	if err != nil {
		return nil, err
	}
	return &liveLearner{
		la:   la,
		bs:   bs,
		root: captureRoot(bs),
		stop: make(chan struct{}),
		done: make(chan struct{}),
	}, nil
}

func (l *liveLearner) start() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.started {
		return
	}
	l.started = true
	go l.run()
}

func (l *liveLearner) run() {
	defer close(l.done)
	for {
		select {
		case <-l.stop:
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

	l.root.apply(l.bs)
	l.la.pending = actionCandidate{}
	l.la.planner = doublePlanner{}
	l.la.beginEpisode()
	if err := l.bs.execute(); err != nil {
		return err
	}
	l.la.endEpisode(monteCarloReward(l.bs))
	return nil
}

// rebase makes the current state of src the root of all further episodes.
// It must be called from the goroutine that owns src.
func (l *liveLearner) rebase(src battleState) {
	root := captureRoot(src)
	l.mu.Lock()
	l.root = root
	l.mu.Unlock()
}

func (l *liveLearner) signalStop() {
	l.stopOnce.Do(func() { close(l.stop) })
}

func (l *liveLearner) close() {
	l.signalStop()
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
			consider(slices.Index(h.keys, action), entry)
			continue
		}
		first, second, _ := strings.Cut(action, compositeSeparator)
		if !h.second {
			consider(slices.Index(h.keys, first), entry)
		} else if first == h.first {
			consider(slices.Index(h.keys, second), entry)
		}
	}
	return best, best >= 0
}
