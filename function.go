package main

import (
	"errors"
	"fmt"
	"os"
)

type LoxFunction struct {
	declaration Function
	closure     Environment
}

func NewLoxFunction(declaration Function, closure Environment) LoxFunction {
	return LoxFunction{declaration, closure}
}

func (lf LoxFunction) Call(interpreter *Interpreter, args []Object) Object {
	environment := NewEnclosedEnvinronment(&lf.closure)

	for i, p := range lf.declaration.Params {
		a := args[i]
		environment.Define(p.Lexeme, a)
	}

	_, err := interpreter.executeBlock(lf.declaration.Body, environment)
	if err == nil {
		return nil
	}

	var returnThis ReturnThis
	if ok := errors.As(err, &returnThis); !ok {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	return returnThis.Value
}

func (lf LoxFunction) Arity() int {
	return len(lf.declaration.Params)
}

func (lf LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", lf.declaration.Name.Lexeme)
}
