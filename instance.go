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
	if obj, ok := i.fields[name.Lexeme]; ok {
		return obj, nil
	}

	return nil, ErrorAtToken(name, "Undefined property '"+name.Lexeme+"'.")
}

func (i LoxInstance) String() string {
	return i.class.Name + " instance"
}
