package ro

import (
	"testing"
)

func TestMemoize(t *testing.T) {
	var count int
	f := func() int {
		count++
		return count
	}
	g := Memoize(f)
	if g() != 1 {
		t.Fail()
	}
	if g() != 1 {
		t.Fail()
	}
	if g() != 1 {
		t.Fail()
	}
}
