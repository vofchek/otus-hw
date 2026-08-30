package hw04lrucache

import "errors"

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

// контейнер для значения + его ключ.
type cachedValue struct {
	Value interface{}
	Key   Key
}

var ErrNotCachedValue = errors.New("value must be cached value, but it is not")

// возвращает *cachedValue или паникует.
func mustGetCachedValue(v interface{}) *cachedValue {
	cV, ok := v.(*cachedValue)

	if !ok {
		panic(ErrNotCachedValue)
	}

	return cV
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

func (lruCache *lruCache) Set(key Key, value interface{}) bool {
	item, inMap := lruCache.items[key]

	if inMap {
		cV := mustGetCachedValue(item.Value)
		cV.Value = value
		lruCache.queue.MoveToFront(item)
		return true
	}

	item = lruCache.queue.PushFront(&cachedValue{Value: value, Key: key})
	if lruCache.queue.Len() > lruCache.capacity {
		oldItem := lruCache.queue.Back()
		cV := mustGetCachedValue(oldItem.Value)
		delete(lruCache.items, cV.Key)
	}

	lruCache.items[key] = item

	return false
}

func (lruCache *lruCache) Get(key Key) (interface{}, bool) {
	item, inMap := lruCache.items[key]

	if inMap {
		lruCache.queue.MoveToFront(item)
		cV := mustGetCachedValue(item.Value)
		return cV.Value, true
	}

	return nil, false
}

func (lruCache *lruCache) Clear() {
	clear(lruCache.items)
	lruCache.queue = NewList()
}
