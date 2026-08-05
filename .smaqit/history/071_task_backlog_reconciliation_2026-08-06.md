# Task Backlog Reconciliation

**Date:** 2026-08-06
**Session Focus:** Review three stale-looking Active tasks (097, 095, 070) against current codebase state, resolving each to Abandoned or Completed as warranted, plus one direct doc fix and a research-map refresh.
**Tasks Referenced:** 097, 095, 070
**Tasks Completed:** 095
**Tasks Abandoned:** 097, 070

---

## Actions Taken

### Task 097 — Lightweight Task-Lifecycle Entry Point for Infrastructure-Only Work
- User had no memory of the task and asked whether it was still relevant. Investigation found its core premise wrong: the task assumed `smaqit.task-start` lived under `skills/` in this repo, but task-lifecycle skills (`task-start`, `task-create`, `task-complete`, `task-list`, `utils.worktree`) are owned entirely by the sibling `smaqit-extensions` repo and never shipped from `smaqit` itself.
- The underlying discoverability gap (nothing tells an operator when standalone `task.start` is appropriate vs. `smaqit.feature-new`) was confirmed still real, but the smaqit-side fix reduces to two small doc edits — not enough to justify tracked-task overhead. Abandoned, with reasoning recorded in the task file.
- Directly fixed the stale pointer this surfaced: `skills/smaqit.new-greenfield-project/SKILL.md:216` named "individual smaqit agents" for post-MVP work instead of `smaqit.feature-new`, which has existed since task 089. Corrected to name it explicitly.

### Task 095 — `smaqit.feature-new` Per-Phase Worktree Spawn
- User asked what remained to do in smaqit now that `smaqit-extensions` had shipped parent-owned task lifecycle (v1.10.0). Traced the actual implementation history: task 095's design and the corresponding rewrite of `skills/smaqit.feature-new/SKILL.md` (Phase 0 parent-first creation, child-aware `task.start`/`task.complete` across Phases 1–5, release-PR-as-sole-vehicle gate) had already shipped together in commit `2bec633` ("v2.0.0") — the task file was simply never run through `task.complete`.
- Verified all Acceptance Criteria against the current `SKILL.md` content directly. 9 of 10 confirmed satisfied; the 10th (a Gotcha naming "the downstream project's task 022") was intentionally left unimplemented since it would violate CONTRIBUTING.md's later no-downstream-naming rule.
- Attempted standard completion; the lifecycle resolver correctly refused (`Task 095 must be In Progress in a registered worktree before completion`) since the task had never been started via `task.start` and had no branch/worktree to merge or clean up. Followed the same precedent used for the earlier task 097 (history 070): completed task bookkeeping directly, with no git lifecycle operations, since the code was already merged to `main`.

