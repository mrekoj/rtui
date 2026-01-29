Goal (incl. success criteria):
- Implement RTUI_PHASES.md; Phase 3 completed with report + tests.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- Each phase completes only with report + tests/scripts.

Key decisions:
- Phase 3 completed per plan (add-path, commit flow, prompts, fetch/refresh, help, filter).

State:
  Done:
  - Phase 1: scaffold + config + git CLI core + tests + report.
  - Phase 2: UI shell + layout + tests + report.
  - Phase 3: user flows + tests + report.
  Now:
  - Await instruction to start Phase 4.
  Next:
  - Phase 4: hardening + polish + expanded tests.

Open questions (UNCONFIRMED if needed):
- UNCONFIRMED: CI environment preference for running tests.

Working set (files/ids/commands):
- internal/ui/update.go
- internal/ui/view.go
- internal/ui/styles.go
- internal/ui/update_test.go
- scripts/phase3_tests.sh
- reports/PHASE-3.md
- CONTINUITY.md
