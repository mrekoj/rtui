Goal (incl. success criteria):
- Remove header/title line per user preference.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.

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
  Now:
  - Await commit/push instruction.
  Next:
  - Commit and push if requested.

Open questions (UNCONFIRMED if needed):
- None.

Working set (files/ids/commands):
- internal/ui/view.go
- internal/ui/styles.go
- CONTINUITY.md
