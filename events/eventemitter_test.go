package events

import (
	"testing"
	"time"
)

func TestEventEmitters(t *testing.T) {
	event := Default()
	event2 := Default()
	t.Log(event, event2, event == event2)

	event.On("a", func(data EventData) {
		t.Log("on A:a 1 data=", data)
	}, "A")
	event.Once("a", func(data EventData) {
		t.Log("on A:a 2 data=", data)
	}, "A")

	listenB := func(data EventData) {
		t.Log("on B data=", data)
	}
	listenC := func(data EventData) {
		t.Log("on C data=", data)
	}
	event.On("b", listenB, "B")
	event.On("b", listenC, "C")
	event.On("c", listenC, "C")
	event.On("d", listenC, "C")

	t.Log("OffByTarget C")
	event.OffByTarget("C")
	//event.Emit("a", "hello a")
	event.Emit("b", "hello b")
	event.Emit("c", "hello c")
	event.Emit("d", "hello d")
	//event.OffByTarget("A")
	event.Clear()
	//time.Sleep(time.Second)
	event.Emit("b", "hello a 2")

	//event.Emit("b", "hello b 1")
	//event.Off("b", listenB, "B")
	//event.Off("b", listenB2, "C")
	//event.OffByTarget("C")
	//event.Emit("b", "hello b 2")
	time.Sleep(time.Second)
}
