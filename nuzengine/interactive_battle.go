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
	return ib, nil
}

// Run executes the battle, writing guided AI prompts to out. Each prompt ends with "> ".
func (ib *InteractiveBattle) Run(out io.Writer) error {
	ib.ai.output = out
	return ib.bs.execute()
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
	})
}

func (ib *InteractiveBattle) PlayerWon() bool {
	return !ib.bs.getPlayerTrainer().lost
}
