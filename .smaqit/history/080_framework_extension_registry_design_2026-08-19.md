# Framework Extension Registry Design

## Metadata

- **Date:** 2026-08-19
- **Focus:** ADK/product boundary cleanup in `smaqit`; registry-driven principle propagation design filed as Task 032 in the sibling `smaqit-adk` repo
- **Tasks referenced:** smaqit-adk Task 032 (new, filed this session), smaqit-adk Task 018 and Task 019 (abandoned this session, superseded by 032)

## Actions Taken

- Added a new "No Grandfathering" principle to `framework/SMAQIT.md` via `smaqit.new-principle` → `smaqit-L0`
- Investigated a suspected "project bleed" between `smaqit` and the sibling `smaqit-adk` repo; confirmed `smaqit-adk` is a real, separate repo and the original L0/L1/L2 extraction (its Task 054, 2026-04-09) was a deliberate, correct architectural decision
- Discovered `smaqit`'s own 9 product agents have no principle→agent compiler at all — `templates/agents/compiled/*.rules.md` were orphaned, never read by the actual build (`scripts/generate-agents.py`), and every agent/skill is hand-authored
- Planned and filed a registry-driven auto-propagation system as Task 032 in `smaqit-adk`'s own `.smaqit/tasks/`, superseding its Task 019 (Cross-Level Compilation) and folding in Task 018's (Level Skills Completion) remaining gaps; marked both Abandoned in `smaqit-adk`'s `PLANNING.md`
- Deleted 18 files from `smaqit` confirmed to be dead, generic ADK-owned content that escaped an earlier cleanup pass (Task 054): `templates/agents/compiled/*.rules.md` (8 smaqit-domain files, orphaned rather than misplaced), `templates/skills/base-skill.template.md` + `compiled/base.rules.md`, and the parallel duplicate set under `.smaqit/templates/agents/` and `.smaqit/templates/skills/`
- Fixed a stale precondition check in `skills/smaqit.new-greenfield-project/SKILL.md` that gated `smaqit.create-skill`'s availability on a project-local path instead of its real (global) dependency
- Verified `smaqit`'s own `framework/*.md` is not itself ADK bleed — diffed all 5 overlapping-named files against `smaqit-adk`'s base set (96%+ of lines differ, 2–9x larger) and grepped for residual ADK/Level-agent vocabulary (zero hits)
- Course-corrected an initial "merge everything into smaqit-adk" proposal into a base/extension inheritance model instead, and folded it into Task 032's design in `smaqit-adk`

## Problems Solved

- **Wrong propagation mechanism proposed initially:** first suggested chaining `smaqit-L1`→`smaqit-L2` to propagate a principle into smaqit's own agents; inspection of both agents' actual input/output contracts showed neither touches smaqit's canonical `agents/*.md` at all — they target a different, generic custom-agent output shape
- **Ambiguous "bleed" diagnosis:** separated three previously-conflated categories — (1) legitimate architectural separation (L0/L1/L2 living in smaqit-adk), (2) genuine dead ADK-generic files sitting in smaqit's repo, (3) smaqit's own orphaned-but-not-misplaced compiled-rules experiment
- **Merge-into-smaqit-adk would have recontaminated the ADK:** the user's first framing ("there can be only one set of framework files, they should live in smaqit-adk") was walked back after establishing it would reverse smaqit-adk's own Task 001 ("Clean L2 Agent Contamination") and Task 054 decisions to keep the ADK domain-agnostic

## Decisions Made

- Task 032 (smaqit-adk) supersedes Task 019 outright rather than extending or sequencing after it — automatic propagation obsoletes 019's manual `smaqit.compile.*` chain premise
- Registry scope excludes smaqit's own hand-authored product agents (`agents/*.md`, compiled via `scripts/generate-agents.py`) — a separate, unrelated pipeline Task 032 does not touch
- Framework content model: smaqit-adk's `framework/*.md` remains the single base; a consuming project's product-domain principles live in that project's own repo at `.smaqit/framework/*.md` as an **extension**, deduplicated against the base rather than merged into or duplicated from it
- `smaqit.new-principle` is designed (in Task 032) to become scope-aware (base vs. extension) and to enforce non-duplication going forward when authoring new extension principles
- The 8 orphaned `templates/agents/compiled/*.rules.md` files were deleted rather than kept as a starting point for a future smaqit-side compiler
- smaqit's own migration (`framework/*.md` → `.smaqit/framework/*.md` + dedup pass) is deliberately left undone this session — it depends on Task 032's mechanism shipping in `smaqit-adk` first, and belongs in a future smaqit-side task, not this session's file changes

## Files Modified

| File | Repo | Action |
|------|------|--------|
| `framework/SMAQIT.md` | smaqit | Added "No Grandfathering" principle |
| `templates/skills/base-skill.template.md` | smaqit | Deleted — dead ADK-generic content |
| `templates/skills/compiled/base.rules.md` | smaqit | Deleted — dead ADK-generic content |
| `templates/agents/compiled/{business,coverage,deploy,develop,functional,infrastructure,stack,validate}.rules.md` | smaqit | Deleted — orphaned, unused by `generate-agents.py` |
| `.smaqit/templates/agents/{base-agent,implementation-agent,specification-agent}.template.md` | smaqit | Deleted — dead ADK-generic duplicate |
| `.smaqit/templates/agents/compiled/{base,implementation,specification}.rules.md` | smaqit | Deleted — dead ADK-generic duplicate |
| `.smaqit/templates/skills/base-skill.template.md` | smaqit | Deleted — dead ADK-generic duplicate |
| `.smaqit/templates/skills/compiled/skill.rules.md` | smaqit | Deleted — dead ADK-generic duplicate |
| `skills/smaqit.new-greenfield-project/SKILL.md` | smaqit | Fixed stale project-local availability check |
| `.smaqit/tasks/032_registry_driven_principle_propagation.md` | smaqit-adk | Created, then revised for the base/extension course correction |
| `.smaqit/tasks/PLANNING.md` | smaqit-adk | Added Task 032; moved Tasks 018 and 019 to Abandoned |
| `.smaqit/tasks/018_level_skills_completion.md` | smaqit-adk | Status marked Abandoned, superseded by 032 |
| `.smaqit/tasks/019_cross_level_compilation.md` | smaqit-adk | Status marked Abandoned, superseded by 032 |

## Next Steps

- Update 3 live wiki docs still describing the now-dead `templates/agents/compiled/*.rules.md` mechanism as current: `docs/wiki/workflows/extending-smaqit.md`, `docs/wiki/designs/hierarchical-levels.md`, `docs/wiki/designs/level-up-compilation.md` — awaiting a decision on doing this now vs. filing it as a task
- File a smaqit-side task for the `framework/*.md` → `.smaqit/framework/*.md` migration and dedup pass, once smaqit-adk's Task 032 mechanism exists
- Nothing from this session is committed in either repo — both `smaqit` and `smaqit-adk` working trees carry uncommitted changes pending review
- The "No Grandfathering" principle added in Step 1 of this session remains unpropagated to smaqit's own agents/skills; Task 032 explicitly does not solve this (registry excludes hand-authored product agents) — it needs its own resolution once the extension migration lands

## Session Metrics

- Files modified/deleted (smaqit): 10 (1 modified, 9 deletions covering 18 total file paths)
- Files created/modified (smaqit-adk): 4
- Tasks filed: 1 (smaqit-adk Task 032)
- Tasks abandoned: 2 (smaqit-adk Tasks 018, 019)
- Commits made: 0
