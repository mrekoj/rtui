# Phase 21 Report

Date: January 30, 2026
Scope: Bottom panel focus, CHANGES/GRAPH toggle, and scrollable lists.

## What changed
- Added panel focus (1/2) and Tab toggle between CHANGES and GRAPH.
- CHANGES and GRAPH lists scroll with j/k and PgUp/PgDn.
- Graph view uses git log graph lines; default bottom view is CHANGES.
- Updated docs and tests.

## Files changed
- internal/ui/model.go
- internal/ui/update.go
- internal/ui/view.go
- internal/ui/panel.go
- internal/ui/graph_cmds.go
- internal/ui/update_test.go
- internal/ui/panel_test.go
- internal/git/git.go
- internal/git/git_integration_test.go
- docs/RTUI_PRODUCT_DOC.md
- docs/RTUI_TESTING.md
- reports/PHASE-21.md

## Tests
- scripts/phase4_tests.sh (PASS)
