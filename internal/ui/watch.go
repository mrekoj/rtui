package ui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/git"
	"rtui/internal/watch"
)

type watchStartedMsg struct {
	manager watch.Runner
}

type repoUpdatedMsg struct {
	repo git.Repo
}

func startWatcherCmd() tea.Cmd {
	return func() tea.Msg {
		cfg := watch.Config{
			Debounce:  500 * time.Millisecond,
			WatchHead: true,
		}
		manager, err := watch.NewManager(cfg)
		if err != nil {
			return watchErrMsg(err)
		}
		manager.Start()
		return watchStartedMsg{manager: manager}
	}
}

func (m Model) watchEventsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.watcher == nil {
			return nil
		}
		ev, ok := <-m.watcher.Events()
		if !ok {
			return nil
		}
		return watchEventMsg(ev)
	}
}

func (m Model) watchErrorsCmd() tea.Cmd {
	return func() tea.Msg {
		if m.watcher == nil {
			return nil
		}
		err, ok := <-m.watcher.Errors()
		if !ok {
			return nil
		}
		return watchErrMsg(err)
	}
}

func (m Model) watchAddReposCmd(repos []git.Repo) tea.Cmd {
	return func() tea.Msg {
		if m.watcher == nil {
			return nil
		}
		for _, r := range repos {
			if r.Path == "" {
				continue
			}
			if err := m.watcher.AddRepo(r.Path); err != nil {
				return watchErrMsg(err)
			}
		}
		return nil
	}
}

func (m Model) refreshRepoCmd(path string) tea.Cmd {
	return func() tea.Msg {
		repo := git.GetRepoStatus(path)
		return repoUpdatedMsg{repo: repo}
	}
}

func (m *Model) applyRepoUpdate(updated git.Repo) {
	for i := range m.repos {
		if m.repos[i].Path == updated.Path {
			m.repos[i] = updated
			return
		}
	}
}
