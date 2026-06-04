package main

type LoxInstance struct {
	class  *LoxClass
	fields map[string]Object
}

func NewLoxInstance(cls *LoxClass) LoxInstance {
	return LoxInstance{
		class:  cls,
		fields: make(map[string]Object),
	}
}

func (i LoxInstance) Get(name Token) (Object, error) {
	if field, ok := i.fields[name.Lexeme]; ok {
		return field, nil
	}

	method, exist := i.class.FindMethod(name.Lexeme)
	if exist {
		return method.bind(i), nil
	}

	return nil, ErrorAtToken(name, "Undefined property '"+name.Lexeme+"'.")
}

func (i LoxInstance) Set(name Token, value Object) {
	i.fields[name.Lexeme] = value
}

func (i LoxInstance) String() string {
	return i.class.Name + " instance"
}
