package main

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
	Id() uint64
}

type ExprVisitor interface {
	VisitLiteralExpr(expr Literal) (any, error)
	VisitGroupingExpr(expr Grouping) (any, error)
	VisitUnaryExpr(expr Unary) (any, error)
	VisitVariableExpr(expr Variable) (any, error)
	VisitAssignmentExpr(expr Assignment) (any, error)
	VisitBinaryExpr(expr Binary) (any, error)
	VisitLogicalExpr(expr Logical) (any, error)
	VisitCallExpr(expr Call) (any, error)
}

type Literal struct {
	Value    any
	Identity uint64
}

func NewLiteral(value any) Literal {
	node := Literal{
		Value: value,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(self)
}
func (self Literal) Id() uint64 {
	return self.Identity
}

type Grouping struct {
	Expression Expr
	Identity   uint64
}

func NewGrouping(expression Expr) Grouping {
	node := Grouping{
		Expression: expression,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(self)
}
func (self Grouping) Id() uint64 {
	return self.Identity
}

type Unary struct {
	Operator Token
	Right    Expr
	Identity uint64
}

func NewUnary(operator Token, right Expr) Unary {
	node := Unary{
		Operator: operator,
		Right:    right,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(self)
}
func (self Unary) Id() uint64 {
	return self.Identity
}

type Variable struct {
	Name     Token
	Identity uint64
}

func NewVariable(name Token) Variable {
	node := Variable{
		Name: name,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(self)
}
func (self Variable) Id() uint64 {
	return self.Identity
}

type Assignment struct {
	Name     Token
	Value    Expr
	Identity uint64
}

func NewAssignment(name Token, value Expr) Assignment {
	node := Assignment{
		Name:  name,
		Value: value,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Assignment) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignmentExpr(self)
}
func (self Assignment) Id() uint64 {
	return self.Identity
}

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
	Identity uint64
}

func NewBinary(left Expr, operator Token, right Expr) Binary {
	node := Binary{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(self)
}
func (self Binary) Id() uint64 {
	return self.Identity
}

type Logical struct {
	Left     Expr
	Operator Token
	Right    Expr
	Identity uint64
}

func NewLogical(left Expr, operator Token, right Expr) Logical {
	node := Logical{
		Left:     left,
		Operator: operator,
		Right:    right,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Logical) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLogicalExpr(self)
}
func (self Logical) Id() uint64 {
	return self.Identity
}

type Call struct {
	Callee    Expr
	Paren     Token
	Arguments []Expr
	Identity  uint64
}

func NewCall(callee Expr, paren Token, arguments []Expr) Call {
	node := Call{
		Callee:    callee,
		Paren:     paren,
		Arguments: arguments,
	}
	node.Identity = nodeID.Add(1)
	return node
}

func (self Call) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitCallExpr(self)
}
func (self Call) Id() uint64 {
	return self.Identity
}
