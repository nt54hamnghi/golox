package parser

import (
	"encoding/gob"
	"fmt"

	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
	Id() NodeID
}

func init() {
	gob.Register(Literal{})
	gob.Register(Call{})
	gob.Register(Get{})
	gob.Register(Set{})
	gob.Register(Super{})
	gob.Register(This{})
	gob.Register(Grouping{})
	gob.Register(Unary{})
	gob.Register(Variable{})
	gob.Register(Assignment{})
	gob.Register(Binary{})
	gob.Register(Logical{})
}

type ExprVisitor interface {
	VisitLiteralExpr(expr Literal) (any, error)
	VisitCallExpr(expr Call) (any, error)
	VisitGetExpr(expr Get) (any, error)
	VisitSetExpr(expr Set) (any, error)
	VisitSuperExpr(expr Super) (any, error)
	VisitThisExpr(expr This) (any, error)
	VisitGroupingExpr(expr Grouping) (any, error)
	VisitUnaryExpr(expr Unary) (any, error)
	VisitVariableExpr(expr Variable) (any, error)
	VisitAssignmentExpr(expr Assignment) (any, error)
	VisitBinaryExpr(expr Binary) (any, error)
	VisitLogicalExpr(expr Logical) (any, error)
}

type Literal struct {
	Value any
	id    NodeID
}

func NewLiteral(value any) Literal {
	node := Literal{
		Value: value,
	}

	tmp := struct{ Value any }{Value: node.Value}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(self)
}

func (self Literal) Id() NodeID {
	tmp := struct{ Value any }{Value: self.Value}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Call struct {
	Callee    Expr
	Paren     token.Token
	Arguments []Expr
	id        NodeID
}

func NewCall(callee Expr, paren token.Token, arguments []Expr) Call {
	node := Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}

	tmp := struct {
		Callee    Expr
		Paren     token.Token
		Arguments []Expr
	}{Callee: node.Callee, Paren: node.Paren, Arguments: node.Arguments}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Call) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(self)
}

func (self Call) Id() NodeID {
	tmp := struct {
		Callee    Expr
		Paren     token.Token
		Arguments []Expr
	}{Callee: self.Callee, Paren: self.Paren, Arguments: self.Arguments}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Get struct {
	Object Expr
	Name   token.Token
	id     NodeID
}

func NewGet(object Expr, name token.Token) Get {
	node := Get{
		Object: object,
		Name:   name,
	}

	tmp := struct {
		Object Expr
		Name   token.Token
	}{Object: node.Object, Name: node.Name}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Get) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGetExpr(self)
}

func (self Get) Id() NodeID {
	tmp := struct {
		Object Expr
		Name   token.Token
	}{Object: self.Object, Name: self.Name}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Set struct {
	Object Expr
	Name   token.Token
	Value  Expr
	id     NodeID
}

func NewSet(object Expr, name token.Token, value Expr) Set {
	node := Set{
		Object: object,
		Name:   name,
		Value:  value,
	}

	tmp := struct {
		Object Expr
		Name   token.Token
		Value  Expr
	}{Object: node.Object, Name: node.Name, Value: node.Value}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Set) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSetExpr(self)
}

func (self Set) Id() NodeID {
	tmp := struct {
		Object Expr
		Name   token.Token
		Value  Expr
	}{Object: self.Object, Name: self.Name, Value: self.Value}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Super struct {
	Keyword token.Token
	Method  token.Token
	id      NodeID
}

func NewSuper(keyword token.Token, method token.Token) Super {
	node := Super{
		Keyword: keyword,
		Method:  method,
	}

	tmp := struct {
		Keyword token.Token
		Method  token.Token
	}{Keyword: node.Keyword, Method: node.Method}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Super) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitSuperExpr(self)
}

func (self Super) Id() NodeID {
	tmp := struct {
		Keyword token.Token
		Method  token.Token
	}{Keyword: self.Keyword, Method: self.Method}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type This struct {
	Keyword token.Token
	id      NodeID
}

func NewThis(keyword token.Token) This {
	node := This{
		Keyword: keyword,
	}

	tmp := struct{ Keyword token.Token }{Keyword: node.Keyword}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self This) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitThisExpr(self)
}

func (self This) Id() NodeID {
	tmp := struct{ Keyword token.Token }{Keyword: self.Keyword}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Grouping struct {
	Expression Expr
	id         NodeID
}

func NewGrouping(expression Expr) Grouping {
	node := Grouping{
		Expression: expression,
	}

	tmp := struct{ Expression Expr }{Expression: node.Expression}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(self)
}

func (self Grouping) Id() NodeID {
	tmp := struct{ Expression Expr }{Expression: self.Expression}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Unary struct {
	Operator token.Token
	Right    Expr
	id       NodeID
}

func NewUnary(operator token.Token, right Expr) Unary {
	node := Unary{
		Operator: operator,
		Right:    right,
	}

	tmp := struct {
		Operator token.Token
		Right    Expr
	}{Operator: node.Operator, Right: node.Right}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(self)
}

func (self Unary) Id() NodeID {
	tmp := struct {
		Operator token.Token
		Right    Expr
	}{Operator: self.Operator, Right: self.Right}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Variable struct {
	Name token.Token
	id   NodeID
}

func NewVariable(name token.Token) Variable {
	node := Variable{
		Name: name,
	}

	tmp := struct{ Name token.Token }{Name: node.Name}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(self)
}

func (self Variable) Id() NodeID {
	tmp := struct{ Name token.Token }{Name: self.Name}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Assignment struct {
	Name  token.Token
	Value Expr
	id    NodeID
}

func NewAssignment(name token.Token, value Expr) Assignment {
	node := Assignment{
		Name:  name,
		Value: value,
	}

	tmp := struct {
		Name  token.Token
		Value Expr
	}{Name: node.Name, Value: node.Value}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Assignment) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignmentExpr(self)
}

func (self Assignment) Id() NodeID {
	tmp := struct {
		Name  token.Token
		Value Expr
	}{Name: self.Name, Value: self.Value}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Binary struct {
	Left     Expr
	Operator token.Token
	Right    Expr
	id       NodeID
}

func NewBinary(left Expr, operator token.Token, right Expr) Binary {
	node := Binary{
		Left:     left,
		Operator: operator,
		Right:    right,
	}

	tmp := struct {
		Left     Expr
		Operator token.Token
		Right    Expr
	}{Left: node.Left, Operator: node.Operator, Right: node.Right}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(self)
}

func (self Binary) Id() NodeID {
	tmp := struct {
		Left     Expr
		Operator token.Token
		Right    Expr
	}{Left: self.Left, Operator: self.Operator, Right: self.Right}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}

type Logical struct {
	Left     Expr
	Operator token.Token
	Right    Expr
	id       NodeID
}

func NewLogical(left Expr, operator token.Token, right Expr) Logical {
	node := Logical{
		Left:     left,
		Operator: operator,
		Right:    right,
	}

	tmp := struct {
		Left     Expr
		Operator token.Token
		Right    Expr
	}{Left: node.Left, Operator: node.Operator, Right: node.Right}
	node.id = NewNodeIDFrom(tmp)
	return node
}

func (self Logical) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(self)
}

func (self Logical) Id() NodeID {
	tmp := struct {
		Left     Expr
		Operator token.Token
		Right    Expr
	}{Left: self.Left, Operator: self.Operator, Right: self.Right}
	if nodeDigest(self.id.id, tmp) != self.id.digest {
		panic(fmt.Sprintf("node id hash mismatch, a copied value was modified: %#v", self))
	}
	return self.id
}
