# RTUI Implementation Phases

> Each phase is complete only when both a report and tests/scripts exist.

## Phase 1: Core Foundations

Scope:
- Project scaffold (`cmd/`, `internal/`), config load/save/append, git CLI wrapper (scan, status porcelain parse, branch, ahead/behind).

Deliverables:
- Working CLI build.
- Config and git modules usable by UI.

Tests/scripts required:
- Unit tests for `internal/config` and git porcelain parsing.
- Script: `scripts/phase1_tests.sh` (runs `go test` targets).

Report required:
- `reports/PHASE-1.md` (what was built, risks, test output summary).

## Phase 2: UI Shell + Layout

Scope:
- Bubble Tea model/update/view skeleton.
- Repo list rendering, header/footer, responsive layout (40-60 cols target).

Deliverables:
- UI renders list, responds to resize.

Tests/scripts required:
- Layout/view tests (widths 35/45/60/80).
- Script: `scripts/phase2_tests.sh`.

Report required:
- `reports/PHASE-2.md`.

## Phase 3: Primary Flows

Scope:
- Add-path modal, commit flow (auto-stage all), pull/push guards, fetch/refresh, status/error handling.

Deliverables:
- End-to-end user flows in TUI.

Tests/scripts required:
- Flow tests + integration tests with temp repos.
- Script: `scripts/phase3_tests.sh`.

Report required:
- `reports/PHASE-3.md`.

## Phase 4: Hardening + Polish

Scope:
- Performance on many repos, edge cases, docs alignment, smoke checks.

Deliverables:
- Stable v1 release behavior.

Tests/scripts required:
- Full suite per `docs/RTUI_TESTING.md` + smoke script.
- Script: `scripts/phase4_tests.sh`.

Report required:
- `reports/PHASE-4.md`.

*Last updated: January 30, 2026*
