package main

import "fmt"

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

	interpreter.executeBlock(lf.declaration.Body, environment)
	return nil
}

func (lf LoxFunction) Arity() int {
	return len(lf.declaration.Params)
}

func (lf LoxFunction) String() string {
	return fmt.Sprintf("<fn %s>", lf.declaration.Name.Lexeme)
}
