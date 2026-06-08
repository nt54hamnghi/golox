package parser

import (
	"encoding/gob"
	"fmt"

	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

type Stmt interface {
	Accept(visitor StmtVisitor) (any, error)
	Id() NodeID
}

func init() {
	gob.Register(Expression{})
	gob.Register(Print{})
	gob.Register(Var{})
	gob.Register(Class{})
	gob.Register(Function{})
	gob.Register(If{})
	gob.Register(While{})
	gob.Register(Return{})
	gob.Register(Block{})
}

type StmtVisitor interface {
	VisitExpressionStmt(stmt Expression) (any, error)
	VisitPrintStmt(stmt Print) (any, error)
	VisitVarStmt(stmt Var) (any, error)
	VisitClassStmt(stmt Class) (any, error)
	VisitFunctionStmt(stmt Function) (any, error)
	VisitIfStmt(stmt If) (any, error)
	VisitWhileStmt(stmt While) (any, error)
	VisitReturnStmt(stmt Return) (any, error)
	VisitBlockStmt(stmt Block) (any, error)
}

type Expression struct {
	Expression Expr
	id         NodeID
}

func NewExpression(expression Expr) Expression {
	node := Expression{
		Expression: expression,
	}

	tmp := struct{ Expression Expr }{Expression: node.Expression}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Expression) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitExpressionStmt(self)
}

func (self Expression) Id() NodeID {
	tmp := struct{ Expression Expr }{Expression: self.Expression}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Print struct {
	Expression Expr
	id         NodeID
}

func NewPrint(expression Expr) Print {
	node := Print{
		Expression: expression,
	}

	tmp := struct{ Expression Expr }{Expression: node.Expression}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Print) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitPrintStmt(self)
}

func (self Print) Id() NodeID {
	tmp := struct{ Expression Expr }{Expression: self.Expression}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Var struct {
	Name        token.Token
	Initializer Expr
	id          NodeID
}

func NewVar(name token.Token, initializer Expr) Var {
	node := Var{
		Name:        name,
		Initializer: initializer,
	}

	tmp := struct {
		Name        token.Token
		Initializer Expr
	}{Name: node.Name, Initializer: node.Initializer}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Var) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitVarStmt(self)
}

func (self Var) Id() NodeID {
	tmp := struct {
		Name        token.Token
		Initializer Expr
	}{Name: self.Name, Initializer: self.Initializer}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Class struct {
	Name       token.Token
	Superclass *Variable
	Methods    []Function
	id         NodeID
}

func NewClass(name token.Token, superclass *Variable, methods []Function) Class {
	node := Class{
		Name:       name,
		Superclass: superclass,
		Methods:    methods,
	}

	tmp := struct {
		Name       token.Token
		Superclass *Variable
		Methods    []Function
	}{Name: node.Name, Superclass: node.Superclass, Methods: node.Methods}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Class) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitClassStmt(self)
}

func (self Class) Id() NodeID {
	tmp := struct {
		Name       token.Token
		Superclass *Variable
		Methods    []Function
	}{Name: self.Name, Superclass: self.Superclass, Methods: self.Methods}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Function struct {
	Name   token.Token
	Params []token.Token
	Body   []Stmt
	id     NodeID
}

func NewFunction(name token.Token, params []token.Token, body []Stmt) Function {
	node := Function{
		Name:   name,
		Params: params,
		Body:   body,
	}

	tmp := struct {
		Name   token.Token
		Params []token.Token
		Body   []Stmt
	}{Name: node.Name, Params: node.Params, Body: node.Body}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Function) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitFunctionStmt(self)
}

func (self Function) Id() NodeID {
	tmp := struct {
		Name   token.Token
		Params []token.Token
		Body   []Stmt
	}{Name: self.Name, Params: self.Params, Body: self.Body}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type If struct {
	Condition  Expr
	ThenBranch Stmt
	ElseBranch Stmt
	id         NodeID
}

func NewIf(condition Expr, thenbranch Stmt, elsebranch Stmt) If {
	node := If{
		Condition:  condition,
		ThenBranch: thenbranch,
		ElseBranch: elsebranch,
	}

	tmp := struct {
		Condition  Expr
		ThenBranch Stmt
		ElseBranch Stmt
	}{Condition: node.Condition, ThenBranch: node.ThenBranch, ElseBranch: node.ElseBranch}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self If) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitIfStmt(self)
}

func (self If) Id() NodeID {
	tmp := struct {
		Condition  Expr
		ThenBranch Stmt
		ElseBranch Stmt
	}{Condition: self.Condition, ThenBranch: self.ThenBranch, ElseBranch: self.ElseBranch}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type While struct {
	Condition Expr
	Body      Stmt
	id        NodeID
}

func NewWhile(condition Expr, body Stmt) While {
	node := While{
		Condition: condition,
		Body:      body,
	}

	tmp := struct {
		Condition Expr
		Body      Stmt
	}{Condition: node.Condition, Body: node.Body}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self While) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitWhileStmt(self)
}

func (self While) Id() NodeID {
	tmp := struct {
		Condition Expr
		Body      Stmt
	}{Condition: self.Condition, Body: self.Body}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Return struct {
	Keyword token.Token
	Value   Expr
	id      NodeID
}

func NewReturn(keyword token.Token, value Expr) Return {
	node := Return{
		Keyword: keyword,
		Value:   value,
	}

	tmp := struct {
		Keyword token.Token
		Value   Expr
	}{Keyword: node.Keyword, Value: node.Value}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Return) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitReturnStmt(self)
}

func (self Return) Id() NodeID {
	tmp := struct {
		Keyword token.Token
		Value   Expr
	}{Keyword: self.Keyword, Value: self.Value}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Block struct {
	Stmts []Stmt
	id    NodeID
}

func NewBlock(stmts []Stmt) Block {
	node := Block{
		Stmts: stmts,
	}

	tmp := struct{ Stmts []Stmt }{Stmts: node.Stmts}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Block) Accept(visitor StmtVisitor) (any, error) {
	return visitor.VisitBlockStmt(self)
}

func (self Block) Id() NodeID {
	tmp := struct{ Stmts []Stmt }{Stmts: self.Stmts}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}
