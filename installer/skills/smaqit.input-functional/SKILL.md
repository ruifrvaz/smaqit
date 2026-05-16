---
name: smaqit.input-functional
description: Validate and elicit functional requirements before Functional spec generation. Invoke automatically when beginning Functional layer specification work to ensure requirements are sufficient.
metadata:
  version: "1.0.0"
---

# Functional Input

Validate that session context contains sufficient functional requirements before generating Functional specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, and Business specs
2. **Check existing specs** — Scan `specs/functional/` for existing documents. If specs already exist, confirm whether the intent is to add new specs or revise existing ones before proceeding.
3. **Assess requirements** — Determine if content is specific enough to produce testable behavior descriptions
4. **Check for conflicts** — If requirements contradict Business specs or contain internal conflicts, flag them explicitly before proceeding; do not silently resolve.
5. **Elicit gaps** — Ask the targeted question for each insufficient section; collect one section at a time
6. **Confirm readiness** — Once all required sections are satisfied, proceed directly to Functional spec generation

## Required Sections

### User Experience
**Question:** "What experience should users have when interacting with this system? Describe the expected flow from the user's perspective."  
**Sufficient when:** The expected user-facing behavior is described well enough to derive functional requirements.

### Behaviors
**Question:** "What should the system do? Describe each distinct system action or response."  
**Sufficient when:** At least one system behavior is described with enough precision to define acceptance criteria.

## Optional Sections

### Interactions
**Question:** "How do users or external systems interact? Describe any input/output exchanges."  
**Include when:** The system accepts user input or calls external interfaces.

### Data Models
**Question:** "What data does the system work with? Name each entity and its key attributes."  
**Include when:** Data structure affects functional behavior.

### State Transitions
**Question:** "How does the system change state over time? List the key states and what triggers each transition."  
**Include when:** System has non-trivial lifecycle (not just request/response).

### API Contracts
**Question:** "What interfaces exist between components? Describe each with its inputs and outputs."  
**Include when:** Component boundaries need to be explicitly defined.

## Readiness Condition

Both required sections (User Experience, Behaviors) have substantive answers. When satisfied, proceed to Functional spec generation.
