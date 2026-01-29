package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
)

func TestAddPathPaste(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeAddPath

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("/tmp/repo\n"), Paste: true}
	m2, _ := m.handleAddPath(msg)
	m = m2.(Model)

	if m.addPathInput != "/tmp/repo" {
		t.Fatalf("expected pasted path, got %q", m.addPathInput)
	}
}

func TestCommitPaste(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeCommitInput

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("msg line\n"), Paste: true}
	m2, _ := m.handleCommitInput(msg)
	m = m2.(Model)

	if m.commitMsg != "msg line" {
		t.Fatalf("expected commit msg, got %q", m.commitMsg)
	}
}
