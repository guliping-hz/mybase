package mybase

import (
	"github.com/bytedance/sonic"
	"testing"
)

func TestSameTransfer(t *testing.T) {
	str := `{"0": [1000, 2000, 3000], "7": [1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 20000], "8": [20000, 30000, 40000, 50000, 100000, 200000, 300000, 400000, 500000, 600000, 800000, 1000000], "9": [1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000, 20000], "10": [20000, 30000, 40000, 50000, 100000, 200000, 300000, 400000, 500000, 600000, 800000, 1000000]}`
	cfg := H{}
	_ = sonic.Unmarshal([]byte(str), &cfg)

	vip0Forts := make([]int64, 0)
	cfg.Get("0", &vip0Forts)
	t.Log(vip0Forts)
}

func TestSliceOver65535(t *testing.T) {
	arra := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	t.Log("test1")
	SliceOver65535(len(arra), 7, func(i, j int) {
		t.Log(arra[i:j])
	})
	t.Log("test2")
	SliceOver65535(len(arra), 5, func(i, j int) {
		t.Log(arra[i:j])
	})
	t.Log("test3")
	SliceOver65535(len(arra), 14, func(i, j int) {
		t.Log(arra[i:j])
	})
	t.Log("test4")
	SliceOver65535(len(arra), 20, func(i, j int) {
		t.Log(arra[i:j])
	})

	t.Logf("x=0x%x\n", []byte("012"))
}
