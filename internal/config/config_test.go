package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := NormalizePath(" ~/repo ")
	want := filepath.Join(home, "repo")
	if got != want {
		t.Fatalf("NormalizePath() = %q, want %q", got, want)
	}
}

func TestAppendPathCreatesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoPath := filepath.Join(home, "proj")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cfg := DefaultConfig()
	if err := AppendPath(&cfg, repoPath); err != nil {
		t.Fatalf("AppendPath: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(loaded.Paths))
	}
	if loaded.Paths[0] != repoPath {
		t.Fatalf("expected %q, got %q", repoPath, loaded.Paths[0])
	}
}

func TestAppendPathDuplicate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	repoPath := filepath.Join(home, "proj")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cfg := DefaultConfig()
	if err := AppendPath(&cfg, repoPath); err != nil {
		t.Fatalf("AppendPath: %v", err)
	}
	if err := AppendPath(&cfg, repoPath); err != nil {
		t.Fatalf("AppendPath duplicate: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(loaded.Paths) != 1 {
		t.Fatalf("expected 1 path, got %d", len(loaded.Paths))
	}
}

func TestAppendPathMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	missing := filepath.Join(home, "missing")
	cfg := DefaultConfig()
	if err := AppendPath(&cfg, missing); err == nil {
		t.Fatal("expected error for missing path")
	}

	path := filepath.Join(home, ".config", "rtui", "config.toml")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no config file, got %v", err)
	}
}

func TestLoadNormalizesPaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfgDir := filepath.Join(home, ".config", "rtui")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	configFile := filepath.Join(cfgDir, "config.toml")
	content := "paths = [\"~/repo\"]\n"
	if err := os.WriteFile(configFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := filepath.Join(home, "repo")
	if len(loaded.Paths) != 1 || loaded.Paths[0] != want {
		t.Fatalf("expected %q, got %v", want, loaded.Paths)
	}
}

func TestConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	got := ConfigPath()
	want := filepath.Join(home, ".config", "rtui", "config.toml")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestSaveWritesMultilinePaths(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg := DefaultConfig()
	cfg.Paths = []string{
		filepath.Join(home, "repo1"),
		filepath.Join(home, "repo2"),
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	content, err := os.ReadFile(ConfigPath())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	got := string(content)
	if !strings.Contains(got, "paths = [\n") {
		t.Fatalf("expected multiline paths, got:\n%s", got)
	}
	if !strings.Contains(got, "repo1") || !strings.Contains(got, "repo2") {
		t.Fatalf("expected repo paths in config, got:\n%s", got)
	}
}
