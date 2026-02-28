package main

import (
	"fmt"
)

type Object any

type Interpreter struct {
	environment Environment
}

func NewInterpreter() Interpreter {
	return Interpreter{
		environment: NewEnvironment(),
	}
}

func (i *Interpreter) Interpret(prog []Stmt) error {
	for _, stmt := range prog {
		_, err := i.execute(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) execute(stmt Stmt) (any, error) {
	return stmt.Accept(i)
}

func (i *Interpreter) evaluate(expr Expr) (Object, error) {
	return expr.Accept(i)
}

func (i *Interpreter) executeBlock(block Block) (any, error) {
	outter := i.environment
	i.environment = NewEnclosedEnvinronment(&outter)
	defer func() {
		i.environment = outter
	}()

	for _, s := range block.Stmts {
		if _, err := i.execute(s); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// VisitBlockStmt implements [StmtVisitor].
func (i *Interpreter) VisitBlockStmt(stmt Block) (any, error) {
	return i.executeBlock(stmt)
}

// VisitVarStmt implements [StmtVisitor].
func (i *Interpreter) VisitVarStmt(stmt Var) (any, error) {
	var (
		value Object
		err   error
	)

	if stmt.Initializer != nil {
		value, err = i.evaluate(stmt.Initializer)
		if err != nil {
			return nil, err
		}
	}

	i.environment.Define(stmt.Name.Lexeme, value)
	return nil, nil
}

// VisitExpressionStmt implements [StmtVisitor].
func (i *Interpreter) VisitExpressionStmt(stmt Expression) (any, error) {
	_, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// VisitPrintStmt implements [StmtVisitor].
func (i *Interpreter) VisitPrintStmt(stmt Print) (any, error) {
	v, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(stringify(v))
	return nil, nil
}

// VisitAssignmentExpr implements [ExprVisitor].
func (i *Interpreter) VisitAssignmentExpr(expr Assignment) (any, error) {
	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	i.environment.Assign(expr.Name, value)
	return value, nil
}

// VisitVariableExpr implements [ExprVisitor].
func (i Interpreter) VisitVariableExpr(expr Variable) (any, error) {
	return i.environment.Get(expr.Name)
}

// VisitLiteralExpr implements [ExprVisitor].
func (i Interpreter) VisitLiteralExpr(expr Literal) (any, error) {
	return expr.Value, nil
}

// VisitGroupingExpr implements [ExprVisitor].
func (i *Interpreter) VisitGroupingExpr(expr Grouping) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitUnaryExpr implements [ExprVisitor].
func (i *Interpreter) VisitUnaryExpr(expr Unary) (any, error) {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case MINUS:
		if value, ok := right.(float64); ok {
			return -value, nil
		} else {
			return nil, RuntimeError{
				expr.Operator,
				"Operand must be a number.",
			}
		}
	case BANG:
		return !isTruthy(right), nil
	}

	panic("unreachable")
}

// VisitBinaryExpr implements [ExprVisitor].
func (i *Interpreter) VisitBinaryExpr(expr Binary) (any, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	op := expr.Operator.Type
	switch op {
	case PLUS:
		if l, r, err := checkOperands[float64](left, right, expr.Operator); err == nil {
			return l + r, err
		}
		if l, r, err := checkOperands[string](left, right, expr.Operator); err == nil {
			return l + r, err
		}
	case MINUS, STAR, SLASH, GREATER, GREATER_EQUAL, LESS, LESS_EQUAL:
		l, r, err := checkOperands[float64](left, right, expr.Operator)
		if err != nil {
			return nil, err
		}
		switch op {
		case MINUS:
			return l - r, nil
		case STAR:
			return l * r, nil
		case SLASH:
			if r == 0 {
				return nil, RuntimeError{
					expr.Operator,
					"Division by zero.",
				}
			}
			return l / r, nil
		case GREATER:
			return l > r, nil
		case GREATER_EQUAL:
			return l >= r, nil
		case LESS:
			return l < r, nil
		case LESS_EQUAL:
			return l <= r, nil
		}
	case BANG_EQUAL:
		return left != right, nil
	case EQUAL_EQUAL:
		// https://go.dev/ref/spec#Comparison_operators
		return left == right, nil
	}

	panic("unimplemented")
}

// isTruthy returns whether obj should be considered true in a boolean context.
// In Lox , false and nil are falsey, and everything else is truthy.
func isTruthy(obj Object) bool {
	if obj == nil {
		return false
	}
	if boolean, ok := obj.(bool); ok {
		return boolean
	}

	return true
}

// checkOperands verifies that both operands are the same expected runtime type.
// T is limited to float64 (number) or string and is used to type-assert both
// values. On success it returns the typed operands; otherwise it returns a
// RuntimeError associated with token.
func checkOperands[T float64 | string](left, right Object, token Token) (T, T, error) {
	if leftNum, ok := left.(T); ok {
		if rightNum, ok := right.(T); ok {
			return leftNum, rightNum, nil
		}
	}

	var (
		zero T
		typ  string
	)
	switch fmt.Sprintf("%T", zero) {
	case "float64":
		typ = "number"
	case "string":
		typ = "string"
	}

	return zero, zero, RuntimeError{
		token,
		fmt.Sprintf("Operands must be %ss.", typ),
	}
}

func stringify(obj Object) string {
	if obj == nil {
		return "nil"
	}
	return fmt.Sprint(obj)
}
