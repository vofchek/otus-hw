package hw04lrucache

import "fmt"

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	length int
	first  *ListItem
	last   *ListItem
}

func NewList() List {
	return new(list)
}

func (list *list) Len() int {
	return list.length
}

func (list *list) Front() *ListItem {
	return list.first
}

func (list *list) Back() *ListItem {
	return list.last
}

func (list *list) PushFront(v interface{}) *ListItem {
	item := wrapInListItem(v)

	if list.first != nil {
		list.first.Prev = item
	}
	item.Next = list.first
	list.first = item
	if list.last == nil {
		list.last = item
	}
	list.length++

	return item
}

func (list *list) PushBack(v interface{}) *ListItem {
	item := wrapInListItem(v)

	if list.last != nil {
		list.last.Next = item
	}
	item.Prev = list.last
	list.last = item
	list.length++
	if list.first == nil {
		list.first = item
	}

	return item
}

func (list *list) Remove(i *ListItem) {
	if i.Prev != nil {
		i.Prev.Next = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	}

	if list.first == i {
		list.first = i.Next
	} else if list.last == i {
		list.last = i.Prev
	}

	i.Prev = nil
	i.Next = nil
	list.length--
}

func (list *list) MoveToFront(i *ListItem) {
	list.Remove(i)
	list.PushFront(i)
}

// завернуть значение в ListItem, если оно ещё не ListItem.
func wrapInListItem(v interface{}) *ListItem {
	var item *ListItem

	switch realV := v.(type) {
	case *ListItem:
		item = realV
	default:
		item = &ListItem{Value: v, Prev: nil, Next: nil}
	}

	return item
}

// вспомогательная функция для визуализации списка для тестирования.
func PrintList(l List) {
	item := l.Front()
	if item == nil {
		fmt.Println("---empty---")
	}

	fmt.Print("\n---list---\n")

	for {
		fmt.Print("item: ", item, "\n")

		item = item.Next

		if item == nil {
			break
		}
	}

	fmt.Print("---list---\n")
}
