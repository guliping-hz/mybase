package mybase

import (
	"math"
	"sync"
	"time"
)

type MyRand struct {
	mutex sync.Mutex
	seed  uint64
}

func (m *MyRand) Seed(seed int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.seed = uint64(seed)
}

func (m *MyRand) int31() int32 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.seed = (48271*m.seed + 1) % 0x7FFFFFFF //2147483647
	return int32(m.seed)
}

func (m *MyRand) Float64() float64 {
again:
	f := float64(m.int31()) / 2147483647.0
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return f
}

func (m *MyRand) Int63() int64 {
	return int64(math.Floor(m.Float64() * float64(0x7FFFFFFFFFFFFFFF)))
}

func (m *MyRand) Intn(n int) int {
	return int(m.Int63() % int64(n))
}

func (m *MyRand) Int31(n int32) int32 {
	return int32(m.Int63() % int64(n))
}

func (m *MyRand) Float32() float32 {
again:
	f := float32(m.Float64())
	if f == 1 {
		goto again // resample; this branch is taken O(very rarely)
	}
	return f
}

func NewMyRand() *MyRand {
	return &MyRand{seed: uint64(time.Now().UnixNano())}
}
