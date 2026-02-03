package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"strconv"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Paths           []string `toml:"paths"`
	Editor          string   `toml:"editor"`
	EditorArgs      []string `toml:"editor_args"`
	RefreshInterval int      `toml:"refresh_interval"`
	ShowClean       bool     `toml:"show_clean"`
	ScanDepth       int      `toml:"scan_depth"`
}

func DefaultConfig() Config {
	return Config{
		Paths:           []string{},
		Editor:          getDefaultEditor(),
		EditorArgs:      []string{"--profile", "Minimalist"},
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

// ConfigPath returns the resolved config file path.
func ConfigPath() string {
	return configPath()
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
	content := formatConfig(cfg)
	return os.WriteFile(path, []byte(content), 0o644)
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

func formatConfig(cfg Config) string {
	var b strings.Builder
	if len(cfg.Paths) == 0 {
		b.WriteString("paths = []\n\n")
	} else {
		b.WriteString("paths = [\n")
		for _, p := range cfg.Paths {
			b.WriteString("  ")
			b.WriteString(strconv.Quote(p))
			b.WriteString(",\n")
		}
		b.WriteString("]\n\n")
	}

	b.WriteString("editor = ")
	b.WriteString(strconv.Quote(cfg.Editor))
	b.WriteString("\n")
	b.WriteString("editor_args = ")
	if len(cfg.EditorArgs) == 0 {
		b.WriteString("[]\n")
	} else {
		b.WriteString("[")
		for i, arg := range cfg.EditorArgs {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(strconv.Quote(arg))
		}
		b.WriteString("]\n")
	}
	b.WriteString("refresh_interval = ")
	b.WriteString(strconv.Itoa(cfg.RefreshInterval))
	b.WriteString("\n")
	b.WriteString("show_clean = ")
	b.WriteString(strconv.FormatBool(cfg.ShowClean))
	b.WriteString("\n")
	b.WriteString("scan_depth = ")
	b.WriteString(strconv.Itoa(cfg.ScanDepth))
	b.WriteString("\n")
	return b.String()
}
