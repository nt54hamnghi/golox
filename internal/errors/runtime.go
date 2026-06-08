package errors

import (
	"fmt"

	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

// RuntimeError represents a runtime evaluation error at a specific token.
// It implements the error interface.
type RuntimeError struct {
	token   token.Token
	message string
}

// RuntimeErrorAtToken constructs a RuntimeError tied to a token location.
func RuntimeErrorAtToken(token token.Token, message string) RuntimeError {
	return RuntimeError{token, message}
}

// Error returns the runtime error message.
func (r RuntimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", r.message, r.token.Line)
}
