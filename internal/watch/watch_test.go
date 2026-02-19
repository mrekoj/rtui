package watch

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeClock struct {
	now time.Time
}

type fakeTimer struct {
	ch     chan time.Time
	active bool
}

type timerFactory struct {
	created []*fakeTimer
}

func (f *timerFactory) NewTimer(d time.Duration) *fakeTimer {
	t := &fakeTimer{ch: make(chan time.Time, 1), active: true}
	f.created = append(f.created, t)
	return t
}

func (t *fakeTimer) Stop() bool {
	was := t.active
	t.active = false
	return was
}

func (t *fakeTimer) C() <-chan time.Time {
	return t.ch
}

func (t *fakeTimer) Fire() {
	if t.active {
		t.ch <- time.Now()
	}
}

type debounceHarness struct {
	debounce time.Duration
	factory  *timerFactory
	pending  map[string]*fakeTimer
	fired    []string
}

func newDebounceHarness(debounce time.Duration, factory *timerFactory) *debounceHarness {
	return &debounceHarness{
		debounce: debounce,
		factory:  factory,
		pending:  map[string]*fakeTimer{},
	}
}

func (h *debounceHarness) Schedule(repo string) {
	if t, ok := h.pending[repo]; ok {
		t.Stop()
	}
	t := h.factory.NewTimer(h.debounce)
	h.pending[repo] = t
}

func (h *debounceHarness) FireAll() {
	for repo, t := range h.pending {
		t.Fire()
		_ = repo
	}
}

func TestDebounceCoalescesPerRepo(t *testing.T) {
	factory := &timerFactory{}
	h := newDebounceHarness(500*time.Millisecond, factory)

	h.Schedule("repoA")
	h.Schedule("repoA")
	h.Schedule("repoA")

	if len(factory.created) != 3 {
		t.Fatalf("expected 3 timers created, got %d", len(factory.created))
	}
	if len(h.pending) != 1 {
		t.Fatalf("expected 1 pending timer, got %d", len(h.pending))
	}
}

func TestDebounceSeparatesRepos(t *testing.T) {
	factory := &timerFactory{}
	h := newDebounceHarness(500*time.Millisecond, factory)

	h.Schedule("repoA")
	h.Schedule("repoB")

	if len(h.pending) != 2 {
		t.Fatalf("expected 2 pending timers, got %d", len(h.pending))
	}
}

func TestIgnoreRules(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"/repo/node_modules", true},
		{"/repo/.git/objects/abc", true},
		{"/repo/.git/logs/HEAD", true},
		{"/repo/.git/index", false},
		{"/repo/.git/HEAD", false},
		{"/repo/node_modules/pkg", true},
		{"/repo/dist", true},
		{"/repo/dist/app.js", true},
		{"/repo/build", true},
		{"/repo/target", true},
		{"/repo/target/debug", true},
		{"/repo/src/main.go", false},
	}

	for _, c := range cases {
		if got := shouldIgnorePath(c.path); got != c.want {
			t.Fatalf("path %q: expected %v, got %v", c.path, c.want, got)
		}
	}
}

func TestRepoForPath(t *testing.T) {
	rootA := filepath.Join("/repos", "a")
	rootB := filepath.Join("/repos", "b")
	repos := []string{rootA, rootB}

	cases := []struct {
		path string
		want string
		ok   bool
	}{
		{filepath.Join(rootA, "file.txt"), rootA, true},
		{filepath.Join(rootA, "dir", "file.txt"), rootA, true},
		{filepath.Join(rootB, ".git", "index"), rootB, true},
		{filepath.Join("/other", "x.txt"), "", false},
	}

	for _, c := range cases {
		got, ok := repoForPath(c.path, repos)
		if ok != c.ok {
			t.Fatalf("path %q: expected ok=%v, got %v", c.path, c.ok, ok)
		}
		if got != c.want {
			t.Fatalf("path %q: expected repo %q, got %q", c.path, c.want, got)
		}
	}
}

func TestIsIgnorableWatchError(t *testing.T) {
	if !isIgnorableWatchError(&os.PathError{
		Op:   "watch",
		Path: "/tmp/missing",
		Err:  os.ErrNotExist,
	}) {
		t.Fatal("expected not-exist path error to be ignorable")
	}

	if isIgnorableWatchError(errors.New("permission denied")) {
		t.Fatal("expected non-not-exist error to be non-ignorable")
	}
}

func TestAddRepoAddsRootAndGitMetaOnly(t *testing.T) {
	root := t.TempDir()
	repoDir := filepath.Join(root, "repo")
	if err := os.Mkdir(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	if err := os.Mkdir(filepath.Join(repoDir, "src"), 0o755); err != nil {
		t.Fatalf("mkdir src: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repoDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".git", "index"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoDir, ".git", "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatalf("write HEAD: %v", err)
	}

	m, err := NewManager(Config{
		Debounce:  10 * time.Millisecond,
		WatchHead: true,
		Ignore:    func(string) bool { return false },
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer m.Close()

	if err := m.AddRepo(repoDir); err != nil {
		t.Fatalf("add repo: %v", err)
	}

	m.mu.Lock()
	_, hasRoot := m.watched[filepath.Clean(repoDir)]
	_, hasIndex := m.watched[filepath.Clean(filepath.Join(repoDir, ".git", "index"))]
	_, hasHead := m.watched[filepath.Clean(filepath.Join(repoDir, ".git", "HEAD"))]
	_, hasSrc := m.watched[filepath.Clean(filepath.Join(repoDir, "src"))]
	watchedCount := len(m.watched)
	m.mu.Unlock()

	if !hasRoot || !hasIndex || !hasHead {
		t.Fatalf("expected root/index/HEAD watched, got root=%v index=%v head=%v", hasRoot, hasIndex, hasHead)
	}
	if hasSrc {
		t.Fatal("expected src dir not to be watched in low-resource mode")
	}
	if watchedCount != 3 {
		t.Fatalf("expected exactly 3 watched paths, got %d", watchedCount)
	}
}

func TestAddRepoWithoutGitMetaWatchesRootOnly(t *testing.T) {
	repoDir := t.TempDir()
	m, err := NewManager(Config{
		Debounce: 10 * time.Millisecond,
		Ignore:   func(string) bool { return false },
	})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	defer m.Close()
	if err := m.AddRepo(repoDir); err != nil {
		t.Fatalf("add repo: %v", err)
	}
	m.mu.Lock()
	count := len(m.watched)
	_, hasRoot := m.watched[filepath.Clean(repoDir)]
	m.mu.Unlock()
	if !hasRoot {
		t.Fatal("expected root watch")
	}
	if count != 1 {
		t.Fatalf("expected one watched path, got %d", count)
	}
}
