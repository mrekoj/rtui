package ui

import "testing"

func TestCalculateLayoutCompact(t *testing.T) {
	m := Model{width: 35}
	layout := m.calculateLayout()

	if layout.Branch != 0 {
		t.Fatalf("expected branch hidden in compact, got %d", layout.Branch)
	}
	if layout.Status != 8 {
		t.Fatalf("expected status width 8 in compact, got %d", layout.Status)
	}
	if layout.Sync != 6 {
		t.Fatalf("expected sync width 6 in compact, got %d", layout.Sync)
	}
}

func TestCalculateLayoutNarrow(t *testing.T) {
	m := Model{width: 45}
	layout := m.calculateLayout()

	if layout.Branch <= 0 {
		t.Fatalf("expected branch visible, got %d", layout.Branch)
	}
	if layout.Name <= 0 {
		t.Fatalf("expected name width > 0, got %d", layout.Name)
	}
	if layout.Status != 8 {
		t.Fatalf("expected status width 8, got %d", layout.Status)
	}
	if layout.Sync != 6 {
		t.Fatalf("expected sync width 6, got %d", layout.Sync)
	}
}

func TestCalculateLayoutNormal(t *testing.T) {
	m := Model{width: 80}
	layout := m.calculateLayout()

	if layout.Branch <= 0 {
		t.Fatalf("expected branch visible, got %d", layout.Branch)
	}
	if layout.Name <= 0 {
		t.Fatalf("expected name width > 0, got %d", layout.Name)
	}
}
