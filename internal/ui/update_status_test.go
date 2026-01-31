package ui

import (
	"testing"
	"time"

	"rtui/internal/config"
	"rtui/internal/git"

	tea "github.com/charmbracelet/bubbletea"
)

func TestReposLoadedClearsRefreshing(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m.loading = true
	m.statusMsg = "Refreshing..."

	msg := reposLoadedMsg{repos: []git.Repo{}, usedCWD: false, cwd: ""}
	m2, _ := m.Update(msg)
	m = m2.(Model)

	if m.loading {
		t.Fatal("expected loading false")
	}
	if m.statusMsg != "Refreshed" {
		t.Fatalf("expected status 'Refreshed', got %q", m.statusMsg)
	}
}

func TestStatusAutoClearInfo(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m = m.setStatusInfo("Refreshed")
	if m.statusMsg == "" {
		t.Fatal("expected status set")
	}
	tick := statusTickMsg{now: m.statusUntil.Add(1 * time.Second)}
	m2, _ := m.Update(tick)
	m = m2.(Model)
	if m.statusMsg != "" {
		t.Fatalf("expected status cleared, got %q", m.statusMsg)
	}
}

func TestStatusErrorPersistsUntilKey(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m = m.setStatusError("Error: failed")
	tick := statusTickMsg{now: time.Now().Add(10 * time.Second)}
	m2, _ := m.Update(tick)
	m = m2.(Model)
	if m.statusMsg == "" {
		t.Fatal("expected error status to persist")
	}
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = m2.(Model)
	if m.statusMsg != "" {
		t.Fatalf("expected error cleared on key, got %q", m.statusMsg)
	}
}
