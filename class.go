package main

type LoxClass struct {
	Name string
}

func NewLoxClass(name string) *LoxClass {
	return &LoxClass{name}
}

func (cls *LoxClass) String() string {
	return cls.Name
}

func (cls *LoxClass) Call(interpreter *Interpreter, arguments []Object) Object {
	return NewLoxInstance(cls)
}

func (cls *LoxClass) Arity() int {
	return 0
}
