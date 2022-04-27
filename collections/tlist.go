package collections

import (
	"log"
	"sync"
	"time"
)

// TList Struct is to keep buffer with required time interval of incoming data in memory
// May be used to keep data for calculating moving average
type TList[T any] struct {
	mu    sync.Mutex
	d     time.Duration //time duration to keep
	first *TListNode[T]
	last  *TListNode[T]
	size  uint
}

type TListNode[T any] struct {
	time time.Time
	data T
	next *TListNode[T]
}

func (n TListNode[T]) GetData() T {
	return n.data
}

func (n TListNode[T]) Next() *TListNode[T] {
	return n.next
}

// Append data to the end of list and remove first elements out of data
func (t *TList[T]) Append(data T, tm time.Time) {
	t.mu.Lock()
	defer t.mu.Unlock()
	node := TListNode[T]{data: data, time: tm}
	if t.last == nil {
		//list empty case
		t.first = &node
		t.last = &node
	} else {
		//list has data case
		t.last.next = &node
		t.last = &node
	}
	t.size += 1
	t.removeOutOfTime()
}

func (t *TList[T]) removeOutOfTime() {
	if t.last == nil {
		log.Println("removeOutOfTime called on empty TList...")
		return
	}
	for t.last.time.Sub(t.first.time) > t.d {
		t.removeFirst()
	}
}

// RemoveFirst removes first element from TList
func (t *TList[T]) RemoveFirst() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.removeFirst()
}

func (t *TList[T]) removeFirst() {
	if t.first == nil {
		log.Println("Called removing first on empty TList...")
		//empty case
		return
	}
	if t.first.next == nil {
		//one element case
		t.first.next = nil //remove link to second
		t.first = nil
		t.last = nil
		return
	} else {
		second := t.first.next
		t.first.next = nil //remove link to second
		t.first = second
	}
	t.size -= 1
}

func (t TList[T]) GetSize() uint {
	return t.size
}

func (t TList[T]) First() *TListNode[T] {
	return t.first
}

func (t TList[T]) Last() *TListNode[T] {
	return t.last
}

func NewTList[T any](d time.Duration) TList[T] {
	return TList[T]{
		d: d,
	}
}
