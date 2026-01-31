# TUI Testing Playbook (Shared)

Layers:
- Unit: pure logic, no external IO
- Integration: git/fs interactions with temp repos
- View: layout snapshots at multiple sizes

Guard tests (must cover):
- Missing config, empty paths
- Add invalid path, add duplicate
- Dirty or conflicting repo guards for sync
- No upstream handling

Manual smoke (minimum):
- Navigate, resize, scroll
- Trigger each primary action
- Verify bottom panel and modal behavior

Naming:
- scripts/phaseN_tests.sh per phase
- reports/PHASE-N.md per phase

*Last updated: January 30, 2026*
