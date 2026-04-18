package main

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
	Id() uint64
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt Expression) (any, error)
	VisitPrintStmt(stmt Print) (any, error)
	VisitVarStmt(stmt Var) (any, error)
	VisitFunctionStmt(stmt Function) (any, error)
	VisitIfStmt(stmt If) (any, error)
	VisitWhileStmt(stmt While) (any, error)
	VisitReturnStmt(stmt Return) (any, error)
	VisitBlockStmt(stmt Block) (any, error)
}

type Expression struct {
	Expression Expr
	Identity   uint64
}

func NewExpression(expression Expr) Expression {
	node := Expression{
		Expression: expression,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(self)
}
func (self Expression) Id() uint64 {
	return self.Identity
}

type Print struct {
	Expression Expr
	Identity   uint64
}

func NewPrint(expression Expr) Print {
	node := Print{
		Expression: expression,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Print) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(self)
}
func (self Print) Id() uint64 {
	return self.Identity
}

type Var struct {
	Name        Token
	Initializer Expr
	Identity    uint64
}

func NewVar(name Token, initializer Expr) Var {
	node := Var{
		Name:        name,
		Initializer: initializer,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Var) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitVarStmt(self)
}
func (self Var) Id() uint64 {
	return self.Identity
}

type Function struct {
	Name     Token
	Params   []Token
	Body     []Stmt
	Identity uint64
}

func NewFunction(name Token, params []Token, body []Stmt) Function {
	node := Function{
		Name:   name,
		Params: params,
		Body:   body,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Function) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitFunctionStmt(self)
}
func (self Function) Id() uint64 {
	return self.Identity
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
	Identity   uint64
}

func NewIf(condition Expr, thenBranch Stmt, elseBranch Stmt) If {
	node := If{
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self If) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitIfStmt(self)
}
func (self If) Id() uint64 {
	return self.Identity
}

type While struct {
	Condition Expr
	Body      Stmt
	Identity  uint64
}

func NewWhile(condition Expr, body Stmt) While {
	node := While{
		Condition: condition,
		Body:      body,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self While) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitWhileStmt(self)
}
func (self While) Id() uint64 {
	return self.Identity
}

type Return struct {
	Keyword  Token
	Value    Expr
	Identity uint64
}

func NewReturn(keyword Token, value Expr) Return {
	node := Return{
		Keyword: keyword,
		Value:   value,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Return) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitReturnStmt(self)
}
func (self Return) Id() uint64 {
	return self.Identity
}

type Block struct {
	Stmts    []Stmt
	Identity uint64
}

func NewBlock(stmts []Stmt) Block {
	node := Block{
		Stmts: stmts,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(self)
}
func (self Block) Id() uint64 {
	return self.Identity
}
