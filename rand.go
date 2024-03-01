package mybase

import (
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

func (m *MyRand) Uint64() uint64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.seed = (48271*m.seed + 1) % uint64(1<<31-1)
	return m.seed
}

func (m *MyRand) Int63() int64 {
	return int64(m.Uint64() & 0x7fffffffffffffff)
}

func (m *MyRand) Intn(n int) int {
	return int(m.Uint64() % uint64(n))
}

func (m *MyRand) Float64() float64 {
again:
	f := float64(m.Int63()) / (1 << 63)
	if f == 1 {
		goto again // resample; this branch is taken O(never)
	}
	return f
}

func (m *MyRand) Float32() float32 {
	return float32(m.Float64())
}

func NewMyRand() *MyRand {
	return &MyRand{seed: uint64(time.Now().UnixNano())}
}
