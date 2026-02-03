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
		bottomMax := m.bottomPanelMaxLines()
		if len(m.visibleRepos()) > 0 && bottomMax >= 3 {
			b.WriteString("\n")
			b.WriteString(m.renderBottomPanel(bottomMax))
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

	start, end := m.repoWindow(len(repos))
	window := repos
	if len(repos) > 0 {
		window = repos[start:end]
	}
	for i, repo := range window {
		index := start + i
		line := m.renderRepoLine(repo, index == m.cursor, layout)
		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (m Model) renderRepoSectionHeader() string {
	title := "REPOSITORIES"
	if m.filterDirty {
		title += " (dirty only)"
	}
	label := panelLabel("1", m.panelFocus == FocusRepos)
	space := " "
	labelW := lipgloss.Width(label)
	maxTitleW := m.width
	includeLabel := true
	if labelW > 0 {
		maxTitleW = m.width - labelW - 1
		if maxTitleW < 1 {
			includeLabel = false
			maxTitleW = m.width
		}
	}
	titlePlain := title
	if lipgloss.Width(titlePlain) > maxTitleW {
		titlePlain = truncate(titlePlain, maxTitleW)
	}
	headerTitle := sectionTitleStyle.Render(titlePlain)
	header := headerTitle
	if includeLabel {
		header = headerTitle + space + label
	}

	status := m.statusMsg
	if m.loading {
		status = "Loading..."
	}
	if status == "" {
		return header
	}

	leftW := lipgloss.Width(header)
	maxRight := m.width - leftW - 1
	if maxRight < 4 {
		return header
	}
	if lipgloss.Width(status) > maxRight {
		status = truncate(status, maxRight)
	}
	right := footerStyle.Render(status)
	gap := m.width - leftW - lipgloss.Width(right)
	if gap < 1 {
		return header
	}
	return header + strings.Repeat(" ", gap) + right
}

func panelLabel(key string, focused bool) string {
	label := "[" + key + "]"
	if focused {
		return selectedRepoStyle.Render(label)
	}
	return footerStyle.Render(label)
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

func (m Model) renderBottomPanel(maxLines int) string {
	if m.bottomView == BottomGraph {
		return m.renderGraphPanel(maxLines)
	}
	return m.renderChangesPanel(maxLines)
}

func (m Model) renderChangesPanel(maxLines int) string {
	repo := m.currentRepo()
	if repo == nil {
		return ""
	}

	var b strings.Builder
	header := sectionTitleStyle.Render("CHANGES") + " " + panelLabel("2", m.panelFocus == FocusBottom)
	b.WriteString(header)
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

	lines := m.changesLines(staged, modified, untracked)
	contentMax := maxLines - 2
	if contentMax < 1 {
		return b.String()
	}
	start := clamp(m.changesScroll, 0, maxScroll(len(lines), contentMax))
	end := min(start+contentMax, len(lines))
	m.writePanelLines(&b, lines[start:end], contentMax)
	return b.String()
}

func (m Model) renderAddPath() string {
	msg := "Add repo path"
	maxW := max(m.width-4, 20)
	targetW := int(float64(m.width) * 0.7)
	boxW := min(maxW, max(30, targetW))
	inputW := boxW - 4
	input := m.addPathInput + "█"
	if len(input) > inputW {
		input = input[len(input)-inputW:]
	}
	body := msg + "\n\n" + input + "\n\n[Enter]=save  [Esc]=cancel"
	modal := boxStyle.Width(boxW).Render(body)
	return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, modal)
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

func (m Model) renderGraphPanel(maxLines int) string {
	repo := m.currentRepo()
	if repo == nil {
		return ""
	}
	var b strings.Builder
	header := sectionTitleStyle.Render("GRAPH") + " " + panelLabel("2", m.panelFocus == FocusBottom)
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", m.width))
	b.WriteString("\n")

	contentMax := maxLines - 2
	if contentMax < 1 {
		return b.String()
	}
	lines := m.graphLines
	if len(lines) == 0 {
		lines = []string{footerStyle.Render("  No commits")}
	}
	start := clamp(m.graphScroll, 0, maxScroll(len(lines), contentMax))
	end := min(start+contentMax, len(lines))
	window := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		window = append(window, truncate(line, m.width))
	}
	m.writePanelLines(&b, window, contentMax)
	return b.String()
}

func (m Model) changesLines(staged, modified, untracked []git.ChangedFile) []string {
	maxPathW := m.width - 4
	lines := []string{
		stagedStyle.Render(fmt.Sprintf("Staged (%d)", len(staged))),
	}
	for _, f := range staged {
		lines = append(lines, "  "+truncatePath(f.Path, maxPathW))
	}
	lines = append(lines, modifiedStyle.Render(fmt.Sprintf("Modified (%d)", len(modified))))
	for _, f := range modified {
		lines = append(lines, "  "+truncatePath(f.Path, maxPathW))
	}
	lines = append(lines, untrackedStyle.Render(fmt.Sprintf("Untracked (%d)", len(untracked))))
	for _, f := range untracked {
		lines = append(lines, "  "+truncatePath(f.Path, maxPathW))
	}
	return lines
}

func (m Model) changesTotalLines() int {
	repo := m.currentRepo()
	if repo == nil {
		return 0
	}
	staged := 0
	modified := 0
	untracked := 0
	for _, f := range repo.ChangedFiles {
		switch f.Status {
		case git.StatusStaged:
			staged++
		case git.StatusModified:
			modified++
		case git.StatusUntracked:
			untracked++
		}
	}
	return 3 + staged + modified + untracked
}

func (m Model) writePanelLines(b *strings.Builder, lines []string, contentMax int) {
	for i := 0; i < contentMax; i++ {
		if i < len(lines) {
			b.WriteString(lines[i])
		}
		b.WriteString("\n")
	}
}

func (m Model) bottomPanelMaxLines() int {
	repoLines := m.repoListLineCount()
	bodyMax := m.bodyMaxLines()
	maxLines := bodyMax - repoLines - 1
	if maxLines <= 0 {
		return 0
	}
	return maxLines
}

func (m Model) repoListLineCount() int {
	repos := m.visibleRepos()
	if len(repos) == 0 {
		return 3
	}
	return m.repoFixedLines() + m.repoWindowSize(len(repos))
}

func (m Model) repoFixedLines() int {
	return 4
}

func (m Model) repoWindowSize(total int) int {
	if total <= 0 {
		return 0
	}
	start, end := m.repoWindow(total)
	if end < start {
		return 0
	}
	return end - start
}

func (m Model) repoWindow(total int) (start, end int) {
	if total <= 0 {
		return 0, 0
	}
	maxRows := m.maxRepoRows(total)
	if maxRows <= 0 {
		return 0, 0
	}
	start, end = branchWindow(total, m.cursor, maxRows)
	return start, end
}

func (m Model) maxRepoRows(total int) int {
	if total <= 0 {
		return 0
	}
	fixed := m.repoFixedLines()
	available := m.bodyMaxLines() - fixed
	if available < 1 {
		return 0
	}
	reserve := 4 // gap + min bottom panel lines (3)
	maxRows := available
	if available > reserve {
		maxRows = available - reserve
	}
	if maxRows < 1 {
		maxRows = 1
	}
	if maxRows > total {
		maxRows = total
	}
	return maxRows
}

func (m Model) bodyMaxLines() int {
	footerLines := m.footerLineCount()
	bodyMax := m.height - footerLines
	if bodyMax < 0 {
		return 0
	}
	return bodyMax
}

func (m Model) footerLineCount() int {
	footer := m.renderFooter()
	if footer == "" {
		return 0
	}
	return strings.Count(footer, "\n") + 1
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
  s       Settings (open config in editor)
  p       Pull
  P       Push
  f       Fetch all
  r       Refresh

Panels
  1       Focus repo list
  2       Focus bottom panel
  Tab     Toggle Changes/Graph (bottom panel)

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
	return footerActions(m.width)
}

type footerToken struct {
	plain  string
	styled string
}

func footerActions(width int) string {
	if width <= 0 {
		return ""
	}

	tokens := footerTokens()
	gap := 3
	lines := wrapFooterTokens(tokens, width, 2, gap)
	if len(lines) == 0 {
		return ""
	}

	styledSpace := footerStyle.Render(strings.Repeat(" ", gap))
	var out []string
	for _, lineTokens := range lines {
		if len(lineTokens) == 0 {
			continue
		}
		line := lineTokens[0].styled
		for i := 1; i < len(lineTokens); i++ {
			line += styledSpace + lineTokens[i].styled
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func footerTokens() []footerToken {
	return []footerToken{
		footerTokenHotkey("a", "add path"),
		footerTokenHotkey("b", "branch"),
		footerTokenHotkey("c", "commit"),
		footerTokenHotkey("P", "push"),
		footerTokenHotkey("o", "open"),
		footerTokenHotkey("p", "pull"),
		footerTokenHotkey("r", "refresh"),
		footerTokenHotkey("?", "?"),
	}
}

func footerTokenHotkey(key, label string) footerToken {
	keyLower := strings.ToLower(key)
	labelLower := strings.ToLower(label)
	index := strings.Index(labelLower, keyLower)
	if index < 0 {
		plain := key + label
		styled := hotkeyStyle.Render(key) + footerStyle.Render(label)
		return footerToken{plain: plain, styled: styled}
	}

	prefix := label[:index]
	match := label[index : index+1]
	suffix := label[index+1:]

	plain := label
	styled := footerStyle.Render(prefix) + hotkeyStyle.Render(match) + footerStyle.Render(suffix)
	return footerToken{plain: plain, styled: styled}
}

func wrapFooterTokens(tokens []footerToken, width, maxLines, gap int) [][]footerToken {
	if width <= 0 {
		return nil
	}
	var lines [][]footerToken
	var current []footerToken
	currentWidth := 0

	for _, tok := range tokens {
		tokW := lipgloss.Width(tok.plain)
		if len(current) == 0 {
			current = append(current, tok)
			currentWidth = tokW
			continue
		}
		if currentWidth+gap+tokW <= width {
			current = append(current, tok)
			currentWidth += gap + tokW
			continue
		}
		lines = append(lines, current)
		current = []footerToken{tok}
		currentWidth = tokW
	}
	if len(current) > 0 {
		lines = append(lines, current)
	}

	if maxLines > 0 && len(lines) > maxLines {
		line1, line2 := packFooterTokens(tokens, width, gap)
		lines = [][]footerToken{line1, line2}
	}

	return lines
}

func packFooterTokens(tokens []footerToken, width, gap int) ([]footerToken, []footerToken) {
	var line1 []footerToken
	var line2 []footerToken
	w1 := 0
	w2 := 0

	for _, tok := range tokens {
		tokW := lipgloss.Width(tok.plain)
		if len(line1) == 0 || w1+gap+tokW <= width {
			if len(line1) == 0 {
				line1 = append(line1, tok)
				w1 = tokW
			} else {
				line1 = append(line1, tok)
				w1 += gap + tokW
			}
			continue
		}
		if len(line2) == 0 || w2+gap+tokW <= width {
			if len(line2) == 0 {
				line2 = append(line2, tok)
				w2 = tokW
			} else {
				line2 = append(line2, tok)
				w2 += gap + tokW
			}
		}
	}

	if len(tokens) == 0 {
		return line1, line2
	}
	help := tokens[len(tokens)-1]
	if !containsFooterToken(line1, help) && !containsFooterToken(line2, help) {
		helpW := lipgloss.Width(help.plain)
		if helpW <= width {
			for len(line2) > 0 && w2+gap+helpW > width {
				removed := line2[len(line2)-1]
				w2 -= lipgloss.Width(removed.plain)
				if len(line2) > 1 {
					w2 -= gap
				}
				line2 = line2[:len(line2)-1]
			}
			if len(line2) == 0 {
				line2 = append(line2, help)
			} else if w2+gap+helpW <= width {
				line2 = append(line2, help)
			}
		}
	}

	return line1, line2
}

func containsFooterToken(tokens []footerToken, target footerToken) bool {
	for _, tok := range tokens {
		if tok.plain == target.plain {
			return true
		}
	}
	return false
}

func padToBottom(height int, body, footer string) string {
	bodyLines := lineCount(body)
	footerLines := lineCount(footer)
	if footer == "" {
		footerLines = 0
	}
	gap := height - bodyLines - footerLines
	if gap < 0 {
		gap = 0
	}
	sep := ""
	if body != "" && !strings.HasSuffix(body, "\n") {
		sep = "\n"
	}
	return sep + strings.Repeat("\n", gap) + footer
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

func lineCount(s string) int {
	if s == "" {
		return 0
	}
	return strings.Count(s, "\n") + 1
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
