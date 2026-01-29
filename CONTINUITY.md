Goal (incl. success criteria):
- Plan modular fsnotify watcher with tests-first and per-phase test+report+doc updates.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- User request: docs should not include code; code lives only in repo.
- New requirement: tests first; after each phase run tests, write report, update doc.

Key decisions:
- Treat bracketed paste as KeyRunes with Paste=true; strip newlines.
- Render footer after padding content to terminal height.

State:
  Done:
  - Added paste handling in add-path and commit input.
  - Added paste tests.
  - Ran scripts/phase4_tests.sh.
  - Pinned footer/status line to bottom of terminal.
  - Ran scripts/phase4_tests.sh after footer change.
  - Changed selected row to arrow "→" + text-only highlight on name/branch.
  - Ran scripts/phase4_tests.sh after selection change.
  - Removed header/title line.
  - Ran scripts/phase4_tests.sh after header removal.
  - Committed and pushed changes to origin/main.
  - Added repo table layout with | separators and header row.
  - Ran scripts/phase4_tests.sh after table change.
  - Set Status=8, Sync=6 widths and added "-" placeholder for empty Sync.
  - Updated layout tests.
  - Ran scripts/phase4_tests.sh after width change.
  - Switched to commit-only flow with auto-stage all; push is separate.
  - Mapped keys: c commit, o open editor, p push; removed confirm-stage mode.
  - Added commit/push completion messages and auto-refresh after pull/push/commit.
  - Updated RTUI_PRODUCT_DOC.md for new flow and keybindings.
  - Ran scripts/phase4_tests.sh after UX change.
  - Moved status text to REPOSITORIES header line (right-aligned).
  - Simplified footer to actions only.
  - Updated RTUI_PRODUCT_DOC.md UI spec and layout snippets for header status.
  - Ran scripts/phase4_tests.sh (escalated to access Go build cache).
  - Removed all Go code blocks from RTUI_PRODUCT_DOC.md and replaced with business-logic text/tables.
  - Simplified Implementation Guide to module responsibilities and runtime flows.
  - Filled UI spec rules (responsive, truncation, resize) without code.
  - Committed and pushed docs cleanup + header status update.
  - Found references: gh-dash uses polling interval defaults; gh run watch polls at 3s; watchman provides file-watching; gitwatch uses inotify/fswatch.
  - Chosen: fsnotify watcher-only, debounce 500ms, watch repo root + .git/index + .git/HEAD.
  - Phase 0 tests added for watcher debounce and ignore rules.
  - Phase 0 report written.
  - RTUI_PRODUCT_DOC.md updated with watcher-only flow and fsnotify dependency.
  - RTUI_TESTING.md updated with watcher tests + manual check.
  - Phase 1 watcher module implemented (fsnotify backend).
  - Added repo-path test and fsnotify dependency.
  - Phase 1 report written.
  - Ran scripts/phase4_tests.sh after Phase 1.
  - Phase 2 integration: watcher lifecycle wired into UI.
  - Added watcher integration tests.
  - Phase 2 report written.
  - Ran scripts/phase4_tests.sh after Phase 2.
  Now:
  - Commit and push watcher phases (0-2) changes.
  Next:
  - Await next request.

Open questions (UNCONFIRMED if needed):
- None.

Working set (files/ids/commands):
- internal/ui/view.go
- internal/ui/styles.go
- RTUI_PRODUCT_DOC.md
- CONTINUITY.md
