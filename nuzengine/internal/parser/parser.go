package parser

import (
	"fmt"
	"strconv"
	"strings"
)

type parser struct {
	tokens     []token
	current    token
	peek       token
	pos        int
	lineNumber int
}

func (p *parser) newParserError(line pokemonLine, message string) parseError {
	errorMessage := "line " + strconv.Itoa(p.lineNumber)
	if message != "" {
		errorMessage += ", " + message
	}

	return parseError{
		message: errorMessage,
		line:    line,
	}
}

func newParser(tokens []token) *parser {
	p := &parser{
		tokens:     tokens,
		peek:       tokens[0],
		current:    token{Type: t_START},
		lineNumber: 1,
	}

	return p
}

func (p *parser) parsePokemon() (ParsedPokemon, error) {
	res := ParsedPokemon{}
	res.IVs = make(map[string]int)

	p.consumeNewlines()

	// parse name @ item line
	err := p.parseNameLine(&res)
	if err != nil {
		return ParsedPokemon{}, err
	}

	// parse level line
	err = p.parseLevelLine(&res)
	if err != nil {
		return ParsedPokemon{}, err
	}

	// parse nature line
	err = p.parseNatureLine(&res)
	if err != nil {
		return ParsedPokemon{}, err
	}

	// parse ability line
	err = p.parseAbilityLine(&res)
	if err != nil {
		return ParsedPokemon{}, err
	}

	// parse status line
	if p.expectedToken(t_STATUS) {
		err = p.parseStatusLine(&res)
		if err != nil {
			return ParsedPokemon{}, err
		}
	}

	// parse predamage line
	if p.expectedToken(t_HP) {
		err = p.parsePredamageLine(&res)
		if err != nil {
			return ParsedPokemon{}, err
		}
	} else {
		res.HP = -1
	}

	// parse IVs line
	if p.expectedToken((t_IVS)) {
		err = p.parseIVsLine(&res)
		if err != nil {
			return ParsedPokemon{}, err
		}
	}

	// parse move lines, max 4
	var moves int
	for p.expectedToken(t_MOVE) && moves < 4 {
		err = p.parseMoveLine(&res)
		if err != nil {
			return ParsedPokemon{}, err
		}
		moves++
	}

	return res, nil
}

func (p *parser) parseNameLine(pokemon *ParsedPokemon) error {
	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineName, "")
	}

	pokemon.Name = p.current.Literal

	p.nextToken()

	if p.expectedToken(t_ITEM) {
		pokemon.Item = p.current.Literal

		p.nextToken()
	}

	p.consumeNewlines()

	return nil
}

func (p *parser) parseLevelLine(pokemon *ParsedPokemon) error {
	if !p.expectedToken(t_LEVEL) {
		return p.newParserError(pokemonLineLevel, "")
	}

	p.nextToken()

	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineLevel, "")
	}

	val, err := strconv.Atoi(p.current.Literal)
	if err != nil {
		if p.current.Literal == "" {
			return p.newParserError(pokemonLineLevel, fmt.Sprintf("level is missing for %s", pokemon.Name))
		}
		return p.newParserError(pokemonLineLevel, fmt.Sprintf("%s could not be converted to a number for %s", p.current.Literal, pokemon.Name))
	}

	pokemon.Level = val

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) parseNatureLine(pokemon *ParsedPokemon) error {
	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineNature, "")
	}

	parts := strings.Fields(p.current.Literal)

	if len(parts) < 2 {
		return p.newParserError(pokemonLineNature, fmt.Sprintf("nature line not enough fields for %s: '%s'", pokemon.Name, p.current.Literal))
	}
	if parts[1] != "Nature" {
		return p.newParserError(pokemonLineNature, fmt.Sprintf("incorrect nature line format for %s: '%s'", pokemon.Name, p.current.Literal))
	}
	if _, ok := validNature[parts[0]]; !ok {
		return p.newParserError(pokemonLineNature, fmt.Sprintf("unrecognized nature for %s: '%s'", pokemon.Name, p.current.Literal))
	}

	pokemon.Nature = strings.ToLower(parts[0])

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) parseAbilityLine(pokemon *ParsedPokemon) error {
	if !p.expectedToken(t_ABILITY) {
		return p.newParserError(pokemonLineAbility, "")
	}

	p.nextToken()

	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineAbility, "")
	}

	pokemon.Ability = p.current.Literal

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) parseStatusLine(pokemon *ParsedPokemon) error {
	p.nextToken()

	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineStatus, "")
	}
	if _, ok := validAilments[p.current.Literal]; !ok {
		return p.newParserError(pokemonLineStatus, fmt.Sprintf("unrecognized status for %s: '%s'", pokemon.Name, p.current.Literal))
	}

	pokemon.Status = strings.ToLower(p.current.Literal)

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) parsePredamageLine(pokemon *ParsedPokemon) error {
	p.nextToken()

	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLinePredamage, "")
	}

	val, err := strconv.Atoi(p.current.Literal)
	if err != nil {
		if p.current.Literal == "" {
			return p.newParserError(pokemonLinePredamage, fmt.Sprintf("predamage is missing for %s", pokemon.Name))
		}
		return p.newParserError(pokemonLinePredamage, fmt.Sprintf("%s could not be converted to a number for %s", p.current.Literal, pokemon.Name))
	}

	pokemon.HP = val

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) parseIVsLine(pokemon *ParsedPokemon) error {
	p.nextToken()

	for !p.expectedToken(t_NEWLINE) {
		if !p.expectedToken(t_IDENT) {
			return p.newParserError(pokemonLineIVs, "")
		}
		parts := strings.Fields(p.current.Literal)

		if len(parts) < 2 {
			return p.newParserError(pokemonLineIVs, fmt.Sprintf("iv not enough fields for %s: '%s'", pokemon.Name, p.current.Literal))
		}
		iv := parts[1]

		statName, ok := validStat[iv]
		if !ok {
			return p.newParserError(pokemonLineIVs, fmt.Sprintf("unrecognized iv for %s: '%s'", pokemon.Name, iv))
		}

		val, err := strconv.Atoi(parts[0])
		if err != nil {
			if p.current.Literal == "" {
				return p.newParserError(pokemonLineIVs, fmt.Sprintf("%s is missing for %s", iv, pokemon.Name))
			}
			return p.newParserError(pokemonLineIVs, fmt.Sprintf("%s could not be converted to a number for %s", parts[0], pokemon.Name))
		}

		pokemon.IVs[statName] = val

		p.nextToken()
	}

	p.consumeNewlines()
	return nil
}

func (p *parser) parseMoveLine(pokemon *ParsedPokemon) error {
	p.nextToken()

	if !p.expectedToken(t_IDENT) {
		return p.newParserError(pokemonLineMove, "")
	}

	pokemon.Moves = append(pokemon.Moves, p.current.Literal)

	p.nextToken()

	p.consumeNewlines()

	return nil
}

func (p *parser) consumeNewlines() {
	for p.current.Type == t_NEWLINE {
		p.nextToken()
		p.lineNumber++
	}
}

func (p *parser) nextToken() {
	p.current = p.peek

	if p.current.Type == t_EOF {
		return
	}

	p.pos++
	p.peek = p.tokens[p.pos]
}

func (p *parser) expectedToken(t tokenType) bool {
	return p.current.Type == t
}
