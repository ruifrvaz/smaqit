---
name: smaqit.input-business
description: Validate and elicit business requirements before Business spec generation. Invoke automatically when beginning Business layer specification work to ensure requirements are sufficient.
metadata:
  version: "1.1.0"
---

# Business Input

Validate that session context contains sufficient business requirements before generating Business specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, and open tasks for existing requirements
2. **Check existing specs** — Scan `specs/business/` for existing documents. If specs already exist and neither session context nor `answers` says whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Check whether content is substantive: not generic, not a placeholder, actionable enough to produce testable acceptance criteria
4. **Check for conflicts** — If requirements contain internal contradictions or conflict with each other, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from session context or from `answers` passed back on a later invocation, proceed directly to Business spec generation without requesting further confirmation

## Required Sections

### Actors
**Question:** "Who interacts with this system? List each actor with a brief description of their role."  
**Sufficient when:** At least one actor is named with enough context to scope their interactions.

### Use Cases
**Question:** "What do users want to accomplish? Describe each distinct use case — 'As [actor], I want to [action] so that [outcome]' or plain language equivalent."  
**Sufficient when:** At least one use case is described with a clear actor and goal.

### Success Metrics
**Question:** "How will you know the system is working correctly? List at least one measurable outcome."  
**Sufficient when:** At least one criterion is expressed in a form that can be objectively tested.

## Optional Sections

### Business Goals
**Question:** "What broader objectives does this system support?"  
**Include when:** The use cases serve a strategic purpose that should shape the spec scope.

### Constraints
**Question:** "Are there any business-level limitations? (budget, regulatory, brand, content restrictions)"  
**Include when:** Omission would lead to specs that violate known boundaries.

## Readiness Condition

All three required sections (Actors, Use Cases, Success Metrics) have substantive answers. When satisfied, proceed to Business spec generation.

## NEEDS-INPUT Return

When step 5 cannot reach the user, end the invocation with this block and write nothing else:

```
NEEDS-INPUT
source: smaqit.input-business (smaqit.business agent)
layer: business
written: nothing
questions:
  - id: actors
    question: "<the targeted question for that section>"
    why: "<what is missing or contradictory in the current requirements>"
next: Ask the user these questions, then resume this agent (or invoke it again) with the answers as `answers`, keyed by id.
```

Question ids: the required-section slugs (`actors`, `use-cases`, `success-metrics`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as session-context requirements and run this gate again from step 1.
