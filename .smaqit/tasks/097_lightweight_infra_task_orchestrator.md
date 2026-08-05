# Lightweight Task-Lifecycle Entry Point for Infrastructure-Only Work

**Status:** Abandoned
**Created:** 2026-07-30

**Abandoned:** 2026-08-05
**Reason:** Reviewed cold five days after filing with no clear sense of what shipping it would actually change. Investigation confirmed the underlying discoverability gap is real (nothing tells an operator when standalone `task.start` is appropriate vs. `smaqit.feature-new`; `smaqit.new-greenfield-project`'s post-MVP pointer at SKILL.md:216 is stale and predates `feature-new`'s existence), but the task's own premise was wrong — the task-lifecycle skills (`task-start`, `task-create`, etc.) are not owned by this repo at all; they live in the sibling `smaqit-extensions` repo, and only ship there. That leaves the smaqit-side fix as two small, low-risk documentation edits (a stop-condition in `feature-new`'s preamble, a corrected pointer in `new-greenfield-project`) with no mechanics, no new skill, and no cross-cutting design decision — not enough substance to justify tracked task overhead. If picked up again, do it as a direct doc fix rather than through the task lifecycle; the `smaqit.task-start` description change belongs in `smaqit-extensions`, not here.

## Description

`smaqit.feature-new` already covers infrastructure-only post-MVP work correctly at the mechanics level — its Phase 3 resolves `provisioning_mode` (including `existing-unmanaged`, task 096) and its Phase 0 creates a dedicated worktree via `smaqit.task-start` before anything else happens. The problem isn't missing capability; it's discoverability and naming. On a downstream project, an operator picked up a pure infrastructure task (add a second deployment target, no Business/Functional/Stack spec changes involved) and executed it entirely ad hoc — no `task-start`, no worktree, direct edits and `git`/`gh` commands in the shared primary checkout — because neither "`smaqit.feature-new`" (framed around *feature* work) nor "run `task-start` standalone for small things" (a fact that has to be remembered, not something the tooling surfaces) read as the obvious next step for "this is just infrastructure work."

The consequence was concrete, not hypothetical: mid-task, a `git reset --hard` run in that same shared primary checkout (to get a clean base for a later commit) discarded uncommitted bookkeeping belonging to a *different*, concurrent task's session — another operator's `PLANNING.md` row and VS Code workspace-folder entry, both of which existed only as uncommitted state in that shared checkout. No real work was lost (the concurrent task's actual files lived safely in its own worktree), but the incident is a direct, traceable consequence of not having a dedicated worktree in the first place — exactly what `task-start` exists to provide, and exactly what would never have been at risk had it been invoked.

## Design Decisions

TBD — to be resolved by whoever picks this up. Open questions:

- **Shape of the fix:** a genuinely new skill (`smaqit.deploy-new` or similar), versus a documentation-only fix that makes `smaqit.task-start`'s standalone path more prominent for infra-only work, versus a very thin wrapper skill whose entire body is "call `task-start`, do the work, call `task-complete`" with no mechanics of its own (no `provisioning_mode` handling, no credential logic — all of that stays exactly where task 096 put it, in the existing infra skill family). The investigation that led here explicitly rejected building a mechanics-duplicating skill once already (the `existing-unmanaged` provisioning-mode work) — this task must not re-litigate that and land in the same place under a different name. If a new skill is the answer, its only job is orchestration/naming, not mechanics.
- **Trigger condition:** what distinguishes "infra-only, use the lightweight path" from "touches Business/Functional/Stack too, use `smaqit.feature-new`"? Likely: no spec changes outside `specs/infrastructure/`, matching how task 096's own downstream consumer task had zero Business/Functional/Stack touches. Worth stating as an explicit, checkable condition rather than left to judgment.
- **Does this even need a new command**, or is the real fix just making `smaqit.task-start`'s already-existing standalone (non-`feature-new`) path more discoverable — e.g. via a description/summary change to `smaqit.task-start` itself, or a mention in `smaqit.new-greenfield-project`'s own "post-MVP" pointer — rather than adding another named entry point to remember?

## Implementation Steps

TBD — expected to touch `skills/smaqit.task-start/` (or wherever its description lives) and possibly a new skill file, exact shape depends on Design Decisions above.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] The chosen fix does not duplicate any `provisioning_mode` or credential mechanics already owned by the infra skill family (task 084/096) — orchestration/discoverability only
- [ ] A worked example: an infra-only task (no Business/Functional/Stack spec changes) run end to end using the fixed path produces a dedicated worktree before any file is touched, exactly as `smaqit.feature-new`/`task-start` already do for the cases that do invoke them
- [ ] Explicit, checkable trigger condition documented for when to use the lightweight path vs. `smaqit.feature-new`

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| TBD — depends on Design Decisions | TBD |

## Notes

Filed retroactively after a real incident on a downstream project: an infra-only task was executed entirely outside the task-lifecycle tooling (no `task-start`, no worktree), and a later `git reset --hard` in the shared primary checkout discarded another concurrent task's uncommitted bookkeeping as a direct result. No data was actually lost (the concurrent task's real work lived in its own worktree), but the near-miss is the clearest possible demonstration of why worktree isolation exists. Related: task 096 (`existing-unmanaged` provisioning mode) already closed the *mechanics* gap this same incident's originating task needed — this task is specifically about the *orchestration/discoverability* gap that let the mechanics get invoked without the lifecycle wrapper around them.
