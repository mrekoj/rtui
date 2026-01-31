# TUI Runtime Flow (Shared)

Purpose: describe the standard data flow for Bubble Tea based TUIs.

Flow (high level):
- Input (keys, resize) -> Update loop
- Update loop -> state change + side effects
- Side effects -> data fetch (git, fs, config)
- State -> View assembly
- View -> Styling/layout
- Render -> Terminal

Component map (typical):

| Component | Role | Library |
|----------|------|---------|
| Event loop | process input, dispatch updates | Bubble Tea |
| State model | in-memory app state | Go structs |
| IO layer | git, fs, config | go-git + git CLI + fsnotify + TOML |
| Layout | spacing, borders, align | Lip Gloss |
| Rendering | string output | Bubble Tea view |

*Last updated: January 30, 2026*
