package parser

import "fmt"

type pokemonLine int

const (
	pokemonLineName = iota
	pokemonLineLevel
	pokemonLineNature
	pokemonLineAbility
	pokemonLineStatus
	pokemonLinePredamage
	pokemonLineIVs
	pokemonLineMove
)

type parseError struct {
	message string
	line    pokemonLine
}

func (e parseError) Error() string {
	switch e.line {
	case pokemonLineName:
		e.message = fmt.Sprintf("%s, expected name line: <name> [@ item]", e.message)
	case pokemonLineLevel:
		e.message = fmt.Sprintf("%s, expected level line: Level: <number>", e.message)
	case pokemonLineNature:
		e.message = fmt.Sprintf("%s, expected nature line: <nature> Nature", e.message)
	case pokemonLineAbility:
		e.message = fmt.Sprintf("%s, expected ability line: Ability: <ability>", e.message)
	case pokemonLineStatus:
		e.message = fmt.Sprintf("%s, expected status line: Status: <status>", e.message)
	case pokemonLinePredamage:
		e.message = fmt.Sprintf("%s, expected predamage line: HP: <number>", e.message)
	case pokemonLineIVs:
		e.message = fmt.Sprintf("%s, expected IVs line: IVs: <number> <stat> / <number> <stat> / ...", e.message)
	case pokemonLineMove:
		e.message = fmt.Sprintf("%s, expected move line: - <move>", e.message)
	}
	return e.message
}
