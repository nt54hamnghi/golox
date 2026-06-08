package errors

import (
	"fmt"

	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

// StaticError represents a scanner/parser error with source location context.
// It implements the error interface.
type StaticError struct {
	line    int
	where   string
	message string
}

// Error formats the report as:
// [line N] Error{where}: {message}
func (r StaticError) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s", r.line, r.where, r.message)
}

// StaticErrorAtLine constructs a Report tied to a specific line without token context.
func StaticErrorAtLine(line int, message string) StaticError {
	return StaticError{line, "", message}
}

// StaticErrorAtToken constructs a Report tied to a token location.
// If token is EOF, the location is reported as "at end";
// otherwise it is reported as "at '<lexeme>'".
func StaticErrorAtToken(t token.Token, message string) StaticError {
	if t.Type == token.EOF {
		return StaticError{t.Line, " at end", message}
	} else {
		at := fmt.Sprintf(" at '%s'", t.Lexeme)
		return StaticError{t.Line, at, message}
	}
}
