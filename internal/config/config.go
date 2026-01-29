package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Paths           []string `toml:"paths"`
	Editor          string   `toml:"editor"`
	RefreshInterval int      `toml:"refresh_interval"`
	ShowClean       bool     `toml:"show_clean"`
	ScanDepth       int      `toml:"scan_depth"`
}

func DefaultConfig() Config {
	return Config{
		Paths:           []string{},
		Editor:          getDefaultEditor(),
		RefreshInterval: 30,
		ShowClean:       true,
		ScanDepth:       1,
	}
}

func getDefaultEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	return "code"
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "rtui", "config.toml")
}

// Load reads config from file, returns defaults if not found.
func Load() (Config, error) {
	cfg := DefaultConfig()

	path := configPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}

	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return cfg, err
	}

	for i, p := range cfg.Paths {
		cfg.Paths[i] = NormalizePath(p)
	}

	return cfg, nil
}

// Save writes config to disk.
func Save(cfg Config) error {
	path := configPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(cfg)
}

// AppendPath normalizes and appends a path, then saves config.
// Requires existing path and ignores duplicates.
func AppendPath(cfg *Config, path string) error {
	p := NormalizePath(path)
	if p == "" {
		return fmt.Errorf("empty path")
	}
	if _, err := os.Stat(p); err != nil {
		return err
	}
	for _, existing := range cfg.Paths {
		if NormalizePath(existing) == p {
			return nil
		}
	}
	cfg.Paths = append(cfg.Paths, p)
	return Save(*cfg)
}

// NormalizePath expands ~ and cleans path.
func NormalizePath(path string) string {
	p := strings.TrimSpace(path)
	if p == "" {
		return ""
	}
	if p[0] == '~' {
		home, _ := os.UserHomeDir()
		p = filepath.Join(home, p[1:])
	}
	return filepath.Clean(p)
}
