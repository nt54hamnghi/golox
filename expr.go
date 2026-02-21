package main

type Expr[T any] interface {
	Accept(visitor Visitor[T]) (T, error)
}

type Visitor[T any] interface {
	VisitLiteralExpr(expr Literal[T]) (T, error)
	VisitGroupingExpr(expr Grouping[T]) (T, error)
	VisitUnaryExpr(expr Unary[T]) (T, error)
	VisitBinaryExpr(expr Binary[T]) (T, error)
}

type Literal[T any] struct {
	Value any
}

func NewLiteral[T any](value any) Literal[T] {
	return Literal[T]{
		Value: value,
	}
}

func (self Literal[T]) Accept(visitor Visitor[T]) (T, error) {
	return visitor.VisitLiteralExpr(self)
}

type Grouping[T any] struct {
	Expression Expr[T]
}

func NewGrouping[T any](expression Expr[T]) Grouping[T] {
	return Grouping[T]{
		Expression: expression,
	}
}

func (self Grouping[T]) Accept(visitor Visitor[T]) (T, error) {
	return visitor.VisitGroupingExpr(self)
}

type Unary[T any] struct {
	Operator Token
	Right    Expr[T]
}

func NewUnary[T any](operator Token, right Expr[T]) Unary[T] {
	return Unary[T]{
		Operator: operator,
		Right:    right,
	}
}

func (self Unary[T]) Accept(visitor Visitor[T]) (T, error) {
	return visitor.VisitUnaryExpr(self)
}

type Binary[T any] struct {
	Left     Expr[T]
	Operator Token
	Right    Expr[T]
}

func NewBinary[T any](left Expr[T], operator Token, right Expr[T]) Binary[T] {
	return Binary[T]{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
}

func (self Binary[T]) Accept(visitor Visitor[T]) (T, error) {
	return visitor.VisitBinaryExpr(self)
}
