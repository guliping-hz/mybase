package mybase

import (
	"fmt"
	"testing"
)

func TestAtomicSet_Contain(t *testing.T) {
	saveSet := AtomicSet{}
	saveSet.Insert(1)
	saveSet.Insert(1)
	saveSet.Insert(2)
	saveSet.Insert(3)
	saveSet.Remove(2)

	saveSet.Range(func(val any) bool {
		fmt.Printf("v=%v\n", val)
		return true
	})
}
