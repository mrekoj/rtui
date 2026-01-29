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

	if err := m.addRecursive(root); err != nil {
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

	if ev.Op&fsnotify.Create == fsnotify.Create {
		if isDir(path) && !shouldSkipDir(path) {
			_ = m.watcher.Add(path)
		}
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

func (m *Manager) addRecursive(root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			return nil
		}
		if shouldSkipDir(path) {
			return filepath.SkipDir
		}
		if m.cfg.Ignore != nil && m.cfg.Ignore(path) {
			return filepath.SkipDir
		}
		return m.watcher.Add(path)
	})
}

func (m *Manager) watchGitMeta(root string) error {
	indexPath := filepath.Join(root, ".git", "index")
	if fileExists(indexPath) {
		if err := m.watcher.Add(indexPath); err != nil {
			return err
		}
	}
	if m.cfg.WatchHead {
		headPath := filepath.Join(root, ".git", "HEAD")
		if fileExists(headPath) {
			if err := m.watcher.Add(headPath); err != nil {
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

func shouldSkipDir(path string) bool {
	base := filepath.Base(path)
	return base == ".git" || base == ".hg" || base == ".svn"
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
