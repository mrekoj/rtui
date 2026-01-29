package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
)

type Model struct {
	repos     []git.Repo
	config    config.Config
	cursor    int
	width     int
	height    int
	loading   bool
	statusMsg string
	err       error
}

// Messages
type reposLoadedMsg []git.Repo
type errMsg error

func NewModel(cfg config.Config) Model {
	return Model{
		config: cfg,
		cursor: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadRepos()
}

func (m Model) loadRepos() tea.Cmd {
	return func() tea.Msg {
		paths := m.config.Paths
		if len(paths) == 0 {
			if cwd, err := os.Getwd(); err == nil {
				paths = []string{cwd}
			}
		}
		repos := git.ScanRepos(paths, m.config.ScanDepth)
		return reposLoadedMsg(repos)
	}
}

func (m Model) visibleRepos() []git.Repo {
	return m.repos
}
