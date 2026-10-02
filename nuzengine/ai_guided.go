package nuzengine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
)

type guidedAi struct {
	input   *bufio.Reader
	output  io.Writer
	pending actionCandidate
}

func newGuidedAI(input io.Reader, output io.Writer) *guidedAi {
	if input == nil {
		input = os.Stdin
	}
	return &guidedAi{input: bufio.NewReader(input), output: output}
}

func (ga *guidedAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	candidates := buildCandidates(bs, slot, actions)

	ga.print("Choose an action:\n")
	for i, c := range candidates {
		if c.move != nil {
			ga.print("%d. %s against %s\n", i+1, c.move.move.Move, c.move.targetSlot.mon.base.Name)
			continue
		}
		ga.print("%d. switch to %s (%d/%d HP)\n", i+1, c.target.base.Name, c.target.hp, c.target.MaxHP())
	}
	numberOfChoices := len(candidates)

	for {
		ga.print("> ")
		line, err := ga.input.ReadString('\n')
		if err != nil && len(line) == 0 {
			ga.print("error: guided AI input ended before a valid choice was entered\n")
			continue
		}

		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			ga.print("error: invalid input, please enter a number between 1 and %d\n", numberOfChoices)
			continue
		}
		if choice < 1 || choice > numberOfChoices {
			ga.print("error: choice out of range, please enter a number between 1 and %d\n", numberOfChoices)
			continue
		}

		chosen := candidates[choice-1]
		ga.pending = chosen
		if chosen.move != nil {
			return chosen.move, 1
		}
		return actions[0], -1
	}
}

func (ga *guidedAi) evaluteSwitchIns(bs battleState, mons []*pokemon, opponentSlot *slot) *pokemon {
	if ga.pending.target != nil && slices.Contains(mons, ga.pending.target) {
		return ga.pending.target
	}

	ga.print("Choose a replacement:\n")
	for i, mon := range mons {
		ga.print("%d. %s (%d/%d HP)\n", i+1, mon.base.Name, mon.hp, mon.MaxHP())
	}

	for {
		ga.print("> ")
		line, err := ga.input.ReadString('\n')
		if err != nil && len(line) == 0 {
			inputErr := fmt.Errorf("error: guided AI input ended before a valid replacement was chosen")
			ga.print("%s\n", inputErr)
			bs.setError(inputErr)
			return nil
		}

		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil {
			ga.print("error: invalid input, please enter a number between 1 and %d\n", len(mons))
			continue
		}
		if choice < 1 || choice > len(mons) {
			ga.print("error: choice out of range, please enter a number between 1 and %d\n", len(mons))
			continue
		}
		return mons[choice-1]
	}
}

func (ga *guidedAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	return ga.pending.move == nil
}

func (ga *guidedAi) print(format string, args ...any) {
	output := ga.output
	if output == nil {
		output = os.Stdout
	}
	fmt.Fprintf(output, format, args...)
}
