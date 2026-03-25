package main

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt Expression) (any, error)
	VisitPrintStmt(stmt Print) (any, error)
	VisitVarStmt(stmt Var) (any, error)
	VisitFunctionStmt(stmt Function) (any, error)
	VisitIfStmt(stmt If) (any, error)
	VisitWhileStmt(stmt While) (any, error)
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

type Function struct {
	Name   Token
	Params []Token
	Body   []Stmt
}

func NewFunction(name Token, params []Token, body []Stmt) Function {
	return Function{
		Name:   name,
		Params: params,
		Body:   body,
	}
}

func (self Function) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitFunctionStmt(self)
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
}

func NewIf(condition Expr, thenBranch Stmt, elseBranch Stmt) If {
	return If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
}

func (self If) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitIfStmt(self)
}

type While struct {
	Condition Expr
	Body      Stmt
}

func NewWhile(condition Expr, body Stmt) While {
	return While{
		Condition: condition,
		Body:      body,
	}
}

func (self While) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitWhileStmt(self)
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
