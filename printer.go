package main

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (p AstPrinter) String(expr Expr[string]) string {
	repr, _ := expr.Accept(p)
	return repr
}

func (p AstPrinter) VisitLiteralExpr(expr Literal[string]) (string, error) {
	if expr.Value == nil {
		return "nil", nil
	}
	return fmt.Sprintf("%v", expr.Value), nil
}

func (p AstPrinter) VisitGroupingExpr(expr Grouping[string]) (string, error) {
	return p.parenthesize("group", expr.Expression)
}

func (p AstPrinter) VisitUnaryExpr(expr Unary[string]) (string, error) {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p AstPrinter) VisitBinaryExpr(expr Binary[string]) (string, error) {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) parenthesize(name string, expr ...Expr[string]) (string, error) {
	var b strings.Builder

	b.WriteString("(" + name)
	for _, e := range expr {
		b.WriteString(" " + p.String(e))
	}
	b.WriteString(")")

	return b.String(), nil
}

func printExample() {
	var printer AstPrinter
	expr := NewBinary(
		NewUnary(
			NewToken(MINUS, "-", nil, 0),
			NewLiteral[string](123),
		),
		NewToken(STAR, "*", nil, 0),
		NewGrouping(NewLiteral[string](45.67)),
	)
	repr := printer.String(expr)
	fmt.Println(repr)
}
