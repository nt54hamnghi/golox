package main

import "fmt"

type Environment struct {
	values map[string]Object
}

func NewEnvironment() Environment {
	return Environment{
		values: make(map[string]Object),
	}
}

// Define adds or updates a variable in the environment.
// It does not check if the name already exists, so defining the same name redefines it.
func (e *Environment) Define(name string, value Object) {
	e.values[name] = value
}

// Assign updates an existing variable in the current environment.
// If the variable is not already defined, it returns a runtime error.
// Unlike Define, this method does not create new bindings.
func (e *Environment) Assign(name Token, value Object) error {
	if _, exist := e.values[name.Lexeme]; exist {
		e.values[name.Lexeme] = value
		return nil
	}

	return RuntimeError{
		name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	}
}

// Get retrieves a variable when an identifier is evaluated.
// If the variable has not been defined at evaluation time, it returns a runtime error.
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

	return nil, RuntimeError{
		name,
		fmt.Sprintf("Undefined variable '%s'.", name.Lexeme),
	}
}
