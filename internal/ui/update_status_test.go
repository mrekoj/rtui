package ui

import (
	"testing"

	"rtui/internal/config"
	"rtui/internal/git"
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
