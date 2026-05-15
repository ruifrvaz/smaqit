---
name: smaqit.input-coverage
description: Validate and elicit verification requirements before Coverage spec generation. Invoke automatically when beginning Coverage layer specification work to ensure testing requirements are sufficient.
metadata:
  version: "1.0.0"
---

# Coverage Input

Validate that session context contains sufficient verification requirements before generating Coverage specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, and all upstream specs
2. **Check existing specs** — Scan `specs/coverage/` for existing documents. If specs already exist, confirm whether the intent is to add new specs or revise existing ones before proceeding.
3. **Assess requirements** — Determine if testing requirements are specific enough to produce coverage specs
4. **Check for conflicts** — If verification requirements conflict with or are untrackable against upstream acceptance criteria, flag them explicitly before proceeding; do not silently resolve.
5. **Elicit gaps** — Ask the targeted question for each insufficient section; collect one section at a time
6. **Confirm readiness** — Once all required sections are satisfied, proceed directly to Coverage spec generation

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
