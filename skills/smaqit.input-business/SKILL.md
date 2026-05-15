---
name: smaqit.input-business
description: Validate and elicit business requirements before Business spec generation. Invoke automatically when beginning Business layer specification work to ensure requirements are sufficient.
metadata:
  version: "1.0.0"
---

# Business Input

Validate that session context contains sufficient business requirements before generating Business specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, and open tasks for existing requirements
2. **Check existing specs** — Scan `specs/business/` for existing documents. If specs already exist, confirm whether the intent is to add new specs or revise existing ones before proceeding.
3. **Assess requirements** — Check whether content is substantive: not generic, not a placeholder, actionable enough to produce testable acceptance criteria
4. **Check for conflicts** — If requirements contain internal contradictions or conflict with each other, flag them explicitly before proceeding; do not silently resolve.
5. **Elicit gaps** — For each insufficient section, ask the targeted question below; collect one section at a time, not all at once
6. **Confirm readiness** — Once all required sections are satisfied, proceed directly to Business spec generation without requesting further confirmation

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
