# RTUI - Implementation Guide

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
- Bottom panel toggles between CHANGES and GRAPH views
- Switches branches with a picker
- Allows quick commit and push without leaving the dashboard
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
mkdir -p rtui/cmd/rtui rtui/internal/{config,git,ui}
cd rtui

# 2. Initialize Go module
go mod init rtui

# 3. Add dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/BurntSushi/toml

# 4. Create files (see Section 6 for code)
# ... implement each file ...

# 5. Build and run
go build ./cmd/rtui
./rtui

# Or run directly during development
go run ./cmd/rtui
```

---

## 4. Architecture

### Project Structure

> 📁 Following [Go project layout conventions](https://go.dev/doc/modules/layout#package-or-command-with-supporting-packages)

```
rtui/
├── cmd/rtui/
│   └── main.go              # Entry point, starts Bubble Tea
├── internal/
│   ├── config/
│   │   └── config.go        # Load ~/.config/rtui/config.toml (TOML)
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

RTUI follows this loop with a small state machine (modes) and a single render function.


### Data Flow
```
┌──────────────────────────────────────────────────────────────┐
│                                                              │
│  Load config ──► Scan repos ──► UI model ──► View          │
│                                                              │
│  User presses 'c' ──► Update() ──► commit input              │
│                                └─► commit (stage all)       │
│  User presses 'p' ──► Update() ──► pull                      │
│  User presses 'P' ──► Update() ──► push                      │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## 5. Data Structures

> 📚 See [Go struct documentation](https://go.dev/tour/moretypes/2) and [git status --porcelain](https://git-scm.com/docs/git-status#_short_format)

### Core Types (define in `internal/git/git.go`)

**FileStatus values**
- Staged: file has staged changes
- Modified: file has unstaged changes
- Untracked: file is new/untracked
- Conflict: file has merge conflict markers

**ChangedFile**
| Field | Meaning |
|-------|---------|
| Path | Repo-relative file path |
| Status | One of the FileStatus values above |

**Repo**
| Field | Meaning |
|-------|---------|
| Name | Folder name (repo display name) |
| Path | Full filesystem path |
| Branch | Current branch or detached label |
| Staged | Count of staged files |
| Modified | Count of modified files |
| Untracked | Count of untracked files |
| Ahead | Commits ahead of upstream |
| Behind | Commits behind upstream |
| HasConflict | Any merge conflicts present |
| ChangedFiles | List of files with changes |

**Derived state**
- Dirty: any of Staged, Modified, Untracked > 0
- Clean: none of the above

### Config Type (define in `internal/config/config.go`)

| Field | Purpose | Default |
|-------|---------|---------|
| paths | Folders to scan for repos | empty (falls back to CWD) |
| editor | Command to open repo | "code" |
| refresh_interval | Reserved for future polling (unused in watcher-only). | 0 |
| show_clean | Show clean repos | true |
| scan_depth | Max depth under each path | 1 |

### UI Model (define in `internal/ui/model.go`)

**Modes**
- Normal: repo list + changes panel
- AddPath: add-path modal open
- CommitInput: commit message input
- BranchPicker: local/remote branch picker
- ConfirmStash: stash-and-switch confirmation
- Help: help modal

**State fields**
| Field | Meaning |
|-------|---------|
| repos | Current repo list |
| config | Loaded config |
| cursor | Selected repo index |
| mode | Current UI mode |
| addPathInput | Text buffer for add-path |
| commitMsg | Text buffer for commit message |
| filterDirty | Show only dirty repos |
| panelFocus | Which panel is focused (repo list or bottom panel) |
| bottomView | CHANGES or GRAPH |
| changesScroll | Scroll offset for changes list |
| graphScroll | Scroll offset for graph list |
| graphLines | Cached graph lines for current repo |
| width / height | Terminal size |
| statusMsg | Status text shown in header line |
| loading | True during refresh |
| err | Last error (if any) |

### Layout Type (define in `internal/ui/view.go`)

| Field | Meaning |
|-------|---------|
| Name | Width for repo name column |
| Branch | Width for branch column (0 = hidden) |
| Status | Width for status column |
| Sync | Width for sync column |

---

## 6. Implementation Guide

> 💡 **Implementation Order:** Create files in this order: `main.go` → `config.go` → `git.go` → `styles.go` → `model.go` → `update.go` → `view.go`
>
> 📖 **Reference while coding:**
> - [Bubble Tea API](https://pkg.go.dev/github.com/charmbracelet/bubbletea)
> - [Lip Gloss API](https://pkg.go.dev/github.com/charmbracelet/lipgloss)
> - [TOML library](https://pkg.go.dev/github.com/BurntSushi/toml)

### Module Responsibilities

| Module | Responsibility |
|--------|----------------|
| `cmd/rtui/main.go` | Load config, start Bubble Tea program |
| `internal/config` | Read/write TOML config, path normalization |
| `internal/git` | All git status/commit/push/pull/fetch calls |
| `internal/watch` | File system watcher for auto-refresh (fsnotify) |
| `internal/ui/model` | Holds UI state and modes |
| `internal/ui/update` | Handles key events and async commands |
| `internal/ui/view` | Renders list, panels, and modals |
| `internal/ui/styles` | Colors and typography rules |

### Runtime Flows

- Startup: load config -> scan repos -> render list
- Refresh: `r` triggers rescan and updates header status
- Auto-refresh (watcher-only): file events trigger per-repo refresh after 500ms debounce
- Commit: `c` opens commit input; commit auto-stages all
- Branch switch: `b` opens picker; select branch and switch; remote creates tracking
- Pull: `p` pulls current repo; blocked if repo is dirty; after pull, auto-refresh
- Push: `P` pushes current repo; blocked if dirty or behind; after push, auto-refresh
- Add path: `a` opens input; append path, rescan
- Bottom panel: `Tab` toggles CHANGES/GRAPH; `1`/`2` switch focus
- Settings: `s` opens the config file in VS Code

### Auto-refresh (watcher-only)

- Watch scope: repo root + `.git/index` + `.git/HEAD`
- Debounce: 500ms per repo (coalesce rapid changes)
- No polling; manual refresh (`r`) remains available
- On watcher error: show header status and rely on manual refresh


---

## 7. UI Specification

### Responsive Design Philosophy

The UI adapts to any terminal size. No fixed widths - everything scales dynamically.

Primary target: run in the right 1/3 of a terminal while coding on the left 2/3.
Design for 40-60 columns as the common panel width.
Action bar must never overflow; it uses full labels with the hotkey letter colored (cyan) and wraps to two lines when needed.

### Layout Structure
```
┌─────────────────────────────────────────────────────────────────────────┐
│ REPOSITORIES [1]                                          Refreshed    │
│─────────────────────────────────────────────────────────────────────────│
│   Name             | Branch         | Status   | Sync                   │
│ → miwiz-api        | main           | 2M       | ↓3                     │
│   miwiz-web        | feature/auth   | 1S       | ↑2                     │
│   miwiz-cms        | develop        | ✓        | -                      │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ CHANGES [2]                                                             │
│─────────────────────────────────────────────────────────────────────────│
│ Staged (1)                                                              │
│   src/components/Header.tsx                                             │
│ Modified (0)                                                            │
│ Untracked (0)                                                           │
├─────────────────────────────────────────────────────────────────────────┤
│ add path   branch   commit   push   open   pull   refresh   ?          │
│ (hotkey letter is colored; wraps to second line when needed)           │
└─────────────────────────────────────────────────────────────────────────┘
```

### Add Path Modal

Press `a` to add a repo path. Show a centered modal input box (about 70% of panel width; clamp to 30-72 cols), then append the path to config and rescan.
Rules: path must already exist; duplicates are ignored.

```
┌─────────────────────────────────────┐
│ Add repo path                       │
│ /Users/you/SourceCode              │
│                                     │
│ [Enter]=save  [Esc]=cancel          │
└─────────────────────────────────────┘
```

### Branch Picker Modal

Press `b` to switch branches for the selected repo. Default view is Local.

Behavior:
- Tabs for Local and Remote branches.
- Local tab shows only local branches; Remote tab shows only remotes.
- Current branch is highlighted.
- Type to filter (case-insensitive). Filter is stored per tab.
- Scroll markers appear only when items are hidden above/below.
- Long lists scroll; selection stays visible.

### Bottom Panel Toggle

- Bottom panel defaults to CHANGES view.
- `Tab` toggles CHANGES <-> GRAPH when bottom panel is focused (`2`).
- `1` focuses repo list; `2` focuses bottom panel.
- `j/k` scrolls the focused panel; `PgUp/PgDn` fast scrolls.
- GRAPH view shows `git log --graph --oneline` for the selected repo.
- Long lists scroll; scroll position is preserved per view.
- Panel height is fixed to the available space; switching views does not shift layout.
- `Enter` switches to the selected branch.
- If dirty: prompt to stash and then switch.
- Selecting a remote creates a local tracking branch automatically.

```
┌─────────────────────────────────────┐
│ Switch branch  [Local] [Remote]     │
│ Filter: feat                        │
│                                     │
│ → feature/auth                      │
│   feature/api                       │
│                                     │
│                                     │
│ [Enter]=switch  [Esc]=cancel        │
└─────────────────────────────────────┘
```

**Stash confirm**
```
┌─────────────────────────────────────┐
│ Repo has uncommitted changes.       │
│ Stash and switch?                   │
│                                     │
│ [s]tash  [c]ancel                   │
└─────────────────────────────────────┘
```

### Responsive Column Layout

Columns use **fixed/flexible widths** tuned for narrow panels (40-60 cols):

| Column | Width | Behavior | Priority |
|--------|-------|----------|----------|
| Cursor | 2 chars | Fixed | Always shown |
| Name | Flexible | Gets remaining space | High |
| Branch | Flexible | Uses remainder after Name | High |
| Status | 8 chars | Fixed (icons + counts) | Medium |
| Sync | 6 chars | Fixed | High |

### Breakpoints (target: right 1/3 panel)

| Terminal Width | Behavior |
|----------------|----------|
| < 40 chars | Compact mode: hide branch, shorten status/sync |
| 40-60 chars | Narrow mode: show branch, truncate names |
| > 60 chars | Normal mode: show full columns, extra padding |

### Responsive Implementation

Rules:
- Cursor width fixed at 2
- Status width fixed at 8, Sync width fixed at 6
- Name/Branch split 55/45 of remaining space
- Minimums: Name >= 10, Branch >= 8 (when visible)
- Wide terminals: cap Name at 30, Branch at 25
- Compact (< 40 cols): hide Branch, keep Name >= 8

### Vertical Responsiveness

Rules:
- Bottom panel renders only when there is room for header plus at least one content line.
- If remaining space after repo list is < 3 lines, the bottom panel is hidden.
- When shown, the bottom panel uses the remaining lines and content scrolls.


### Truncation Rules

When content exceeds column width:
- Repo name and branch truncate from the end with ellipsis
- File paths truncate from the left to keep filename visible

**Examples:**
- `"my-very-long-repo-name"` → `"my-very-long-re…"` (15 chars)
- `"src/components/Header.tsx"` → `"…ponents/Header.tsx"` (20 chars, path)

### Dynamic Separator Lines

Separators should span the full width:
use a full-width line matching terminal width.


### Window Resize Handling

The UI automatically re-renders on terminal resize:
recompute layout and reflow columns without changing selection.


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
| `git log --graph --oneline -n N` | [git-log](https://git-scm.com/docs/git-log) | Graph view lines |
| `git add -A` | [git-add](https://git-scm.com/docs/git-add) | Stage all changes |
| `git commit -m "msg"` | [git-commit](https://git-scm.com/docs/git-commit) | Create commit |
| `git push` | [git-push](https://git-scm.com/docs/git-push) | Push to remote |
| `git pull` | [git-pull](https://git-scm.com/docs/git-pull) | Fetch and merge |
| `git fetch --all` | [git-fetch](https://git-scm.com/docs/git-fetch) | Fetch all remotes |
| `git branch --list` | [git-branch](https://git-scm.com/docs/git-branch) | List local branches |
| `git branch -r` | [git-branch](https://git-scm.com/docs/git-branch) | List remote branches |
| `git checkout <branch>` | [git-checkout](https://git-scm.com/docs/git-checkout) | Switch local branch |
| `git checkout -t <remote>` | [git-checkout](https://git-scm.com/docs/git-checkout) | Create tracking branch |
| `git stash push -u` | [git-stash](https://git-scm.com/docs/git-stash) | Stash dirty changes |

Note: `git add -A` runs automatically when the user commits.

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

# List local branches
git branch --list
# Docs: https://git-scm.com/docs/git-branch

# List remote branches
git branch -r
# Docs: https://git-scm.com/docs/git-branch

# Switch local branch
git checkout <branch>
# Docs: https://git-scm.com/docs/git-checkout

# Create tracking branch from remote
git checkout -t <remote>
# Docs: https://git-scm.com/docs/git-checkout

# Stash dirty changes (including untracked)
git stash push -u
# Docs: https://git-scm.com/docs/git-stash
```

