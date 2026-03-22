package main

import (
	"fmt"
	"os"
	"slices"
)

type Parser struct {
	tokens  []Token
	current int
}

func NewParser(tokens []Token) Parser {
	return Parser{tokens, 0}
}

// program → declaration* EOF ;
func (p Parser) Parse() []Stmt {
	stmts := make([]Stmt, 0)

	for !p.isAtEnd() {
		s, err := p.declaration()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			p.synchronize()
		} else {
			stmts = append(stmts, s)
		}
	}

	return stmts
}

// declaration → varDecl | statement ;
func (p *Parser) declaration() (Stmt, error) {
	if p.match(VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

// varDecl → "var" IDENTIFIER ( "=" expression )? ";" ;
func (p *Parser) varDeclaration() (Stmt, error) {
	ident, err := p.consume(IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var init Expr
	if p.match(EQUAL) {
		if init, err = p.expression(); err != nil {
			return nil, err
		}
	}

	if err := p.expectSemicolon(); err != nil {
		return nil, err
	}

	return NewVar(ident, init), nil
}

// statement → exprStmt | ifStmt | whileStmt | printStmt | block ;
func (p *Parser) statement() (Stmt, error) {
	switch {
	case p.match(FOR):
		return p.forStatement()
	case p.match(WHILE):
		return p.whileStatement()
	case p.match(IF):
		return p.ifStatement()
	case p.match(PRINT):
		return p.printStatement()
	case p.match(LEFT_BRACE):
		return p.block()
	default:
		return p.expressionStatement()
	}
}

// forStmt → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";"  expression? ")" statement ;
func (p *Parser) forStatement() (Stmt, error) {
	var err error
	_, err = p.consume(LEFT_PAREN, "Expect '(' after 'for'.")
	if err != nil {
		return nil, err
	}

	// parse initializers
	var initializer Stmt
	switch {
	case p.match(SEMICOLON):
		// initializer is omitted
	case p.match(VAR):
		initializer, err = p.varDeclaration()
		if err != nil {
			return nil, err
		}
	default:
		initializer, err = p.expressionStatement()
		if err != nil {
			return nil, err
		}
	}

	// parse condition
	var condition Expr
	if !p.check(SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(SEMICOLON, "Expect ';' after loop condition.")
	if err != nil {
		return nil, err
	}

	// parse increment
	var increment Expr
	if !p.check(RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(RIGHT_PAREN, "Expect ')' after for clauses.")
	if err != nil {
		return nil, err
	}

	// parse body
	body, err := p.statement()
	if err != nil {
		return nil, err
	}

	if increment != nil {
		body = NewBlock([]Stmt{body, NewExpression(increment)})
	}
	if condition == nil {
		condition = NewLiteral(true)
	}
	body = NewWhile(condition, body)
	if initializer != nil {
		body = NewBlock([]Stmt{initializer, body})
	}

	return body, nil
}

// whileStmt → "while" "(" expression ")" statement ;
func (p *Parser) whileStatement() (Stmt, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after 'while'."); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after condition."); err != nil {
		return nil, err
	}
	body, err := p.statement()
	if err != nil {
		return nil, err
	}
	return NewWhile(condition, body), nil
}

// ifStmt → "if" "(" expression ")" statement ( "else" statement )? ;
func (p *Parser) ifStatement() (Stmt, error) {
	if _, err := p.consume(LEFT_PAREN, "Expect '(' after 'if'."); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(RIGHT_PAREN, "Expect ')' after if condition."); err != nil {
		return nil, err
	}
	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if p.match(ELSE) {
		if elseBranch, err = p.statement(); err != nil {
			return nil, err
		}
	}

	return NewIf(condition, thenBranch, elseBranch), nil
}

// block → "{" declaration* "}" ;
func (p *Parser) block() (Stmt, error) {
	stmts := make([]Stmt, 0)

	for !p.check(RIGHT_BRACE) && !p.isAtEnd() {
		s, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}

	if _, err := p.consume(RIGHT_BRACE, "Expect '}' after block."); err != nil {
		return nil, err
	}

	return NewBlock(stmts), nil
}

// printStmt → "print" expression ";" ;
func (p *Parser) printStatement() (Stmt, error) {
	expr, err := p.expression()
	if err != nil {
		return nil, err
	}
	if err := p.expectSemicolon(); err != nil {
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
	if err := p.expectSemicolon(); err != nil {
		return nil, err
	}
	return NewExpression(expr), nil
}

// expression → assignment ;
func (p *Parser) expression() (Expr, error) {
	return p.assignment()
}

// assignment → IDENTIFIER "=" assignment | logic_or ;
func (p *Parser) assignment() (Expr, error) {
	expr, err := p.logic_or()
	if err != nil {
		return nil, err
	}

	if p.match(EQUAL) {
		equal := p.previous()
		// assignment is right-associative, so recursively call assignment()
		// to have it resolves first before we create a new Assignment expression
		value, err := p.assignment()
		if err != nil {
			return nil, err
		}

		if variable, ok := expr.(Variable); ok {
			name := variable.Name
			return NewAssignment(name, value), nil
		}

		return nil, ErrorAtToken(equal, "Invalid assignment target.")
	}

	return expr, nil
}

// logic_or → logic_and ( "or" logic_and )* ;
func (p *Parser) logic_or() (Expr, error) {
	expr, err := p.logic_and()
	if err != nil {
		return nil, err
	}

	for p.match(OR) {
		operator := p.previous()
		right, err := p.logic_and()
		if err != nil {
			return nil, err
		}
		expr = NewLogical(expr, operator, right)

	}

	return expr, nil
}

// logic_and → equality ( "and" equality )* ;
func (p *Parser) logic_and() (Expr, error) {
	expr, err := p.equality()
	if err != nil {
		return nil, err
	}

	for p.match(AND) {
		operator := p.previous()
		right, err := p.equality()
		if err != nil {
			return nil, err
		}
		expr = NewLogical(expr, operator, right)

	}

	return expr, nil
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

// primary → NUMBER | STRING | "true" | "false" | "nil"| "(" expression ")" | IDENTIFIER ;
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

	if p.match(IDENTIFIER) {
		return NewVariable(p.previous()), nil
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

// expectSemicolon consumes a required ';' token.
// It returns an error when the current token is not a semicolon.
func (p *Parser) expectSemicolon() error {
	if _, err := p.consume(SEMICOLON, "Expect ';' after value."); err != nil {
		return err
	}
	return nil
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
func (p *Parser) synchronize() {
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
