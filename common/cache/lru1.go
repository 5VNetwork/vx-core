package cache

import (
	"container/list"
	"sync"
)

type lru1 struct {
	capacity    int
	list        *list.List
	keyToElem   map[interface{}]*list.Element
	valueToElem map[interface{}]*list.Element
	mu          sync.Mutex
}

type lru1Entry struct {
	key   interface{}
	value interface{}
}

// NewLru1 creates a new memory-efficient LRU cache
func NewLru1(cap int) Lru {
	return &lru1{
		capacity:    cap,
		list:        list.New(),
		keyToElem:   make(map[interface{}]*list.Element, cap),
		valueToElem: make(map[interface{}]*list.Element, cap),
	}
}

func (l *lru1) Get(key interface{}) (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if elem, ok := l.keyToElem[key]; ok {
		l.list.MoveToFront(elem)
		return elem.Value.(*lru1Entry).value, true
	}
	return nil, false
}

func (l *lru1) GetKeyFromValue(value interface{}) (interface{}, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if elem, ok := l.valueToElem[value]; ok {
		l.list.MoveToFront(elem)
		return elem.Value.(*lru1Entry).key, true
	}
	return nil, false
}

func (l *lru1) ForEach(fn func(key, value interface{}) bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for e := l.list.Front(); e != nil; e = e.Next() {
		entry := e.Value.(*lru1Entry)
		if !fn(entry.key, entry.value) {
			return
		}
	}
}

func (l *lru1) Put(key, value interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, ok := l.keyToElem[key]; ok {
		entry := elem.Value.(*lru1Entry)
		// Fix: remove old value from reverse map before updating
		delete(l.valueToElem, entry.value)
		entry.value = value
		l.valueToElem[value] = elem
		l.list.MoveToFront(elem)
		return
	}

	entry := &lru1Entry{key: key, value: value}
	elem := l.list.PushFront(entry)
	l.keyToElem[key] = elem
	l.valueToElem[value] = elem

	if l.list.Len() > l.capacity {
		oldest := l.list.Back()
		if oldest != nil {
			l.list.Remove(oldest)
			e := oldest.Value.(*lru1Entry)
			delete(l.keyToElem, e.key)
			delete(l.valueToElem, e.value)
		}
	}
}
