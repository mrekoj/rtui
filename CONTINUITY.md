Goal (incl. success criteria):
- Reorganize docs into docs/ and update all references; commit+push.

Constraints/Assumptions:
- Follow AGENTS.md: start with hi + confirm doc scope; telegraph style; minimal tokens; ASCII only.
- Docs are business logic only; no code snippets in RTUI_PRODUCT_DOC.md.
- Update "Last updated" on material doc edits with absolute date.

Key decisions:
- Docs relocated to docs/: RTUI_PRODUCT_DOC, RTUI_TESTING, RTUI_PHASES, BREW_RELEASE.
- References updated to docs/ paths (AGENTS, README, reports, docs).

State:
  Done:
  - Docs moved into docs/ and references updated.
  - Verified references via rg; checked git status.
  Now:
  - Await commit+push for docs reorg; untracked include README.md, docs/*, dist/*.
  Next:
  - Await user follow-up requests.

Open questions (UNCONFIRMED if needed):
- None.

Working set (files/ids/commands):
- docs/RTUI_PRODUCT_DOC.md
- docs/RTUI_TESTING.md
- docs/RTUI_PHASES.md
- docs/BREW_RELEASE.md
- README.md
- AGENTS.md
- reports/PHASE-*.md
- CONTINUITY.md
