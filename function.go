package main

import (
	"errors"
	"fmt"
	"os"
)

type LoxFunction struct {
	declaration   Function
	closure       Environment
	isInitializer bool
}

func NewLoxFunction(declaration Function, closure Environment, isInitializer bool) LoxFunction {
	return LoxFunction{
		declaration,
		closure,
		isInitializer,
	}
}

func (lf LoxFunction) bind(this LoxInstance) LoxFunction {
	env := NewEnclosedEnvinronment(&lf.closure)
	env.Define("this", this)
	return LoxFunction{
		lf.declaration,
		env,
		lf.isInitializer,
	}
}

// Call implements [LoxCallable].
func (lf LoxFunction) Call(interpreter *Interpreter, args []Object) Object {
	environment := NewEnclosedEnvinronment(&lf.closure)

	for i, p := range lf.declaration.Params {
		a := args[i]
		environment.Define(p.Lexeme, a)
	}

	var value Object
	_, err := interpreter.executeBlock(lf.declaration.Body, environment)
	if err == nil {
		if lf.isInitializer {
			value = lf.closure.GetAt(0, "this")
		}
		return value
	}

	var returnThis ReturnThis
	if ok := errors.As(err, &returnThis); ok {
		value = returnThis.Value
	} else {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	if lf.isInitializer {
		value = lf.closure.GetAt(0, "this")
	}
	return value
}

// Arity implements [LoxCallable].
func (lf LoxFunction) Arity() int {
	return len(lf.declaration.Params)
}

func (lf LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", lf.declaration.Name.Lexeme)
}
