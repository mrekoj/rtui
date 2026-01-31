# RTUI

Minimal TUI dashboard to monitor and manage multiple git repos.

## Why
- See repo status, branch, and sync at a glance
- Quick commit/pull/push from one screen
- Toggle CHANGES and GRAPH for the selected repo
- Designed to run in the right 1/3 of your terminal

## Install (macOS)

Via Homebrew:
```bash
brew tap mrekoj/rtui
brew install rtui
```

## Run
```bash
rtui
```

## Config
Config file:
```
~/.config/rtui/config.toml
```

Add repo paths inside `paths = [...]` or press `a` in the app.

Example:
```toml
paths = [
  "~/SourceCode/Miwiz",
  "~/SourceCode/Personal",
]

editor = "code"
refresh_interval = 0
show_clean = true
scan_depth = 1
```

## Keybindings (core)

Navigation
- `j/k` or arrows: move
- `1`: focus repo list
- `2`: focus bottom panel
- `Tab`: toggle CHANGES <-> GRAPH (bottom panel)

Actions
- `a`: add path
- `b`: switch branch
- `c`: commit (stages all)
- `p`: pull
- `P`: push
- `f`: fetch
- `r`: refresh
- `o`: open repo in editor
- `s`: open config in VS Code
- `?`: help
- `q`: quit

## Notes
- Auto-refresh uses file watcher (fsnotify). Manual `r` still available.
- Push is blocked if repo is dirty or behind; pull is blocked if dirty.

## Docs
- Product spec: `docs/RTUI_PRODUCT_DOC.md`
- Test plan: `docs/RTUI_TESTING.md`
- Homebrew release: `docs/BREW_RELEASE.md`
