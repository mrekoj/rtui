package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"rtui/internal/config"
	"rtui/internal/git"
)

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

func TestFooterActionsFitWidth(t *testing.T) {
	widths := []int{20, 30, 40, 60}
	for _, w := range widths {
		m := Model{width: w}
		line := m.renderFooter()
		if lipgloss.Width(line) > w {
			t.Fatalf("footer width %d exceeds %d", lipgloss.Width(line), w)
		}
		if w > 0 && line == "" {
			t.Fatalf("expected footer text for width %d", w)
		}
	}
}

func TestFooterWrapsAtNarrowWidth(t *testing.T) {
	m := Model{width: 40}
	line := m.renderFooter()
	if !strings.Contains(line, "\n") {
		t.Fatal("expected footer to wrap at width 40")
	}
}

func TestAddPathModalCentered(t *testing.T) {
	m := Model{width: 80}
	out := m.renderAddPath()
	lines := strings.Split(out, "\n")
	if len(lines) == 0 {
		t.Fatal("expected modal output")
	}
	first := lines[0]
	leftPad := len(first) - len(strings.TrimLeft(first, " "))
	if leftPad == 0 {
		t.Fatal("expected modal to be centered with left padding")
	}
	for _, line := range lines {
		if lipgloss.Width(line) > m.width {
			t.Fatalf("line width %d exceeds %d", lipgloss.Width(line), m.width)
		}
	}
}

func TestBottomPanelHeightStable(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.width = 60
	m.height = 20
	m.repos = []git.Repo{{Name: "repo", Path: "/tmp/repo"}}

	maxLines := m.bottomPanelMaxLines()
	if maxLines < 3 {
		t.Fatal("expected bottom panel to be visible")
	}

	changes := m.renderChangesPanel(maxLines)
	if got := strings.Count(changes, "\n"); got != maxLines {
		t.Fatalf("expected changes panel lines %d, got %d", maxLines, got)
	}

	m.graphLines = []string{"* commit 1", "* commit 2"}
	graph := m.renderGraphPanel(maxLines)
	if got := strings.Count(graph, "\n"); got != maxLines {
		t.Fatalf("expected graph panel lines %d, got %d", maxLines, got)
	}
}
