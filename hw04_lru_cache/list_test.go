package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("check front value is set", func(t *testing.T) {
		l := NewList()

		l.PushFront(1)

		first := l.Front()
		require.Equal(t, 1, first.Value)
	})

	t.Run("check back value is set", func(t *testing.T) {
		l := NewList()

		l.PushBack(1)

		back := l.Back()
		require.Equal(t, 1, back.Value)
	})

	t.Run("check front equal back in one item list", func(t *testing.T) {
		l := NewList()

		l.PushFront(1)

		front := l.Front()
		back := l.Back()
		require.Equal(t, front, back)
	})

	t.Run("check list structure", func(t *testing.T) {
		l := NewList()

		last := l.PushFront(1)
		middle := l.PushFront(2)
		first := l.PushFront(3)

		require.Nil(t, last.Next, "last can`t have next item")
		require.Equal(t, middle.Next, last, "middle next item must be last")
		require.Equal(t, middle.Prev, first, "middle prev item must be first")
		require.Nil(t, first.Prev, "first can`t have prev item")
	})
}
