# RepoTUI - Implementation Guide

> A minimal TUI dashboard to monitor and manage multiple git repos from a parent folder.

---

## Table of Contents

1. [Overview](#1-overview)
2. [Prerequisites](#2-prerequisites)
3. [Quick Start](#3-quick-start)
4. [Architecture](#4-architecture)
5. [Data Structures](#5-data-structures)
6. [Implementation Guide](#6-implementation-guide)
7. [UI Specification](#7-ui-specification)
8. [Git Commands Reference](#8-git-commands-reference)
9. [Keybindings](#9-keybindings)
10. [Config File](#10-config-file)
11. [Color Scheme](#11-color-scheme)
12. [Error Handling](#12-error-handling)
13. [Testing](#13-testing)
14. [Resources & Reference Links](#14-resources--reference-links)

---

## 1. Overview

### Problem
Developers managing multiple git repos (microservices, related projects) waste time running `git status` in each folder repeatedly.

### Solution
A single TUI dashboard that:
- Shows all repos at a glance (name, branch, dirty status)
- Shows ahead/behind remote counts
- Allows quick commit+push without leaving the dashboard
- Opens repos in your editor with one keypress

### Target User
- Developers with 5-30+ git repos in organized folders
- Comfortable with terminal/TUI applications
- Uses vim-style keybindings (j/k navigation)

---

## 2. Prerequisites

### Required Software

| Software | Version | Installation | Documentation |
|----------|---------|--------------|---------------|
| [Go](https://go.dev/) | 1.21+ | [Install Guide](https://go.dev/doc/install) | [Go Docs](https://go.dev/doc/) |
| [Git](https://git-scm.com/) | 2.0+ | Usually pre-installed | [Git Book](https://git-scm.com/book/en/v2) |

```bash
# Check Go installation (requires Go 1.21+)
go version

# If not installed (macOS)
brew install go

# If not installed (Linux)
sudo apt install golang-go  # Debian/Ubuntu
sudo dnf install golang     # Fedora

# Windows - Download from https://go.dev/dl/
```

### Required Knowledge

| Topic | What You Need | Learn Here |
|-------|---------------|------------|
| Go Basics | Structs, interfaces, error handling | [Go Tour](https://go.dev/tour/) |
| Go Modules | `go mod`, dependency management | [Go Modules](https://go.dev/blog/using-go-modules) |
| Git | status, commit, push, fetch, remote | [Git Basics](https://git-scm.com/book/en/v2/Git-Basics-Getting-a-Git-Repository) |
| Terminal | Navigation, environment variables | [Command Line Crash Course](https://developer.mozilla.org/en-US/docs/Learn/Tools_and_testing/Understanding_client-side_tools/Command_line) |

### Recommended Reading (15-30 min each)

**Essential (Read Before Starting):**
1. [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics) - The Elm Architecture for TUIs
2. [Bubble Tea README](https://github.com/charmbracelet/bubbletea#bubble-tea) - Overview and examples
3. [Lip Gloss README](https://github.com/charmbracelet/lipgloss#lip-gloss) - Styling terminal output

**Helpful (Reference as Needed):**
4. [Git Status Porcelain](https://git-scm.com/docs/git-status#_short_format) - Parseable status output
5. [TOML Spec](https://toml.io/en/) - Config file format
6. [ANSI Escape Codes](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors) - Terminal colors

---

## 3. Quick Start

```bash
# 1. Create project
mkdir -p repotui/cmd/repotui repotui/internal/{config,git,ui}
cd repotui

# 2. Initialize Go module
go mod init repotui

# 3. Add dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/BurntSushi/toml

# 4. Create files (see Section 6 for code)
# ... implement each file ...

# 5. Build and run
go build ./cmd/repotui
./repotui

# Or run directly during development
go run ./cmd/repotui
```

---

## 4. Architecture

### Project Structure

> 📁 Following [Go project layout conventions](https://go.dev/doc/modules/layout#package-or-command-with-supporting-packages)

```
repotui/
├── cmd/repotui/
│   └── main.go              # Entry point, starts Bubble Tea
├── internal/
│   ├── config/
│   │   └── config.go        # Load ~/.config/repotui/config.toml (TOML)
│   ├── git/
│   │   └── git.go           # All git operations (git CLI)
│   └── ui/
│       ├── model.go         # App state (tea.Model interface)
│       ├── update.go        # Handle events (Update func)
│       ├── view.go          # Render UI (View func + Lip Gloss)
│       └── styles.go        # Colors and formatting (Lip Gloss)
├── go.mod
└── go.sum
```

### How [Bubble Tea](https://github.com/charmbracelet/bubbletea) Works (The [Elm Architecture](https://guide.elm-lang.org/architecture/))

[Bubble Tea](https://github.com/charmbracelet/bubbletea) uses a simple loop:
```
┌─────────────────────────────────────────────────┐
│                                                 │
│   Model ──────► View() ──────► Terminal        │
│     │                              │            │
│     │                              │            │
│     └──── Update(msg) ◄───── User Input        │
│                                                 │
└─────────────────────────────────────────────────┘
```

1. **Model** - Your app's state (cursor position, active repo, etc.)
2. **View(model)** - Returns a string to render based on current state
3. **Update(model, msg)** - Handles events (keypresses), returns new model

```go
// Minimal Bubble Tea program structure
type model struct {
    cursor int
    items  []string
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q":
            return m, tea.Quit
        case "j":
            m.cursor++
        }
    }
    return m, nil
}

func (m model) View() string {
    return "Hello World"
}
```

### Data Flow
```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│  config.Load() ──► git.ScanRepos() ──► ui.Model ──► View()  │
│                                                              │
│  User presses 'p' ──► Update() ──► confirm pull (if behind) │
│                                └─► confirm stage all        │
│                                └─► git.CommitAndPush()      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 5. Data Structures

> 📚 See [Go struct documentation](https://go.dev/tour/moretypes/2) and [git status --porcelain](https://git-scm.com/docs/git-status#_short_format)

### Core Types (define in `internal/git/git.go`)

```go
// FileStatus represents a changed file
type FileStatus int

const (
    StatusStaged FileStatus = iota
    StatusModified
    StatusUntracked
    StatusConflict
)

// ChangedFile represents a file with changes
type ChangedFile struct {
    Path   string
    Status FileStatus
}

// Repo represents a git repository
type Repo struct {
    Name         string        // Folder name (e.g., "miwiz-api")
    Path         string        // Full path (e.g., "/Users/dev/miwiz-api")
    Branch       string        // Current branch (e.g., "main")
    Staged       int           // Count of staged files
    Modified     int           // Count of modified files
    Untracked    int           // Count of untracked files
    Ahead        int           // Commits ahead of remote
    Behind       int           // Commits behind remote
    HasConflict  bool          // Has merge conflicts
    ChangedFiles []ChangedFile // List of changed files
}

// IsDirty returns true if repo has any changes
func (r Repo) IsDirty() bool {
    return r.Staged > 0 || r.Modified > 0 || r.Untracked > 0
}

// IsClean returns true if repo has no changes
func (r Repo) IsClean() bool {
    return !r.IsDirty()
}
```

### Config Type (define in `internal/config/config.go`)

```go
// Config holds application settings
type Config struct {
    Paths           []string `toml:"paths"`            // Folders to scan
    Editor          string   `toml:"editor"`           // Editor command
    RefreshInterval int      `toml:"refresh_interval"` // Seconds
    ShowClean       bool     `toml:"show_clean"`       // Show clean repos
    ScanDepth       int      `toml:"scan_depth"`       // How deep to scan
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
    return Config{
        Paths:           []string{},
        Editor:          "code", // VS Code
        RefreshInterval: 30,
        ShowClean:       true,
        ScanDepth:       1,
    }
}
```

### UI Model (define in `internal/ui/model.go`)

```go
// ViewMode represents the current UI mode
type ViewMode int

const (
    ModeNormal ViewMode = iota
    ModeAddPath
    ModeCommitInput
    ModeConfirmStage
    ModeConfirmPull
    ModeHelp
)

// Model holds all application state
type Model struct {
    // Data
    repos    []git.Repo
    config   config.Config

    // UI State
    cursor       int      // Highlighted repo index (active repo)
    mode         ViewMode // Current mode
    addPathInput string   // Path input for add-path modal
    commitMsg    string   // Commit message being typed
    filterDirty  bool     // Only show dirty repos

    // Terminal (for responsive layout)
    width  int  // Current terminal width
    height int  // Current terminal height

    // Status
    statusMsg string // Bottom status message
    loading   bool   // Show loading indicator
    err       error  // Last error
}
```

### Layout Type (define in `internal/ui/view.go`)

```go
// Layout holds calculated responsive column widths
// Recalculated on every render based on terminal width
type Layout struct {
    Name   int  // Width for repo name column
    Branch int  // Width for branch column (0 = hidden in compact mode)
    Status int  // Width for status column (M/S/U counts)
    Sync   int  // Width for sync column (↑/↓)
}
```

---

## 6. Implementation Guide

> 💡 **Implementation Order:** Create files in this order: `main.go` → `config.go` → `git.go` → `styles.go` → `model.go` → `update.go` → `view.go`
>
> 📖 **Reference while coding:**
> - [Bubble Tea API](https://pkg.go.dev/github.com/charmbracelet/bubbletea)
> - [Lip Gloss API](https://pkg.go.dev/github.com/charmbracelet/lipgloss)
> - [TOML library](https://pkg.go.dev/github.com/BurntSushi/toml)

### File 1: `cmd/repotui/main.go`

> 📖 Reference: [tea.NewProgram](https://pkg.go.dev/github.com/charmbracelet/bubbletea#NewProgram), [Program options](https://pkg.go.dev/github.com/charmbracelet/bubbletea#ProgramOption)

```go
package main

import (
    "fmt"
    "os"

    tea "github.com/charmbracelet/bubbletea"
    "repotui/internal/config"
    "repotui/internal/ui"
)

func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Config error: %v\n", err)
        os.Exit(1)
    }

    // Create and run program
    p := tea.NewProgram(
        ui.NewModel(cfg),
        tea.WithAltScreen(),       // Use alternate screen buffer
        tea.WithMouseCellMotion(), // Optional: mouse support
    )

    if _, err := p.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

### File 2: `internal/config/config.go`

> 📖 Reference: [BurntSushi/toml](https://pkg.go.dev/github.com/BurntSushi/toml), [os.UserHomeDir](https://pkg.go.dev/os#UserHomeDir), [TOML spec](https://toml.io/)

```go
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

// configPath returns ~/.config/repotui/config.toml
func configPath() string {
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "repotui", "config.toml")
}

// Load reads config from file, returns defaults if not found
func Load() (Config, error) {
    cfg := DefaultConfig()

    path := configPath()
    if _, err := os.Stat(path); os.IsNotExist(err) {
        // Config doesn't exist, use defaults
        return cfg, nil
    }

    _, err := toml.DecodeFile(path, &cfg)
    if err != nil {
        return cfg, err
    }

    // Expand ~ in paths
    for i, p := range cfg.Paths {
        cfg.Paths[i] = NormalizePath(p)
    }

    return cfg, nil
}

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

// AppendPath normalizes and appends a path, then saves config (requires existing path, ignores duplicates)
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
            return nil // already present
        }
    }
    cfg.Paths = append(cfg.Paths, p)
    return Save(*cfg)
}

// NormalizePath expands ~ and cleans path
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
```

### File 3: `internal/git/git.go`

> 📖 Reference: [git-status --porcelain](https://git-scm.com/docs/git-status#_short_format), [git-rev-parse](https://git-scm.com/docs/git-rev-parse), [os/exec](https://pkg.go.dev/os/exec), [filepath.WalkDir](https://pkg.go.dev/path/filepath#WalkDir)

```go
package git

import (
    "bufio"
    "bytes"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "strings"
)

type FileStatus int

const (
    StatusStaged FileStatus = iota
    StatusModified
    StatusUntracked
    StatusConflict
)

type ChangedFile struct {
    Path   string
    Status FileStatus
}

type Repo struct {
    Name         string
    Path         string
    Branch       string
    Staged       int
    Modified     int
    Untracked    int
    Ahead        int
    Behind       int
    HasConflict  bool
    ChangedFiles []ChangedFile
}

func (r Repo) IsDirty() bool {
    return r.Staged > 0 || r.Modified > 0 || r.Untracked > 0
}

// ScanRepos finds all git repos in given paths up to depth
func ScanRepos(paths []string, depth int) []Repo {
    var repos []Repo

    for _, basePath := range paths {
        // Walk directory
        filepath.WalkDir(basePath, func(path string, d os.DirEntry, err error) error {
            if err != nil {
                return nil // Skip errors
            }

            // Check depth
            relPath, _ := filepath.Rel(basePath, path)
            currentDepth := strings.Count(relPath, string(os.PathSeparator))
            if currentDepth > depth {
                return filepath.SkipDir
            }

            // Check if it's a git repo
            if d.IsDir() && isGitRepo(path) {
                repo := GetRepoStatus(path)
                repos = append(repos, repo)
                return filepath.SkipDir // Don't descend into repo
            }

            return nil
        })
    }

    return repos
}

func isGitRepo(path string) bool {
    gitPath := filepath.Join(path, ".git")
    _, err := os.Stat(gitPath)
    return err == nil // allow dir or file (.git file for worktrees)
}

// GetRepoStatus gets full status for a repo
func GetRepoStatus(path string) Repo {
    repo := Repo{
        Name: filepath.Base(path),
        Path: path,
    }

    // Get branch name
    repo.Branch = getBranch(path)

    // Get file statuses using git CLI
    if out, err := gitOutput(path, "status", "--porcelain=v1"); err == nil {
        parsePorcelain(&repo, out)
    }

    // Get ahead/behind (requires git CLI for simplicity)
    repo.Ahead, repo.Behind = getAheadBehind(path)

    return repo
}

func getBranch(path string) string {
    out, err := gitOutput(path, "rev-parse", "--abbrev-ref", "HEAD")
    if err != nil {
        return "unknown"
    }
    branch := strings.TrimSpace(out)
    if branch == "HEAD" {
        sha, _ := gitOutput(path, "rev-parse", "--short", "HEAD")
        sha = strings.TrimSpace(sha)
        if sha == "" {
            return "detached"
        }
        return "detached@" + sha
    }
    return branch
}

func parsePorcelain(repo *Repo, out string) {
    scanner := bufio.NewScanner(strings.NewReader(out))
    for scanner.Scan() {
        line := scanner.Text()
        if len(line) < 3 {
            continue
        }
        code := line[:2]
        path := strings.TrimSpace(line[2:])
        cf := ChangedFile{Path: path}

        if code == "??" {
            repo.Untracked++
            cf.Status = StatusUntracked
            repo.ChangedFiles = append(repo.ChangedFiles, cf)
            continue
        }

        if isConflict(code) {
            repo.HasConflict = true
            cf.Status = StatusConflict
            repo.ChangedFiles = append(repo.ChangedFiles, cf)
            continue
        }

        if code[0] != ' ' {
            repo.Staged++
            cf.Status = StatusStaged
            repo.ChangedFiles = append(repo.ChangedFiles, cf)
        }
        if code[1] != ' ' {
            repo.Modified++
            cf.Status = StatusModified
            repo.ChangedFiles = append(repo.ChangedFiles, cf)
        }
    }
}

func isConflict(code string) bool {
    switch code {
    case "UU", "AA", "DD", "AU", "UA", "DU", "UD":
        return true
    default:
        return false
    }
}

func gitOutput(path string, args ...string) (string, error) {
    cmd := exec.Command("git", args...)
    cmd.Dir = path
    var out bytes.Buffer
    cmd.Stdout = &out
    err := cmd.Run()
    return out.String(), err
}

// getAheadBehind uses git CLI for reliable remote comparison
func getAheadBehind(path string) (ahead, behind int) {
    cmd := exec.Command("git", "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
    cmd.Dir = path

    out, err := cmd.Output()
    if err != nil {
        return 0, 0 // No upstream or error
    }

    parts := strings.Fields(string(out))
    if len(parts) == 2 {
        ahead, _ = strconv.Atoi(parts[0])
        behind, _ = strconv.Atoi(parts[1])
    }

    return
}

// CommitAndPush stages all, commits, and pushes (call after user confirms stage-all)
func CommitAndPush(path, message string) error {
    // git add -A
    addCmd := exec.Command("git", "add", "-A")
    addCmd.Dir = path
    if err := addCmd.Run(); err != nil {
        return err
    }

    // git commit -m "message"
    commitCmd := exec.Command("git", "commit", "-m", message)
    commitCmd.Dir = path
    if err := commitCmd.Run(); err != nil {
        return err
    }

    // git push
    pushCmd := exec.Command("git", "push")
    pushCmd.Dir = path
    return pushCmd.Run()
}

// Pull runs git pull
func Pull(path string) error {
    cmd := exec.Command("git", "pull")
    cmd.Dir = path
    return cmd.Run()
}

// FetchAll runs git fetch --all
func FetchAll(path string) error {
    cmd := exec.Command("git", "fetch", "--all")
    cmd.Dir = path
    return cmd.Run()
}

// OpenInEditor opens the repo in the configured editor
func OpenInEditor(path, editor string) error {
    cmd := exec.Command(editor, path)
    return cmd.Start() // Don't wait for editor to close
}
```

### File 4: `internal/ui/styles.go`

> 📖 Reference: [Lip Gloss NewStyle](https://pkg.go.dev/github.com/charmbracelet/lipgloss#NewStyle), [Colors](https://github.com/charmbracelet/lipgloss#colors), [ANSI color codes](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors)

```go
package ui

import "github.com/charmbracelet/lipgloss"

// Colors
var (
    colorCyan    = lipgloss.Color("6")
    colorMagenta = lipgloss.Color("5")
    colorGreen   = lipgloss.Color("2")
    colorYellow  = lipgloss.Color("3")
    colorRed     = lipgloss.Color("1")
    colorGray    = lipgloss.Color("8")
    colorWhite   = lipgloss.Color("15")
)

// Styles
var (
    // Header
    titleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(colorCyan)

    // Repo list (cursor highlight)
    cursorStyle = lipgloss.NewStyle().
        Bold(true).
        Reverse(true)

    cleanRepoStyle = lipgloss.NewStyle().
        Foreground(colorGray)

    dirtyRepoStyle = lipgloss.NewStyle().
        Foreground(colorWhite)

    // Status indicators
    stagedStyle = lipgloss.NewStyle().
        Foreground(colorGreen)

    modifiedStyle = lipgloss.NewStyle().
        Foreground(colorYellow)

    untrackedStyle = lipgloss.NewStyle().
        Foreground(colorGray)

    conflictStyle = lipgloss.NewStyle().
        Foreground(colorRed).
        Bold(true)

    // Sync indicators
    aheadStyle = lipgloss.NewStyle().
        Foreground(colorCyan)

    behindStyle = lipgloss.NewStyle().
        Foreground(colorMagenta)

    // Sections
    sectionTitleStyle = lipgloss.NewStyle().
        Bold(true).
        Foreground(colorWhite)

    // Footer
    footerStyle = lipgloss.NewStyle().
        Foreground(colorGray)

    // Input
    inputStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1)

    // Status message
    successStyle = lipgloss.NewStyle().
        Foreground(colorGreen)

    errorStyle = lipgloss.NewStyle().
        Foreground(colorRed)
)

// Box drawing
var (
    boxStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        Padding(0, 1)
)
```

### File 5: `internal/ui/model.go`

> 📖 Reference: [tea.Model interface](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Model), [tea.Cmd](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Cmd), [tea.Tick](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Tick)

```go
package ui

import (
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "repotui/internal/config"
    "repotui/internal/git"
)

type ViewMode int

const (
    ModeNormal ViewMode = iota
    ModeAddPath
    ModeCommitInput
    ModeConfirmStage
    ModeConfirmPull
    ModeHelp
)

type Model struct {
    repos       []git.Repo
    config      config.Config
    cursor      int
    mode        ViewMode
    addPathInput string
    commitMsg   string
    filterDirty bool
    width       int
    height      int
    statusMsg   string
    loading     bool
    err         error
}

// Messages
type reposLoadedMsg []git.Repo
type tickMsg time.Time
type statusMsg string
type errMsg error
type pullDoneMsg string

func NewModel(cfg config.Config) Model {
    return Model{
        config: cfg,
        cursor: 0,
    }
}

func (m Model) Init() tea.Cmd {
    return tea.Batch(
        m.loadRepos(),
        m.tickCmd(),
    )
}

func (m Model) loadRepos() tea.Cmd {
    return func() tea.Msg {
        repos := git.ScanRepos(m.config.Paths, m.config.ScanDepth)
        return reposLoadedMsg(repos)
    }
}

func (m Model) tickCmd() tea.Cmd {
    if m.config.RefreshInterval <= 0 {
        return nil
    }
    return tea.Tick(
        time.Duration(m.config.RefreshInterval)*time.Second,
        func(t time.Time) tea.Msg { return tickMsg(t) },
    )
}

// visibleRepos returns repos after filtering
func (m Model) visibleRepos() []git.Repo {
    if !m.filterDirty {
        return m.repos
    }

    var result []git.Repo
    for _, r := range m.repos {
        // Filter dirty
        if m.filterDirty && r.IsClean() {
            continue
        }
        result = append(result, r)
    }
    return result
}

// currentRepo returns the repo under the cursor or nil
func (m Model) currentRepo() *git.Repo {
    repos := m.visibleRepos()
    if len(repos) == 0 || m.cursor < 0 || m.cursor >= len(repos) {
        return nil
    }
    return &repos[m.cursor]
}
```

### File 6: `internal/ui/update.go`

> 📖 Reference: [tea.KeyMsg](https://pkg.go.dev/github.com/charmbracelet/bubbletea#KeyMsg), [tea.WindowSizeMsg](https://pkg.go.dev/github.com/charmbracelet/bubbletea#WindowSizeMsg), [Key handling examples](https://github.com/charmbracelet/bubbletea/tree/master/examples/simple)

```go
package ui

import (
    tea "github.com/charmbracelet/bubbletea"
    "repotui/internal/config"
    "repotui/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {

    // Window resize
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        return m, nil

    // Repos loaded
    case reposLoadedMsg:
        m.repos = msg
        m.loading = false
        return m, nil

    // Auto-refresh tick
    case tickMsg:
        return m, tea.Batch(m.loadRepos(), m.tickCmd())

    // Status message
    case statusMsg:
        m.statusMsg = string(msg)
        return m, nil

    // Error
    case errMsg:
        m.err = msg
        m.statusMsg = "Error: " + msg.Error()
        return m, nil
    
    // Pull done
    case pullDoneMsg:
        m.statusMsg = "Pulled " + string(msg)
        m.mode = ModeConfirmStage
        return m, nil

    // Keyboard
    case tea.KeyMsg:
        return m.handleKey(msg)
    }

    return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Handle different modes
    switch m.mode {
    case ModeAddPath:
        return m.handleAddPath(msg)
    case ModeCommitInput:
        return m.handleCommitInput(msg)
    case ModeConfirmStage:
        return m.handleConfirmStage(msg)
    case ModeHelp:
        return m.handleHelp(msg)
    case ModeConfirmPull:
        return m.handleConfirmPull(msg)
    }

    // Normal mode
    switch msg.String() {

    // Quit
    case "q", "ctrl+c":
        return m, tea.Quit

    // Navigation
    case "j", "down":
        visible := m.visibleRepos()
        if m.cursor < len(visible)-1 {
            m.cursor++
        }
    case "k", "up":
        if m.cursor > 0 {
            m.cursor--
        }

    // Add path
    case "a":
        m.mode = ModeAddPath
        m.addPathInput = ""

    // Open in editor
    case "c":
        if repo := m.currentRepo(); repo != nil {
            git.OpenInEditor(repo.Path, m.config.Editor)
            m.statusMsg = "Opened " + repo.Name + " in " + m.config.Editor
        }

    // Commit + Push
    case "p":
        if repo := m.currentRepo(); repo != nil {
            if repo.HasConflict {
                m.statusMsg = "Cannot push: repo has conflicts"
                return m, nil
            }
            if !repo.IsDirty() {
                m.statusMsg = "Nothing to commit"
                return m, nil
            }
            if repo.Behind > 0 {
                m.mode = ModeConfirmPull
                return m, nil
            }
            m.mode = ModeConfirmStage
        }

    // Refresh
    case "r":
        m.loading = true
        m.statusMsg = "Refreshing..."
        return m, m.loadRepos()

    // Fetch all
    case "f":
        if repo := m.currentRepo(); repo != nil {
            m.statusMsg = "Fetching..."
            return m, func() tea.Msg {
                if err := git.FetchAll(repo.Path); err != nil {
                    return errMsg(err)
                }
                return statusMsg("Fetched " + repo.Name)
            }
        }

    // Toggle dirty filter
    case "d":
        m.filterDirty = !m.filterDirty
        m.cursor = 0

    // Help
    case "?":
        m.mode = ModeHelp
    }

    return m, nil
}

func (m Model) handleCommitInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        m.mode = ModeNormal
        m.commitMsg = ""
    case "enter":
        if m.commitMsg == "" {
            return m, nil
        }
        repo := m.currentRepo()
        if repo == nil {
            m.mode = ModeNormal
            return m, nil
        }

        m.statusMsg = "Pushing..."
        m.mode = ModeNormal
        commitMsg := m.commitMsg
        m.commitMsg = ""

        return m, func() tea.Msg {
            if err := git.CommitAndPush(repo.Path, commitMsg); err != nil {
                return errMsg(err)
            }
            return statusMsg("Pushed to " + repo.Name)
        }
    case "backspace":
        if len(m.commitMsg) > 0 {
            m.commitMsg = m.commitMsg[:len(m.commitMsg)-1]
        }
    default:
        // Add character to message
        if len(msg.String()) == 1 {
            m.commitMsg += msg.String()
        }
    }
    return m, nil
}

func (m Model) handleHelp(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    // Any key exits help
    m.mode = ModeNormal
    return m, nil
}

func (m Model) handleAddPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        m.mode = ModeNormal
        m.addPathInput = ""
    case "enter":
        if err := config.AppendPath(&m.config, m.addPathInput); err != nil {
            return m, errMsg(err)
        }
        m.statusMsg = "Path added"
        m.mode = ModeNormal
        m.addPathInput = ""
        return m, m.loadRepos()
    case "backspace":
        if len(m.addPathInput) > 0 {
            m.addPathInput = m.addPathInput[:len(m.addPathInput)-1]
        }
    default:
        if len(msg.String()) == 1 {
            m.addPathInput += msg.String()
        }
    }
    return m, nil
}

func (m Model) handleConfirmStage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y":
        m.mode = ModeCommitInput
        m.commitMsg = ""
    case "n", "c", "esc":
        m.mode = ModeNormal
        m.statusMsg = "Commit canceled"
    }
    return m, nil
}

func (m Model) handleConfirmPull(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "y":
        repo := m.currentRepo()
        if repo != nil {
            m.statusMsg = "Pulling..."
            m.mode = ModeNormal
            return m, func() tea.Msg {
                if err := git.Pull(repo.Path); err != nil {
                    return errMsg(err)
                }
                return pullDoneMsg(repo.Name)
            }
        }
    case "n":
        m.mode = ModeConfirmStage
    case "c", "esc":
        m.mode = ModeNormal
    }
    return m, nil
}
```

### File 7: `internal/ui/view.go`

> 📖 Reference: [Lip Gloss Width](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Width), [Style.Render](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Style.Render), [strings.Builder](https://pkg.go.dev/strings#Builder), [fmt.Sprintf](https://pkg.go.dev/fmt#Sprintf)

```go
package ui

import (
    "fmt"
    "strings"

    "github.com/charmbracelet/lipgloss"
    "repotui/internal/git"
)

// Layout holds calculated column widths for responsive design
type Layout struct {
    Name   int
    Branch int
    Status int
    Sync   int
}

// calculateLayout computes responsive column widths based on terminal width
func (m Model) calculateLayout() Layout {
    w := m.width

    // Fixed columns
    cursorW := 2
    syncW := 8
    statusW := 12
    gaps := 4 // spaces between columns

    // Remaining space for name and branch
    remaining := w - cursorW - syncW - statusW - gaps

    // Compact mode for very small terminals
    if w < 40 {
        return Layout{
            Name:   max(remaining-2, 8),
            Branch: 0, // Hide branch in compact mode
            Status: 8,
            Sync:   6,
        }
    }

    // Split remaining between name and branch (60/40)
    nameW := int(float64(remaining) * 0.55)
    branchW := remaining - nameW

    // Apply minimums
    nameW = max(nameW, 12)
    branchW = max(branchW, 10)

    // For wide terminals, cap maximums
    if w > 120 {
        nameW = min(nameW, 30)
        branchW = min(branchW, 25)
    }

    return Layout{
        Name:   nameW,
        Branch: branchW,
        Status: statusW,
        Sync:   syncW,
    }
}

func (m Model) View() string {
    if m.width == 0 {
        return "Loading..."
    }

    var b strings.Builder

    // Header
    b.WriteString(m.renderHeader())
    b.WriteString("\n")

    // Main content based on mode
    switch m.mode {
    case ModeHelp:
        b.WriteString(m.renderHelp())
    case ModeAddPath:
        b.WriteString(m.renderRepoList())
        b.WriteString("\n")
        b.WriteString(m.renderAddPath())
    case ModeCommitInput:
        b.WriteString(m.renderRepoList())
        b.WriteString("\n")
        b.WriteString(m.renderCommitInput())
    case ModeConfirmStage:
        b.WriteString(m.renderRepoList())
        b.WriteString("\n")
        b.WriteString(m.renderStageConfirm())
    case ModeConfirmPull:
        b.WriteString(m.renderRepoList())
        b.WriteString("\n")
        b.WriteString(m.renderPullConfirm())
    default:
        b.WriteString(m.renderRepoList())
        if len(m.visibleRepos()) > 0 && m.height >= 15 {
            b.WriteString("\n")
            b.WriteString(m.renderChangesPanel())
        }
    }

    // Footer
    b.WriteString("\n")
    b.WriteString(m.renderFooter())

    return b.String()
}

func (m Model) renderHeader() string {
    title := titleStyle.Render("● RepoTUI")
    hints := footerStyle.Render("[r]fresh  [q]uit")

    // Right-align hints dynamically
    titleW := lipgloss.Width(title)
    hintsW := lipgloss.Width(hints)
    gap := m.width - titleW - hintsW

    if gap < 1 {
        // Terminal too narrow, stack vertically or truncate
        if m.width < 30 {
            return title
        }
        gap = 1
    }

    return title + strings.Repeat(" ", gap) + hints
}

func (m Model) renderRepoList() string {
    var b strings.Builder

    // Section header
    header := "REPOSITORIES"
    if m.filterDirty {
        header += " (dirty only)"
    }
    b.WriteString(sectionTitleStyle.Render(header))
    b.WriteString("\n")
    b.WriteString(strings.Repeat("─", m.width))
    b.WriteString("\n")

    repos := m.visibleRepos()

    if len(repos) == 0 {
        b.WriteString(footerStyle.Render("  No repositories found"))
        return b.String()
    }

    // Calculate layout once for all rows
    layout := m.calculateLayout()

    for i, repo := range repos {
        line := m.renderRepoLine(repo, i == m.cursor, layout)
        b.WriteString(line)
        b.WriteString("\n")
    }

    return b.String()
}

func (m Model) renderRepoLine(repo git.Repo, isCursor bool, layout Layout) string {
    // Cursor indicator (fixed 2 chars)
    cursor := "  "
    if isCursor {
        cursor = "▶ "
    }

    // Repo name (dynamic width)
    name := fmt.Sprintf("%-*s", layout.Name, truncate(repo.Name, layout.Name))

    // Branch (dynamic width, hidden in compact mode)
    var branch string
    if layout.Branch > 0 {
        branch = fmt.Sprintf("%-*s", layout.Branch, truncate(repo.Branch, layout.Branch))
    }

    // Status counts (styled individually)
    var status string
    if repo.IsDirty() {
        parts := []string{}
        if repo.Modified > 0 {
            parts = append(parts, modifiedStyle.Render(fmt.Sprintf("%dM", repo.Modified)))
        }
        if repo.Staged > 0 {
            parts = append(parts, stagedStyle.Render(fmt.Sprintf("%dS", repo.Staged)))
        }
        if repo.Untracked > 0 {
            parts = append(parts, untrackedStyle.Render(fmt.Sprintf("%dU", repo.Untracked)))
        }
        status = strings.Join(parts, " ")
    } else {
        status = stagedStyle.Render("✓")
    }

    // Pad status to fixed width for alignment
    statusPadded := fmt.Sprintf("%-*s", layout.Status, status)

    // Sync status (right-aligned)
    var sync string
    if repo.HasConflict {
        sync = conflictStyle.Render("CONFLICT")
    } else {
        if repo.Ahead > 0 {
            sync += aheadStyle.Render(fmt.Sprintf("↑%d", repo.Ahead))
        }
        if repo.Behind > 0 {
            sync += behindStyle.Render(fmt.Sprintf("↓%d", repo.Behind))
        }
    }

    // Build line with dynamic spacing
    var line string
    if layout.Branch > 0 {
        line = cursor + name + " " + branch + " " + statusPadded + " " + sync
    } else {
        // Compact mode: no branch
        line = cursor + name + " " + statusPadded + " " + sync
    }

    // Apply row style
    if isCursor {
        line = cursorStyle.Render(line)
    } else if repo.IsClean() {
        line = cleanRepoStyle.Render(line)
    }

    return line
}

func (m Model) renderChangesPanel() string {
    repo := m.currentRepo()
    if repo == nil {
        return ""
    }

    var b strings.Builder

    // Header
    header := fmt.Sprintf("CHANGES: %s (%s)", repo.Name, repo.Branch)
    b.WriteString(sectionTitleStyle.Render(header))
    b.WriteString("\n")
    b.WriteString(strings.Repeat("─", m.width))
    b.WriteString("\n")

    // Group files by status
    staged := []git.ChangedFile{}
    modified := []git.ChangedFile{}
    untracked := []git.ChangedFile{}

    for _, f := range repo.ChangedFiles {
        switch f.Status {
        case git.StatusStaged:
            staged = append(staged, f)
        case git.StatusModified:
            modified = append(modified, f)
        case git.StatusUntracked:
            untracked = append(untracked, f)
        }
    }

    // Calculate max path width (leave room for indent)
    maxPathW := m.width - 4

    // Render each group
    b.WriteString(stagedStyle.Render(fmt.Sprintf("Staged (%d)", len(staged))))
    b.WriteString("\n")
    for _, f := range staged {
        b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
    }

    b.WriteString(modifiedStyle.Render(fmt.Sprintf("Modified (%d)", len(modified))))
    b.WriteString("\n")
    for _, f := range modified {
        b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
    }

    b.WriteString(untrackedStyle.Render(fmt.Sprintf("Untracked (%d)", len(untracked))))
    b.WriteString("\n")
    for _, f := range untracked {
        b.WriteString("  " + truncatePath(f.Path, maxPathW) + "\n")
    }

    return b.String()
}

func (m Model) renderCommitInput() string {
    var b strings.Builder
    b.WriteString("Commit message:\n")

    // Input box adapts to terminal width
    inputW := min(m.width-4, 60)
    input := m.commitMsg + "█"
    if len(input) > inputW {
        input = input[len(input)-inputW:]
    }

    b.WriteString(inputStyle.Width(inputW).Render(input))
    b.WriteString("\n")
    b.WriteString(footerStyle.Render("[Enter] commit  [Esc] cancel"))
    return b.String()
}

func (m Model) renderAddPath() string {
    msg := "Add repo path"
    boxW := min(m.width-4, 50)
    inputW := boxW - 4
    input := m.addPathInput + "█"
    if len(input) > inputW {
        input = input[len(input)-inputW:]
    }
    body := msg + "\n\n" + input + "\n\n[Enter]=save  [Esc]=cancel"
    return boxStyle.Width(boxW).Render(body)
}

func (m Model) renderStageConfirm() string {
    repo := m.currentRepo()
    if repo == nil {
        return ""
    }

    msg := "Stage all changes and continue?"
    boxW := min(m.width-4, 50)
    return boxStyle.Width(boxW).Render(msg + "\n\n[y]es  [n]o  [c]ancel")
}

func (m Model) renderPullConfirm() string {
    repo := m.currentRepo()
    if repo == nil {
        return ""
    }

    msg := fmt.Sprintf("Repo is %d commits behind. Pull first?", repo.Behind)
    boxW := min(m.width-4, 50)
    return boxStyle.Width(boxW).Render(msg + "\n\n[y]es  [n]o  [c]ancel")
}

func (m Model) renderHelp() string {
    help := `KEYBINDINGS

Navigation
  j/↓     Next repo
  k/↑     Previous repo

Actions
  a       Add path
  c       Open in editor
  p       Commit + Push (prompts)
  f       Fetch all
  r       Refresh

Filters
  d       Toggle dirty-only

Other
  ?       This help
  q       Quit

Press any key to close...`

    // Center help box, adapt width
    boxW := min(m.width-4, 40)
    return boxStyle.Width(boxW).Render(help)
}

func (m Model) renderFooter() string {
    // Left: action hints
    actions := "[a]dd path  [c]ode  [p]ush  [r]efresh  [?]help"

    // Compact mode: shorter hints
    if m.width < 50 {
        actions = "a:add c:code p:push r:ref ?:help"
    }

    // Right: status message
    status := m.statusMsg
    if m.loading {
        status = "Loading..."
    }

    left := footerStyle.Render(actions)
    right := footerStyle.Render(status)

    leftW := lipgloss.Width(left)
    rightW := lipgloss.Width(right)
    gap := m.width - leftW - rightW

    if gap < 1 {
        // Too narrow: show only status or truncate
        if m.width < 40 {
            return right
        }
        gap = 1
    }

    return left + strings.Repeat(" ", gap) + right
}

// truncate shortens string from the end with ellipsis
func truncate(s string, maxW int) string {
    if len(s) <= maxW {
        return s
    }
    if maxW <= 1 {
        return s[:maxW]
    }
    return s[:maxW-1] + "…"
}

// truncatePath shortens path from the start (keeps filename visible)
func truncatePath(path string, maxW int) string {
    if len(path) <= maxW {
        return path
    }
    if maxW <= 3 {
        return path[len(path)-maxW:]
    }
    return "…" + path[len(path)-maxW+1:]
}

// Helper functions
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

---

## 7. UI Specification

### Responsive Design Philosophy

The UI adapts to any terminal size. No fixed widths - everything scales dynamically.

Primary target: run in the right 1/3 of a terminal while coding on the left 2/3.
Design for 40-60 columns as the common panel width.

### Layout Structure
```
┌─────────────────────────────────────────────────────────────────────────┐
│ ● RepoTUI                                          [r]fresh  [q]uit     │
├─────────────────────────────────────────────────────────────────────────┤
│ REPOSITORIES                                                            │
│─────────────────────────────────────────────────────────────────────────│
│   miwiz-api          main           2M 1S                          ↓3   │
│ ▶ miwiz-web          feature/auth   1S                             ↑2   │
│   miwiz-cms          develop        ✓                                   │
│   istudy-api         main           3U                             ↓1   │
│   istudy-ios         main           ✓                                   │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ CHANGES: miwiz-web (feature/auth)                                       │
│─────────────────────────────────────────────────────────────────────────│
│ Staged (1)                                                              │
│   src/components/Header.tsx                                             │
│ Modified (0)                                                            │
│ Untracked (0)                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│ [a]dd path  [c]ode  [p]ush  [r]efresh  [?]help                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Add Path Modal

Press `a` to add a repo path. Show a small modal input box, then append the path to config and rescan.
Rules: path must already exist; duplicates are ignored.

```
┌─────────────────────────────────────┐
│ Add repo path                       │
│ /Users/you/SourceCode              │
│                                     │
│ [Enter]=save  [Esc]=cancel          │
└─────────────────────────────────────┘
```

### Responsive Column Layout

Columns use **fixed/flexible widths** tuned for narrow panels (40-60 cols):

| Column | Width | Behavior | Priority |
|--------|-------|----------|----------|
| Cursor | 2 chars | Fixed | Always shown |
| Name | Flexible | Gets remaining space | High |
| Branch | Flexible | Uses remainder after Name | High |
| Status | 8-12 chars | Fixed-ish (icons + counts) | Medium |
| Sync | 6-8 chars | Right-aligned, fixed-ish | High |

### Breakpoints (target: right 1/3 panel)

| Terminal Width | Behavior |
|----------------|----------|
| < 40 chars | Compact mode: hide branch, shorten status/sync |
| 40-60 chars | Narrow mode: show branch, truncate names |
| > 60 chars | Normal mode: show full columns, extra padding |

### Responsive Implementation

```go
// Calculate dynamic column widths based on terminal width
func (m Model) calculateLayout() Layout {
    w := m.width

    // Fixed columns
    cursorW := 2
    syncW := 8
    statusW := 12

    // Flexible columns share remaining space
    remaining := w - cursorW - syncW - statusW - 4 // 4 for gaps

    // Split remaining between name and branch (55/45)
    nameW := int(float64(remaining) * 0.55)
    branchW := remaining - nameW

    // Apply minimums
    if nameW < 10 {
        nameW = 10
    }
    if branchW < 8 {
        branchW = 8
    }

    // Compact mode for very small terminals
    if w < 40 {
        branchW = 0 // Hide branch
        nameW = w - cursorW - syncW - 6
    }

    return Layout{
        Cursor: cursorW,
        Name:   nameW,
        Branch: branchW,
        Status: statusW,
        Sync:   syncW,
    }
}
```

### Vertical Responsiveness

| Terminal Height | Repo List | Details Panel |
|-----------------|-----------|---------------|
| < 15 rows | Full height, no details | Hidden |
| 15-30 rows | 60% height | 40% height |
| > 30 rows | Auto (content-based) | Auto (content-based) |

```go
// Calculate panel heights
func (m Model) calculatePanelHeights() (repoHeight, detailHeight int) {
    h := m.height
    headerH := 2  // Title + separator
    footerH := 2  // Actions + separator
    available := h - headerH - footerH

    if len(m.visibleRepos()) == 0 || h < 15 {
        // No repos or tiny terminal: full height for repos
        return available, 0
    }

    // Split based on content
    repoCount := len(m.visibleRepos())
    repoNeeded := repoCount + 2 // +2 for section header

    if h > 30 {
        // Large terminal: content-based
        detailHeight = min(available/3, 15) // Max 15 lines for details
        repoHeight = available - detailHeight
    } else {
        // Medium terminal: 60/40 split
        repoHeight = int(float64(available) * 0.6)
        detailHeight = available - repoHeight
    }

    return repoHeight, detailHeight
}
```

### Truncation Rules

When content exceeds column width:

```go
func truncate(s string, maxWidth int) string {
    if len(s) <= maxWidth {
        return s
    }
    if maxWidth <= 3 {
        return s[:maxWidth]
    }
    return s[:maxWidth-1] + "…"
}

// For file paths, truncate from the left (keep filename)
func truncatePath(path string, maxWidth int) string {
    if len(path) <= maxWidth {
        return path
    }
    if maxWidth <= 3 {
        return path[len(path)-maxWidth:]
    }
    return "…" + path[len(path)-maxWidth+1:]
}
```

**Examples:**
- `"my-very-long-repo-name"` → `"my-very-long-re…"` (15 chars)
- `"src/components/Header.tsx"` → `"…ponents/Header.tsx"` (20 chars, path)

### Dynamic Separator Lines

Separators should span the full width:

```go
func (m Model) renderSeparator() string {
    return strings.Repeat("─", m.width)
}
```

### Window Resize Handling

The UI automatically re-renders on terminal resize:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
        // Recalculate layout on resize
        m.layout = m.calculateLayout()
        return m, nil
    }
    // ...
}
```

### Dynamic Behavior
- **No repos or tiny terminal**: Repo list takes full available height, details hidden
- **Repos available**: Split view (ratio depends on terminal height)
- **Tiny terminal** (< 40 wide): Compact mode, essential info only
- **Resize**: Instant re-layout, no flicker

---

## 8. Git Commands Reference

> 📖 Full documentation: [Git Reference Manual](https://git-scm.com/docs)

Commands used internally:

| Command | Documentation | Purpose |
|---------|---------------|---------|
| `git status --porcelain=v1` | [git-status](https://git-scm.com/docs/git-status) | Parse file changes |
| `git rev-parse --abbrev-ref HEAD` | [git-rev-parse](https://git-scm.com/docs/git-rev-parse) | Current branch |
| `git rev-list --left-right --count HEAD...@{upstream}` | [git-rev-list](https://git-scm.com/docs/git-rev-list) | Get ahead/behind counts |
| `git add -A` | [git-add](https://git-scm.com/docs/git-add) | Stage all changes |
| `git commit -m "msg"` | [git-commit](https://git-scm.com/docs/git-commit) | Create commit |
| `git push` | [git-push](https://git-scm.com/docs/git-push) | Push to remote |
| `git pull` | [git-pull](https://git-scm.com/docs/git-pull) | Fetch and merge |
| `git fetch --all` | [git-fetch](https://git-scm.com/docs/git-fetch) | Fetch all remotes |

Note: `git add -A` is only run after the user confirms stage-all.

```bash
# Get status (parseable)
git status --porcelain=v1
# Docs: https://git-scm.com/docs/git-status

# Get current branch
git rev-parse --abbrev-ref HEAD
# Docs: https://git-scm.com/docs/git-rev-parse

# Get ahead/behind counts
git rev-list --left-right --count HEAD...@{upstream}
# Output: "3    1" (3 ahead, 1 behind)
# Docs: https://git-scm.com/docs/git-rev-list

# Stage all changes
git add -A
# Docs: https://git-scm.com/docs/git-add

# Commit
git commit -m "message"
# Docs: https://git-scm.com/docs/git-commit

# Push
git push
# Docs: https://git-scm.com/docs/git-push

# Pull
git pull
# Docs: https://git-scm.com/docs/git-pull

# Fetch all remotes
git fetch --all
# Docs: https://git-scm.com/docs/git-fetch
```

**Understanding `@{upstream}`:** See [gitrevisions](https://git-scm.com/docs/gitrevisions#Documentation/gitrevisions.txt-emltaboranchnamegt64telemerename93telemeregt)

---

## 9. Keybindings

| Key | Action | Mode |
|-----|--------|------|
| `j` / `↓` | Next repo | Normal |
| `k` / `↑` | Previous repo | Normal |
| `a` | Add repo path | Normal |
| `c` | Open repo in editor | Normal |
| `p` | Start commit+push flow | Normal |
| `f` | Fetch all remotes | Normal |
| `r` | Refresh status | Normal |
| `d` | Toggle dirty-only filter | Normal |
| `?` | Show help | Normal |
| `q` | Quit | Normal |
| `Enter` | Confirm commit | Commit Input |
| `Esc` | Cancel | Commit Input |
| `y` | Yes (stage all) | Confirm Stage |
| `n` | No (cancel commit) | Confirm Stage |
| `c` | Cancel | Confirm Stage |
| `y` | Yes (pull) | Confirm Pull |
| `n` | No (skip pull) | Confirm Pull |
| `c` | Cancel | Confirm Pull |

---

## 10. Config File

> 📖 Reference: [TOML Specification](https://toml.io/en/), [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html), [TOML Validator](https://www.toml.io/en/validator)

**Location:** `~/.config/repotui/config.toml`

```toml
# Folders to scan for git repos
# Supports ~ expansion
# TOML array syntax: https://toml.io/en/v1.0.0#array
paths = [
    "~/SourceCode/Miwiz",
    "~/SourceCode/Personal",
]

# Editor command (default: $EDITOR or "code")
# Common editors: "code", "zed", "nvim", "hx", "subl"
editor = "zed"

# Auto-refresh interval in seconds (0 to disable)
# TOML integer: https://toml.io/en/v1.0.0#integer
refresh_interval = 30

# Show clean repos (true) or dirty only (false)
# TOML boolean: https://toml.io/en/v1.0.0#boolean
show_clean = true

# Max depth to scan for repos (default: 1)
# 1 = direct children only
# 2 = children and grandchildren
scan_depth = 1
```

Note: If the config file is missing or `paths` is empty, RepoTUI scans the current working directory (CWD) and shows a banner with the path.
Note: The UI `[a]dd path` appends a normalized path to `paths` and writes the config file. Path must exist; duplicates are ignored.

**Create config manually:**
```bash
mkdir -p ~/.config/repotui
cat > ~/.config/repotui/config.toml << 'EOF'
paths = ["~/your/repos/folder"]
editor = "code"
EOF
```

---

## 11. Color Scheme

> 📖 Reference: [ANSI escape codes](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors), [Lip Gloss Colors](https://github.com/charmbracelet/lipgloss#colors), [Terminal color chart](https://misc.flogisoft.com/bash/tip_colors_and_formatting)

| Element | ANSI Color | Hex (approx) | Usage |
|---------|------------|--------------|-------|
| Cursor row | Reverse | - | Highlighted row |
| Clean repo | 8 (gray) | `#808080` | No action needed |
| Dirty repo | 15 (white) | `#FFFFFF` | Needs attention |
| Staged | 2 (green) | `#00FF00` | Ready to commit |
| Modified | 3 (yellow) | `#FFFF00` | Changed, not staged |
| Untracked | 8 (gray) | `#808080` | New files |
| Ahead (↑) | 6 (cyan) | `#00FFFF` | Can push |
| Behind (↓) | 5 (magenta) | `#FF00FF` | Needs pull |
| Conflict | 1 (red) | `#FF0000` | Blocked |

**Lip Gloss color usage:**
```go
// ANSI 256 colors (0-255)
lipgloss.Color("6")    // Cyan

// Hex colors (for modern terminals)
lipgloss.Color("#00FFFF")

// Adaptive colors (light/dark mode)
lipgloss.AdaptiveColor{Light: "0", Dark: "15"}
```

---

## 12. Error Handling

### Common Scenarios

| Scenario | Handling |
|----------|----------|
| No config file | Use defaults, scan CWD, show banner with path |
| Paths empty | Scan CWD, show banner with path |
| Invalid config path | Show error, use defaults |
| Path doesn't exist | Skip silently |
| Not a git repo | Skip silently |
| No remote upstream | Show "–" for ahead/behind |
| Add path is empty/invalid | Show error, keep config unchanged |
| Config write fails | Show error, keep config unchanged |
| Add path already exists | Show status message, no change |
| Push fails | Show error message in footer |
| Network error | Show error, allow retry |

### Error Display
- Errors show in the footer/status area
- Use red color for errors
- Auto-clear after 5 seconds or on next action

---

## 13. Testing

See `REPOTUI_TESTING.md` for the automated test plan, guard checks, and manual/responsive checklists.

---

## 14. Resources & Reference Links

### Core Dependencies

| Library | GitHub | Documentation | Used For |
|---------|--------|---------------|----------|
| Bubble Tea | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) | TUI framework (Model-View-Update) |
| Lip Gloss | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/lipgloss) | Terminal styling & colors |
| Git CLI | [git/git](https://github.com/git/git) | [git-scm.com/docs](https://git-scm.com/docs) | Git operations via CLI |
| TOML | [BurntSushi/toml](https://github.com/BurntSushi/toml) | [pkg.go.dev](https://pkg.go.dev/github.com/BurntSushi/toml) | Config file parsing |

### Bubble Tea (TUI Framework)

| Resource | Link | Description |
|----------|------|-------------|
| GitHub Repo | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Source code, issues, discussions |
| Basics Tutorial | [tutorials/basics](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/basics) | **Start here** - Step-by-step guide |
| Commands Tutorial | [tutorials/commands](https://github.com/charmbracelet/bubbletea/tree/master/tutorials/commands) | Async operations, tea.Cmd |
| Official Examples | [examples/](https://github.com/charmbracelet/bubbletea/tree/master/examples) | 30+ example programs |
| API Reference | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) | Full API documentation |
| Blog Tutorial | [charm.sh/blog](https://charm.sh/blog/bubbletea-tutorial/) | Building your first TUI |
| The Elm Architecture | [elm-lang.org/guide](https://guide.elm-lang.org/architecture/) | Original pattern (JavaScript) |

**Key Bubble Tea Concepts:**
- [tea.Model interface](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Model) - Your app's state
- [tea.Cmd](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Cmd) - Async operations
- [tea.Msg](https://pkg.go.dev/github.com/charmbracelet/bubbletea#Msg) - Events (keyboard, window, custom)
- [tea.KeyMsg](https://pkg.go.dev/github.com/charmbracelet/bubbletea#KeyMsg) - Keyboard input handling

### Lip Gloss (Styling)

| Resource | Link | Description |
|----------|------|-------------|
| GitHub Repo | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Source code, examples |
| Style Guide | [README#styles](https://github.com/charmbracelet/lipgloss#defining-styles) | How to define styles |
| Colors | [README#colors](https://github.com/charmbracelet/lipgloss#colors) | ANSI, hex, adaptive colors |
| Layout | [README#layout](https://github.com/charmbracelet/lipgloss#layout) | Width, height, alignment |
| Borders | [README#borders](https://github.com/charmbracelet/lipgloss#borders) | Box drawing characters |
| API Reference | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/lipgloss) | Full API documentation |

**Key Lip Gloss Functions:**
- [lipgloss.NewStyle()](https://pkg.go.dev/github.com/charmbracelet/lipgloss#NewStyle) - Create a style
- [Style.Render()](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Style.Render) - Apply style to string
- [lipgloss.Width()](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Width) - Get rendered width
- [lipgloss.Color()](https://pkg.go.dev/github.com/charmbracelet/lipgloss#Color) - Define colors

### Git CLI (Porcelain)

| Resource | Link | Description |
|----------|------|-------------|
| git-status | [git-status](https://git-scm.com/docs/git-status) | Status output and porcelain format |
| git-rev-parse | [git-rev-parse](https://git-scm.com/docs/git-rev-parse) | Branch and HEAD info |
| git-rev-list | [git-rev-list](https://git-scm.com/docs/git-rev-list) | Ahead/behind counts |

**Key Git CLI Calls:**
- `git status --porcelain=v1` - Parse file changes
- `git rev-parse --abbrev-ref HEAD` - Current branch
- `git rev-list --left-right --count HEAD...@{upstream}` - Ahead/behind

### Go Language

| Resource | Link | Description |
|----------|------|-------------|
| Official Site | [go.dev](https://go.dev/) | Downloads, docs, blog |
| Go Tour | [go.dev/tour](https://go.dev/tour/) | Interactive tutorial |
| Effective Go | [go.dev/doc/effective_go](https://go.dev/doc/effective_go) | Best practices |
| Go Modules | [go.dev/blog/using-go-modules](https://go.dev/blog/using-go-modules) | Dependency management |
| Standard Library | [pkg.go.dev/std](https://pkg.go.dev/std) | Built-in packages |
| Go by Example | [gobyexample.com](https://gobyexample.com/) | Practical examples |
| Playground | [go.dev/play](https://go.dev/play/) | Try Go online |

**Relevant Go Packages:**
- [os/exec](https://pkg.go.dev/os/exec) - Run external commands (git CLI)
- [path/filepath](https://pkg.go.dev/path/filepath) - File path manipulation
- [strings](https://pkg.go.dev/strings) - String operations
- [strconv](https://pkg.go.dev/strconv) - String conversions
- [time](https://pkg.go.dev/time) - Time and timers (auto-refresh)

### Git

| Resource | Link | Description |
|----------|------|-------------|
| Pro Git Book | [git-scm.com/book](https://git-scm.com/book/en/v2) | Comprehensive guide |
| Git Reference | [git-scm.com/docs](https://git-scm.com/docs) | Command documentation |
| git-status | [git-scm.com/docs/git-status](https://git-scm.com/docs/git-status) | Status output format |
| git-rev-list | [git-scm.com/docs/git-rev-list](https://git-scm.com/docs/git-rev-list) | Ahead/behind counts |
| Porcelain vs Plumbing | [git-scm.com/book/.../plumbing](https://git-scm.com/book/en/v2/Git-Internals-Plumbing-and-Porcelain) | Git internals |

### TOML Configuration

| Resource | Link | Description |
|----------|------|-------------|
| TOML Spec | [toml.io](https://toml.io/en/) | Official specification |
| TOML Wiki | [github.com/toml-lang/toml](https://github.com/toml-lang/toml) | Examples, FAQ |
| Go Library | [github.com/BurntSushi/toml](https://github.com/BurntSushi/toml) | Parser we use |
| Validator | [toml.io/en/validator](https://www.toml.io/en/validator) | Online TOML validator |

### Terminal & ANSI

| Resource | Link | Description |
|----------|------|-------------|
| ANSI Escape Codes | [Wikipedia](https://en.wikipedia.org/wiki/ANSI_escape_code) | Color codes reference |
| Terminal Colors | [misc.flogisoft.com](https://misc.flogisoft.com/bash/tip_colors_and_formatting) | Color examples |
| XTerm Control Sequences | [invisible-island.net](https://invisible-island.net/xterm/ctlseqs/ctlseqs.html) | Complete reference |
| Box Drawing | [Wikipedia](https://en.wikipedia.org/wiki/Box-drawing_character) | Border characters |

### Example TUI Projects (Study These)

| Project | Link | Complexity | Learn From |
|---------|------|------------|------------|
| **Glow** | [charmbracelet/glow](https://github.com/charmbracelet/glow) | High | File browser, markdown rendering |
| **Soft Serve** | [charmbracelet/soft-serve](https://github.com/charmbracelet/soft-serve) | High | Git TUI, SSH server |
| **Lazygit** | [jesseduffield/lazygit](https://github.com/jesseduffield/lazygit) | High | Git operations, complex UI |
| **gh-dash** | [dlvhdr/gh-dash](https://github.com/dlvhdr/gh-dash) | Medium | GitHub dashboard, tables |
| **Slides** | [maaslalani/slides](https://github.com/maaslalani/slides) | Low | Simple Bubble Tea app |
| **Charm Examples** | [charmbracelet/bubbletea/examples](https://github.com/charmbracelet/bubbletea/tree/master/examples) | Varies | Official examples |

### Charm Ecosystem (All Compatible)

| Tool | Link | Description |
|------|------|-------------|
| Bubble Tea | [bubbletea](https://github.com/charmbracelet/bubbletea) | TUI framework |
| Lip Gloss | [lipgloss](https://github.com/charmbracelet/lipgloss) | Styling |
| Bubbles | [bubbles](https://github.com/charmbracelet/bubbles) | Pre-built components (inputs, lists, spinners) |
| Harmonica | [harmonica](https://github.com/charmbracelet/harmonica) | Animations |
| Log | [log](https://github.com/charmbracelet/log) | Pretty logging |
| Wish | [wish](https://github.com/charmbracelet/wish) | SSH server for TUIs |

### Troubleshooting

| Problem | Solution | Reference |
|---------|----------|-----------|
| `go: module not found` | Run `go mod tidy` | [Go Modules](https://go.dev/ref/mod#go-mod-tidy) |
| Colors not showing | Check `$TERM` supports colors | [ANSI Colors](https://en.wikipedia.org/wiki/ANSI_escape_code#Colors) |
| Keys not working | Try different terminal | [Bubble Tea FAQ](https://github.com/charmbracelet/bubbletea#faq) |
| Git commands slow | Increase refresh interval | [git](https://git-scm.com/docs) |
| Resize flickers | Use `tea.WithAltScreen()` | [Bubble Tea Options](https://pkg.go.dev/github.com/charmbracelet/bubbletea#WithAltScreen) |
| Import cycle | Check package dependencies | [Go Import Cycles](https://go.dev/doc/faq#import_cycle) |

### Quick Reference Card

```
╭─────────────────────────────────────────────────────────────╮
│                    QUICK REFERENCE                          │
├─────────────────────────────────────────────────────────────┤
│ Go Commands                                                 │
│   go mod init repotui     # Initialize module               │
│   go mod tidy             # Clean dependencies              │
│   go get <pkg>            # Add dependency                  │
│   go build ./cmd/repotui  # Build binary                    │
│   go run ./cmd/repotui    # Build and run                   │
├─────────────────────────────────────────────────────────────┤
│ Git Commands (used internally)                              │
│   git rev-list --left-right --count HEAD...@{upstream}     │
│   git add -A && git commit -m "msg" && git push            │
│   git fetch --all                                           │
├─────────────────────────────────────────────────────────────┤
│ Bubble Tea Pattern                                          │
│   Model.Init()   → Initial command (load data)              │
│   Model.Update() → Handle events, return new model          │
│   Model.View()   → Render UI as string                      │
├─────────────────────────────────────────────────────────────┤
│ Lip Gloss Pattern                                           │
│   style := lipgloss.NewStyle().Foreground(color).Bold(true) │
│   rendered := style.Render("text")                          │
│   width := lipgloss.Width(rendered)                         │
╰─────────────────────────────────────────────────────────────╯
```

---

## Appendix: Complete File List

After implementation, you should have:

```
repotui/
├── cmd/repotui/
│   └── main.go           (~30 lines)
├── internal/
│   ├── config/
│   │   └── config.go     (~70 lines)
│   ├── git/
│   │   └── git.go        (~180 lines)
│   └── ui/
│       ├── model.go      (~100 lines)
│       ├── update.go     (~170 lines)
│       ├── view.go       (~320 lines)  <- Includes responsive layout logic
│       └── styles.go     (~80 lines)
├── go.mod
└── go.sum

Total: ~950 lines of Go code
```

### File Responsibilities

| File | Primary Responsibility |
|------|------------------------|
| `main.go` | Entry point, program setup |
| `config.go` | Load TOML config, defaults |
| `git.go` | All git operations, repo scanning |
| `model.go` | State management, Bubble Tea model |
| `update.go` | Event handling, key bindings |
| `view.go` | Rendering, **responsive layout calculations** |
| `styles.go` | Lip Gloss color/style definitions |

---

*Last updated: January 29, 2026*
