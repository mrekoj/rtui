# RTUI Testing Guide

> Automated-first testing plan with safety guards and responsive checks.

## 1. Quick Start (Automated)

```bash
# Run all tests
GOFLAGS="-count=1" go test ./...

# Fast unit-only runs (recommended during dev)
GOFLAGS="-count=1" go test ./internal/config ./internal/git ./internal/ui

# Optional: run UI layout tests only
GOFLAGS="-count=1" go test ./internal/ui -run TestLayout
```

Environment variables used by tests (define in CI or local shell):
- `RTUI_TEST_REPOS_ROOT` (path where fixtures are created)

## 2. Test Layers

### 2.1 Unit Tests (fast, pure)
Focus: deterministic logic, no external git.
- `config.NormalizePath`: trims, expands `~`, cleans path.
- `config.AppendPath`: requires existing path, ignores duplicates, writes config.
- `config.Load/Save`: round-trip, preserves values.
- UI model state transitions: `ModeAddPath`, `ModeConfirmStage`, `ModeConfirmPull`, `ModeCommitInput`.

### 2.2 Integration Tests (git + filesystem)
Focus: git status counts and safety flows.
- `git.GetRepoStatus` on clean/dirty/staged/untracked/conflict repos.
- Ahead/behind counts with and without upstream.
- Add-path flow writes config, then rescan picks up new path.

### 2.3 View/Layout Tests (snapshot)
Focus: render stability for right-panel widths.
- Render widths: 35, 45, 60, 80; height 25.
- Assert: compact hides branch <40; narrow shows branch at 40-60; status/sync align.
- Store golden strings under `internal/ui/testdata/` (when implemented).

## 3. Guard Tests (must cover)

| Guard | Trigger | Expected |
|-------|---------|----------|
| No config | Missing config file | Scan CWD, show banner with path |
| Empty paths | `paths = []` | Scan CWD, show banner |
| Add path missing | Add non-existent path | Error; no config change |
| Add path duplicate | Add existing path again | Status message; no change |
| Add path ok | Add valid existing path | Config saved; rescan |
| Behind remote | `repo.Behind > 0` then `p` | Prompt pull first |
| Conflicts | `repo.HasConflict == true` then `p` | Block commit+push |
| No upstream | `git rev-list ... @{upstream}` fails | Show "-" for sync |
| Push/Pull fails | git CLI error | Error message in footer |

## 4. Fixture Setup (automated)

Use a helper that creates temp repos under `RTUI_TEST_REPOS_ROOT`.
Example script logic (for reference in tests):

```bash
ROOT=${RTUI_TEST_REPOS_ROOT:-/tmp/rtui-test-repos}
rm -rf "$ROOT" && mkdir -p "$ROOT"

# clean repo
mkdir "$ROOT/clean" && cd "$ROOT/clean"
git init && echo "# Test" > README.md && git add . && git commit -m "init"

# dirty repo
mkdir "$ROOT/dirty" && cd "$ROOT/dirty"
git init && echo "# Test" > README.md && git add . && git commit -m "init"
echo "changed" >> README.md

# staged repo
mkdir "$ROOT/staged" && cd "$ROOT/staged"
git init && echo "# Test" > README.md && git add . && git commit -m "init"
echo "new" > new.txt && git add new.txt
```

Tests should create and clean fixtures automatically.

## 5. Manual Smoke (minimum)

- Navigate list with `j/k` at width 45 cols.
- Add a path using `a`, verify config update and rescan.
- Commit flow: behind prompt, stage-all prompt, commit message.
- Toggle dirty-only filter.

## 6. Responsive Checks (right panel)

- 35x25: compact mode, branch hidden, hints shortened.
- 45x25: narrow mode, branch shown, actions visible.
- 60x25: narrow/normal boundary, truncated names.
- 80x25: normal layout.

*Last updated: January 29, 2026*
