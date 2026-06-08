package interpreter

import "time"

type Callable interface {
	/// Calls this callable with the given arguments.
	Call(interpreter *Interpreter, args []Object) (Object, error)
	/// Returns the number of arguments this callable expects.
	Arity() int
}

// TODO: A native function that always accept 0 args is not helpful.
// Maybe we can generalize this?
type NativeFun func() (Object, error)

func (f NativeFun) Call(_interpreter *Interpreter, _args []Object) (Object, error) {
	return f()
}

func (f NativeFun) Arity() int {
	return 0
}

func Clock() (Object, error) {
	return float64(time.Now().Unix()), nil
}
