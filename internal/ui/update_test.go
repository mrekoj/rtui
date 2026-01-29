package ui

import (
	ospkg "os"
	pathpkg "path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
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

func TestConfirmStageFlow(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeConfirmStage

	m2, _ := m.handleConfirmStage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	m = m2.(Model)
	if m.mode != ModeCommitInput {
		t.Fatalf("expected ModeCommitInput, got %v", m.mode)
	}

	m.mode = ModeConfirmStage
	m2, _ = m.handleConfirmStage(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = m2.(Model)
	if m.mode != ModeNormal {
		t.Fatalf("expected ModeNormal, got %v", m.mode)
	}
}

func TestConfirmPullNo(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeConfirmPull

	m2, _ := m.handleConfirmPull(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = m2.(Model)
	if m.mode != ModeConfirmStage {
		t.Fatalf("expected ModeConfirmStage, got %v", m.mode)
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
