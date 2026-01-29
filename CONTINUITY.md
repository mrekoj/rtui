Goal (incl. success criteria):
- Complete Phase 4 hardening + polish + expanded tests per RTUI_PHASES.md.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- Each phase completes only with report + tests/scripts.

Key decisions:
- Phase 4 completed per plan.

State:
  Done:
  - Phase 1–4 completed, tests and reports created.
  - Added git integration tests, phase4 scripts, smoke script.
  - Updated RTUI_TESTING.md and RTUI_PRODUCT_DOC.md for CWD banner behavior.
  Now:
  - Await commit/push instruction.
  Next:
  - None.

Open questions (UNCONFIRMED if needed):
- None.

Working set (files/ids/commands):
- internal/git/git.go
- internal/git/git_integration_test.go
- internal/ui/model.go
- internal/ui/update.go
- scripts/smoke.sh
- scripts/phase4_tests.sh
- reports/PHASE-4.md
- RTUI_TESTING.md
- RTUI_PRODUCT_DOC.md
- CONTINUITY.md
