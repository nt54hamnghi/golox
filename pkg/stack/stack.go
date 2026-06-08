package stack

import (
	"iter"
	"slices"
)

type Stack[T any] struct {
	inner []T
}

func NewStack[T any]() Stack[T] {
	return Stack[T]{inner: []T{}}
}

func (s *Stack[T]) All() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range slices.Backward(s.inner) {
			depth := len(s.inner) - 1 - i
			if !yield(depth, v) {
				return
			}
		}
	}
}

func (s *Stack[T]) Push(item T) {
	s.inner = append(s.inner, item)
}

func (s *Stack[T]) Pop() (T, bool) {
	if len(s.inner) == 0 {
		var zero T
		return zero, false
	}
	item := s.inner[len(s.inner)-1]
	s.inner = s.inner[:len(s.inner)-1]
	return item, true
}

func (s *Stack[T]) Peek() (T, bool) {
	if len(s.inner) == 0 {
		var zero T
		return zero, false
	}
	return s.inner[len(s.inner)-1], true
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.inner) == 0
}

func (s *Stack[T]) Size() int {
	return len(s.inner)
}
