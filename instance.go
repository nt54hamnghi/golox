package main

type LoxInstance struct {
	class *LoxClass
}

func NewLoxInstance(cls *LoxClass) LoxInstance {
	return LoxInstance{cls}
}

func (l LoxInstance) String() string {
	return l.class.Name + " instance"
}
