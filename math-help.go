package mybase

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

func MinT[T int64 | int32 | int | float64 | float32](a, b T) T {
	return T(math.Min(float64(a), float64(b)))
}
func MaxT[T int64 | int32 | int | float64 | float32](a, b T) T {
	return T(math.Max(float64(a), float64(b)))
}
func AbsT[T int64 | int32 | int | float64 | float32](a T) T {
	return T(math.Abs(float64(a)))
}
func CeilT[T int64 | int32 | int | float64 | float32](a T) T {
	return T(math.Ceil(float64(a)))
}
func FloorT[T int64 | int32 | int | float64 | float32](a T) T {
	return T(math.Floor(float64(a)))
}

// maxValue > 0
func GetRandom(maxValue int) int {
	if maxValue <= 0 {
		return 0
	}
	return RandInt(0, maxValue)
}

func GetRandomI32(maxValue int) int32 {
	return int32(GetRandom(maxValue))
}

// 区间：[minValue,maxValue)
func RandInt(minValue, maxValue int) int {
	diff := maxValue - minValue
	ret, err := rand.Int(rand.Reader, big.NewInt(int64(diff)))
	if err != nil {
		return 0
	}
	return int(ret.Int64()) + minValue
}

func GetRandSeed() int64 {
	var a = 0 //变量地址当做随机数
	var b = 0 //变量地址当做随机数
	aPtr, _ := strconv.ParseInt(fmt.Sprintf("%p", &a), 0, 64)
	bPtr, _ := strconv.ParseInt(fmt.Sprintf("%p", &b), 0, 64)

	return time.Now().Unix() * aPtr * bPtr
}
