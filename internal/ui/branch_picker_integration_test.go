package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
)

func TestBranchPickerOpens(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{Path: "/repo/a", Branch: "main"}}

	m2, _ := m.handleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	m = m2.(Model)
	if m.mode != ModeBranchPicker {
		t.Fatalf("expected ModeBranchPicker, got %v", m.mode)
	}
}

func TestBranchesLoadedSetsCursorToCurrent(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeBranchPicker
	msg := branchesLoadedMsg{
		current: "dev",
		items:   []BranchItem{{Name: "main"}, {Name: "dev"}, {Name: "origin/feat", IsRemote: true}},
	}
	m2, _ := m.Update(msg)
	m = m2.(Model)
	if m.branchCursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.branchCursor)
	}
}

func TestBranchPickerEnterDirtyShowsStashConfirm(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.repos = []git.Repo{{Path: "/repo/a", Branch: "main", Modified: 1}}
	m.mode = ModeBranchPicker
	m.branchItems = []BranchItem{{Name: "dev"}}

	m2, _ := m.handleBranchPicker(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(Model)
	if m.mode != ModeConfirmStash {
		t.Fatalf("expected ModeConfirmStash, got %v", m.mode)
	}
	if m.pendingBranch.Name != "dev" {
		t.Fatalf("expected pending branch dev, got %q", m.pendingBranch.Name)
	}
}

func TestConfirmStashCancelReturnsPicker(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.mode = ModeConfirmStash

	m2, _ := m.handleConfirmStash(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	m = m2.(Model)
	if m.mode != ModeBranchPicker {
		t.Fatalf("expected ModeBranchPicker, got %v", m.mode)
	}
}
