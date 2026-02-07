package main

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (p AstPrinter) String(expr Expr[string]) string {
	return expr.Accept(p)
}

func (p AstPrinter) VisitBinaryExpr(expr Binary[string]) string {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) VisitGroupingExpr(expr Grouping[string]) string {
	return p.parenthesize("group", expr.Expression)
}

func (p AstPrinter) VisitLiteralExpr(expr Literal[string]) string {
	if expr.Value == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", expr.Value)
}

func (p AstPrinter) VisitUnaryExpr(expr Unary[string]) string {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

func (p AstPrinter) parenthesize(name string, expr ...Expr[string]) string {
	var b strings.Builder

	b.WriteString("(" + name)
	for _, e := range expr {
		b.WriteString(" " + p.String(e))
	}
	b.WriteString(")")

	return b.String()
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
