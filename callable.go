package main

type Callable interface {
	/// Calls this callable with the given arguments.
	Call(interpreter *Interpreter, args []Object) Object
	/// Returns the number of arguments this callable expects.
	Arity() int
}
