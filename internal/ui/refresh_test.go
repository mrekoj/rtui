package ui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
)

func TestRefreshIntervalFromConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RefreshInterval = 7
	m := NewModel(cfg)

	if got := m.refreshInterval(); got != 7*time.Second {
		t.Fatalf("expected 7s refresh interval, got %v", got)
	}
}

func TestRefreshIntervalFallsBackToDefault(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.RefreshInterval = 0
	m := NewModel(cfg)
	if got := m.refreshInterval(); got != defaultRefreshInterval {
		t.Fatalf("expected default interval %v, got %v", defaultRefreshInterval, got)
	}

	cfg.RefreshInterval = -5
	m = NewModel(cfg)
	if got := m.refreshInterval(); got != defaultRefreshInterval {
		t.Fatalf("expected default interval %v, got %v", defaultRefreshInterval, got)
	}
}

func TestRefreshTickMessageSchedulesReloadAndNextTick(t *testing.T) {
	m := NewModel(config.DefaultConfig())
	m2, cmd := m.Update(refreshTickMsg{})
	_ = m2.(Model)

	if cmd == nil {
		t.Fatal("expected command batch from refresh tick")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}
	if len(batch) != 2 {
		t.Fatalf("expected 2 batched commands, got %d", len(batch))
	}
}
