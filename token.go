package main

import (
	"fmt"
)

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
}

func NewToken(kind TokenType, lexeme string, literal any, line int) Token {
	return Token{
		kind,
		lexeme,
		literal,
		line,
	}
}

func (t Token) String() string {
	literal := "null"
	if t.Literal != nil {
		literal = fmt.Sprintf("%v", t.Literal)
	}

	return fmt.Sprintf("%s %s %v", t.Type, t.Lexeme, literal)
}
