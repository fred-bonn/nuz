package nuzengine

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type guidedAi struct {
	input   *bufio.Reader
	output  io.Writer
	pending actionCandidate
	// aborted is only set for interactive battles; when it reports true, an ended input aborts the battle.
	aborted func() bool
	// onPrompt, when set, runs before a choice prompt is printed; replacement is true for switch-in prompts.
	onPrompt func(bs battleState, s *slot, candidates []actionCandidate, replacement bool)
	// onChoices, when set, receives each choice prompt as structured data right before it is printed.
	onChoices func(title string, options []string)
}

func newGuidedAI(input io.Reader, output io.Writer) *guidedAi {
	if input == nil {
		input = os.Stdin
	}
	return &guidedAi{input: bufio.NewReader(input), output: output}
}

func (ga *guidedAi) evaluateActions(bs battleState, slot *slot, actions []*moveAction) (*moveAction, int) {
	candidates := buildCandidates(bs, slot, actions)
	if ga.onPrompt != nil {
		ga.onPrompt(bs, slot, candidates, false)
	}

	options := make([]string, len(candidates))
	for i, c := range candidates {
		if c.move != nil {
			options[i] = fmt.Sprintf("%s against %s", c.move.move.Move, c.move.targetSlot.mon.base.Name)
			continue
		}
		options[i] = fmt.Sprintf("switch to %s (%d/%d HP)", c.target.base.Name, c.target.hp, c.target.MaxHP())
	}
	ga.announce("Choose an action:", options)
	numberOfChoices := len(candidates)

	for {
		ga.print("> ")
		line, err := ga.input.ReadString('\n')
		if err != nil && len(line) == 0 {
			if ga.aborted != nil && ga.aborted() {
				bs.setError(fmt.Errorf("guided AI input ended before a valid choice was entered"))
				return actions[0], 1
			}
			ga.print("error: guided AI input ended before a valid choice was entered\n")
			continue
		}

		choice, err := parseChoice(line)
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

	if ga.onPrompt != nil {
		ga.onPrompt(bs, opponentSlot, buildSwitchCandidates(mons), true)
	}

	options := make([]string, len(mons))
	for i, mon := range mons {
		options[i] = fmt.Sprintf("%s (%d/%d HP)", mon.base.Name, mon.hp, mon.MaxHP())
	}
	ga.announce("Choose a replacement:", options)

	for {
		ga.print("> ")
		line, err := ga.input.ReadString('\n')
		if err != nil && len(line) == 0 {
			inputErr := fmt.Errorf("error: guided AI input ended before a valid replacement was chosen")
			ga.print("%s\n", inputErr)
			bs.setError(inputErr)
			return nil
		}

		choice, err := parseChoice(line)
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

var choicePattern = regexp.MustCompile(`^[0-9]{1,9}$`)

// parseChoice accepts a line consisting of a single non-negative number and nothing else
// (apart from the line terminator).
func parseChoice(line string) (int, error) {
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if !choicePattern.MatchString(line) {
		return 0, fmt.Errorf("input is not a single number")
	}
	return strconv.Atoi(line)
}

func (ga *guidedAi) shouldSwitch(bs battleState, slot *slot, score int, party []*pokemon) bool {
	if ga.aborted != nil && ga.aborted() {
		return false
	}
	return ga.pending.move == nil
}

func (ga *guidedAi) announce(title string, options []string) {
	if ga.onChoices != nil {
		ga.onChoices(title, options)
	}
	ga.print("%s\n", title)
	for i, option := range options {
		ga.print("%d. %s\n", i+1, option)
	}
}

func (ga *guidedAi) print(format string, args ...any) {
	output := ga.output
	if output == nil {
		output = os.Stdout
	}
	fmt.Fprintf(output, format, args...)
}
