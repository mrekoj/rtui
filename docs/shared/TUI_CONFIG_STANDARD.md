# TUI Config Standard (Shared)

Location:
- Use XDG: ~/.config/<app>/config.toml

Format:
- TOML (human friendly, comments allowed)

Common keys:
- paths: array of strings (folders to scan)
- editor: command string (used by open action)
- refresh_interval: int seconds (0 disables polling)
- show_clean: bool (list clean items)
- scan_depth: int (directory depth)

Conventions:
- Support ~ expansion in paths
- Save paths as a multi-line TOML array
- Ignore duplicates and non-existent paths

Sample TOML:

paths = [
  "~/SourceCode/Work",
  "~/SourceCode/Personal",
]
editor = "code"
refresh_interval = 0
show_clean = true
scan_depth = 1

*Last updated: January 30, 2026*
