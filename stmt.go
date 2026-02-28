package main

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt Expression) (any, error)
	VisitPrintStmt(stmt Print) (any, error)
	VisitVarStmt(stmt Var) (any, error)
	VisitBlockStmt(stmt Block) (any, error)
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

type Var struct {
	Name        Token
	Initializer Expr
}

func NewVar(name Token, initializer Expr) Var {
	return Var{
		Name:        name,
		Initializer: initializer,
	}
}

func (self Var) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitVarStmt(self)
}

type Block struct {
	Stmts []Stmt
}

func NewBlock(stmts []Stmt) Block {
	return Block{
		Stmts: stmts,
	}
}

func (self Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(self)
}
