package events

import (
	"fmt"
	"testing"
	"time"
)

func TestEventEmitters(t *testing.T) {
	event := Default()
	event2 := Default()
	fmt.Println(event, event2, event == event2)

	event.On("a", func(data EventData) {
		fmt.Println("on a:1 data=", data)
	}, "A")
	event.Once("a", func(data EventData) {
		fmt.Println("on a:2 data=", data)
	}, "A")

	listenB := func(data EventData) {
		fmt.Println("on b:3 data=", data)
	}
	listenB2 := func(data EventData) {
		fmt.Println("on b:4 data=", data)
	}
	event.On("b", listenB, "B")
	event.On("b", listenB, "C")
	event.On("b", listenB2, "C")

	event.Emit("a", "hello a 1")
	//event.OffByTarget("A")
	event.Clear()
	//time.Sleep(time.Second)
	event.Emit("a", "hello a 2")

	//event.Emit("b", "hello b 1")
	//event.Off("b", listenB, "B")
	//event.Off("b", listenB2, "C")
	//event.OffByTarget("C")
	//event.Emit("b", "hello b 2")
	time.Sleep(time.Second)
}
