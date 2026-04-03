package stack

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStackEmptyState(t *testing.T) {
	r := require.New(t)

	s := NewStack[int]()

	r.True(s.IsEmpty())
	r.Equal(0, s.Size())

	peeked, ok := s.Peek()
	r.False(ok)
	r.Zero(peeked)

	popped, ok := s.Pop()
	r.False(ok)
	r.Zero(popped)
}

func TestStackPushPeekPop(t *testing.T) {
	r := require.New(t)

	s := NewStack[string]()
	s.Push("foo")
	s.Push("bar")
	s.Push("baz")

	r.False(s.IsEmpty())
	r.Equal(3, s.Size())

	peeked, ok := s.Peek()
	r.True(ok)
	r.Equal("baz", peeked)
	r.Equal(3, s.Size())

	tests := []struct {
		name     string
		wantItem string
		wantSize int
	}{
		{
			name:     "pop top item first",
			wantItem: "baz",
			wantSize: 2,
		},
		{
			name:     "pop second item next",
			wantItem: "bar",
			wantSize: 1,
		},
		{
			name:     "pop bottom item last",
			wantItem: "foo",
			wantSize: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item, ok := s.Pop()

			r.True(ok)
			r.Equal(tt.wantItem, item)
			r.Equal(tt.wantSize, s.Size())
		})
	}

	r.True(s.IsEmpty())
}

func TestStackAllIteratesFromTopToBottom(t *testing.T) {
	r := require.New(t)

	s := NewStack[string]()
	s.Push("bottom")
	s.Push("middle")
	s.Push("top")

	var gotIndices []int
	var gotItems []string

	for i, item := range s.All() {
		gotIndices = append(gotIndices, i)
		gotItems = append(gotItems, item)
	}

	r.Equal([]int{0, 1, 2}, gotIndices)
	r.Equal([]string{"top", "middle", "bottom"}, gotItems)
}

func TestStackAllYieldsTraversalIndex(t *testing.T) {
	r := require.New(t)

	s := NewStack[string]()
	s.Push("alpha")
	s.Push("beta")
	s.Push("gamma")

	got := make(map[string]int)
	for i, item := range s.All() {
		got[item] = i
	}

	r.Equal(map[string]int{
		"gamma": 0,
		"beta":  1,
		"alpha": 2,
	}, got)
}

func TestStackAllStopsWhenYieldReturnsFalse(t *testing.T) {
	r := require.New(t)

	s := NewStack[int]()
	s.Push(10)
	s.Push(20)
	s.Push(30)

	var got []int

	for _, item := range s.All() {
		got = append(got, item)
		if len(got) == 2 {
			break
		}
	}

	r.Equal([]int{30, 20}, got)
}

func TestStackAllOnEmptyStack(t *testing.T) {
	r := require.New(t)

	s := NewStack[int]()

	called := false
	for range s.All() {
		called = true
	}

	r.False(called)
}
