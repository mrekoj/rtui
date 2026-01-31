package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type StatusKind int

const (
	StatusInfo StatusKind = iota
	StatusError
)

const (
	statusAutoClear   = 5 * time.Second
	statusTickInterval = 1 * time.Second
)

type statusTickMsg struct {
	now time.Time
}

func (m Model) statusTickCmd() tea.Cmd {
	return tea.Tick(statusTickInterval, func(t time.Time) tea.Msg {
		return statusTickMsg{now: t}
	})
}

func (m Model) setStatusInfo(msg string) Model {
	m.statusMsg = msg
	m.statusKind = StatusInfo
	if msg == "" {
		m.statusUntil = time.Time{}
		return m
	}
	m.statusUntil = time.Now().Add(statusAutoClear)
	return m
}

func (m Model) setStatusError(msg string) Model {
	m.statusMsg = msg
	m.statusKind = StatusError
	m.statusUntil = time.Time{}
	return m
}

func (m Model) clearStatus() Model {
	m.statusMsg = ""
	m.statusKind = StatusInfo
	m.statusUntil = time.Time{}
	return m
}
