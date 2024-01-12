package mybase

import "testing"

func TestInitLogBigFile(t *testing.T) {
	InitLogBigFile(false, "./bin", "test", 10, 10)

	LogBig("0\n")
	LogBig("123456789\n")

}
