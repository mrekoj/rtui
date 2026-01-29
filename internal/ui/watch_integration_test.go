package ui

import (
	"testing"

	"rtui/internal/git"
	"rtui/internal/watch"
)

type fakeWatcher struct {
	added  []string
	events chan watch.Event
	errors chan error
}

func newFakeWatcher() *fakeWatcher {
	return &fakeWatcher{
		events: make(chan watch.Event, 1),
		errors: make(chan error, 1),
	}
}

func (f *fakeWatcher) Start()                     {}
func (f *fakeWatcher) Close() error               { return nil }
func (f *fakeWatcher) Events() <-chan watch.Event { return f.events }
func (f *fakeWatcher) Errors() <-chan error       { return f.errors }
func (f *fakeWatcher) AddRepo(path string) error {
	f.added = append(f.added, path)
	return nil
}

func TestWatchAddReposCmdAddsPaths(t *testing.T) {
	fw := newFakeWatcher()
	m := Model{watcher: fw}
	repos := []git.Repo{
		{Path: "/repo/a"},
		{Path: "/repo/b"},
	}

	cmd := m.watchAddReposCmd(repos)
	if cmd == nil {
		t.Fatal("expected non-nil cmd")
	}
	_ = cmd()

	if len(fw.added) != 2 {
		t.Fatalf("expected 2 repos added, got %d", len(fw.added))
	}
}

func TestApplyRepoUpdate(t *testing.T) {
	m := Model{repos: []git.Repo{{Path: "/repo/a", Branch: "main"}}}
	updated := git.Repo{Path: "/repo/a", Branch: "dev"}
	m.applyRepoUpdate(updated)

	if m.repos[0].Branch != "dev" {
		t.Fatalf("expected branch updated, got %q", m.repos[0].Branch)
	}
}

func TestRepoUpdatedMsgUpdatesModel(t *testing.T) {
	m := Model{repos: []git.Repo{{Path: "/repo/a", Branch: "main"}}}
	msg := repoUpdatedMsg{repo: git.Repo{Path: "/repo/a", Branch: "dev"}}
	m2, _ := m.Update(msg)
	m = m2.(Model)

	if m.repos[0].Branch != "dev" {
		t.Fatalf("expected branch updated, got %q", m.repos[0].Branch)
	}
}
