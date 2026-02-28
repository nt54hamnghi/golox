package main

import (
	"fmt"
	"strings"
)

type AstPrinter struct{}

func (p AstPrinter) String(expr Expr) string {
	repr, _ := expr.Accept(p)
	if v, ok := repr.(string); ok {
		return v
	} else {
		panic("AstPrinter: expected string result from expr.Accept")
	}
}

// VisitAssignmentExpr implements [ExprVisitor].
func (p AstPrinter) VisitAssignmentExpr(expr Assignment) (any, error) {
	panic("unimplemented")
}

// VisitVariableExpr implements [ExprVisitor].
func (p AstPrinter) VisitVariableExpr(expr Variable) (any, error) {
	return expr.Name.Lexeme, nil
}

// VisitLiteralExpr implements [ExprVisitor].
func (p AstPrinter) VisitLiteralExpr(expr Literal) (any, error) {
	if expr.Value == nil {
		return "nil", nil
	}
	return fmt.Sprintf("%v", expr.Value), nil
}

// VisitGroupingExpr implements [ExprVisitor].
func (p AstPrinter) VisitGroupingExpr(expr Grouping) (any, error) {
	return p.parenthesize("group", expr.Expression)
}

// VisitUnaryExpr implements [ExprVisitor].
func (p AstPrinter) VisitUnaryExpr(expr Unary) (any, error) {
	return p.parenthesize(expr.Operator.Lexeme, expr.Right)
}

// VisitBinaryExpr implements [ExprVisitor].
func (p AstPrinter) VisitBinaryExpr(expr Binary) (any, error) {
	return p.parenthesize(expr.Operator.Lexeme, expr.Left, expr.Right)
}

func (p AstPrinter) parenthesize(name string, expr ...Expr) (any, error) {
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
			NewLiteral(123),
		),
		NewToken(STAR, "*", nil, 0),
		NewGrouping(NewLiteral(45.67)),
	)
	repr := printer.String(expr)
	fmt.Println(repr)
}
