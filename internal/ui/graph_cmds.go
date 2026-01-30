package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/git"
)

const graphLimit = 50

func (m Model) loadGraphCmd(path string) tea.Cmd {
	return func() tea.Msg {
		lines, err := git.GetGraph(path, graphLimit)
		return graphLoadedMsg{lines: lines, err: err}
	}
}
