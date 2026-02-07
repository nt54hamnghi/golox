package main

import (
	"errors"
	"slices"
)

var ParseError = errors.New("Parsing error")

type Parser[T any] struct {
	tokens   []Token
	current  int
	_phantom *T
}

func NewParser[T any](tokens []Token) Parser[T] {
	return Parser[T]{tokens, 0, new(T)}
}

func (p Parser[T]) Parse() (expr Expr[T]) {
	defer func() {
		r := recover()
		if err, ok := r.(error); ok && errors.Is(err, ParseError) {
			expr = nil
		}
	}()

	expr = p.expression()
	return expr
}

// expression → equality ;
func (p *Parser[T]) expression() Expr[T] {
	return p.equality()
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser[T]) equality() Expr[T] {
	expr := p.comparison()

	for p.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := p.previous()
		right := p.comparison()
		expr = NewBinary(expr, operator, right)
	}

	return expr
}

// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser[T]) comparison() Expr[T] {
	expr := p.term()

	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right := p.term()
		expr = NewBinary(expr, operator, right)
	}

	return expr
}

// term → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser[T]) term() Expr[T] {
	expr := p.factor()

	for p.match(MINUS, PLUS) {
		operator := p.previous()
		right := p.factor()
		expr = NewBinary(expr, operator, right)
	}

	return expr
}

// factor → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser[T]) factor() Expr[T] {
	expr := p.unary()

	for p.match(SLASH, STAR) {
		operator := p.previous()
		right := p.unary()
		expr = NewBinary(expr, operator, right)
	}

	return expr
}

// unary → ( "!" | "-" ) unary | primary ;
func (p *Parser[T]) unary() Expr[T] {
	if p.match(BANG, MINUS) {
		operator := p.previous()
		right := p.unary()
		return NewUnary(operator, right)
	}
	return p.primary()
}

// primary → NUMBER | STRING | "true" | "false" | "nil"| "(" expression ")" ;
func (p *Parser[T]) primary() Expr[T] {
	if p.match(FALSE) {
		return NewLiteral[T](false)
	}
	if p.match(TRUE) {
		return NewLiteral[T](true)
	}
	if p.match(NIL) {
		return NewLiteral[T](nil)
	}
	if p.match(NUMBER, STRING) {
		return NewLiteral[T](p.previous().Literal)
	}

	if p.match(LEFT_PAREN) {
		expr := p.expression()
		p.consume(RIGHT_PAREN, "Expext ')' after expression")
		return NewGrouping(expr)
	}

	err := p.error(p.peek(), "Expected expression.")
	panic(err)
}

// match checks to see if the current token has any of the given types.
// If so, it consumes the token and returns true.
// Otherwise, it returns false, leaving the current token alone.
func (p *Parser[T]) match(types ...TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// consume checks if the next token is of the expected type.
// If so, it consumes and returns the token. If not, it panics with an error.
func (p *Parser[T]) consume(expected TokenType, message string) Token {
	if p.check(expected) {
		return p.advance()
	}
	err := p.error(p.peek(), message)
	panic(err)
}

// check returns true if the current token is of the given type.
// It never consumes the token.
func (p Parser[T]) check(expected TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == expected
}

// advance consumes the current token and returns it.
func (p *Parser[T]) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we’ve run out of tokens to parse.
func (p Parser[T]) isAtEnd() bool {
	return p.peek().Type == EOF
}

// peek returns the current token we have yet to consume.
func (p Parser[T]) peek() Token {
	return p.tokens[p.current]
}

// previous returns the most recently consumed token.
func (p Parser[T]) previous() Token {
	return p.tokens[p.current-1]
}

func (p Parser[T]) error(token Token, message string) error {
	errorAtToken(token, message)
	return ParseError
}

// synchronize attempts to recover from a parsing error
// by discarding tokens until it has found a statement boundary.
func (p Parser[T]) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == SEMICOLON {
			return
		}

		switch p.peek().Type {
		case CLASS, FUN, VAR, FOR, IF, WHILE, PRINT, RETURN:
			return
		}
	}

	p.advance()
}
