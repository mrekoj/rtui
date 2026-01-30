package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"rtui/internal/git"
)

type Layout struct {
	Name   int
	Branch int
	Status int
	Sync   int
}

func (m Model) calculateLayout() Layout {
	w := m.width

	cursorW := 2
	syncW := 6
	statusW := 8
	gaps := 9

	remaining := w - cursorW - syncW - statusW - gaps

	if w < 40 {
		gaps = 6
		remaining = w - cursorW - syncW - statusW - gaps
		return Layout{
			Name:   max(remaining-2, 8),
			Branch: 0,
			Status: 8,
			Sync:   6,
		}
	}

	nameW := int(float64(remaining) * 0.55)
	branchW := remaining - nameW

	nameW = max(nameW, 10)
	branchW = max(branchW, 8)

	if w > 120 {
		nameW = min(nameW, 30)
		branchW = min(branchW, 25)
	}

	return Layout{
		Name:   nameW,
		Branch: branchW,
		Status: statusW,
		Sync:   syncW,
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	switch m.mode {
	case ModeHelp:
		b.WriteString(m.renderHelp())
	case ModeAddPath:
		b.WriteString(m.renderRepoList())
		b.WriteString("\n")
		b.WriteString(m.renderAddPath())
	case ModeCommitInput:
		b.WriteString(m.renderRepoList())
		b.WriteString("\n")
		b.WriteString(m.renderCommitInput())
	case ModeConfirmPull:
		b.WriteString(m.renderRepoList())
		b.WriteString("\n")
		b.WriteString(m.renderPullConfirm())
	case ModeBranchPicker:
		b.WriteString(m.renderRepoList())
		b.WriteString("\n")
		b.WriteString(m.renderBranchPicker())
	case ModeConfirmStash:
		b.WriteString(m.renderRepoList())
		b.WriteString("\n")
		b.WriteString(m.renderStashConfirm())
	default:
		b.WriteString(m.renderRepoList())
		if len(m.visibleRepos()) > 0 && m.height >= 15 {
			b.WriteString("\n")
			b.WriteString(m.renderChangesPanel())
		}
	}

	footer := m.renderFooter()
	b.WriteString(padToBottom(m.height, b.String(), footer))

	return b.String()
}

func (m Model) renderRepoList() string {
	var b strings.Builder

	b.WriteString(m.renderRepoSectionHeader())
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	repos := m.visibleRepos()
	if len(repos) == 0 {
		b.WriteString(footerStyle.Render("  No repositories found"))
		return b.String()
	}

	layout := m.calculateLayout()

	b.WriteString(m.renderRepoHeader(layout))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	for i, repo := range repos {
		line := m.renderRepoLine(repo, i == m.cursor, layout)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderRepoSectionHeader() string {
	header := "REPOSITORIES"
	if m.filterDirty {
		header += " (dirty only)"
	}

	status := m.statusMsg
	if m.loading {
		status = "Loading..."
	}
	if status == "" {
		return sectionTitleStyle.Render(header)
	}

	right := footerStyle.Render(status)
	rightW := lipgloss.Width(right)
	maxLeft := m.width - rightW - 1
	if maxLeft <= 0 {
		return right
	}
	if lipgloss.Width(header) > maxLeft {
		header = truncate(header, maxLeft)
	}
	left := sectionTitleStyle.Render(header)
	gap := m.width - lipgloss.Width(left) - rightW
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func (m Model) renderRepoHeader(layout Layout) string {
	cursor := "  "
	name := padRight("Name", layout.Name)
	status := padRight("Status", layout.Status)
	sync := padRight("Sync", layout.Sync)

	var branch string
	if layout.Branch > 0 {
		branch = padRight("Branch", layout.Branch)
	}

	var line string
	if layout.Branch > 0 {
		line = cursor + name + " | " + branch + " | " + status + " | " + sync
	} else {
		line = cursor + name + " | " + status + " | " + sync
	}

	return footerStyle.Render(line)
}

func (m Model) renderRepoLine(repo git.Repo, isCursor bool, layout Layout) string {
	cursor := "  "
	if isCursor {
		cursor = "→ "
	}

	name := padRight(truncate(repo.Name, layout.Name), layout.Name)
	if isCursor {
		name = selectedRepoStyle.Render(name)
	}

	var branch string
	if layout.Branch > 0 {
		branch = padRight(truncate(repo.Branch, layout.Branch), layout.Branch)
		if isCursor {
			branch = selectedRepoStyle.Render(branch)
		}
	}

	var status string
	if repo.IsDirty() {
		parts := []string{}
		if repo.Modified > 0 {
			parts = append(parts, modifiedStyle.Render(fmt.Sprintf("%dM", repo.Modified)))
		}
		if repo.Staged > 0 {
			parts = append(parts, stagedStyle.Render(fmt.Sprintf("%dS", repo.Staged)))
		}
		if repo.Untracked > 0 {
			parts = append(parts, untrackedStyle.Render(fmt.Sprintf("%dU", repo.Untracked)))
		}
		status = strings.Join(parts, " ")
	} else {
		status = stagedStyle.Render("✓")
	}

	statusPadded := padRight(status, layout.Status)

	var sync string
	if repo.HasConflict {
		sync = conflictStyle.Render("CONFLICT")
	} else {
		if repo.Ahead > 0 {
			sync += aheadStyle.Render(fmt.Sprintf("↑%d", repo.Ahead))
		}
		if repo.Behind > 0 {
			sync += behindStyle.Render(fmt.Sprintf("↓%d", repo.Behind))
		}
	}
	if strings.TrimSpace(sync) == "" {
		sync = footerStyle.Render("-")
	}
	sync = padRight(sync, layout.Sync)

	var line string
	if layout.Branch > 0 {
		line = cursor + name + " | " + branch + " | " + statusPadded + " | " + sync
	} else {
		line = cursor + name + " | " + statusPadded + " | " + sync
	}

	if repo.IsDirty() {
		line = dirtyRepoStyle.Render(line)
	} else {
		line = cleanRepoStyle.Render(line)
	}

	return line
}

func (m Model) renderChangesPanel() string {
	repo := m.currentRepo()
	if repo == nil {
		return ""
	}

	var b strings.Builder
	header := fmt.Sprintf("CHANGES: %s (%s)", repo.Name, repo.Branch)
	b.WriteString(sectionTitleStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	staged := []git.ChangedFile{}
	modified := []git.ChangedFile{}
	untracked := []git.ChangedFile{}

	for _, f := range repo.ChangedFiles {
		switch f.Status {
		case git.StatusStaged:
			staged = append(staged, f)
		case git.StatusModified:
			modified = append(modified, f)
		case git.StatusUntracked:
			untracked = append(untracked, f)
		}
	}

	maxPathW := m.width - 4

	b.WriteString(stagedStyle.Render(fmt.Sprintf("Staged (%d)", len(staged))))
	b.WriteString("\n")
	for _, f := range staged {
		b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
	}

	b.WriteString(modifiedStyle.Render(fmt.Sprintf("Modified (%d)", len(modified))))
	b.WriteString("\n")
	for _, f := range modified {
		b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
	}

	b.WriteString(untrackedStyle.Render(fmt.Sprintf("Untracked (%d)", len(untracked))))
	b.WriteString("\n")
	for _, f := range untracked {
		b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
	}

	return b.String()
}

func (m Model) renderAddPath() string {
	msg := "Add repo path"
	boxW := min(m.width-4, 50)
	inputW := boxW - 4
	input := m.addPathInput + "█"
	if len(input) > inputW {
		input = input[len(input)-inputW:]
	}
	body := msg + "\n\n" + input + "\n\n[Enter]=save  [Esc]=cancel"
	return boxStyle.Width(boxW).Render(body)
}

func (m Model) renderCommitInput() string {
	var b strings.Builder
	b.WriteString("Commit message (stages all):\n")

	inputW := min(m.width-4, 60)
	input := m.commitMsg + "█"
	if len(input) > inputW {
		input = input[len(input)-inputW:]
	}

	b.WriteString(inputStyle.Width(inputW).Render(input))
	b.WriteString("\n")
	b.WriteString(footerStyle.Render("[Enter] commit  [Esc] cancel"))
	return b.String()
}

func (m Model) renderBranchPicker() string {
	var b strings.Builder
	tabLocal := "Local"
	tabRemote := "Remote"
	if m.branchTab == BranchTabLocal {
		tabLocal = "[" + tabLocal + "]"
	} else {
		tabRemote = "[" + tabRemote + "]"
	}
	b.WriteString("Switch branch  " + tabLocal + " " + tabRemote + "\n")
	b.WriteString(footerStyle.Render("Filter: " + m.branchFilter()))
	b.WriteString("\n\n")

	boxW := min(m.width-4, 60)
	contentW := boxW - 4

	items := m.filteredBranches()
	maxList := m.height - 8
	if maxList < 3 {
		maxList = 3
	}
	start, end, showTop, showBottom := branchWindowInfo(len(items), m.branchCursor, maxList)
	window := items
	if len(items) > 0 {
		window = items[start:end]
	}

	if showTop {
		b.WriteString(footerStyle.Render("  ↑ more"))
		b.WriteString("\n")
	}
	current := ""
	if repo := m.currentRepo(); repo != nil {
		current = repo.Branch
	}
	if len(items) == 0 {
		b.WriteString(footerStyle.Render("  No branches"))
	} else {
		for i, item := range window {
			cursor := "  "
			if i+start == m.branchCursor {
				cursor = "→ "
			}
			name := item.Name
			if name == current {
				name = stagedStyle.Render(name)
			}
			line := cursor + truncate(name, contentW-2)
			b.WriteString(line + "\n")
		}
	}
	if showBottom {
		b.WriteString(footerStyle.Render("  ↓ more"))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(footerStyle.Render("[Enter] switch  [Esc] cancel"))

	return boxStyle.Width(boxW).Render(b.String())
}

func (m Model) renderStashConfirm() string {
	msg := "Repo has uncommitted changes. Stash and switch?"
	boxW := min(m.width-4, 60)
	return boxStyle.Width(boxW).Render(msg + "\n\n[s]tash  [c]ancel")
}

func (m Model) renderPullConfirm() string {
	repo := m.currentRepo()
	if repo == nil {
		return ""
	}

	msg := fmt.Sprintf("Repo is %d commits behind. Pull first?", repo.Behind)
	boxW := min(m.width-4, 50)
	return boxStyle.Width(boxW).Render(msg + "\n\n[y]es  [n]o  [c]ancel")
}

func (m Model) renderHelp() string {
	help := `KEYBINDINGS

Navigation
  j/↓     Next repo
  k/↑     Previous repo

Actions
  a       Add path
  c       Commit (stages all)
  b       Switch branch
  o       Open in editor
  p       Push
  f       Fetch all
  r       Refresh

Filters
  d       Toggle dirty-only

Other
  ?       This help
  q       Quit

Press any key to close...`

	boxW := min(m.width-4, 40)
	return boxStyle.Width(boxW).Render(help)
}

func (m Model) renderFooter() string {
	actions := "[a]dd path  [c]ommit  [b]ranch  [o]pen  [p]ush  [r]efresh  [?]help"
	if m.width < 50 {
		actions = "a:add c:com b:br o:open p:push r:ref ?:help"
	}

	return footerStyle.Render(actions)
}

func padToBottom(height int, body, footer string) string {
	lines := strings.Split(body, "\n")
	if height <= len(lines)+1 {
		return "\n" + footer
	}
	gap := height - len(lines) - 1
	return "\n" + strings.Repeat("\n", gap) + footer
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, maxW int) string {
	if len(s) <= maxW {
		return s
	}
	if maxW <= 1 {
		return s[:maxW]
	}
	return s[:maxW-1] + "…"
}

func truncatePath(path string, maxW int) string {
	if len(path) <= maxW {
		return path
	}
	if maxW <= 3 {
		return path[len(path)-maxW:]
	}
	return "…" + path[len(path)-maxW+1:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
