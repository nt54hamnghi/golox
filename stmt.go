package main

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type StmtVisitor interface {
	VisitExpressionStmt(expr Expression) (any, error)
	VisitPrintStmt(expr Print) (any, error)
}

type Expression struct {
	Expression Expr
}

func NewExpression(expression Expr) Expression {
	return Expression{
		Expression: expression,
	}
}

func (self Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(self)
}

type Print struct {
	Expression Expr
}

func NewPrint(expression Expr) Print {
	return Print{
		Expression: expression,
	}
}

func (self Print) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(self)
}

