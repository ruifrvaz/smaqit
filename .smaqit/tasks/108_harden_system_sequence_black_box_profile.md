# Harden System-Sequence Black-Box Profile

**Status:** In Progress
**Created:** 2026-08-14
**Mode:** Assisted
**Started:** 2026-08-14

## Description

Make the Functional `system-sequence` design profile fully deterministic. It must render one
actor and one visible black-box participant named only `System` (case-insensitive), without
repeated participant footboxes or inferred participants from undeclared message endpoints.

This is a strict forward-only policy. Legacy diagrams are not retained for compatibility and no
migration behavior is required.

## Design Decisions

- **Visible system identity:** The declared system participant's display label and alias must both
  be `System`, case-insensitively.
- **No repeated footboxes:** Every system-sequence source must include PlantUML's `hide footbox`.
  Literal PlantUML `footer` directives are forbidden.
- **Explicit-only participants:** Message endpoints must resolve only to the one declared actor or
  declared `System`; undeclared endpoints fail rather than letting PlantUML infer participants.
- **No backwards compatibility:** Existing designs that violate this profile are invalid; the
  framework will not grandfather or transform them.

## Implementation Steps

1. Extend the system-sequence source validator to parse visible declaration labels, aliases,
   footbox/footer directives, and arrow endpoints.
2. Reject labels or aliases other than `System` (case-insensitive), a missing `hide footbox`,
   literal footer directives, multiple actors, extra participants, and undeclared message endpoints.
3. Update the Functional design template and agent instructions to emit the strict profile.
4. Add regression coverage for every new failure and for the canonical template.
5. Regenerate installer payloads and run the relevant Go tests.

## Known Issues Triage

**Triaged:** 2026-08-14
**Tools searched:** PlantUML
**Result:** Clear

### Blocking Issues
- None.

### Advisory Issues
- None.

### Historical (Closed)
- None.

### Unresolvable Tools
- None.

## Acceptance Criteria

- [ ] A system-sequence passes only when it has one explicit actor and one explicitly declared,
  visibly named `System` participant, both with a `System` alias where applicable.
- [ ] A system-sequence lacking `hide footbox`, containing a literal footer directive, declaring
  additional actors or participants, or referencing an undeclared arrow endpoint fails with a
  deterministic design-validation error.
- [ ] The canonical Functional template emits `participant "System" as System` and `hide footbox`.
- [ ] Functional-agent instructions describe the strict visible-System and no-footbox contract.
- [ ] Regression tests cover the new profile and the installer test suite passes.

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
| `installer/design.go` | Modify |
| `installer/design_test.go` | Modify |
| `templates/designs/functional.template.md` | Modify |
| `agents/functional.md` | Modify |

## Notes

The task may require updating any additional canonical documentation that still permits a
descriptive visible system label. No downstream design artifacts are in scope.
