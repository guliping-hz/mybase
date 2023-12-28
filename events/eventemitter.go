package events

import (
	"container/list"
	"fmt"
	"strconv"
	"sync"
)

type EventData any
type Handler func(EventData)

func (h Handler) Address() int64 {
	//rv := reflect.ValueOf(h)
	address, _ := strconv.ParseInt(fmt.Sprintf("%p", h), 0, 64)
	//fmt.Println("Address", rv, rv.String(), rv.Pointer(), address)
	/*
		Address 0x11f5d60 <events.Handler Value> 18832736 18832736
		Address 0x11f5d60 <events.Handler Value> 18832736 18832736
	*/
	return address
}

//type HandlerOnce struct {
//	handler Handler
//	once    bool //true 表示只监听一次
//}

//type Handlers list.List //HandlerOnce

type PosElement struct {
	pos *list.Element
	evt string
}

type EventEmitters struct {
	mutex sync.Mutex

	dictHandlers map[string]*list.List
	dictTargets  map[string]map[int64][]*PosElement

	catch func()
}

func (e *EventEmitters) SetCatch(catch func()) {
	e.catch = catch
}

func (e *EventEmitters) safeDo(handler Handler, data EventData) {
	if e.catch != nil {
		defer e.catch()
	}
	if handler != nil {
		handler(data)
	}
}

func (e *EventEmitters) Emit(key string, data EventData) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	handlers, ok := e.dictHandlers[key]
	if !ok {
		return
	}
	p := handlers.Front()
	for p != nil {
		handler := p.Value.(Handler)
		p = p.Next()
		//go handler(data)
		go e.safeDo(handler, data)
	}
}

func (e *EventEmitters) Once(key string, handler Handler, target string) bool {
	address := int64(0)
	var innerHandler Handler = func(data EventData) {
		e.off(key, address, target)
		//go handler(data)
		go e.safeDo(handler, data)
	}
	address = innerHandler.Address()
	//fmt.Printf("Once address=%d\n", address)
	return e.On(key, innerHandler, target)
}

func (e *EventEmitters) Off(key string, handler Handler, target string) {
	if key == "" {
		return
	}
	e.off(key, handler.Address(), target)
}

func (e *EventEmitters) OffByTarget(target string) {
	e.off("", 0, target)
}

/*
*
session = 0 表示清空
*/
func (e *EventEmitters) off(key string, session int64, target string) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	handlerMap, ok := e.dictTargets[target]
	if !ok {
		return
	}

	if key != "" {
		handlers, ok := e.dictHandlers[key]
		if !ok {
			return
		}

		if session == 0 {
			panic("please check the code; must not do this")
		}

		if elements, ok := handlerMap[session]; ok {
			for i := range elements {
				handlers.Remove(elements[i].pos)
			}
			delete(e.dictTargets[target], session)
		}
		return
	}

	for k := range handlerMap {
		elements := handlerMap[k]
		for i := range elements {
			handlers, ok := e.dictHandlers[elements[i].evt]
			if !ok {
				continue
			}
			handlers.Remove(elements[i].pos)
		}

	}
	delete(e.dictTargets, target)
}

func (e *EventEmitters) On(key string, handler Handler, target string) bool {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	//fmt.Printf("handler=%d\n", handler.Address())

	if e.dictHandlers == nil {
		//初始化创建
		e.dictHandlers = make(map[string]*list.List)
		e.dictTargets = make(map[string]map[int64][]*PosElement)
	}

	handlers, ok := e.dictHandlers[key]
	if !ok {
		//创建列表
		handlers = &list.List{}
		e.dictHandlers[key] = handlers
	}

	pos := handlers.PushBack(handler)
	handlerMap, ok := e.dictTargets[target]
	if !ok {
		//创建字典
		handlerMap = make(map[int64][]*PosElement)
		e.dictTargets[target] = handlerMap
	}
	hAddress := handler.Address()
	//这里改成列表，是因为可能多个key共享一个handler。
	handlerMap[hAddress] = append(handlerMap[hAddress], &PosElement{pos, key})
	return true
}

func (e *EventEmitters) Clear() {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.dictHandlers = nil
	e.dictTargets = nil
}

var defaultEE = &EventEmitters{}

func Default() *EventEmitters {
	return defaultEE
}

func NewEventEmitters() *EventEmitters {
	return &EventEmitters{}
}
