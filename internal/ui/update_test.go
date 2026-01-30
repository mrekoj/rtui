package ui

import (
	ospkg "os"
	pathpkg "path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
)

func TestToggleDirtyFilter(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.filterDirty = false

	m2, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	m = m2.(Model)
	if !m.filterDirty {
		t.Fatal("expected filterDirty true")
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor reset to 0, got %d", m.cursor)
	}
}

func TestAddPathFlow(t *testing.T) {
	home := t.TempDir()
	if err := ospkg.Setenv("HOME", home); err != nil {
		t.Fatalf("setenv: %v", err)
	}

	repoPath := pathpkg.Join(home, "repo")
	if err := ospkg.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	m := NewModel(config.DefaultConfig())
	m.mode = ModeAddPath
	m.addPathInput = repoPath

	m2, cmd := m.handleAddPath(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(Model)
	if cmd == nil {
		t.Fatal("expected loadRepos cmd")
	}
	if m.mode != ModeNormal {
		t.Fatalf("expected ModeNormal, got %v", m.mode)
	}
	if len(m.config.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(m.config.Paths))
	}
}

func TestAddPathDuplicate(t *testing.T) {
	home := t.TempDir()
	if err := ospkg.Setenv("HOME", home); err != nil {
		t.Fatalf("setenv: %v", err)
	}

	repoPath := pathpkg.Join(home, "repo")
	if err := ospkg.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cfg := config.DefaultConfig()
	cfg.Paths = []string{repoPath}
	m := NewModel(cfg)
	m.mode = ModeAddPath
	m.addPathInput = repoPath

	m2, cmd := m.handleAddPath(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(Model)
	if cmd != nil {
		t.Fatal("expected no cmd on duplicate")
	}
	if m.statusMsg != "Path already exists" {
		t.Fatalf("expected duplicate status, got %q", m.statusMsg)
	}
}

func TestPullBlockedWhenDirty(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	m.repos = []git.Repo{{
		Name:     "repo",
		Path:     "/tmp/repo",
		Modified: 1,
	}}

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = m2.(Model)
	if cmd != nil {
		t.Fatal("expected no cmd when pull is blocked")
	}
	if m.statusMsg != "Cannot pull: repo has uncommitted changes" {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}

func TestCommitInputBackspace(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeCommitInput
	m.commitMsg = "ab"

	m2, _ := m.handleCommitInput(tea.KeyMsg{Type: tea.KeyBackspace})
	m = m2.(Model)
	if m.commitMsg != "a" {
		t.Fatalf("expected commitMsg 'a', got %q", m.commitMsg)
	}
}

func TestPullStartsWhenClean(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{
		Name: "repo",
		Path: "/tmp/repo",
	}}

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	m = m2.(Model)
	if cmd == nil {
		t.Fatal("expected pull cmd")
	}
	if m.statusMsg != "Pulling..." {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}

func TestPushBlockedWhenBehind(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{
		Name:   "repo",
		Path:   "/tmp/repo",
		Behind: 2,
	}}

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = m2.(Model)
	if cmd != nil {
		t.Fatal("expected no cmd when push is blocked")
	}
	if m.statusMsg != "Cannot push: behind remote (pull first)" {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}

func TestPushBlockedWhenDirty(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{
		Name:     "repo",
		Path:     "/tmp/repo",
		Modified: 1,
	}}

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = m2.(Model)
	if cmd != nil {
		t.Fatal("expected no cmd when push is blocked")
	}
	if m.statusMsg != "Cannot push: repo has uncommitted changes" {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}

func TestPushStartsWhenUpToDate(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{
		Name: "repo",
		Path: "/tmp/repo",
	}}

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	m = m2.(Model)
	if cmd == nil {
		t.Fatal("expected push cmd")
	}
	if m.statusMsg != "Pushing..." {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}

func TestPanelFocusKeys(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	m2, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = m2.(Model)
	if m.panelFocus != FocusBottom {
		t.Fatalf("expected focus bottom, got %v", m.panelFocus)
	}

	m2, _ = m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}})
	m = m2.(Model)
	if m.panelFocus != FocusRepos {
		t.Fatalf("expected focus repos, got %v", m.panelFocus)
	}
}

func TestBottomViewToggle(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	if m.bottomView != BottomChanges {
		t.Fatalf("expected default bottom view changes, got %v", m.bottomView)
	}

	m2, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyTab})
	m = m2.(Model)
	if m.bottomView != BottomGraph {
		t.Fatalf("expected bottom view graph, got %v", m.bottomView)
	}
}

func TestOpenSettingsKey(t *testing.T) {
	m := NewModel(config.DefaultConfig())

	m2, cmd := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = m2.(Model)
	if cmd == nil {
		t.Fatal("expected settings cmd")
	}
	if m.statusMsg != "Opening settings in VS Code..." {
		t.Fatalf("unexpected status: %q", m.statusMsg)
	}
}
