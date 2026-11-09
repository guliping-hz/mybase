package mybase

import (
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"testing"
)

func TestRandInt63(t *testing.T) {
	myR := NewMyRand()
	total := 1000000
	cnt := 0
	for i := 0; i < total; i++ {
		if myR.Int63()%10000000 < 5000000 {
			cnt++
		}
	}
	t.Logf("total: %d, cnt: %d %f", total, cnt, float64(cnt)/float64(total))
}

func TestRandInt32(t *testing.T) {
	myR := NewMyRand()
	total := 1000000
	cnt := 0
	for i := 0; i < total; i++ {
		if myR.Int31(10000000) < 5000000 {
			cnt++
		}
	}
	t.Logf("total: %d, cnt: %d %f", total, cnt, float64(cnt)/float64(total))
}

func TestRandIntn(t *testing.T) {
	myR := NewMyRand()
	total := 10000000
	cnt := 0
	cnt2 := 0
	cnt3 := 0

	a := 1000000
	ah := a / 2
	for range total {
		if myR.Intn(a) < ah {
			cnt++
		}

		//if rand.Intn(a) < ah {
		//	cnt2++
		//}

		if rand.IntN(a) < ah {
			cnt3++
		}

	}
	t.Logf("total: %d, cnt: %d %f", total, cnt, float64(cnt)/float64(total))
	t.Logf("total: %d, cnt: %d %f", total, cnt2, float64(cnt2)/float64(total))
	t.Logf("total: %d, cnt: %d %f", total, cnt3, float64(cnt3)/float64(total))
}

func TestRandFloat32(t *testing.T) {
	myR := NewMyRand()
	total := 1000000
	cnt := 0
	for i := 0; i < total; i++ {
		if myR.Float32() < 0.5 {
			cnt++
		}
	}
	t.Logf("total: %d, cnt: %d %f", total, cnt, float64(cnt)/float64(total))
}

func TestFloat(t *testing.T) {
	myR := NewMyRand()
	less05 := 0
	less01 := 0
	less00001 := 0
	for i := 0; i < 1000000; i++ {
		//c := float64(myR.Int63()) / float64((1<<63)-1)
		c := myR.Float64()
		if c < 0.5 {
			less05++

			if c < 0.1 {
				less01++

				if c < 0.00001 {
					less00001++
				}
			}
		}
	}
	t.Log("less05", less05)
	t.Log("less01", less01)
	t.Log("less00001", less00001)
}

func TestRandInt(t *testing.T) {
	//myRand := NewMyRand()
	myRand2 := NewMyRand()
	_ = myRand2

	//检查100亿次的随机数1~1亿的分布情况，理论上是1~1亿每个数各100左右
	const n = 10000000
	m := make([]int, n)
	for i := range n * 100 {
		if n > 1000000 && i%1000000 == 0 {
			t.Log("scan", i)
		}
		//rnd := myRand.Intn(n)
		//rnd := myRand2.Intn(n)
		rnd := rand.IntN(n)
		m[rnd]++
	}

	mm := make(map[int][]int)
	low := []int{1, 10, 20, 30, 40, 50, 60, 70, 80, 90}           //检查某个数随机次数小于指定数的
	up := []int{110, 120, 130, 140, 150, 160, 170, 180, 190, 200} //检查某个随机数次数大于指定数的
	for i := range m {
		v := m[i]
		find := false
		for j := range low {
			vv := low[j]
			if v <= vv {
				mm[vv] = append(mm[vv], i)
				find = true
			}
		}
		for j := len(up) - 1; j >= 0; j-- {
			vv := up[j]
			if v >= vv {
				mm[vv] = append(mm[vv], i)
				find = true
			}
		}
		if !find {
			mm[100] = append(mm[100], i)
		}
	}

	f, err := os.Create(fmt.Sprintf("D:\\myRand-out%d.csv", n))
	if err == nil {
		for i := range mm {
			fmt.Fprintf(f, "%d,%d\n", i, len(mm[i]))
		}
		f.Close()
	}
}

// TestRandDiceDistribution 测试模拟 1~6 的骰子分布情况
func TestRandDiceDistribution(t *testing.T) {
	myR := NewMyRand()
	total := 10000000        // 1000 万次
	counts := make([]int, 7) // 索引 1-6 对应骰子点数

	for i := 0; i < total; i++ {
		//dice := myR.int31()%6 + 1
		dice := myR.Intn(6) + 1
		counts[dice]++
	}

	t.Logf("=== 骰子 1~6 分布测试 (总次数: %d) ===", total)
	expect := float64(total) / 6.0

	// 允许的误差范围（1%）
	tolerance := 0.01

	allPass := true
	for i := 1; i <= 6; i++ {
		ratio := float64(counts[i]) / float64(total)
		diff := math.Abs(float64(counts[i])-expect) / expect
		status := "PASS"
		if diff > tolerance {
			status = "FAIL"
			allPass = false
		}
		t.Logf("点数 %d: 出现 %8d 次, 占比 %.4f%% (期望 16.6667%%) -> 误差 %.4f%% [%s]",
			i, counts[i], ratio*100, diff*100, status)
	}

	if !allPass {
		t.Fatalf("骰子分布不均匀，误差超过 %0.1f%%", tolerance*100)
	}
}
