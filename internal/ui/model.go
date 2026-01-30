package ui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/config"
	"rtui/internal/git"
	"rtui/internal/watch"
)

type Model struct {
	repos              []git.Repo
	config             config.Config
	cursor             int
	width              int
	height             int
	mode               ViewMode
	panelFocus         PanelFocus
	bottomView         BottomView
	addPathInput       string
	commitMsg          string
	filterDirty        bool
	branchItems        []BranchItem
	branchFilterLocal  string
	branchFilterRemote string
	branchCursor       int
	branchTab          BranchTab
	pendingBranch      BranchItem
	changesScroll      int
	graphScroll        int
	graphLines         []string
	loading            bool
	statusMsg          string
	err                error
	watcher            watch.Runner
}

// Messages
type reposLoadedMsg struct {
	repos   []git.Repo
	usedCWD bool
	cwd     string
}
type errMsg error
type statusMsg string
type pullDoneMsg string
type commitDoneMsg string
type pushDoneMsg string
type watchEventMsg watch.Event
type watchErrMsg error
type branchesLoadedMsg struct {
	items   []BranchItem
	current string
}
type graphLoadedMsg struct {
	lines []string
	err   error
}

type ViewMode int

const (
	ModeNormal ViewMode = iota
	ModeAddPath
	ModeCommitInput
	ModeBranchPicker
	ModeConfirmStash
	ModeHelp
)

type PanelFocus int

const (
	FocusRepos PanelFocus = iota
	FocusBottom
)

type BottomView int

const (
	BottomChanges BottomView = iota
	BottomGraph
)

func NewModel(cfg config.Config) Model {
	return Model{
		config:      cfg,
		cursor:      0,
		mode:        ModeNormal,
		panelFocus:  FocusRepos,
		bottomView:  BottomChanges,
		branchTab:   BranchTabLocal,
		changesScroll: 0,
		graphScroll:   0,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadRepos(),
		startWatcherCmd(),
	)
}

func (m Model) loadRepos() tea.Cmd {
	return func() tea.Msg {
		paths := m.config.Paths
		usedCWD := false
		cwd := ""
		if len(paths) == 0 {
			if cwd, err := os.Getwd(); err == nil {
				paths = []string{cwd}
				usedCWD = true
				cwd = cwd
			}
		}
		repos := git.ScanRepos(paths, m.config.ScanDepth)
		return reposLoadedMsg{repos: repos, usedCWD: usedCWD, cwd: cwd}
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
