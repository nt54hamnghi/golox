package interpreter

import (
	"github.com/nt54hamnghi/golox/internal/errors"
	"github.com/nt54hamnghi/golox/internal/scanner/token"
)

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

func (i LoxInstance) Get(name token.Token) (Object, error) {
	if field, ok := i.fields[name.Lexeme]; ok {
		return field, nil
	}

	method, exist := i.class.FindMethod(name.Lexeme)
	if exist {
		return method.bind(i), nil
	}

	return nil, errors.RuntimeErrorAtToken(
		name,
		"Undefined property '"+name.Lexeme+"'.",
	)
}

func (i LoxInstance) Set(name token.Token, value Object) {
	i.fields[name.Lexeme] = value
}

func (i LoxInstance) String() string {
	return i.class.Name + " instance"
}
