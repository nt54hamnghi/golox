package main

type LoxClass struct {
	Name    string
	methods map[string]LoxFunction
}

func NewLoxClass(name string, methods map[string]LoxFunction) *LoxClass {
	return &LoxClass{name, methods}
}

func (cls *LoxClass) FindMethod(name string) *LoxFunction {
	if method, ok := cls.methods[name]; ok {
		return &method
	}
	return nil
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
