package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
)

type Model struct {
	repos        []git.Repo
	config       config.Config
	cursor       int
	width        int
	height       int
	mode         ViewMode
	addPathInput string
	commitMsg    string
	filterDirty  bool
	loading      bool
	statusMsg    string
	err          error
}

// Messages
type reposLoadedMsg []git.Repo
type errMsg error
type statusMsg string
type pullDoneMsg string

type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeAddPath
	ModeConfirmStage
	ModeCommitInput
	ModeConfirmPull
	ModeHelp
)

func NewModel(cfg config.Config) Model {
	return Model{
		config: cfg,
		cursor: 0,
		mode:   ModeNormal,
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
	if !m.filterDirty {
		return m.repos
	}
	var result []git.Repo
	for _, r := range m.repos {
		if r.IsDirty() {
			result = append(result, r)
		}
	}
	return result
}

func (m Model) currentRepo() *git.Repo {
	repos := m.visibleRepos()
	if len(repos) == 0 || m.cursor < 0 || m.cursor >= len(repos) {
		return nil
	}
	return &repos[m.cursor]
}
