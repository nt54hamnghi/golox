package main

import (
	"slices"
)

type Parser[T any] struct {
	tokens   []Token
	current  int
	_phantom *T
}

func NewParser[T any](tokens []Token) Parser[T] {
	return Parser[T]{tokens, 0, new(T)}
}

func (p Parser[T]) Parse() (Expr[T], error) {
	return p.expression()
}

// expression → equality ;
func (p *Parser[T]) expression() (Expr[T], error) {
	return p.equality()
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser[T]) equality() (Expr[T], error) {
	expr, err := p.comparison()
	if err != nil {
		return nil, err
	}

	for p.match(BANG_EQUAL, EQUAL_EQUAL) {
		operator := p.previous()
		right, err := p.comparison()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right)
	}

	return expr, nil
}

// comparison → term ( ( ">" | ">=" | "<" | "<=" ) term )* ;
func (p *Parser[T]) comparison() (Expr[T], error) {
	expr, err := p.term()
	if err != nil {
		return nil, err
	}

	for p.match(GREATER, GREATER_EQUAL, LESS, LESS_EQUAL) {
		operator := p.previous()
		right, err := p.term()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right)
	}

	return expr, nil
}

// term → factor ( ( "-" | "+" ) factor )* ;
func (p *Parser[T]) term() (Expr[T], error) {
	expr, err := p.factor()
	if err != nil {
		return nil, err
	}

	for p.match(MINUS, PLUS) {
		operator := p.previous()
		right, err := p.factor()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right)
	}

	return expr, nil
}

// factor → unary ( ( "/" | "*" ) unary )* ;
func (p *Parser[T]) factor() (Expr[T], error) {
	expr, err := p.unary()
	if err != nil {
		return nil, err
	}

	for p.match(SLASH, STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right)
	}

	return expr, nil
}

// unary → ( "!" | "-" ) unary | primary ;
func (p *Parser[T]) unary() (Expr[T], error) {
	if p.match(BANG, MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right), nil
	}
	return p.primary()
}

// primary → NUMBER | STRING | "true" | "false" | "nil"| "(" expression ")" ;
func (p *Parser[T]) primary() (Expr[T], error) {
	if p.match(FALSE) {
		return NewLiteral[T](false), nil
	}
	if p.match(TRUE) {
		return NewLiteral[T](true), nil
	}
	if p.match(NIL) {
		return NewLiteral[T](nil), nil
	}
	if p.match(NUMBER, STRING) {
		return NewLiteral[T](p.previous().Literal), nil
	}

	if p.match(LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}
		return NewGrouping(expr), nil
	}

	return nil, p.error("Expect expression.")
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
// If so, it consumes and returns the token. Otherwise, it returns an error.
func (p *Parser[T]) consume(expected TokenType, message string) (Token, error) {
	if p.check(expected) {
		return p.advance(), nil
	}
	return Token{}, p.error(message)
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

func (p Parser[T]) error(message string) error {
	return ErrorAtToken(p.peek(), message)
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

		p.advance()
	}
}
