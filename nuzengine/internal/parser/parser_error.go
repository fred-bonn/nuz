package parser

import "fmt"

type parseError struct {
	previous token
	current  token
}

func (e parseError) Error() string {
	return fmt.Sprintf("did not expect token %s after %s", e.current, e.previous)
}
