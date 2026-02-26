package main

import (
	"slices"
)

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) Parser {
	return Parser{tokens, 0}
}

// program → statement* EOF ;
func (p Parser) Parse() ([]Stmt, error) {
	stmts := make([]Stmt, 0)

	for !p.isAtEnd() {
		s, err := p.statement()
		if err != nil {
			return nil, err
		}

		stmts = append(stmts, s)
	}

	return stmts, nil
}

// statement → exprStmt | printStmt ;
func (p *Parser) statement() (Stmt, error) {
	if p.match(PRINT) {
		return p.printStatement()
	}
	return p.expressionStatement()
}

// printStmt → "print" expression ";" ;
func (p *Parser) printStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(SEMICOLON, "Expect ';' after value."); err != nil {
		return nil, err
	}
	return NewPrint(expr), nil
}

// exprStmt → expression ";" ;
func (p *Parser) expressionStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(SEMICOLON, "Expect ';' after value."); err != nil {
		return nil, err
	}
	return NewExpression(expr), nil
}

// expression → equality ;
func (p *Parser) expression() (Expr, error) {
	return p.equality()
}

// equality → comparison ( ( "!=" | "==" ) comparison )* ;
func (p *Parser) equality() (Expr, error) {
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
func (p *Parser) comparison() (Expr, error) {
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
func (p *Parser) term() (Expr, error) {
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
func (p *Parser) factor() (Expr, error) {
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
func (p *Parser) unary() (Expr, error) {
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
func (p *Parser) primary() (Expr, error) {
	if p.match(FALSE) {
		return NewLiteral(false), nil
	}
	if p.match(TRUE) {
		return NewLiteral(true), nil
	}
	if p.match(NIL) {
		return NewLiteral(nil), nil
	}
	if p.match(NUMBER, STRING) {
		return NewLiteral(p.previous().Literal), nil
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
func (p *Parser) match(types ...TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// consume checks if the next token is of the expected type.
// If so, it consumes and returns the token. Otherwise, it returns an error.
func (p *Parser) consume(expected TokenType, message string) (Token, error) {
	if p.check(expected) {
		return p.advance(), nil
	}
	return Token{}, p.error(message)
}

// check returns true if the current token is of the given type.
// It never consumes the token.
func (p Parser) check(expected TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == expected
}

// advance consumes the current token and returns it.
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we’ve run out of tokens to parse.
func (p Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

// peek returns the current token we have yet to consume.
func (p Parser) peek() Token {
	return p.tokens[p.current]
}

// previous returns the most recently consumed token.
func (p Parser) previous() Token {
	return p.tokens[p.current-1]
}

func (p Parser) error(message string) error {
	return ErrorAtToken(p.peek(), message)
}

// synchronize attempts to recover from a parsing error
// by discarding tokens until it has found a statement boundary.
func (p Parser) synchronize() {
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
