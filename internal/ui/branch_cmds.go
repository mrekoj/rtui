package ui

import (
	"sort"

	tea "github.com/charmbracelet/bubbletea"

	"rtui/internal/git"
)

func (m Model) loadBranchesCmd(path string) tea.Cmd {
	return func() tea.Msg {
		locals, remotes, current, err := git.ListBranches(path)
		if err != nil {
			return errMsg(err)
		}
		sort.Strings(locals)
		sort.Strings(remotes)
		items := make([]BranchItem, 0, len(locals)+len(remotes))
		for _, b := range locals {
			items = append(items, BranchItem{Name: b})
		}
		for _, b := range remotes {
			items = append(items, BranchItem{Name: b, IsRemote: true})
		}
		return branchesLoadedMsg{items: items, current: current}
	}
}

func (m Model) switchBranchCmd(path string, item BranchItem, stash bool) tea.Cmd {
	return func() tea.Msg {
		if stash {
			if err := git.StashPush(path); err != nil {
				return errMsg(err)
			}
		}
		if item.IsRemote {
			if err := git.CheckoutRemoteBranch(path, item.Name); err != nil {
				return errMsg(err)
			}
		} else {
			if err := git.CheckoutBranch(path, item.Name); err != nil {
				return errMsg(err)
			}
		}
		repo := git.GetRepoStatus(path)
		return repoUpdatedMsg{repo: repo}
	}
}
