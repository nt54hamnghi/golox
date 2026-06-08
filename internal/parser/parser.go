package parser

import (
	"fmt"
	"os"
	"slices"

	"github.com/nt54hamnghi/golox/internal/errors"
	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

type Parser struct {
	tokens  []token.Token
	current int
}

func NewParser(tokens []token.Token) Parser {
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

// declaration → classDecl | funDecl | varDecl | statement ;
func (p *Parser) declaration() (Stmt, error) {
	if p.match(token.FUN) {
		return p.function("function")
	}
	if p.match(token.CLASS) {
		return p.classDeclaration()
	}
	if p.match(token.VAR) {
		return p.varDeclaration()
	}
	return p.statement()
}

// classDecl → "class" IDENTIFIER ( "<" IDENTIFIER )? "{" function* "}";
func (p *Parser) classDeclaration() (Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, "Expect superclass name.")
	if err != nil {
		return nil, err
	}

	var superclass *Variable
	if p.match(token.LESS) {
		name, err := p.consume(token.IDENTIFIER, "Expect class name.")
		if err != nil {
			return nil, err
		}
		variable := NewVariable(name)
		superclass = &variable
	}

	if _, err = p.consume(token.LEFT_BRACE, "Expect '{' before class body."); err != nil {
		return nil, err
	}

	methods := make([]Function, 0)
	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		stmt, err := p.function("method")
		if err != nil {
			return nil, err
		}
		method, ok := stmt.(Function)
		if !ok {
			panic("unexpected stmt type, while parsing class methods")
		}
		methods = append(methods, method)
	}

	if _, err = p.consume(token.RIGHT_BRACE, "Expect '}' after class body."); err != nil {
		return nil, err
	}

	return NewClass(name, superclass, methods), nil
}