### Task 070 — E2E Boundary Enforcement Validation
- User asked whether this January-2026 validation task was still relevant. Found it written entirely against mechanics that no longer exist: `.github/prompts/smaqit.business.prompt.md` and `/smaqit.business` as a direct slash command (both removed by task 081's prompt deprecation and the orchestration-first/Task-delegated subagent migration, tasks 082/073), and `docs/tasks/PLANNING.md` (moved to `.smaqit/tasks/PLANNING.md` long ago). It also targeted a since-redesigned "assessment skill" with a fixed 6-component output shape that no longer matches what shipped as `smaqit.session-assess`. The task file itself was also corrupted — garbled, interleaved text across its checklist sections.
- Confirmed the underlying capability it wanted to check (System Actor / technical-verb leakage into Business specs) has zero matches for "System Actor" across `agents/`, `skills/`, `framework/` today, and has had extensive indirect validation via real Business specs generated across many downstream projects since, with no boundary-violation issue ever surfacing.
- Abandoned, with full reasoning recorded in the task file.

### Research Map Refresh
- `session.finish` detected the research map (`.smaqit/references/project-research.md`) was 10 days stale and `installer/go.mod` had changed since the last refresh — triggered `smaqit.project-research` to rebuild it.
- Diffed `installer/go.mod` and the newly-discovered `tools/plantuml/package.json` against the prior map: found three new dependencies from the PlantUML visual-design work (task 098) not yet tracked — `github.com/modelcontextprotocol/go-sdk`, `github.com/tailscale/hujson`, plus the npm packages `@plantuml/mcp-js`, `@resvg/resvg-wasm`, and `@fontsource/noto-sans`.
- Found and fixed a data-format bug while running the map's own `verify-urls.sh`: the script's input format is 3-column (`TOOL\tSECTION\tURL`), not the 4-column (with `LAYER`) format implied by the skill's own SKILL.md step 3 description — the extra column silently broke every `curl` call (all URLs returned status `000`). Corrected the input file and re-ran; all previously-known URLs verified live. The three new npm package URLs returned genuine `403` from npmjs.com (confirmed via direct `curl`, not a script artifact) and are recorded as unreachable rather than dropped.

## Problems Solved

- **Two stale tasks closed out with accurate reasoning** rather than either blindly completing them or leaving them to rot — both 097 and 070 turned out to have wrong premises (repo ownership, deprecated mechanics) that only surfaced through direct investigation against current code, not just re-reading the task text.
- **One already-shipped task (095) reconciled with its own bookkeeping** — the actual `smaqit.feature-new` v2.0.0 implementation has been live since 2026-07-29 with no gaps found; this session only closed the paperwork gap.
- **`verify-urls.sh`'s column-count bug found and worked around** — every URL check was silently failing with status `000` (indistinguishable from "no network access" without inspecting the script) until the extra `LAYER` column was dropped from the input file.

## Decisions Made

- **Task 097 abandoned rather than rescoped into a smaller task.** The remaining smaqit-side fix (a `feature-new` trigger-condition note plus a doc pointer, already done directly) was deemed too small to be worth tracked-task overhead on its own.
- **Task 095 completed via bookkeeping-only path, no retroactive `task.start`.** Spinning up a worktree/branch now for code already merged to `main` would have been theater with nothing to actually branch or merge — same precedent as the earlier task 097's direct-completion handling in history 070.
- **Task 095's task-022 AC (Gotcha naming a downstream project) deliberately left unimplemented.** CONTRIBUTING.md's no-downstream-naming rule postdates the AC and directly forbids it; the AC is treated as superseded, not failed.
- **npm package URLs recorded as unreachable, not silently dropped or treated as dead links.** Confirmed via a direct, out-of-script `curl` that npmjs.com's 403 is real (bot-detection), not a script bug — distinct from the `000` status caused by the actual script defect.

## Files Modified

- `.smaqit/tasks/097_lightweight_infra_task_orchestrator.md` — `Status: Abandoned` + full reasoning
- `.smaqit/tasks/095_feature_new_per_phase_worktree_spawn.md` — Findings populated, 9/10 ACs checked, `Status: Completed`
- `.smaqit/tasks/070_e2e_boundary_enforcement_validation.md` — `Status: Abandoned` + full reasoning
- `.smaqit/tasks/PLANNING.md` — 097 and 070 moved to Abandoned; 095 moved to Completed
- `skills/smaqit.new-greenfield-project/SKILL.md` — line 216 post-MVP pointer corrected to name `smaqit.feature-new`
- `.smaqit/references/project-research.md` — rebuilt; added MCP Go SDK, hujson, and the PlantUML/resvg/font npm packages; refreshed date to 2026-08-06

## Next Steps

- Remaining Active tasks: 099 (Opaque PlantUML PNG Rendering, High priority — appeared mid-session, not yet assessed by this agent), 094 (`feature-new` E2E browser gate), 077 (retroactive specs for brownfield), 074 (extensible-through-templates doc), 071 (Q&A agent + wiki skill)
- The `smaqit-extensions` `smaqit.task-start` description change (advertising its standalone lightweight path) remains an unfiled cross-repo follow-up, noted in task 097's abandonment reasoning but not tracked anywhere formally
- Consider reporting the `verify-urls.sh` column-count mismatch (3-column script vs. 4-column SKILL.md Step 3 description) as a small fix to `smaqit.project-research` itself

## Session Metrics

- **Tasks resolved:** 3 (1 Completed, 2 Abandoned)
- **Files modified:** 6
- **Research map:** rebuilt, 3 new dependencies tracked, 1 script bug found and worked around
