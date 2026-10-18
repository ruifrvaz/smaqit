---
name: smaqit.input-infrastructure
description: Validate and elicit infrastructure requirements before Infrastructure spec generation. Invoke automatically when beginning Infrastructure layer specification work to ensure deployment requirements are sufficient.
metadata:
  version: "1.2.0"
---

# Infrastructure Input

Validate that the user has stated sufficient deployment and environment requirements before generating Infrastructure specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract user-stated requirements** — Collect requirements only from the sources listed under Requirement Sources, and note the `mode` the caller passed (`assisted` when none was passed)
2. **Check existing specs** — Scan `specs/infrastructure/` for existing documents. If specs already exist and neither the user's own words nor `answers` say whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Determine if environment information is specific enough to produce deployment specs
4. **Check for conflicts** — If deployment requirements conflict with Stack or Functional specs, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from the Requirement Sources or from `answers` passed back on a later invocation, proceed directly to Infrastructure spec generation

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

- **`autonomous`:** the existing upstream specs (Business, Functional, Stack) also count as requirement sources; they passed their own gate.
- **`assisted`:** every required section must trace to the user's own statements. The upstream specs (Business, Functional, Stack) are context for cross-layer coherence only, never a source of new requirements for this layer.

A required section is never filled with an assumption. When the sources leave it insufficient, it is an open question for step 5.

## Required Sections

### Target Environment
**Question:** "Where will this system run? (local machine, cloud provider, on-premise server, container, serverless)"  
**Sufficient when:** The deployment target is specific enough to determine the infrastructure tier.

### Hosting Platform
**Question:** "What platform or infrastructure will host the system? If local execution only, state that explicitly."  
**Sufficient when:** The platform is named or 'local execution only' is confirmed.

## Optional Sections

### Service Topology
**Question:** "How is the system structured at the infrastructure level? (single process, microservices, monolith, scheduled jobs)"  
**Include when:** System structure has non-trivial infrastructure implications.

### Resource Constraints
**Question:** "What are the compute, memory, or storage limits?"  
**Include when:** Resource ceilings shape the infrastructure design.

### Scaling Requirements
**Question:** "How should the system handle load? (fixed capacity, auto-scale, peak traffic expectations)"  
**Include when:** Production system; omit for local or dev-only deployments.

### Geographic Constraints
**Question:** "Are there location or data residency requirements?"  
**Include when:** Regulatory or latency requirements restrict deployment geography.

### Budget Constraints
**Question:** "What are the infrastructure cost limits?"  
**Include when:** Budget directly constrains infrastructure choices.

### Integration Points
**Question:** "What existing systems does this need to connect to at the infrastructure level?"  
**Include when:** Existing infrastructure must be accommodated.

## Readiness Condition

Both required sections (Target Environment, Hosting Platform) have substantive answers. When satisfied, proceed to Infrastructure spec generation.

## NEEDS-INPUT Return

When step 5 cannot reach the user, end the invocation with this block and write nothing else:

```
NEEDS-INPUT
source: smaqit.input-infrastructure (smaqit.infrastructure agent)
layer: infrastructure
written: nothing
questions:
  - id: target-environment
    question: "<the targeted question for that section>"
    why: "<what is missing or contradictory in the current requirements>"
next: Ask the user these questions, then resume this agent (or invoke it again) with the answers as `answers`, keyed by id.
```

Question ids: the required-section slugs (`target-environment`, `hosting-platform`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as a user-stated requirement and run this gate again from step 1.
