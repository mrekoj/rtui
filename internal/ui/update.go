package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case reposLoadedMsg:
		wasLoading := m.loading
		m.repos = msg.repos
		m.loading = false
		if msg.usedCWD && msg.cwd != "" {
			m.statusMsg = "Scanning CWD: " + msg.cwd
		} else if wasLoading {
			m.statusMsg = "Refreshed"
		}
		if m.watcher != nil {
			return m, m.watchAddReposCmd(msg.repos)
		}
		return m, nil
	case watchStartedMsg:
		m.watcher = msg.manager
		cmds := []tea.Cmd{m.watchEventsCmd(), m.watchErrorsCmd()}
		if len(m.repos) > 0 {
			cmds = append(cmds, m.watchAddReposCmd(m.repos))
		}
		return m, tea.Batch(cmds...)
	case watchEventMsg:
		return m, tea.Batch(
			m.refreshRepoCmd(msg.Repo),
			m.watchEventsCmd(),
		)
	case watchErrMsg:
		m.statusMsg = "Watcher error: " + msg.Error()
		return m, m.watchErrorsCmd()
	case repoUpdatedMsg:
		m.applyRepoUpdate(msg.repo)
		return m, nil
	case statusMsg:
		m.statusMsg = string(msg)
		return m, nil
	case errMsg:
		m.err = msg
		m.statusMsg = "Error: " + msg.Error()
		return m, nil
	case pullDoneMsg:
		m.statusMsg = "Pulled " + string(msg)
		m.mode = ModeNormal
		return m, m.loadRepos()
	case commitDoneMsg:
		m.statusMsg = "Committed " + string(msg)
		return m, m.loadRepos()
	case pushDoneMsg:
		m.statusMsg = "Pushed " + string(msg)
		return m, m.loadRepos()
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeAddPath:
		return m.handleAddPath(msg)
	case ModeCommitInput:
		return m.handleCommitInput(msg)
	case ModeConfirmPull:
		return m.handleConfirmPull(msg)
	case ModeHelp:
		return m.handleHelp(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		if m.watcher != nil {
			_ = m.watcher.Close()
		}
		return m, tea.Quit
	case "j", "down":
		if m.cursor < len(m.visibleRepos())-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "r":
		m.loading = true
		m.statusMsg = "Refreshing..."
		return m, m.loadRepos()
	case "d":
		m.filterDirty = !m.filterDirty
		m.cursor = 0
	case "a":
		m.mode = ModeAddPath
		m.addPathInput = ""
	case "c":
		if repo := m.currentRepo(); repo != nil {
			if repo.HasConflict {
				m.statusMsg = "Cannot commit: repo has conflicts"
				return m, nil
			}
			if !repo.IsDirty() {
				m.statusMsg = "Nothing to commit"
				return m, nil
			}
			m.mode = ModeCommitInput
			m.commitMsg = ""
		}
	case "o":
		if repo := m.currentRepo(); repo != nil {
			_ = git.OpenInEditor(repo.Path, m.config.Editor)
			m.statusMsg = "Opened " + repo.Name + " in " + m.config.Editor
		}
	case "f":
		if repo := m.currentRepo(); repo != nil {
			m.statusMsg = "Fetching..."
			return m, func() tea.Msg {
				if err := git.FetchAll(repo.Path); err != nil {
					return errMsg(err)
				}
				return statusMsg("Fetched " + repo.Name)
			}
		}
	case "p":
		if repo := m.currentRepo(); repo != nil {
			if repo.HasConflict {
				m.statusMsg = "Cannot push: repo has conflicts"
				return m, nil
			}
			if repo.Behind > 0 {
				m.mode = ModeConfirmPull
				return m, nil
			}
			m.statusMsg = "Pushing..."
			return m, func() tea.Msg {
				if err := git.Push(repo.Path); err != nil {
					return errMsg(err)
				}
				return pushDoneMsg(repo.Name)
			}
		}
	case "?":
		m.mode = ModeHelp
	}

	return m, nil
}

func (m Model) handleAddPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.addPathInput = ""
	case "enter":
		normalized := config.NormalizePath(m.addPathInput)
		if normalized == "" {
			return m, func() tea.Msg { return errMsg(errEmptyPath{}) }
		}
		for _, existing := range m.config.Paths {
			if config.NormalizePath(existing) == normalized {
				m.statusMsg = "Path already exists"
				m.mode = ModeNormal
				m.addPathInput = ""
				return m, nil
			}
		}
		if err := config.AppendPath(&m.config, normalized); err != nil {
			return m, func() tea.Msg { return errMsg(err) }
		}
		m.statusMsg = "Path added"
		m.mode = ModeNormal
		m.addPathInput = ""
		return m, m.loadRepos()
	case "backspace":
		if len(m.addPathInput) > 0 {
			m.addPathInput = m.addPathInput[:len(m.addPathInput)-1]
		}
	default:
		if msg.Paste && msg.Type == tea.KeyRunes {
			m.addPathInput += stripNewlines(string(msg.Runes))
		} else if len(msg.String()) == 1 {
			m.addPathInput += msg.String()
		}
	}

	return m, nil
}

func (m Model) handleCommitInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = ModeNormal
		m.commitMsg = ""
	case "enter":
		if m.commitMsg == "" {
			return m, nil
		}
		repo := m.currentRepo()
		if repo == nil {
			m.mode = ModeNormal
			return m, nil
		}
		m.statusMsg = "Committing..."
		m.mode = ModeNormal
		commitMsg := m.commitMsg
		m.commitMsg = ""
		return m, func() tea.Msg {
			if err := git.CommitAll(repo.Path, commitMsg); err != nil {
				return errMsg(err)
			}
			return commitDoneMsg(repo.Name)
		}
	case "backspace":
		if len(m.commitMsg) > 0 {
			m.commitMsg = m.commitMsg[:len(m.commitMsg)-1]
		}
	default:
		if msg.Paste && msg.Type == tea.KeyRunes {
			m.commitMsg += stripNewlines(string(msg.Runes))
		} else if len(msg.String()) == 1 {
			m.commitMsg += msg.String()
		}
	}
	return m, nil
}

func (m Model) handleConfirmPull(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		repo := m.currentRepo()
		if repo != nil {
			m.statusMsg = "Pulling..."
			m.mode = ModeNormal
			return m, func() tea.Msg {
				if err := git.Pull(repo.Path); err != nil {
					return errMsg(err)
				}
				return pullDoneMsg(repo.Name)
			}
		}
	case "n", "c", "esc":
		m.mode = ModeNormal
		m.statusMsg = "Pull canceled"
	}
	return m, nil
}

func (m Model) handleHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.mode = ModeNormal
	return m, nil
}

type errEmptyPath struct{}

func (errEmptyPath) Error() string { return "empty path" }

func stripNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
