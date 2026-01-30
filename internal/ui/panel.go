package ui

import tea "github.com/charmbracelet/bubbletea"

func maxScroll(total, window int) int {
	if window <= 0 || total <= window {
		return 0
	}
	return total - window
}

func (m *Model) resetBottomScroll() {
	m.changesScroll = 0
	m.graphScroll = 0
}

func (m *Model) toggleBottomView() {
	if m.bottomView == BottomChanges {
		m.bottomView = BottomGraph
	} else {
		m.bottomView = BottomChanges
	}
}

func (m Model) maybeLoadGraph() tea.Cmd {
	if m.bottomView != BottomGraph {
		return nil
	}
	repo := m.currentRepo()
	if repo == nil {
		return nil
	}
	return m.loadGraphCmd(repo.Path)
}

func (m Model) bottomListMaxLines() int {
	maxLines := m.bottomPanelMaxLines()
	if maxLines <= 2 {
		return 0
	}
	return maxLines - 2
}

func (m *Model) scrollBottom(delta int) {
	if delta == 0 {
		return
	}
	window := m.bottomListMaxLines()
	if window <= 0 {
		return
	}
	if m.bottomView == BottomGraph {
		m.graphScroll = clamp(m.graphScroll+delta, 0, maxScroll(len(m.graphLines), window))
		return
	}
	m.changesScroll = clamp(m.changesScroll+delta, 0, maxScroll(m.changesTotalLines(), window))
}

func clamp(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}
