# AGENTS.md

RTUI product documentation. Start: say hi + confirm doc scope. Style: telegraph; minimal tokens; skip filler.

## Continuity Ledger (compaction-safe)

Maintain a single Continuity Ledger for this workspace in `CONTINUITY.md`. The ledger is the canonical session briefing designed to survive context compaction; do not rely on earlier chat text unless it's reflected in the ledger.
### How it works

- At the start of every assistant turn: read `CONTINUITY.md`, update it to reflect the latest goal/constraints/decisions/state, then proceed with the work.
- Update `CONTINUITY.md` again whenever any of these change: goal, constraints/assumptions, key decisions, progress state (Done/Now/Next), or important tool outcomes.
- Keep it short and stable: facts only, no transcripts. Prefer bullets. Mark uncertainty as `UNCONFIRMED` (never guess).
- If you notice missing recall or a compaction/summary event: refresh/rebuild the ledger from visible context, mark gaps `UNCONFIRMED`, ask up to 1–3 targeted questions, then continue.

### `functions.update_plan` vs the Ledger

- `functions.update_plan` is for short-term execution scaffolding while you work (a small 3–7 step plan with pending/in_progress/completed).
- `CONTINUITY.md` is for long-running continuity across compaction (the "what/why/current state"), not a step-by-step task list.
- Keep them consistent: when the plan or state changes, update the ledger at the intent/progress level (not every micro-step).

### In replies

- Begin with a brief "Ledger Snapshot" (Goal + Now/Next + Open Questions). Print the full ledger only when it materially changes or when the user asks.

### `CONTINUITY.md` format (keep headings)

- Goal (incl. success criteria):
- Constraints/Assumptions:
- Key decisions:
- State:
    - Done:
    - Now:
    - Next:
- Open questions (UNCONFIRMED if needed):
- Working set (files/ids/commands):

## Protocol
- Read `RTUI_PRODUCT_DOC.md` first; treat it as source of truth
- Preserve the 14-section numbering and ToC anchors; update ToC when headings change
- Keep tone technical, direct, and consistent with existing sections
- Use ASCII only unless the file already uses non-ASCII
- No destructive ops without explicit user ok
- When making material edits, update the "Last updated" line with an absolute date

## Doc Goals
- Define product scope, UX, and implementation details for RTUI
- Keep requirements, keybindings, config, and error handling aligned
- Provide runnable, correct snippets that match the intended stack

## Stack (reference names used in doc)
| Layer | Choice |
|-------|--------|
| Language | Go 1.21+ |
| TUI | Bubble Tea |
| Styling | Lip Gloss |
| Git | go-git + git CLI |
| Config | TOML |

## Editing Guide
| Change | Update Sections |
|--------|-----------------|
| New feature | Overview, UI Spec, Keybindings, Config, Error Handling, Testing |
| UI layout or colors | UI Spec, Color Scheme |
| Git behavior | Git Commands Reference, Error Handling |
| Config fields | Config File, Data Structures |
| Data model | Data Structures, Implementation Guide |
| Dependencies | Prerequisites, Resources |

## Quality Checks
- Headings and anchors match the ToC
- Tables are aligned and not stale
- Examples use consistent names (RTUI, repo, path, branch)
- No broken or duplicate keybindings
- Commands match the doc (Go 1.21+, module name `rtui`)

## Conventions
- Prefer compact tables over long prose
- Use fenced code blocks for commands and snippets
- Avoid marketing language; focus on concrete behaviors
