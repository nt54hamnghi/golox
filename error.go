package main

import "fmt"

// Report represents a scanner/parser error with source location context.
// It implements the error interface.
type Report struct {
	line    int
	where   string
	message string
}

// Error formats the report as:
// [line N] Error{where}: {message}
func (r Report) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s", r.line, r.where, r.message)
}

// ErrorAtLine constructs a Report tied to a specific line without token context.
func ErrorAtLine(line int, message string) Report {
	return Report{line, "", message}
}

// ErrorAtToken constructs a Report tied to a token location.
// If token is EOF, the location is reported as "at end";
// otherwise it is reported as "at '<lexeme>'".
func ErrorAtToken(token Token, message string) Report {
	if token.Type == EOF {
		return Report{token.Line, " at end", message}
	} else {
		at := fmt.Sprintf(" at '%s'", token.Lexeme)
		return Report{token.Line, at, message}
	}
}

// RuntimeError represents a runtime evaluation error at a specific token.
// It implements the error interface.
type RuntimeError struct {
	token   Token
	message string
}

// TODO: add a constructor for RuntimeError

// Error returns the runtime error message.
func (r RuntimeError) Error() string {
	return fmt.Sprintf("%s\n[line %d]", r.message, r.token.Line)
}
