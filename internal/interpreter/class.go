package interpreter

type LoxClass struct {
	Name       string
	Superclass *LoxClass
	methods    map[string]LoxFunction
}

func NewLoxClass(name string, superclass *LoxClass, methods map[string]LoxFunction) *LoxClass {
	return &LoxClass{name, superclass, methods}
}

func (cls *LoxClass) FindMethod(name string) (LoxFunction, bool) {
	if method, ok := cls.methods[name]; ok {
		return method, true
	}

	if cls.Superclass != nil {
		return cls.Superclass.FindMethod(name)
	}

	return LoxFunction{}, false
}

func (cls *LoxClass) String() string {
	return cls.Name
}

// Call implements [Callable].
func (cls *LoxClass) Call(interpreter *Interpreter, arguments []Object) (Object, error) {
	instance := NewLoxInstance(cls)

	init, exist := cls.FindMethod("init")
	if exist {
		return init.bind(instance).Call(interpreter, arguments)
	}

	return instance, nil
}

// Arity implements [Callable].
func (cls *LoxClass) Arity() int {
	init, exist := cls.FindMethod("init")
	if !exist {
		return 0
	}
	return init.Arity()
}