func (p *Parser) function(kind string) (Stmt, error) {
	name, err := p.consume(token.IDENTIFIER, fmt.Sprintf("Expect %s name.", kind))
	if err != nil {
		return nil, err
	}

	_, err = p.consume(token.LEFT_PAREN, fmt.Sprintf("Expect '(' after %s name.", kind))
	if err != nil {
		return nil, err
	}

	params := make([]token.Token, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(params) >= 225 {
				err = errors.StaticErrorAtToken(p.peek(), "Can't have more than 255 parameters.")
				fmt.Fprint(os.Stderr, err.Error())
			}
			param, err := p.consume(token.IDENTIFIER, "Expect parameter name.")
			if err != nil {
				return nil, err
			}
			params = append(params, param)
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after parameters.")
	if err != nil {
		return nil, err
	}

	// `block` assumes that '{' has already be matched.
	// Also, consuming '{' here lets us report a more precise error message.
	_, err = p.consume(token.LEFT_BRACE, fmt.Sprintf("Expect '{' before %s body.", kind))
	if err != nil {
		return nil, err
	}
	body, err := p.block()
	if err != nil {
		return nil, err
	}

	return NewFunction(name, params, body), nil
}

// varDecl → "var" IDENTIFIER ( "=" expression )? ";" ;
func (p *Parser) varDeclaration() (Stmt, error) {
	ident, err := p.consume(token.IDENTIFIER, "Expect variable name.")
	if err != nil {
		return nil, err
	}

	var init Expr
	if p.match(token.EQUAL) {
		if init, err = p.expression(); err != nil {
			return nil, err
		}
	}

	if err := p.expectSemicolon(); err != nil {
		return nil, err
	}

	return NewVar(ident, init), nil
}

// statement → exprStmt | ifStmt | printStmt | returnStmt | whileStmt | forStmt | block ;
func (p *Parser) statement() (Stmt, error) {
	switch {
	case p.match(token.RETURN):
		return p.returnStatement()
	case p.match(token.FOR):
		return p.forStatement()
	case p.match(token.WHILE):
		return p.whileStatement()
	case p.match(token.IF):
		return p.ifStatement()
	case p.match(token.PRINT):
		return p.printStatement()
	case p.match(token.LEFT_BRACE):
		stmts, err := p.block()
		if err != nil {
			return nil, err
		}
		return NewBlock(stmts), nil
	default:
		return p.expressionStatement()
	}
}

// returnStmt → "return" expression? ";" ;
func (p *Parser) returnStatement() (Stmt, error) {
	keyword := p.previous()
	var (
		value Expr
		err   error
	)
	if !p.check(token.SEMICOLON) {
		value, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after return value.")
	if err != nil {
		return nil, nil
	}
	return NewReturn(keyword, value), nil
}

// forStmt → "for" "(" ( varDecl | exprStmt | ";" ) expression? ";"  expression? ")" statement ;
func (p *Parser) forStatement() (Stmt, error) {
	var err error
	_, err = p.consume(token.LEFT_PAREN, "Expect '(' after 'for'.")
	if err != nil {
		return nil, err
	}

	// parse initializers
	var initializer Stmt
	switch {
	case p.match(token.SEMICOLON):
		// initializer is omitted
	case p.match(token.VAR):
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
	if !p.check(token.SEMICOLON) {
		condition, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.SEMICOLON, "Expect ';' after loop condition.")
	if err != nil {
		return nil, err
	}

	// parse increment
	var increment Expr
	if !p.check(token.RIGHT_PAREN) {
		increment, err = p.expression()
		if err != nil {
			return nil, err
		}
	}
	_, err = p.consume(token.RIGHT_PAREN, "Expect ')' after for clauses.")
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
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'while'."); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after condition."); err != nil {
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
	if _, err := p.consume(token.LEFT_PAREN, "Expect '(' after 'if'."); err != nil {
		return nil, err
	}
	condition, err := p.expression()
	if err != nil {
		return nil, err
	}
	if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after if condition."); err != nil {
		return nil, err
	}
	thenBranch, err := p.statement()
	if err != nil {
		return nil, err
	}

	var elseBranch Stmt
	if p.match(token.ELSE) {
		if elseBranch, err = p.statement(); err != nil {
			return nil, err
		}
	}

	return NewIf(condition, thenBranch, elseBranch), nil
}

// block → "{" declaration* "}" ;
func (p *Parser) block() ([]Stmt, error) {
	stmts := make([]Stmt, 0)

	for !p.check(token.RIGHT_BRACE) && !p.isAtEnd() {
		s, err := p.declaration()
		if err != nil {
			return nil, err
		}
		stmts = append(stmts, s)
	}

	if _, err := p.consume(token.RIGHT_BRACE, "Expect '}' after block."); err != nil {
		return nil, err
	}

	return stmts, nil
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

	if p.match(token.EQUAL) {
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
		} else if get, ok := expr.(Get); ok {
			return NewSet(get.Object, get.Name, value), nil
		}

		return nil, errors.StaticErrorAtToken(equal, "Invalid assignment target.")
	}

	return expr, nil
}

// logic_or → logic_and ( "or" logic_and )* ;
func (p *Parser) logic_or() (Expr, error) {
	expr, err := p.logic_and()
	if err != nil {
		return nil, err
	}

	for p.match(token.OR) {
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

	for p.match(token.AND) {
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

	for p.match(token.BANG_EQUAL, token.EQUAL_EQUAL) {
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

	for p.match(token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL) {
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

	for p.match(token.MINUS, token.PLUS) {
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

	for p.match(token.SLASH, token.STAR) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		expr = NewBinary(expr, operator, right)
	}

	return expr, nil
}

// unary → ( "!" | "-" ) unary | call ;
func (p *Parser) unary() (Expr, error) {
	if p.match(token.BANG, token.MINUS) {
		operator := p.previous()
		right, err := p.unary()
		if err != nil {
			return nil, err
		}
		return NewUnary(operator, right), nil
	}
	return p.call()
}

// call → primary ( "(" arguments? ")" | "." IDENTIFIER )* ;
// arguments → expression ( "," expression )* ;
func (p *Parser) call() (Expr, error) {
	expr, err := p.primary()
	if err != nil {
		return nil, err
	}

	for {
		if p.match(token.LEFT_PAREN) {
			if expr, err = p.finishCall(expr); err != nil {
				return nil, err
			}
		} else if p.match(token.DOT) {
			name, err := p.consume(token.IDENTIFIER, "Expect property name after '.'.")
			if err != nil {
				return nil, err
			}
			expr = NewGet(expr, name)
		} else {
			break
		}
	}

	return expr, nil
}

func (p *Parser) finishCall(callee Expr) (Expr, error) {
	args := make([]Expr, 0)
	if !p.check(token.RIGHT_PAREN) {
		for {
			if len(args) >= 255 {
				err := errors.StaticErrorAtToken(p.peek(), "Can't have more than 255 arguments.")
				fmt.Fprint(os.Stderr, err.Error())
			}
			expr, err := p.expression()
			if err != nil {
				return nil, err
			}
			args = append(args, expr)
			if !p.match(token.COMMA) {
				break
			}
		}
	}

	paren, err := p.consume(token.RIGHT_PAREN, "Expect ')' after arguments.")
	if err != nil {
		return nil, err
	}

	return NewCall(callee, paren, args), nil
}

// primary → "true" | "false" | "nil" | "this"
//
//	| NUMBER | STRING | IDENTIFIER | "(" expression ")"
//	| "super" "." IDENTIFIER ;
func (p *Parser) primary() (Expr, error) {
	if p.match(token.FALSE) {
		return NewLiteral(false), nil
	}
	if p.match(token.TRUE) {
		return NewLiteral(true), nil
	}
	if p.match(token.NIL) {
		return NewLiteral(nil), nil
	}
	if p.match(token.NUMBER, token.STRING) {
		return NewLiteral(p.previous().Literal), nil
	}

	if p.match(token.LEFT_PAREN) {
		expr, err := p.expression()
		if err != nil {
			return nil, err
		}
		if _, err := p.consume(token.RIGHT_PAREN, "Expect ')' after expression."); err != nil {
			return nil, err
		}
		return NewGrouping(expr), nil
	}

	if p.match(token.SUPER) {
		keyword := p.previous()
		_, err := p.consume(token.DOT, "Expect '.' after 'super'.")
		if err != nil {
			return nil, err
		}
		method, err := p.consume(token.IDENTIFIER, "Expect superclass method name.")
		if err != nil {
			return nil, err
		}
		return NewSuper(keyword, method), nil
	}

	if p.match(token.THIS) {
		return NewThis(p.previous()), nil
	}

	if p.match(token.IDENTIFIER) {
		return NewVariable(p.previous()), nil
	}

	return nil, p.error("Expect expression.")
}

// match checks to see if the current token has any of the given types.
// If so, it consumes the token and returns true.
// Otherwise, it returns false, leaving the current token alone.
func (p *Parser) match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// consume checks if the next token is of the expected type.
// If so, it consumes and returns the token. Otherwise, it returns an error.
func (p *Parser) consume(expected token.TokenType, message string) (token.Token, error) {
	if p.check(expected) {
		return p.advance(), nil
	}
	return token.Token{}, p.error(message)
}

// expectSemicolon consumes a required ';' token.
// It returns an error when the current token is not a semicolon.
func (p *Parser) expectSemicolon() error {
	if _, err := p.consume(token.SEMICOLON, "Expect ';' after value."); err != nil {
		return err
	}
	return nil
}

// check returns true if the current token is of the given type.
// It never consumes the token.
func (p Parser) check(expected token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == expected
}

// advance consumes the current token and returns it.
func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

// isAtEnd checks if we’ve run out of tokens to parse.
func (p Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

// peek returns the current token we have yet to consume.
func (p Parser) peek() token.Token {
	return p.tokens[p.current]
}

// previous returns the most recently consumed token.
func (p Parser) previous() token.Token {
	return p.tokens[p.current-1]
}

func (p Parser) error(message string) error {
	return errors.StaticErrorAtToken(p.peek(), message)
}

// synchronize attempts to recover from a parsing error
// by discarding tokens until it has found a statement boundary.
func (p *Parser) synchronize() {
	p.advance()

	for !p.isAtEnd() {
		if p.previous().Type == token.SEMICOLON {
			return
		}

		switch p.peek().Type {
		case token.CLASS, token.FUN, token.VAR, token.FOR, token.IF, token.WHILE, token.PRINT, token.RETURN:
			return
		}

		p.advance()
	}
}
