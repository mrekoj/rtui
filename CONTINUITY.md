Goal (incl. success criteria):
- Implement RTUI_PHASES.md; Phase 2 completed and docs synced; commit and push changes.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- Each phase completes only with report + tests/scripts.

Key decisions:
- Phase 2 completed per plan (UI shell + layout).

State:
  Done:
  - Phase 1: scaffold + config + git CLI core + tests + report.
  - Phase 2: UI shell + layout + tests + report.
  - Synced RTUI_PRODUCT_DOC.md model snippet with CWD fallback.
  Now:
  - Commit and push Phase 1-2 implementation and reports/scripts.
  Next:
  - Phase 3: user flows (add-path, commit flow, prompts) + flow tests.

Open questions (UNCONFIRMED if needed):
- UNCONFIRMED: CI environment preference for running tests.

Working set (files/ids/commands):
- go.mod, go.sum
- cmd/rtui/main.go
- internal/config/config.go
- internal/git/git.go
- internal/ui/*.go
- internal/ui/layout_test.go
- scripts/phase1_tests.sh
- scripts/phase2_tests.sh
- reports/PHASE-1.md
- reports/PHASE-2.md
- RTUI_PRODUCT_DOC.md
- CONTINUITY.md
