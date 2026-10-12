package parser

import (
	"strings"
	"unicode"
)

type lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte

	line   int
	column int
}

func newLexer(input string) *lexer {
	l := &lexer{input: input}
	l.readChar()
	return l
}

func (l *lexer) newLexerError() error {
	return lexError{
		line:   l.line,
		column: l.column,
		char:   l.ch,
	}
}

func (l *lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

func (l *lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *lexer) nextToken() token {
	l.skipWhitespace()

	switch l.ch {
	case '\n':
		l.readChar()
		return token{Type: t_NEWLINE, Literal: "\\n"}

	case '@':
		l.readChar()
		return l.readItem()

	case ':':
		l.readChar()
		return l.readIdent()

	case '-':
		l.readChar()
		return token{Type: t_MOVE, Literal: "-"}

	case '/':
		l.readChar()
		l.skipWhitespace()
		return l.readIdent()

	case 0:
		return token{Type: t_EOF, Literal: ""}

	default:
		if isWordChar(l.ch) {
			return l.readIdent()
		}
	}

	tok := token{Type: t_ILLEGAL, Literal: string(l.ch)}
	l.readChar()
	return tok
}

func (l *lexer) readIdent() token {
	start := l.position

	for isWordChar(l.ch) || l.ch == ' ' {
		l.readChar()
	}

	lit := l.input[start:l.position]
	lit = normalize(lit)

	res := token{
		Type:    t_IDENT,
		Literal: lit,
	}

	if token, ok := literalMap[res.Literal]; ok {
		return token
	}

	return res
}

func (l *lexer) readItem() token {
	l.skipWhitespace()

	start := l.position

	for isWordChar(l.ch) || l.ch == ' ' {
		if l.ch == '\n' || l.ch == 0 {
			break
		}
		l.readChar()
	}

	lit := l.input[start:l.position]
	lit = normalize(lit)

	return token{
		Type:    t_ITEM,
		Literal: lit,
	}
}

func isWordChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) ||
		unicode.IsDigit(rune(ch)) ||
		ch == '-' ||
		ch == '\'' ||
		ch == '.'
}

func normalize(s string) string {
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, "'", "")
	s = strings.TrimSpace(s)
	return s
}
