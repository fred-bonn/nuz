package parser

import "fmt"

type lexError struct {
	line   int
	column int
	char   byte
}

func (e lexError) Error() string {
	return fmt.Sprintf("lexer failed at %d:%d (illegal character: %q)", e.line, e.column, e.char)
}
