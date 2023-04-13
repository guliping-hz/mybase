package mybase

import (
	"container/list"
	"sync"
)

type AtomicSet struct {
	set   map[interface{}]*list.Element
	lst   list.List //双向链表。
	mutex sync.RWMutex

	cur *list.Element
}

func (a *AtomicSet) lazyInit() {
	if a.set == nil {
		a.set = make(map[interface{}]*list.Element)
		a.lst.Init()
	}
}

func (a *AtomicSet) Range(cb func(val interface{}) bool) {
	if cb == nil {
		return
	}

	a.mutex.RLock()
	defer a.mutex.RUnlock()

	p := a.lst.Front()
	for p != nil {
		if !cb(p.Value) {
			break
		}
		p = p.Next()
	}
}

func (a *AtomicSet) Insert(val interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.lazyInit()
	if _, ok := a.set[val]; ok {
		return
	}

	p := a.lst.PushBack(val)
	a.set[val] = p //记录指针位置
}

func (a *AtomicSet) Remove(val interface{}) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if p, ok := a.set[val]; ok {
		delete(a.set, val)
		a.lst.Remove(p)
	}
}

func (a *AtomicSet) Contain(val interface{}) bool {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	_, ok := a.set[val]
	return ok
}

func (a *AtomicSet) Len() int {
	a.mutex.RLock()
	defer a.mutex.RUnlock()
	return a.lst.Len()
}

func (a *AtomicSet) Random() (interface{}, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	idleLen := a.lst.Len()
	if idleLen == 0 {
		return nil, false
	}

	for k := range a.set { //每次遍历map，底层都是随机取的，我们就直接使用第一个。
		return k, true
	}
	return nil, false
}

func (a *AtomicSet) Next() (interface{}, bool) {
	a.mutex.RLock()
	defer a.mutex.RUnlock()

	if a.lst.Len() == 0 { //已经取完了，，没有下一个了
		return nil, false
	}

	if a.cur == nil {
		a.cur = a.lst.Front()
	}

	ret := a.cur.Value
	a.cur = a.cur.Next() //指向下一个。
	return ret, true
}
