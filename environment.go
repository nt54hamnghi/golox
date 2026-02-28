package main

import (
	"fmt"
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

// Define adds or updates a variable in the environment.
// It does not check if the name already exists, so defining the same name redefines it.
// Defining a new variable always happens in the most inner scope, which is the current one.
func (e *Environment) Define(name string, value Object) {
	e.values[name] = value
}

// Assign updates an existing variable by name.
// It first checks the current environment, then walks outward through enclosing scopes.
// If no scope defines the variable, it returns a runtime error.
// Unlike Define, this method does not create new bindings.
func (e *Environment) Assign(name Token, value Object) error {
	if _, exist := e.values[name.Lexeme]; exist {
		e.values[name.Lexeme] = value
		return nil
	}

	if e.enclosing != nil {
		return e.enclosing.Assign(name, value)
	}

	return RuntimeError{
		name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	}
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
func (e Environment) Get(name Token) (Object, error) {
	if obj, exist := e.values[name.Lexeme]; exist {
		return obj, nil
	}

	if e.enclosing != nil {
		return e.enclosing.Get(name)
	}

	return nil, RuntimeError{
		name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	}
}
