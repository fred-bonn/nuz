package engine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type guidedAi struct {
	input  *bufio.Reader
	output io.Writer
}

func NewGuidedAI(input io.Reader, output io.Writer) *guidedAi {
	if input == nil {
		input = os.Stdin
	}
	return &guidedAi{input: bufio.NewReader(input), output: output}
}

func (ga *guidedAi) evaluateActions(bs BattleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	ga.print("Choose an action:\n")
	for i, action := range actions {
		ga.print("%d. %s against %s\n", i+1, action.move.Name, action.targetSlot.mon.Base.Name)
	}
	numberOfChoices := len(actions)
	canSwitch := false
	if !slot.isTrapped() && canReplace(slot.Trainer.PokemonParty) {
		numberOfChoices++
		canSwitch = true
		ga.print("%d: switch\n", numberOfChoices)
	}

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
		if canSwitch && choice == numberOfChoices {
			return nil, -1
		}

		return actions[choice-1], 1
	}
}

func (ga *guidedAi) evaluteSwitchIns(bs BattleState, mons []*Pokemon, opponentSlot *slot) *Pokemon {
	ga.print("Choose a Pokemon to switch in:\n")
	for i, mon := range mons {
		ga.print("%d. %s (%d/%d HP)\n", i+1, mon.Base.Name, mon.HP, mon.MaxHP())
	}
	numberOfChoices := len(mons)

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

		return mons[choice-1]
	}
}

func (ga *guidedAi) shouldSwitch(bs BattleState, slot *slot, score int, party []*Pokemon) bool {
	return score == -1
}

func (ga *guidedAi) print(format string, args ...any) {
	output := ga.output
	if output == nil {
		output = os.Stdout
	}
	fmt.Fprintf(output, format, args...)
}
