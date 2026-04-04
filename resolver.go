package main

import "github.com/nt54hamnghi/golox/stack"

// scope is a map of variable names to boolean values.
// The boolean value indicates whether the initializer
// of a variable has been resolved.
type scope = map[string]bool

type Resolver struct {
	interpreter *Interpreter
	// A stack of scopes, representing nesting lexical scopes.
	// The innermost scope is at the top of the stack, and the
	// outermost scope is at the bottom.
	scopes stack.Stack[scope]
}

func NewResolver(interpreter *Interpreter) Resolver {
	return Resolver{
		interpreter: interpreter,
		scopes:      stack.NewStack[scope](),
	}
}

func (r *Resolver) resolveExpr(expr Expr) (any, error) {
	return expr.Accept(r)
}

func (r *Resolver) resolveStmt(stmt Stmt) (any, error) {
	return stmt.Accept(r)
}

func (r *Resolver) Resolve(stmts []Stmt) (any, error) {
	for _, s := range stmts {
		if _, err := s.Accept(r); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// beginScope enters a new scope by pushing a new scope onto the stack.
func (r *Resolver) beginScope() {
	s := make(scope)
	r.scopes.Push(s)
}

// endScope exits the current scope by popping it from the stack.
func (r *Resolver) endScope() {
	r.scopes.Pop()
}

// VisitBlockStmt implements [StmtVisitor].
func (r *Resolver) VisitBlockStmt(stmt Block) (any, error) {
	r.beginScope()
	if _, err := r.Resolve(stmt.Stmts); err != nil {
		return nil, err
	}
	r.endScope()
	return nil, nil
}

// VisitVarStmt implements [StmtVisitor].
func (r *Resolver) VisitVarStmt(stmt Var) (any, error) {
	r.declare(stmt.Name)
	if stmt.Initializer != nil {
		if _, err := r.resolveExpr(stmt.Initializer); err != nil {
			return nil, err
		}
	}
	r.define(stmt.Name)
	return nil, nil
}

func (r *Resolver) declare(name Token) {
	current, exist := r.scopes.Peek()
	if !exist {
		return
	}
	current[name.Lexeme] = false
}

func (r *Resolver) define(name Token) {
	current, exist := r.scopes.Peek()
	if !exist {
		return
	}
	current[name.Lexeme] = true
}

// VisitExpressionStmt implements [StmtVisitor].
func (r *Resolver) VisitExpressionStmt(stmt Expression) (any, error) {
	return r.resolveExpr(stmt.Expression)
}

// VisitFunctionStmt implements [StmtVisitor].
func (r *Resolver) VisitFunctionStmt(stmt Function) (any, error) {
	r.declare(stmt.Name)
	r.define(stmt.Name)
	r.resolveFunction(stmt)
	return nil, nil
}

func (r *Resolver) resolveFunction(fun Function) (any, error) {
	r.beginScope()
	for _, param := range fun.Params {
		r.declare(param)
		r.define(param)
	}
	if _, err := r.Resolve(fun.Body); err != nil {
		return nil, err
	}
	r.endScope()
	return nil, nil
}

// VisitIfStmt implements [StmtVisitor].
func (r *Resolver) VisitIfStmt(stmt If) (any, error) {
	if _, err := r.resolveExpr(stmt.Condition); err != nil {
		return nil, err
	}
	if _, err := r.resolveStmt(stmt.ThenBranch); err != nil {
		return nil, err
	}
	if stmt.ElseBranch != nil {
		if _, err := r.resolveStmt(stmt.ElseBranch); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// VisitPrintStmt implements [StmtVisitor].
func (r *Resolver) VisitPrintStmt(stmt Print) (any, error) {
	return r.resolveExpr(stmt.Expression)
}

// VisitReturnStmt implements [StmtVisitor].
func (r *Resolver) VisitReturnStmt(stmt Return) (any, error) {
	if stmt.Value != nil {
		return r.resolveExpr(stmt.Value)
	}
	return nil, nil
}

// VisitWhileStmt implements [StmtVisitor].
func (r *Resolver) VisitWhileStmt(stmt While) (any, error) {
	if _, err := r.resolveExpr(stmt.Condition); err != nil {
		return nil, err
	}
	if _, err := r.resolveStmt(stmt.Body); err != nil {
		return nil, err
	}
	return nil, nil
}

// VisitAssignmentExpr implements [ExprVisitor].
func (r *Resolver) VisitAssignmentExpr(expr Assignment) (any, error) {
	if _, err := r.resolveExpr(expr.Value); err != nil {
		return nil, err
	}
	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

// VisitBinaryExpr implements [ExprVisitor].
func (r *Resolver) VisitBinaryExpr(expr Binary) (any, error) {
	if _, err := r.resolveExpr(expr.Left); err != nil {
		return nil, err
	}
	if _, err := r.resolveExpr(expr.Right); err != nil {
		return nil, err
	}
	return nil, nil
}

// VisitCallExpr implements [ExprVisitor].
func (r *Resolver) VisitCallExpr(expr Call) (any, error) {
	if _, err := r.resolveExpr(expr.Callee); err != nil {
		return nil, err
	}
	for _, arg := range expr.Arguments {
		if _, err := r.resolveExpr(arg); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

// VisitGroupingExpr implements [ExprVisitor].
func (r *Resolver) VisitGroupingExpr(expr Grouping) (any, error) {
	return r.resolveExpr(expr.Expression)
}

// VisitLiteralExpr implements [ExprVisitor].
func (r *Resolver) VisitLiteralExpr(expr Literal) (any, error) {
	return nil, nil
}

// VisitLogicalExpr implements [ExprVisitor].
func (r *Resolver) VisitLogicalExpr(expr Logical) (any, error) {
	if _, err := r.resolveExpr(expr.Left); err != nil {
		return nil, err
	}
	if _, err := r.resolveExpr(expr.Right); err != nil {
		return nil, err
	}
	return nil, nil
}

// VisitUnaryExpr implements [ExprVisitor].
func (r *Resolver) VisitUnaryExpr(expr Unary) (any, error) {
	return r.resolveExpr(expr.Right)
}

// VisitVariableExpr implements [ExprVisitor].
func (r *Resolver) VisitVariableExpr(expr Variable) (any, error) {
	if current, exist := r.scopes.Peek(); exist {
		defined, ok := current[expr.Name.Lexeme]
		if ok && !defined {
			// !defined means it would be defined with this variable expression.
			// However, the variable name (i.e., the lexeme value) is the same for
			// both the variable being defined and its initializer, which we consider
			// to be an error.
			return nil, ErrorAtToken(
				expr.Name,
				"Can't read local variable in its own initializer",
			)
		}
	}
	r.resolveLocal(expr, expr.Name)
	return nil, nil
}

func (r *Resolver) resolveLocal(expr Expr, name Token) {
	for i, s := range r.scopes.All() {
		if _, ok := s[name.Lexeme]; ok {
			r.interpreter.resolve(expr, i)
			return
		}
	}
}
