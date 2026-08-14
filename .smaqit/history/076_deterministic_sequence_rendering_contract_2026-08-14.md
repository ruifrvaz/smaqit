# Deterministic Sequence Rendering Contract

**Date:** 2026-08-14
**Previous Session:** [Global Agent Installation Release](075_global_agent_installation_release_2026-08-11.md)
**Session focus:** Verify current PlantUML Functional-design behavior, then make the system-sequence black-box profile deterministic.
**Tasks completed/referenced:** 108 completed; 107 remains in progress and was explicitly skipped for session-close gating.

## Actions Taken

- Verified the installed framework and current release context, then inspected downstream Settings design artifacts to distinguish legacy/nonconforming diagrams from current source behavior.
- Diagnosed that the prior system-sequence validator accepted descriptive System labels and did not prevent inferred message endpoints; existing use-case layout also had no deterministic orientation gate.
- Planned, created, started, implemented, tested, merged, and completed the strict Functional system-sequence black-box task.
- Refreshed the project research map and updated the project compendium with the current system-sequence contract.

## Problems Solved

- Functional diagrams could show a descriptive system-side label such as `Settings Area` while still using the `System` alias.
- PlantUML could render repeated participant footboxes and infer undeclared participants from arrow endpoints.
- The framework lacked deterministic source gates for literal footers, footbox omission, multiple actors, and undeclared endpoints.

## Decisions Made

- A Functional system-sequence now requires exactly one actor and `participant "System" as System`, matching `System` case-insensitively for both label and alias.
- `hide footbox` is mandatory and literal `footer` directives are forbidden.
- Every recognized message endpoint must resolve to the declared actor or System; no backward compatibility or migration behavior is retained for legacy designs.
- A visual clean-room smoke test is required alongside the Go regression suite for this PlantUML rendering contract.

## Files Modified

- `installer/design.go` — enforces the deterministic system-sequence profile.
- `installer/design_test.go` — adds canonical and rejection regression coverage.
- `templates/designs/functional.template.md` — emits the canonical System participant and hides footboxes.
- `agents/functional.md` and `framework/LAYERS.md` — document the authoring contract.
- `.smaqit/tasks/108_harden_system_sequence_black_box_profile.md` and `.smaqit/tasks/PLANNING.md` — record task completion.
- `smaqit.code-workspace` — removes the completed task worktree.
- `.smaqit/references/project-research.md` and `.smaqit/compendium.md` — refresh durable project guidance.

## Next Steps

- Continue or complete Task 107 when its assisted-mode implementation has been reviewed.
- Release the merged Task 108 changes before expecting existing global installations to enforce the new profile.

## Session Metrics

- Tasks completed: 1.
- Source files changed for the feature: 5.
- Verification: `go test ./...`, `go vet ./...`, installer build, clean-room initialization, PNG visual attestation, and four negative render-gate cases.
