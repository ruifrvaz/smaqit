---
name: smaqit.input-functional
description: Validate and elicit functional requirements before Functional spec generation. Invoke automatically when beginning Functional layer specification work to ensure requirements are sufficient.
metadata:
  version: "1.2.0"
---

# Functional Input

Validate that the user has stated sufficient functional requirements before generating Functional specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract user-stated requirements** — Collect requirements only from the sources listed under Requirement Sources, and note the `mode` the caller passed (`assisted` when none was passed)
2. **Check existing specs** — Scan `specs/functional/` for existing documents. If specs already exist and neither the user's own words nor `answers` say whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Determine if content is specific enough to produce testable behavior descriptions
4. **Check for conflicts** — If requirements contradict Business specs or contain internal conflicts, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from the Requirement Sources or from `answers` passed back on a later invocation, proceed directly to Functional spec generation

## Requirement Sources

A required section is satisfied only by requirements the user stated:

- `user_request`: the user's own words, verbatim, as the caller passed them
- `answers`: the user's replies to an earlier `NEEDS-INPUT` from this gate
- The user's own messages and the open tasks they wrote, when this gate reads the conversation directly rather than through a caller's retelling
- Files the user pointed at, including `.smaqit/requirements-extract.md`, which is extracted from the user's own assets

Never sufficient, in any mode:

- A project, repository, or directory name, or a document title. "mario-app" names a project; it states no requirement
- Your own knowledge of the domain, or of anything a name evokes
- A caller's summary, paraphrase, or expansion of the request. When it differs from `user_request`, `user_request` wins

Upstream specifications depend on `mode`:

- **`autonomous`:** the existing Business specs also count as requirement sources; they passed their own gate.
- **`assisted`:** every required section must trace to the user's own statements. The Business specs are context for cross-layer coherence only, never a source of new requirements for this layer.

A required section is never filled with an assumption. When the sources leave it insufficient, it is an open question for step 5.

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

## NEEDS-INPUT Return

When step 5 cannot reach the user, end the invocation with this block and write nothing else:

```
NEEDS-INPUT
source: smaqit.input-functional (smaqit.functional agent)
layer: functional
written: nothing
questions:
  - id: user-experience
    question: "<the targeted question for that section>"
    why: "<what is missing or contradictory in the current requirements>"
next: Ask the user these questions, then resume this agent (or invoke it again) with the answers as `answers`, keyed by id.
```

Question ids: the required-section slugs (`user-experience`, `behaviors`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as a user-stated requirement and run this gate again from step 1.
