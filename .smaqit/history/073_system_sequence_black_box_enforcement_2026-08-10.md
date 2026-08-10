# System Sequence Black-Box Enforcement

**Date:** 2026-08-10
**Session Focus:** Review and harden task 104's structural validation for `system-sequence` design diagrams — from a code review of a prior interrupted session's uncommitted implementation, through a redesigned and simpler enforcement rule, to full task completion.
**Tasks Completed:** 104
**Tasks Referenced:** 102 (owned by a separate, concurrent session; not touched)

---

## Actions Taken

### Review of the Prior Implementation

- Confirmed a prior interrupted session had already implemented task 104 (an arrow-parsing heuristic in `validateSystemSequenceProfile`) uncommitted in the task's own worktree, with `go build`/`go test` passing there.
- Entered the registered task worktree and ran a forked code-review pass that executed the validator directly against crafted PlantUML input. Found and reported 7 confirmed correctness bugs, all triggered by ordinary, non-adversarial PlantUML syntax: decorated arrows (`-[hidden]->`) silently bypassing detection, a participant named `NoteService` silently bypassing detection, a decorated `as` alias resolving to the wrong identifier, PlantUML's `create <type> <name>` shorthand extracting the type keyword instead of the name, unhandled multi-line `title`/`legend` bodies being parsed as diagram content, a dead divider-skip regex, and order-dependent actor/system classification.

### Redesign

- User proposed a stricter, simpler rule: require exactly one actor and exactly one system-side participant per `system-sequence` diagram, with the system participant's identifier fixed to literally `System`. Assessed the proposal (via `smaqit.session-assess`) before implementing — confirmed the shipped `functional.template.md` reference already used exactly this shape, and that the change collapses the validator to pure declaration counting, eliminating the arrow-parsing surface that produced all 7 bugs above.
- Rewrote `validateSystemSequenceProfile` in `installer/design.go` as a declaration-only scan: exactly one `actor` declaration, exactly one system-type declaration (`participant`/`boundary`/`control`/`entity`/`database`/`collections`/`queue`), and that declaration's identifier must be exactly `System`. No arrow/message-line parsing remains.
- Rewrote `installer/design_test.go`'s coverage for the new rule set (actor cardinality, system-participant cardinality, `System` naming, decorated-alias resolution, multi-line title-block skipping), and confirmed the shipped template and multi-design-per-spec fixtures still pass unmodified.
- Updated `framework/LAYERS.md` and `agents/functional.md` guidance to require the one-actor/one-`System` shape and to make per-actor/per-flow diagram splitting mandatory rather than a readability suggestion.
- Updated task 104's own Design Decisions, Implementation Steps, and Acceptance Criteria to record the superseded "any actor count is fine" decision and the new rule.

### Task Completion

- Ran `smaqit.task-complete` for task 104. The lifecycle resolver initially failed: the earlier interrupted session's `task-start` had never committed the "In Progress" status/mode change on the primary checkout, leaving it visible only as an uncommitted change in the worktree. Repaired by committing the missing status/mode/`PLANNING.md` state on primary before retrying.
- Verified all 7 acceptance criteria against actual test evidence, checked them off, and wrote Findings in the task file.
- Committed the implementation in the task worktree and merged the branch into `main` (`--no-ff`). Hit one incidental merge conflict in `PLANNING.md`, caused by a concurrent session's unrelated update to task 102's row; resolved by keeping that session's version untouched.
- Verified the merged primary checkout builds and tests cleanly via `make -C installer test` (which also regenerated a stale, gitignored `installer/templates/designs/` build-staging directory — pre-existing drift unrelated to this task).
- Completed cleanup: status set to Completed, `PLANNING.md` entry moved to the Completed section, task worktree removed, merged branch deleted, workspace file rebuilt (task 102's still-active worktree entry preserved throughout).

## Problems Solved

- Closed the actual motivating gap: a `system-sequence` design with more than one system-side participant — the exact shape that produced two structurally different diagrams for the same spec in a downstream project's pilot — now fails validation deterministically with a remedy-stating error.
- Eliminated an entire class of validator bugs by redesigning around declaration-only scanning instead of incrementally patching an arrow-parsing heuristic.
- Repaired a task-lifecycle data inconsistency (status committed only in a worktree, never on primary) that would otherwise have silently blocked any future `task-complete` attempt on this task.

## Decisions Made

- **The system-side participant's identifier must be exactly `System` (case-sensitive), and exactly one actor is required per diagram.** Deliberately supersedes the task's original "any actor count is fine" decision; chosen because it also reduces the validator's implementation to pure declaration counting with zero arrow-token parsing, and because the convention was already the shipped template's default.
- **Declaration-only detection does not see participants PlantUML would auto-create from a never-declared arrow reference.** Documented as a known, accepted heuristic limitation rather than closed — consistent with the task's existing "lint check, not adversarial-proof" stance, and not expected to be the honest/default-authoring case since the shipped template always declares both participants explicitly.

## Files Modified

- `installer/design.go` — `validateSystemSequenceProfile` redesigned as a declaration-only actor/system-participant scan.
- `installer/design_test.go` — regression coverage rewritten for the one-actor/one-`System` rule.
- `framework/LAYERS.md`, `agents/functional.md` — Functional-layer `system-sequence` authoring guidance updated.
- `.smaqit/tasks/104_strict_black_box_system_sequence_validation.md`, `.smaqit/tasks/PLANNING.md` — Design Decisions/Acceptance Criteria updated, Findings recorded, status moved to Completed.
- `smaqit.code-workspace` — task 104's worktree entry removed after cleanup; task 102's active entry preserved.

## Next Steps

- None beyond the two documented, accepted heuristic limitations (actor mislabeling, undeclared arrow-implicit participants).
- Task 102 remains `In Progress` in a separate, concurrent session/worktree — outside this session's scope.

## Session Metrics

- **Tasks completed:** 1 (104)
- **Confirmed validator bugs found and closed by redesign:** 7
- **Test verification:** full `go test ./...` and `make -C installer test` both passing clean after merge, on both the task worktree and the primary checkout
- **Commits on primary:** 5 (lifecycle repair, merge, completion, workspace refresh, plus the squashed-in worktree implementation commit)
