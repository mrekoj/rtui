package ui

import "testing"

func TestFilterBranchesCaseInsensitive(t *testing.T) {
	items := []BranchItem{
		{Name: "main", IsRemote: false},
		{Name: "feature/auth", IsRemote: false},
		{Name: "origin/feature/api", IsRemote: true},
	}

	out := filterBranches(items, "FeAt")
	if len(out) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(out))
	}
	if out[0].Name != "feature/auth" || out[1].Name != "origin/feature/api" {
		t.Fatalf("unexpected filter order: %#v", out)
	}
}

func TestFilterBranchesKeepsOrder(t *testing.T) {
	items := []BranchItem{
		{Name: "b", IsRemote: false},
		{Name: "a", IsRemote: false},
		{Name: "origin/c", IsRemote: true},
	}

	out := filterBranches(items, "")
	if len(out) != 3 {
		t.Fatalf("expected 3 branches, got %d", len(out))
	}
	if out[0].Name != "b" || out[1].Name != "a" || out[2].Name != "origin/c" {
		t.Fatalf("order changed: %#v", out)
	}
}

func TestIndexOfBranch(t *testing.T) {
	items := []BranchItem{
		{Name: "main", IsRemote: false},
		{Name: "dev", IsRemote: false},
		{Name: "origin/feature", IsRemote: true},
	}

	if idx := indexOfBranch(items, "dev"); idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
	if idx := indexOfBranch(items, "missing"); idx != 0 {
		t.Fatalf("expected idx 0 for missing, got %d", idx)
	}
}

func TestBranchWindow(t *testing.T) {
	cases := []struct {
		total  int
		cursor int
		max    int
		start  int
		end    int
	}{
		{total: 5, cursor: 0, max: 10, start: 0, end: 5},
		{total: 100, cursor: 0, max: 10, start: 0, end: 10},
		{total: 100, cursor: 9, max: 10, start: 0, end: 10},
		{total: 100, cursor: 10, max: 10, start: 1, end: 11},
		{total: 100, cursor: 50, max: 10, start: 41, end: 51},
		{total: 100, cursor: 99, max: 10, start: 90, end: 100},
	}

	for _, c := range cases {
		start, end := branchWindow(c.total, c.cursor, c.max)
		if start != c.start || end != c.end {
			t.Fatalf("total=%d cursor=%d max=%d: expected %d..%d got %d..%d",
				c.total, c.cursor, c.max, c.start, c.end, start, end)
		}
	}
}
