package parser

import (
	"fmt"
	"os"
)

func ReadShowdownFile(path string) ([]ParsedPokemon, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %s", "<path>")
	}

	return ParseShowdown(string(input))
}

// ParseShowdown parses Showdown-style party text directly, without reading it from disk.
func ParseShowdown(input string) ([]ParsedPokemon, error) {
	var tokens []token
	l := newLexer(input)

	for tok := l.nextToken(); tok.Type != t_EOF; tok = l.nextToken() {
		if tok.Type == t_ILLEGAL {
			return nil, l.newLexerError()
		}
		tokens = append(tokens, tok)
	}

	tokens = append(tokens, token{Type: t_EOF})

	var res []ParsedPokemon
	p := newParser(tokens)
	p.nextToken()

	for p.current.Type != t_EOF {
		pokemon, err := p.parsePokemon()
		if err != nil {
			return nil, fmt.Errorf("parser failed: %w", err)
		}

		res = append(res, pokemon)
	}

	return res, nil
}
