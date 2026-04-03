package main

import (
	"fmt"
)

type Object any

// The global environment for the interpreter.
var globals = NewEnvironment()

type Interpreter struct {
	// The currently entered environment.
	environment Environment
}

func (i *Interpreter) resolve(expr Expr, depth int) {
	panic("unimplemented")
}

func NewInterpreter() Interpreter {
	globals.Define("clock", NativeFun(Clock))
	return Interpreter{
		// the interpreter starts with the global environment
		// as its current environment.
		environment: globals,
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

func (i *Interpreter) executeBlock(stmts []Stmt, environment Environment) (any, error) {
	current := i.environment
	i.environment = environment
	defer func() {
		i.environment = current
	}()

	for _, s := range stmts {
		if _, err := i.execute(s); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

// VisitBlockStmt implements [StmtVisitor].
func (i *Interpreter) VisitBlockStmt(stmt Block) (any, error) {
	current := i.environment
	inner := NewEnclosedEnvinronment(&current)
	return i.executeBlock(stmt.Stmts, inner)
}

// VisitWhileStmt implements [StmtVisitor].
func (i *Interpreter) VisitWhileStmt(stmt While) (any, error) {
	for {
		condition, err := i.evaluate(stmt.Condition)
		if err != nil {
			return nil, err
		}
		if !isTruthy(condition) {
			return nil, nil
		}
		if _, err := i.execute(stmt.Body); err != nil {
			return nil, err
		}
	}
}

// VisitFunctionStmt implements [StmtVisitor].
func (i *Interpreter) VisitFunctionStmt(stmt Function) (any, error) {
	function := NewLoxFunction(stmt, i.environment)
	i.environment.Define(stmt.Name.Lexeme, function)
	return nil, nil
}

// VisitReturnStmt implements [StmtVisitor].
func (i *Interpreter) VisitReturnStmt(stmt Return) (any, error) {
	var (
		value Object
		err   error
	)
	if stmt.Value != nil {
		value, err = i.evaluate(stmt.Value)
		if err != nil {
			return nil, err
		}

	}

	// use ReturnIt error to unwind the call stack
	return nil, ReturnThis{value}
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

// VisitIfStmt implements [StmtVisitor].
func (i *Interpreter) VisitIfStmt(stmt If) (any, error) {
	condition, err := i.evaluate(stmt.Condition)
	if err != nil {
		return nil, err
	}

	if isTruthy(condition) {
		return i.execute(stmt.ThenBranch)
	} else if stmt.ElseBranch != nil {
		return i.execute(stmt.ElseBranch)
	}

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

// VisitCallExpr implements [ExprVisitor].
func (i *Interpreter) VisitCallExpr(expr Call) (any, error) {
	callee, err := i.evaluate(expr.Callee)
	if err != nil {
		return nil, err
	}

	args := make([]Object, 0)
	for _, argExpr := range expr.Arguments {
		arg, err := i.evaluate(argExpr)
		if err != nil {
			return nil, err
		}
		args = append(args, arg)
	}

	fun, ok := callee.(Callable)
	if !ok {
		return nil, RuntimeError{
			expr.Paren,
			"Can only call functions and classes.",
		}
	}
	if len(args) != fun.Arity() {
		return nil, RuntimeError{
			expr.Paren,
			fmt.Sprintf("Expected %d arguments but got %d.", fun.Arity(), len(args)),
		}
	}

	return fun.Call(i, args), nil
}

// VisitVariableExpr implements [ExprVisitor].
func (i Interpreter) VisitVariableExpr(expr Variable) (any, error) {
	return i.environment.Get(expr.Name)
}

// VisitLiteralExpr implements [ExprVisitor].
func (i Interpreter) VisitLiteralExpr(expr Literal) (any, error) {
	return expr.Value, nil
}

// VisitLogicalExpr implements [ExprVisitor].
func (i *Interpreter) VisitLogicalExpr(expr Logical) (any, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case OR:
		if isTruthy(left) {
			return left, nil
		}
	case AND:
		if !isTruthy(left) {
			return left, nil
		}
	default:
		panic(fmt.Sprintf("unexpected logical operator: %v", expr.Operator.Type))
	}

	return i.evaluate(expr.Right)
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
