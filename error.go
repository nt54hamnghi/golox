package main

import "fmt"

type Report struct {
	line    int
	where   string
	message string
}

func (r Report) Error() string {
	return fmt.Sprintf("[line %d] Error%s: %s", r.line, r.where, r.message)
}

func ErrorAtLine(line int, message string) Report {
	return Report{line, "", message}
}

func ErrorAtToken(token Token, message string) Report {
	if token.Type == EOF {
		return Report{token.Line, " at end", message}
	} else {
		at := fmt.Sprintf(" at '%s'", token.Lexeme)
		return Report{token.Line, at, message}
	}
}
