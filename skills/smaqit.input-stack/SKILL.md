---
name: smaqit.input-stack
description: Validate and elicit technology requirements before Stack spec generation. Invoke automatically when beginning Stack layer specification work to ensure technology preferences are sufficient.
metadata:
  version: "1.1.0"
---

# Stack Input

Validate that session context contains sufficient technology preferences before generating Stack specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, Business specs, and Functional specs
2. **Check existing specs** — Scan `specs/stack/` for existing documents. If specs already exist and neither session context nor `answers` says whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Determine if technology choices are specific enough to drive architecture decisions
4. **Check for conflicts** — If technology choices conflict with constraints in Business or Functional specs, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from session context or from `answers` passed back on a later invocation, proceed directly to Stack spec generation

## Required Sections

### Technology Preferences
**Question:** "What languages, frameworks, or tools do you want to use? If you have no preference, describe any constraints that would narrow the choice — or say 'agent's recommendation' to defer."  
**Sufficient when:** At least one technology direction is indicated, or explicit openness to agent recommendation is stated.

### Constraints
**Question:** "What are the technology limitations? (target platforms, team expertise, licensing, compatibility requirements) State 'none' if there are no constraints."  
**Sufficient when:** Known hard constraints are listed, or 'none' is explicitly confirmed.

## Optional Sections

### Build Tools
**Question:** "What build system or tooling is needed? (e.g., make, gradle, cargo, npm scripts)"  
**Include when:** Build system is non-obvious from the technology choice.

### Development Environment
**Question:** "What development setup is required? (runtime versions, IDE, local toolchain)"  
**Include when:** Environment requirements affect the spec.

### Dependencies
**Question:** "What external libraries or packages are needed?"  
**Include when:** Specific dependencies are pre-decided.

### Rationale
**Question:** "Why these technology choices? (team familiarity, performance, ecosystem, regulatory)"  
**Include when:** Rationale is non-obvious and useful for future decision-making.

## Readiness Condition

Both required sections (Technology Preferences, Constraints) have substantive answers. When satisfied, proceed to Stack spec generation.

## NEEDS-INPUT Return

When step 5 cannot reach the user, end the invocation with this block and write nothing else:

```
NEEDS-INPUT
source: smaqit.input-stack (smaqit.stack agent)
layer: stack
written: nothing
questions:
  - id: technology-preferences
    question: "<the targeted question for that section>"
    why: "<what is missing or contradictory in the current requirements>"
next: Ask the user these questions, then resume this agent (or invoke it again) with the answers as `answers`, keyed by id.
```

Question ids: the required-section slugs (`technology-preferences`, `constraints`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as session-context requirements and run this gate again from step 1.
