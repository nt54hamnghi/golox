package interpreter

import (
	"fmt"

	"github.com/nt54hamnghi/golox/internal/errors"
	"github.com/nt54hamnghi/golox/internal/parser"
	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

type Object any

// The global environment for the interpreter.
var globals = NewEnvironment()

type Interpreter struct {
	// The currently entered environment.
	environment Environment
	// A map of variable usages (via node identity) to
	// their resolved location in the environment stack.
	locals map[parser.NodeID]int
}

func (i *Interpreter) Resolve(expr parser.Expr, depth int) {
	i.locals[expr.Id()] = depth
}

func NewInterpreter() Interpreter {
	globals.Define("clock", NativeFun(Clock))
	return Interpreter{
		// the interpreter starts with the global environment as its current environment.
		environment: globals,
		locals:      make(map[parser.NodeID]int),
	}
}

func (i *Interpreter) Interpret(prog []parser.Stmt) error {
	for _, stmt := range prog {
		_, err := i.execute(stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func (i *Interpreter) execute(stmt parser.Stmt) (any, error) {
	return stmt.Accept(i)
}

func (i *Interpreter) evaluate(expr parser.Expr) (Object, error) {
	return expr.Accept(i)
}

func (i *Interpreter) executeBlock(stmts []parser.Stmt, environment Environment) (any, error) {
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

// VisitBlockStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitBlockStmt(stmt parser.Block) (any, error) {
	current := i.environment
	inner := NewEnclosedEnvinronment(&current)
	return i.executeBlock(stmt.Stmts, inner)
}

// VisitClassStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitClassStmt(stmt parser.Class) (any, error) {
	i.environment.Define(stmt.Name.Lexeme, nil)

	var superclass *LoxClass
	if stmt.Superclass != nil {
		obj, err := i.evaluate(stmt.Superclass)
		if err != nil {
			return nil, err
		}

		var ok bool
		if superclass, ok = obj.(*LoxClass); !ok {
			return nil, errors.RuntimeErrorAtToken(
				stmt.Superclass.Name,
				"Superclass must be a class.",
			)
		}

		current := i.environment
		i.environment = NewEnclosedEnvinronment(&current)
		i.environment.Define("super", superclass)
	}

	methods := make(map[string]LoxFunction)
	for _, method := range stmt.Methods {
		isInitializer := method.Name.Lexeme == "init"
		methods[method.Name.Lexeme] = NewLoxFunction(method, i.environment, isInitializer)
	}

	class := NewLoxClass(stmt.Name.Lexeme, superclass, methods)

	if stmt.Superclass != nil {
		i.environment = *i.environment.enclosing
	}
	i.environment.Assign(stmt.Name, class)
	return nil, nil
}

// VisitWhileStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitWhileStmt(stmt parser.While) (any, error) {
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

// VisitFunctionStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitFunctionStmt(stmt parser.Function) (any, error) {
	function := NewLoxFunction(stmt, i.environment, false)
	i.environment.Define(stmt.Name.Lexeme, function)
	return nil, nil
}

// VisitReturnStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitReturnStmt(stmt parser.Return) (any, error) {
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

// VisitVarStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitVarStmt(stmt parser.Var) (any, error) {
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

// VisitIfStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitIfStmt(stmt parser.If) (any, error) {
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

// VisitExpressionStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitExpressionStmt(stmt parser.Expression) (any, error) {
	_, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// VisitPrintStmt implements [parser.StmtVisitor].
func (i *Interpreter) VisitPrintStmt(stmt parser.Print) (any, error) {
	v, err := i.evaluate(stmt.Expression)
	if err != nil {
		return nil, err
	}
	fmt.Println(stringify(v))
	return nil, nil
}

// VisitAssignmentExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitAssignmentExpr(expr parser.Assignment) (any, error) {

	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}

	if distance, ok := i.locals[expr.Id()]; ok {
		if err := i.environment.AssignAt(distance, expr.Name, value); err != nil {
			return nil, err
		}
	} else {
		if err := globals.Assign(expr.Name, value); err != nil {
			return nil, err
		}
	}

	return value, nil
}

// VisitCallExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitCallExpr(expr parser.Call) (any, error) {
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
		return nil, errors.RuntimeErrorAtToken(
			expr.Paren,
			"Can only call functions and classes.",
		)
	}
	if len(args) != fun.Arity() {
		return nil, errors.RuntimeErrorAtToken(
			expr.Paren,
			fmt.Sprintf("Expected %d arguments but got %d.", fun.Arity(), len(args)),
		)
	}

	return fun.Call(i, args)
}

// VisitGetExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitGetExpr(expr parser.Get) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	instance, ok := obj.(LoxInstance)
	if !ok {
		return nil, errors.RuntimeErrorAtToken(
			expr.Name,
			"Only instances have properties.",
		)
	}

	return instance.Get(expr.Name)
}

// VisitSetExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitSetExpr(expr parser.Set) (any, error) {
	obj, err := i.evaluate(expr.Object)
	if err != nil {
		return nil, err
	}

	instance, ok := obj.(LoxInstance)
	if !ok {
		return nil, errors.RuntimeErrorAtToken(
			expr.Name,
			"Only instances have fields.",
		)
	}

	value, err := i.evaluate(expr.Value)
	if err != nil {
		return nil, err
	}
	instance.Set(expr.Name, value)

	return value, nil
}

// VisitSuperExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitSuperExpr(expr parser.Super) (any, error) {
	distance, ok := i.locals[expr.Id()]
	if !ok {
		panic("unresolved super expression")
	}

	obj := i.environment.GetAt(distance, "super")
	superclass, ok := obj.(*LoxClass)
	if !ok {
		panic(fmt.Sprintf("expected *LoxClass bound to 'super', got %T", obj))
	}

	obj = i.environment.GetAt(distance-1, "this")
	this, ok := obj.(LoxInstance)
	if !ok {
		panic(fmt.Sprintf("expected LoxInstance bound to 'this', got %T", obj))
	}

	method, ok := superclass.FindMethod(expr.Method.Lexeme)
	if !ok {
		return nil, errors.RuntimeErrorAtToken(
			expr.Method,
			fmt.Sprintf("Undefined property '%s'.", expr.Method.Lexeme),
		)
	}

	return method.bind(this), nil
}

// VisitThisExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitThisExpr(expr parser.This) (any, error) {
	return i.lookUpVariable(expr.Keyword, expr)
}

// VisitVariableExpr implements [parser.ExprVisitor].
func (i Interpreter) VisitVariableExpr(expr parser.Variable) (any, error) {
	return i.lookUpVariable(expr.Name, expr)
}

// lookUpVariable finds the resolved distance in the locals map.
// If we don’t find a distance, it must be global, so we look it up
// directly in the global environment.
func (i Interpreter) lookUpVariable(name token.Token, expr parser.Expr) (any, error) {
	if distance, ok := i.locals[expr.Id()]; ok {
		return i.environment.GetAt(distance, name.Lexeme), nil
	} else {
		return globals.Get(name)
	}
}

// VisitLiteralExpr implements [parser.ExprVisitor].
func (i Interpreter) VisitLiteralExpr(expr parser.Literal) (any, error) {
	return expr.Value, nil
}

// VisitLogicalExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitLogicalExpr(expr parser.Logical) (any, error) {
	left, err := i.evaluate(expr.Left)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.OR:
		if isTruthy(left) {
			return left, nil
		}
	case token.AND:
		if !isTruthy(left) {
			return left, nil
		}
	default:
		panic(fmt.Sprintf("unexpected logical operator: %v", expr.Operator.Type))
	}

	return i.evaluate(expr.Right)
}

// VisitGroupingExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitGroupingExpr(expr parser.Grouping) (any, error) {
	return i.evaluate(expr.Expression)
}

// VisitUnaryExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitUnaryExpr(expr parser.Unary) (any, error) {
	right, err := i.evaluate(expr.Right)
	if err != nil {
		return nil, err
	}

	switch expr.Operator.Type {
	case token.MINUS:
		if value, ok := right.(float64); ok {
			return -value, nil
		} else {
			return nil, errors.RuntimeErrorAtToken(
				expr.Operator,
				"Operand must be a number.",
			)
		}
	case token.BANG:
		return !isTruthy(right), nil
	}

	panic("unreachable")
}

// VisitBinaryExpr implements [parser.ExprVisitor].
func (i *Interpreter) VisitBinaryExpr(expr parser.Binary) (any, error) {
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
	case token.PLUS:
		if l, r, err := checkOperands[float64](left, right, expr.Operator); err == nil {
			return l + r, err
		}
		if l, r, err := checkOperands[string](left, right, expr.Operator); err == nil {
			return l + r, err
		}
	case token.MINUS, token.STAR, token.SLASH, token.GREATER, token.GREATER_EQUAL, token.LESS, token.LESS_EQUAL:
		l, r, err := checkOperands[float64](left, right, expr.Operator)
		if err != nil {
			return nil, err
		}
		switch op {
		case token.MINUS:
			return l - r, nil
		case token.STAR:
			return l * r, nil
		case token.SLASH:
			if r == 0 {
				return nil, errors.RuntimeErrorAtToken(
					expr.Operator,
					"Division by zero.",
				)
			}
			return l / r, nil
		case token.GREATER:
			return l > r, nil
		case token.GREATER_EQUAL:
			return l >= r, nil
		case token.LESS:
			return l < r, nil
		case token.LESS_EQUAL:
			return l <= r, nil
		}
	case token.BANG_EQUAL:
		return left != right, nil
	case token.EQUAL_EQUAL:
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
func checkOperands[T float64 | string](left, right Object, token token.Token) (T, T, error) {
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

	return zero, zero, errors.RuntimeErrorAtToken(
		token,
		fmt.Sprintf("Operands must be %ss.", typ),
	)
}

func stringify(obj Object) string {
	if obj == nil {
		return "nil"
	}
	return fmt.Sprint(obj)
}