**Understanding `@{upstream}`:** See [gitrevisions](https://git-scm.com/docs/gitrevisions#Documentation/gitrevisions.txt-emltaboranchnamegt64telemerename93telemeregt)

---

## 9. Keybindings

| Key | Action | Mode |
|-----|--------|------|
| `j` / `↓` | Next repo | Normal |
| `k` / `↑` | Previous repo | Normal |
| `a` | Add repo path | Normal |
| `b` | Switch branch (picker) | Normal |
| `c` | Commit (stages all) | Normal |
| `o` | Open repo in editor | Normal |
| `s` | Open settings (config.toml) in VS Code | Normal |
| `p` | Pull | Normal |
| `P` | Push | Normal |
| `f` | Fetch all remotes | Normal |
| `r` | Refresh status | Normal |
| `d` | Toggle dirty-only filter | Normal |
| `1` | Focus repo list | Normal |
| `2` | Focus bottom panel | Normal |
| `Tab` | Toggle CHANGES/GRAPH (bottom panel) | Normal |
| `PgUp` / `PgDn` | Fast scroll focused panel | Normal |
| `?` | Show help | Normal |
| `q` | Quit | Normal |
| `Enter` | Confirm commit | Commit Input |
| `Esc` | Cancel | Commit Input |
| `Enter` | Switch to selected branch | Branch Picker |
| `Esc` | Close branch picker | Branch Picker |
| `Tab` / `l` / `r` | Toggle Local/Remote view | Branch Picker |
| `s` | Stash and switch | Confirm Stash |
| `c` | Cancel | Confirm Stash |

---

## 10. Config File

> 📖 Reference: [TOML Specification](https://toml.io/en/), [XDG Base Directory](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html), [TOML Validator](https://www.toml.io/en/validator)

**Location:** `~/.config/rtui/config.toml`

Tip: Press `s` to open the config file in VS Code.

Config keys:

| Key | Type | Default | Notes |
|-----|------|---------|-------|
| `paths` | array[string] | empty | Folders to scan; supports `~` expansion; saved as multi-line TOML array |
| `editor` | string | `$EDITOR` or `code` | Editor command used by `o` and `s` |
| `refresh_interval` | int | 30 | Reserved for polling mode; set `0` in watcher-only |
| `show_clean` | bool | true | Show clean repos in list |
| `scan_depth` | int | 1 | Max depth under each path |

Notes:
- If the config file is missing or `paths` is empty, RTUI scans the current working directory (CWD) and shows a banner with the path.
- The UI `a` (add path) appends a normalized path to `paths` and writes the config file. Path must exist; duplicates are ignored.
- For a sample TOML file and shared config conventions, see `docs/shared/TUI_CONFIG_STANDARD.md`.

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

Use ANSI color IDs from the table; keep base text neutral and reserve bright colors for statuses.

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
| Pull blocked (dirty/conflict) | Show status message; no action |
| Push blocked (dirty/behind/conflict) | Show status message; no action |
| Pull fails | Show error message in header status line |
| Push fails | Show error message in header status line |
| Watcher error | Show status warning, rely on manual refresh |
| Graph load fails | Show status message, keep current view |
| Branch switch fails | Show error message, stay on current branch |
| Stash fails | Show error, keep picker open |
| Network error | Show error, allow retry |

### Error Display
- Errors show in the header status area
- Use red color for errors
- Info/success messages auto-clear after ~5 seconds
- Errors persist until the next user action

---

## 13. Testing

See `docs/RTUI_TESTING.md` for the automated test plan, guard checks, and manual/responsive checklists.

---

## 14. Resources & Reference Links

### Local Docs

- `docs/BREW_RELEASE.md` - Homebrew release steps (macOS)
- `docs/shared/` - Reusable TUI docs (runtime flow, keybindings, config, testing, release)

### Core Dependencies

| Library | GitHub | Documentation | Used For |
|---------|--------|---------------|----------|
| Bubble Tea | [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) | TUI framework (Model-View-Update) |
| Lip Gloss | [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/lipgloss) | Terminal styling & colors |
| Git CLI | [git/git](https://github.com/git/git) | [git-scm.com/docs](https://git-scm.com/docs) | Git operations via CLI |
| TOML | [BurntSushi/toml](https://github.com/BurntSushi/toml) | [pkg.go.dev](https://pkg.go.dev/github.com/BurntSushi/toml) | Config file parsing |
| fsnotify | [fsnotify/fsnotify](https://github.com/fsnotify/fsnotify) | [pkg.go.dev](https://pkg.go.dev/github.com/fsnotify/fsnotify) | File system watcher (auto-refresh) |

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
│   go mod init rtui     # Initialize module               │
│   go mod tidy             # Clean dependencies              │
│   go get <pkg>            # Add dependency                  │
│   go build ./cmd/rtui  # Build binary                    │
│   go run ./cmd/rtui    # Build and run                   │
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
rtui/
├── cmd/rtui/
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

*Last updated: January 31, 2026*
