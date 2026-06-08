package interpreter

import (
	"fmt"

	"github.com/nt54hamnghi/golox/internal/errors"
	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

type Environment struct {
	enclosing *Environment
	values    map[string]Object
}

// NewEnvironment creates the global-scope environment
func NewEnvironment() Environment {
	return Environment{
		values: make(map[string]Object),
	}
}

// NewEnclosedEnvinronment creates a local-scope environment nested inside a parent scope.
func NewEnclosedEnvinronment(enclosing *Environment) Environment {
	return Environment{
		enclosing,
		make(map[string]Object),
	}
}

// Define adds a variable in the environment.
// It does not check if the name already exists, so defining the same name redefines it.
// Defining a new variable always happens in the most inner scope, which is the current one.
func (e *Environment) Define(name string, value Object) {
	e.values[name] = value
}

// Assign updates an existing variable by name.
// It first checks the current environment, then walks outward through enclosing scopes.
// If no scope defines the variable, it returns a runtime error.
// Unlike Define, this method does not create new bindings.
func (e *Environment) Assign(name token.Token, value Object) error {
	if _, exist := e.values[name.Lexeme]; exist {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return undefinedVariable(name)
}

func (e Environment) AssignAt(distance int, name token.Token, value Object) error {
	e.ancestor(distance).values[name.Lexeme] = value
	return nil
}

// Get resolves a variable by name at evaluation time in lexical scope order.
// It starts in the current (innermost) environment, then walks outward through enclosing scopes.
// If no scope defines the variable, it returns a runtime error.
// Merely referring to a variable inside a function body is fine until that code is executed.
//
// Lox code example:
//
//	// runtime error: x is undefined when evaluated
//	print x;
//	var x = "too late!"
//
//	// no error yet: y is referenced, but f has not been called
//	fun f() { print y; }
func (e Environment) Get(name token.Token) (Object, error) {
	if obj, exist := e.values[name.Lexeme]; exist {
		return obj, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return nil, undefinedVariable(name)
}

func (e Environment) GetAt(distance int, name string) Object {
	return e.ancestor(distance).values[name]
}

func (e Environment) ancestor(distance int) Environment {
	curr := e
	for i := 0; i < distance; i++ {
		curr = *curr.enclosing
	}
	return curr
}

func undefinedVariable(name token.Token) errors.RuntimeError {
	return errors.RuntimeErrorAtToken(
		name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	)
}
