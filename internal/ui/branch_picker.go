package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type BranchItem struct {
	Name     string
	IsRemote bool
}

func filterBranches(items []BranchItem, filter string) []BranchItem {
	if filter == "" {
		return items
	}
	needle := strings.ToLower(filter)
	var out []BranchItem
	for _, item := range items {
		if strings.Contains(strings.ToLower(item.Name), needle) {
			out = append(out, item)
		}
	}
	return out
}

func indexOfBranch(items []BranchItem, name string) int {
	for i, item := range items {
		if item.Name == name {
			return i
		}
	}
	return 0
}

func branchWindow(total, cursor, max int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	if max <= 0 {
		max = 1
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}
	start = 0
	end = min(total, max)
	if cursor >= end {
		start = cursor - max + 1
		if start < 0 {
			start = 0
		}
		end = min(total, start+max)
	}
	return start, end
}

func (m Model) filteredBranches() []BranchItem {
	return filterBranches(m.branchItems, m.branchFilter)
}

func (m Model) selectedBranch() (BranchItem, bool) {
	items := m.filteredBranches()
	if len(items) == 0 || m.branchCursor < 0 || m.branchCursor >= len(items) {
		return BranchItem{}, false
	}
	return items[m.branchCursor], true
}

func (m Model) handleBranchPicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		return m, nil
	case "enter":
		item, ok := m.selectedBranch()
		if !ok {
			return m, nil
		}
		repo := m.currentRepo()
		if repo == nil {
			m.mode = ModeNormal
			return m, nil
		}
		if item.Name == repo.Branch {
			m.mode = ModeNormal
			m.statusMsg = "Already on " + item.Name
			return m, nil
		}
		if repo.IsDirty() {
			m.pendingBranch = item
			m.mode = ModeConfirmStash
			return m, nil
		}
		m.mode = ModeNormal
		m.statusMsg = "Switching to " + item.Name + "..."
		return m, m.switchBranchCmd(repo.Path, item, false)
	case "backspace":
		if len(m.branchFilter) > 0 {
			m.branchFilter = m.branchFilter[:len(m.branchFilter)-1]
			m.branchCursor = 0
		}
	case "j", "down":
		items := m.filteredBranches()
		if m.branchCursor < len(items)-1 {
			m.branchCursor++
		}
	case "k", "up":
		if m.branchCursor > 0 {
			m.branchCursor--
		}
	default:
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			m.branchFilter += string(msg.Runes)
			m.branchCursor = 0
		}
	}
	return m, nil
}

func (m Model) handleConfirmStash(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		repo := m.currentRepo()
		if repo == nil {
			m.mode = ModeNormal
			return m, nil
		}
		item := m.pendingBranch
		m.pendingBranch = BranchItem{}
		m.mode = ModeNormal
		m.statusMsg = "Stashing and switching..."
		return m, m.switchBranchCmd(repo.Path, item, true)
	case "c", "esc":
		m.pendingBranch = BranchItem{}
		m.mode = ModeBranchPicker
		return m, nil
	}
	return m, nil
}
