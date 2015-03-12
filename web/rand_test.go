package web

import "testing"

func TestRand(t *testing.T) {
	r1 := randSeq(3)
	r2 := randSeq(3)
	r3 := randSeq(12)

	if len(r1) != len(r2) && len(r2) != 3 {
		t.Error("should be 3")
	}

	if r1 == r2 {
		t.Error("not very random")
	}

	if len(r3) != 12 {
		t.Error("should be 12")
	}
}
