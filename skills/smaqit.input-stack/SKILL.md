---
name: smaqit.input-stack
description: Validate and elicit technology requirements before Stack spec generation. Invoke automatically when beginning Stack layer specification work to ensure technology preferences are sufficient.
metadata:
  version: "1.0.0"
---

# Stack Input

Validate that session context contains sufficient technology preferences before generating Stack specifications. This is a gate — spec generation only proceeds when all required sections are satisfied.

## Steps

1. **Extract from session context** — Scan current session, compacted blocks, open tasks, Business specs, and Functional specs
2. **Assess each required section** — Determine if technology choices are specific enough to drive architecture decisions
3. **Elicit gaps** — Ask the targeted question for each insufficient section; collect one section at a time
4. **Confirm readiness** — Once all required sections are satisfied, proceed directly to Stack spec generation

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
