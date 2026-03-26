package main

import (
	"errors"
	"fmt"
	"os"
)

type LoxFunction struct {
	declaration Function
}

func NewLoxFunction(decl Function) LoxFunction {
	return LoxFunction{decl}
}

func (lf LoxFunction) Call(interpreter *Interpreter, args []Object) Object {
	environment := NewEnclosedEnvinronment(&globals)

	for i, p := range lf.declaration.Params {
		a := args[i]
		environment.Define(p.Lexeme, a)
	}

	var returnThis ReturnThis
	_, err := interpreter.executeBlock(lf.declaration.Body, environment)
	if errors.As(err, &returnThis) {
		return returnThis.Value
	} else {
		fmt.Fprintln(os.Stderr, err.Error())
		return nil
	}
}

func (lf LoxFunction) Arity() int {
	return len(lf.declaration.Params)
}

func (lf LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", lf.declaration.Name.Lexeme)
}
