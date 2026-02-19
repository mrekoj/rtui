package watch

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Config controls watcher behavior.
type Config struct {
	Debounce  time.Duration
	WatchHead bool
	Ignore    func(string) bool
}

// Runner defines the watcher interface for reuse/mocking.
type Runner interface {
	Start()
	Close() error
	Events() <-chan Event
	Errors() <-chan error
	AddRepo(path string) error
}

// Event indicates a repo change.
type Event struct {
	Repo string
	Path string
}

// Manager watches multiple repos and emits debounced repo events.
type Manager struct {
	cfg      Config
	watcher  *fsnotify.Watcher
	events   chan Event
	errors   chan error
	done     chan struct{}
	mu       sync.Mutex
	repos    []string
	watched  map[string]struct{}
	debounce map[string]*time.Timer
	closed   bool
}

// NewManager creates a watcher manager.
func NewManager(cfg Config) (*Manager, error) {
	if cfg.Debounce <= 0 {
		cfg.Debounce = 500 * time.Millisecond
	}
	if cfg.Ignore == nil {
		cfg.Ignore = shouldIgnorePath
	}
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	m := &Manager{
		cfg:      cfg,
		watcher:  w,
		events:   make(chan Event, 64),
		errors:   make(chan error, 8),
		done:     make(chan struct{}),
		watched:  map[string]struct{}{},
		debounce: map[string]*time.Timer{},
	}
	return m, nil
}

// Start begins processing fsnotify events.
func (m *Manager) Start() {
	go m.loop()
}

// Close stops the watcher.
func (m *Manager) Close() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	for _, t := range m.debounce {
		t.Stop()
	}
	m.mu.Unlock()
	close(m.done)
	err := m.watcher.Close()
	close(m.events)
	close(m.errors)
	return err
}

// Events returns a channel of repo change events.
func (m *Manager) Events() <-chan Event {
	return m.events
}

// Errors returns a channel of watcher errors.
func (m *Manager) Errors() <-chan error {
	return m.errors
}

// AddRepo adds a repo root to the watch set.
func (m *Manager) AddRepo(path string) error {
	root := filepath.Clean(path)
	if root == "." || root == string(filepath.Separator) {
		return errors.New("invalid repo path")
	}

	m.mu.Lock()
	for _, r := range m.repos {
		if r == root {
			m.mu.Unlock()
			return nil
		}
	}
	m.repos = append(m.repos, root)
	m.mu.Unlock()

	if err := m.addRoot(root); err != nil {
		return err
	}
	if err := m.watchGitMeta(root); err != nil {
		return err
	}
	return nil
}

func (m *Manager) loop() {
	for {
		select {
		case <-m.done:
			return
		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			if isIgnorableWatchError(err) {
				continue
			}
			m.sendError(err)
		case ev, ok := <-m.watcher.Events:
			if !ok {
				return
			}
			m.handleEvent(ev)
		}
	}
}

func (m *Manager) handleEvent(ev fsnotify.Event) {
	path := filepath.Clean(ev.Name)
	if m.cfg.Ignore != nil && m.cfg.Ignore(path) {
		return
	}

	repo, ok := repoForPath(path, m.reposSnapshot())
	if !ok {
		return
	}
	m.schedule(repo, path)
}

func (m *Manager) schedule(repo, path string) {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return
	}
	if t, ok := m.debounce[repo]; ok {
		t.Stop()
	}
	m.debounce[repo] = time.AfterFunc(m.cfg.Debounce, func() {
		select {
		case m.events <- Event{Repo: repo, Path: path}:
		case <-m.done:
		}
	})
	m.mu.Unlock()
}

func (m *Manager) sendError(err error) {
	select {
	case m.errors <- err:
	case <-m.done:
	}
}

func (m *Manager) addRoot(root string) error {
	if m.cfg.Ignore != nil && m.cfg.Ignore(root) {
		return nil
	}
	return m.addWatch(root)
}

func (m *Manager) watchGitMeta(root string) error {
	indexPath := filepath.Join(root, ".git", "index")
	if fileExists(indexPath) {
		if err := m.addWatch(indexPath); err != nil {
			return err
		}
	}
	if m.cfg.WatchHead {
		headPath := filepath.Join(root, ".git", "HEAD")
		if fileExists(headPath) {
			if err := m.addWatch(headPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *Manager) reposSnapshot() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot := make([]string, len(m.repos))
	copy(snapshot, m.repos)
	return snapshot
}

func repoForPath(path string, repos []string) (string, bool) {
	cleanPath := filepath.Clean(path)
	for _, root := range repos {
		cleanRoot := filepath.Clean(root)
		if cleanPath == cleanRoot {
			return cleanRoot, true
		}
		if strings.HasPrefix(cleanPath, cleanRoot+string(filepath.Separator)) {
			return cleanRoot, true
		}
	}
	return "", false
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isIgnorableWatchError(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

func (m *Manager) addWatch(path string) error {
	cleanPath := filepath.Clean(path)

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	if _, ok := m.watched[cleanPath]; ok {
		m.mu.Unlock()
		return nil
	}
	m.mu.Unlock()

	if err := m.watcher.Add(cleanPath); err != nil {
		if isIgnorableWatchError(err) {
			return nil
		}
		return err
	}

	m.mu.Lock()
	if !m.closed {
		m.watched[cleanPath] = struct{}{}
	}
	m.mu.Unlock()

	return nil
}
