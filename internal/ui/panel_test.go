package ui

import "testing"

func TestMaxScroll(t *testing.T) {
	if maxScroll(5, 10) != 0 {
		t.Fatal("expected zero when window >= total")
	}
	if maxScroll(10, 3) != 7 {
		t.Fatalf("expected 7, got %d", maxScroll(10, 3))
	}
}
