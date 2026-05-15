---
name: smaqit.input-infrastructure
description: Validate and elicit infrastructure requirements before Infrastructure spec generation. Invoke automatically when beginning Infrastructure layer specification work to ensure deployment requirements are sufficient.
metadata:
  version: "1.0.0"
---

# Infrastructure Input

Validate that session context contains sufficient deployment and environment requirements before generating Infrastructure specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, and all Phase 1 specs
2. **Assess each required section** — Determine if environment information is specific enough to produce deployment specs
3. **Elicit gaps** — Ask the targeted question for each insufficient section; collect one section at a time
4. **Confirm readiness** — Once all required sections are satisfied, proceed directly to Infrastructure spec generation

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
