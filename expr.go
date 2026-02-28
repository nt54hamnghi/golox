package main

type Expr interface {
	Accept(visitor ExprVisitor) (any, error)
}

type ExprVisitor interface {
	VisitLiteralExpr(expr Literal) (any, error)
	VisitGroupingExpr(expr Grouping) (any, error)
	VisitUnaryExpr(expr Unary) (any, error)
	VisitVariableExpr(expr Variable) (any, error)
	VisitAssignmentExpr(expr Assignment) (any, error)
	VisitBinaryExpr(expr Binary) (any, error)
}

type Literal struct {
	Value any
}

func NewLiteral(value any) Literal {
	return Literal{
		Value: value,
	}
}

func (self Literal) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitLiteralExpr(self)
}

type Grouping struct {
	Expression Expr
}

func NewGrouping(expression Expr) Grouping {
	return Grouping{
		Expression: expression,
	}
}

func (self Grouping) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitGroupingExpr(self)
}

type Unary struct {
	Operator Token
	Right Expr
}

func NewUnary(operator Token, right Expr) Unary {
	return Unary{
		Operator: operator,
		Right: right,
	}
}

func (self Unary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitUnaryExpr(self)
}

type Variable struct {
	Name Token
}

func NewVariable(name Token) Variable {
	return Variable{
		Name: name,
	}
}

func (self Variable) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitVariableExpr(self)
}

type Assignment struct {
	Name Token
	Value Expr
}

func NewAssignment(name Token, value Expr) Assignment {
	return Assignment{
		Name: name,
		Value: value,
	}
}

func (self Assignment) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitAssignmentExpr(self)
}

type Binary struct {
	Left Expr
	Operator Token
	Right Expr
}

func NewBinary(left Expr, operator Token, right Expr) Binary {
	return Binary{
		Left: left,
		Operator: operator,
		Right: right,
	}
}

func (self Binary) Accept(visitor ExprVisitor) (any, error) {
	return visitor.VisitBinaryExpr(self)
}

