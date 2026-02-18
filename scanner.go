package main

import (
	"fmt"
	"strconv"
	"strings"
)

var keyword map[string]TokenType = map[string]TokenType{
	"and":    AND,
	"class":  CLASS,
	"else":   ELSE,
	"false":  FALSE,
	"for":    FOR,
	"fun":    FUN,
	"if":     IF,
	"nil":    NIL,
	"or":     OR,
	"print":  PRINT,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"var":    VAR,
	"while":  WHILE,
}

type Scanner struct {
	// Raw source code
	source []rune
	// Slice of tokens to be filled as we scan the source code
	tokens []Token
	// Offset into the source code, pointing at the first character of the lexeme being scanned
	start int
	// Offset into the source code, pointing at the character being considered for the current lexeme
	current int
	// Line where the lexeme is located.
	line int
}

func NewScanner(src string) Scanner {
	return Scanner{
		source:  []rune(src),
		tokens:  []Token{},
		start:   0,
		current: 0,
		line:    1,
	}
}

func (s *Scanner) ScanTokens() ([]Token, error) {
	sErr := ScannerError{}

	for !s.isAtEnd() {
		s.start = s.current
		if err := s.scanToken(); err != nil {
			sErr = append(sErr, err)
		}
	}

	s.tokens = append(s.tokens, NewEOFToken(s.line))

	if sErr.empty() {
		return s.tokens, nil
	}

	return s.tokens, sErr
}

func (s *Scanner) scanToken() error {
	switch char := s.advanced(); char {
	case '(':
		s.addToken(LEFT_PAREN, nil)
	case ')':
		s.addToken(RIGHT_PAREN, nil)
	case '{':
		s.addToken(LEFT_BRACE, nil)
	case '}':
		s.addToken(RIGHT_BRACE, nil)
	case ',':
		s.addToken(COMMA, nil)
	case '.':
		s.addToken(DOT, nil)
	case '-':
		s.addToken(MINUS, nil)
	case '+':
		s.addToken(PLUS, nil)
	case ';':
		s.addToken(SEMICOLON, nil)
	case '*':
		s.addToken(STAR, nil)
	case '!':
		var typ TokenType
		if s.match('=') {
			typ = BANG_EQUAL
		} else {
			typ = BANG
		}
		s.addToken(typ, nil)
	case '=':
		var typ TokenType
		if s.match('=') {
			typ = EQUAL_EQUAL
		} else {
			typ = EQUAL
		}
		s.addToken(typ, nil)
	case '<':
		var typ TokenType
		if s.match('=') {
			typ = LESS_EQUAL
		} else {
			typ = LESS
		}
		s.addToken(typ, nil)
	case '>':
		var typ TokenType
		if s.match('=') {
			typ = GREATER_EQUAL
		} else {
			typ = GREATER
		}
		s.addToken(typ, nil)
	case '/':
		if s.match('/') {
			for s.peek() != '\n' && !s.isAtEnd() {
				s.advanced()
			}
		} else {
			s.addToken(SLASH, nil)
		}
	case ' ', '\r', '\t':
		// ignore whitespace
		return nil
	case '\n':
		s.line++
		return nil
	case '"':
		return s.string()
	default:
		if isDigit(char) {
			s.number()
		} else if isAlpha(char) {
			s.identifier()
		} else {
			return ErrorAtLine(s.line, "Unexpected character: "+string(char))
		}
	}

	return nil
}

func isDigit(char rune) bool {
	return '0' <= char && char <= '9'
}

func isAlpha(char rune) bool {
	return ('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_'
}

func isAlphaNumeric(char rune) bool {
	return isAlpha(char) || isDigit(char)
}

func (s *Scanner) advanced() rune {
	char := s.source[s.current]
	s.current++
	return char
}

func (s *Scanner) match(expected rune) bool {
	if s.isAtEnd() {
		return false
	}

	char := s.source[s.current]
	if char != expected {
		return false
	}

	s.current++
	return true
}

func (s *Scanner) peek() rune {
	if s.isAtEnd() {
		return 0 // null character
	}
	return s.source[s.current]
}

func (s *Scanner) string() error {
	for s.peek() != '"' && !s.isAtEnd() {
		// support for multi-line string, updating line
		// counter when encountering a newline
		if s.peek() == '\n' {
			s.line += 1
		}
		s.advanced()
	}

	if s.isAtEnd() {
		return ErrorAtLine(s.line, "Unterminated string.")
	}

	// consume the closing "
	s.advanced()

	// trim the surrounding quotes
	value := s.source[s.start+1 : s.current-1]
	s.addToken(STRING, string(value))

	return nil
}

func (s *Scanner) number() {
	for isDigit(s.peek()) {
		s.advanced()
	}

	if s.peek() == '.' && isDigit(s.peekNext()) {
		// consume the .
		s.advanced()
		for isDigit(s.peek()) {
			s.advanced()
		}
	}

	value := s.source[s.start:s.current]
	number, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		panic("lexeme is not a number")
	}
	s.addToken(NUMBER, number)
}

func (s *Scanner) identifier() {
	for isAlphaNumeric(s.peek()) {
		s.advanced()
	}

	value := string(s.source[s.start:s.current])
	if typ, ok := keyword[value]; ok {
		s.addToken(typ, nil)
	} else {
		s.addToken(IDENTIFIER, nil)
	}
}

func (s *Scanner) peekNext() rune {
	if s.current+1 >= len(s.source) {
		return 0
	}
	return s.source[s.current+1]
}

func (s *Scanner) addToken(typ TokenType, literal any) {
	text := string(s.source[s.start:s.current])
	token := NewToken(typ, text, literal, s.line)
	s.tokens = append(s.tokens, token)
}

func (s Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

// ScannerError is a collection of errors that occurred during scanning.
type ScannerError []error

// Error implements the error interface, returning a string representation
// of all errors in the collection.
func (se ScannerError) Error() string {
	var b strings.Builder

	for _, err := range se {
		if err != nil {
			fmt.Fprintln(&b, err.Error())
		}
	}

	return strings.TrimSpace(b.String())
}

// empty returns true if all errors are nil
func (se ScannerError) empty() bool {
	empty := true
	for _, err := range se {
		empty = empty && (err == nil)
	}
	return empty
}
