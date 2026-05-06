package main

type LoxClass struct {
	Name string
}

func NewLoxClass(name string) LoxClass {
	return LoxClass{Name: name}
}

func (l LoxClass) String() string {
	return l.Name
}
