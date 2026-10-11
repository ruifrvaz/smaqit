---
name: smaqit.input-coverage
description: Validate and elicit verification requirements before Coverage spec generation. Invoke automatically when beginning Coverage layer specification work to ensure testing requirements are sufficient.
metadata:
  version: "1.1.0"
---

# Coverage Input

Validate that session context contains sufficient verification requirements before generating Coverage specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, and all upstream specs
2. **Check existing specs** — Scan `specs/coverage/` for existing documents. If specs already exist and neither session context nor `answers` says whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Determine if testing requirements are specific enough to produce coverage specs
4. **Check for conflicts** — If verification requirements conflict with or are untrackable against upstream acceptance criteria, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from session context or from `answers` passed back on a later invocation, proceed directly to Coverage spec generation

## Required Sections

### Test Scope
**Question:** "What types of testing are needed for this system? (unit, integration, end-to-end, performance, manual acceptance)"  
**Sufficient when:** At least one testing type is specified, or the system's risk profile makes the appropriate scope evident from upstream specs.

### Acceptance Thresholds
**Question:** "What defines acceptable test results? What must pass before the system is considered verified?"  
**Sufficient when:** At least one measurable threshold is defined (e.g., 'all acceptance criteria must pass', '100% of integration tests green').

## Optional Sections

### Test Environment
**Question:** "Where and how should tests run? (CI platform, local only, test containers, specific OS)"  
**Include when:** Test execution environment is constrained or non-obvious from Infrastructure specs.

### Integration Points
**Question:** "What external systems or interfaces need dedicated test coverage?"  
**Include when:** The system integrates with external services beyond what upstream specs make clear.

## Readiness Condition

Both required sections (Test Scope, Acceptance Thresholds) have substantive answers. When satisfied, proceed to Coverage spec generation.

## NEEDS-INPUT Return

When step 5 cannot reach the user, end the invocation with this block and write nothing else:

```
NEEDS-INPUT
source: smaqit.input-coverage (smaqit.coverage agent)
layer: coverage
written: nothing
questions:
  - id: test-scope
    question: "<the targeted question for that section>"
    why: "<what is missing or contradictory in the current requirements>"
next: Ask the user these questions, then resume this agent (or invoke it again) with the answers as `answers`, keyed by id.
```

Question ids: the required-section slugs (`test-scope`, `acceptance-thresholds`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as session-context requirements and run this gate again from step 1.
