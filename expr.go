package main

type Expr[T any] interface {
	Accept(visitor Visitor[T]) T
}

type Visitor[T any] interface {
	VisitBinaryExpr(expr Binary[T]) T
	VisitGroupingExpr(expr Grouping[T]) T
	VisitLiteralExpr(expr Literal[T]) T
	VisitUnaryExpr(expr Unary[T]) T
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

func (self Binary[T]) Accept(visitor Visitor[T]) T {
	return visitor.VisitBinaryExpr(self)
}

type Grouping[T any] struct {
	Expression Expr[T]
}

func NewGrouping[T any](expression Expr[T]) Grouping[T] {
	return Grouping[T]{
		Expression: expression,
	}
}

func (self Grouping[T]) Accept(visitor Visitor[T]) T {
	return visitor.VisitGroupingExpr(self)
}

type Literal[T any] struct {
	Value any
}

func NewLiteral[T any](value any) Literal[T] {
	return Literal[T]{
		Value: value,
	}
}

func (self Literal[T]) Accept(visitor Visitor[T]) T {
	return visitor.VisitLiteralExpr(self)
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

func (self Unary[T]) Accept(visitor Visitor[T]) T {
	return visitor.VisitUnaryExpr(self)
}
