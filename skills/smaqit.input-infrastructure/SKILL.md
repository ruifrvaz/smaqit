---
name: smaqit.input-infrastructure
description: Validate and elicit infrastructure requirements before Infrastructure spec generation. Invoke automatically when beginning Infrastructure layer specification work to ensure deployment requirements are sufficient.
metadata:
  version: "1.1.0"
---

# Infrastructure Input

Validate that session context contains sufficient deployment and environment requirements before generating Infrastructure specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, and all upstream specs
2. **Check existing specs** — Scan `specs/infrastructure/` for existing documents. If specs already exist and neither session context nor `answers` says whether to add new specs or revise existing ones, record that as an open `add-or-revise` question for step 5.
3. **Assess requirements** — Determine if environment information is specific enough to produce deployment specs
4. **Check for conflicts** — If deployment requirements conflict with Stack or Functional specs, flag them explicitly before proceeding; do not silently resolve. Each unresolved conflict becomes an open question for step 5.
5. **Resolve open questions** — Gather every insufficient required section (with its targeted question below), every unresolved conflict from step 4, and any open question from step 2. If you can ask the user directly, ask one section at a time. Otherwise stop before generating anything: write no specification or design, and return `NEEDS-INPUT` (see below) listing every open question at once.
6. **Confirm readiness** — Once all required sections are satisfied, from session context or from `answers` passed back on a later invocation, proceed directly to Infrastructure spec generation

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

Question ids: the required-section slugs (`target-environment`, `hosting-platform`), an optional-section slug when its omission would break the spec, `add-or-revise`, and `conflict-1`, `conflict-2`, … for conflicts. On the next invocation, treat each answer as session-context requirements and run this gate again from step 1.
