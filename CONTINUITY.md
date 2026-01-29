Goal (incl. success criteria):
- Keep RTUI_PRODUCT_DOC.md at business-logic level; no code snippets.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- User request: docs should not include code; code lives only in repo.

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
  Now:
  - Await user review of doc cleanup (business-logic only).
  Next:
  - Commit and push if requested.

Open questions (UNCONFIRMED if needed):
- None.

Working set (files/ids/commands):
- internal/ui/view.go
- internal/ui/styles.go
- RTUI_PRODUCT_DOC.md
- CONTINUITY.md
