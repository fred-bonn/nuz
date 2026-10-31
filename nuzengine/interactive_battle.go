package nuzengine

import (
	"io"
	"sync"
)

// InteractiveBattle is a battle whose player is driven by a guided AI over
// an in-memory channel instead of stdin/stdout.
type InteractiveBattle struct {
	bs        battleState
	ai        *guidedAi
	input     *io.PipeWriter
	inputRead *io.PipeReader
	closeOnce sync.Once
	closed    chan struct{}

	live      *liveLearner
	hintMu    sync.Mutex
	hint      *optionHint
	rootState string
}

func InitInteractiveBattle(battleStateInt int, playerPartyStr, opponentPartyStr string, weatherInt int) (*InteractiveBattle, error) {
	pr, pw := io.Pipe()
	ib := &InteractiveBattle{input: pw, inputRead: pr, closed: make(chan struct{})}
	ga := newGuidedAI(pr, io.Discard)
	ga.aborted = func() bool {
		select {
		case <-ib.closed:
			return true
		default:
			return false
		}
	}
	bs, err := initBattleState(battleStateInt, playerPartyStr, opponentPartyStr, 1, weatherInt, "", ga)
	if err != nil {
		return nil, err
	}
	ib.bs = bs
	ib.ai = ga

	// Hints are optional: without a learner the battle still works.
	live, err := newLiveLearner(battleStateInt, playerPartyStr, opponentPartyStr, weatherInt)
	if err != nil {
		elogf("error: failed initializing live learner: %s", err)
	} else {
		ib.live = live
		ga.onPrompt = ib.onPrompt
	}
	return ib, nil
}

// Run executes the battle, calling onPrompt on the engine goroutine each time the player must choose.
// Options are numbered from 1 when answering with Send.
func (ib *InteractiveBattle) Run(onPrompt func(title string, options []string)) error {
	ib.ai.onChoices = onPrompt
	if ib.live != nil {
		ib.live.start()
		defer ib.live.close()
	}
	return ib.bs.execute()
}

// onPrompt runs on the engine goroutine right before the player is asked to choose.
func (ib *InteractiveBattle) onPrompt(bs battleState, s *slot, candidates []actionCandidate, replacement bool) {
	h := optionHint{state: bs.key()}
	for _, c := range candidates {
		h.keys = append(h.keys, c.key)
	}

	if dbs, ok := bs.(*doubleBattleState); ok && !replacement {
		h.composite = true
		one := dbs.activePlayerSlotOne
		if s == one || one.mon.fainted || one.invulnerableAction != nil {
			ib.live.rebase(bs)
			ib.rootState = h.state
		}
		h.state = ib.rootState
		if s == dbs.activePlayerSlotTwo {
			h.second = true
			h.first = noActionKey
			if s != one && !one.mon.fainted && one.invulnerableAction == nil {
				h.first = ib.ai.pending.key
			}
		}
	} else if !replacement {
		ib.live.rebase(bs)
	}

	ib.hintMu.Lock()
	ib.hint = &h
	ib.hintMu.Unlock()
}

// BestOption returns the 1-based number of the option the learner currently rates best
// for the pending prompt. It is safe to call from any goroutine.
func (ib *InteractiveBattle) BestOption() (int, bool) {
	ib.hintMu.Lock()
	h := ib.hint
	ib.hintMu.Unlock()
	if h == nil || ib.live == nil {
		return 0, false
	}
	idx, ok := ib.live.bestOption(*h)
	return idx + 1, ok
}

// Send answers the pending prompt with a line of input; it blocks until the AI reads it or the battle is closed.
func (ib *InteractiveBattle) Send(line string) error {
	_, err := io.WriteString(ib.input, line+"\n")
	return err
}

// Close aborts the battle; a running Run returns with an error.
func (ib *InteractiveBattle) Close() {
	ib.closeOnce.Do(func() {
		close(ib.closed)
		ib.input.Close()
		ib.inputRead.Close()
		if ib.live != nil {
			ib.live.signalStop()
		}
	})
}

func (ib *InteractiveBattle) PlayerWon() bool {
	return !ib.bs.getPlayerTrainer().lost
}
