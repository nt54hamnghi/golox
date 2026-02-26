package main

type Stmt[T any] interface {
	Accept(visitor StmtVisitor[T]) (T, error)
}

type StmtVisitor[T any] interface {
	VisitExpressionStmt(expr Expression[T]) (T, error)
	VisitPrintStmt(expr Print[T]) (T, error)
}

type Expression[T any] struct {
	Expression Expr[T]
}

func NewExpression[T any](expression Expr[T]) Expression[T] {
	return Expression[T]{
		Expression: expression,
	}
}

func (self Expression[T]) Accept(visitor StmtVisitor[T]) (T, error) {
	return visitor.VisitExpressionStmt(self)
}

type Print[T any] struct {
	Expression Expr[T]
}

func NewPrint[T any](expression Expr[T]) Print[T] {
	return Print[T]{
		Expression: expression,
	}
}

func (self Print[T]) Accept(visitor StmtVisitor[T]) (T, error) {
	return visitor.VisitPrintStmt(self)
}

