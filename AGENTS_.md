# AGENTS.md

Go TUI project. Start: say hi + confirm task scope.
Style: telegraph; min tokens; skip filler.

## Protocol
- Read `../REPOTUI_PRODUCT_DOC.md` first (spec + reference code)
- Keep files <300 LOC; split when needed
- Commits: Conventional Commits (`feat|fix|refactor|docs|test|chore`)
- No destructive ops w/o explicit ok
- Format before commit: `gofmt -w .`

## Stack
| Layer | Choice | Docs |
|-------|--------|------|
| TUI | Bubble Tea | [github](https://github.com/charmbracelet/bubbletea) / [pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/bubbletea) |
| Styling | Lip Gloss | [github](https://github.com/charmbracelet/lipgloss) |
| Git | go-git + CLI | [pkg.go.dev](https://pkg.go.dev/github.com/go-git/go-git/v5) |
| Config | TOML | [toml.io](https://toml.io/) |

## Commands
```bash
go build ./cmd/repotui    # build
go run ./cmd/repotui      # dev run
go mod tidy               # clean deps
go vet ./...              # lint
gofmt -w .                # format
```

## Patterns

**Bubble Tea loop:**
```
Init()   -> startup command
Update() -> handle msg, return (model, cmd)
View()   -> return string
```

**Key rules:**
- Never block in `Update()` - return `tea.Cmd` for async
- Use `m.width`/`m.height` for responsive - no hardcoded sizes
- Use `strings.Builder` for view rendering
- Show errors in status bar, don't panic

## Structure
```
cmd/repotui/main.go       # entry
internal/config/          # settings
internal/git/             # git ops
internal/ui/              # model, update, view, styles
```

## Edit Guide
| Change | File(s) |
|--------|---------|
| Keybinding | `update.go` |
| Layout/render | `view.go` |
| Colors/theme | `styles.go` |
| App state | `model.go` |
| Git command | `git.go` |
| Config option | `config.go` |

## References
- Full spec: `../REPOTUI_PRODUCT_DOC.md`
- Bubble Tea examples: [examples/](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- Similar projects: [lazygit](https://github.com/jesseduffield/lazygit), [glow](https://github.com/charmbracelet/glow)
