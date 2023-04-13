package mybase

import "testing"

func TestInitLogBigFile(t *testing.T) {
	InitLogBigFile(false, "./bin", "test", 5<<10, 10)
}
