# [CONCEPT_NAME]

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## Actors

<!-- Human actors, external systems, or the System itself -->
<!-- Use "System" actor for stakeholder requirements about system properties (uptime, auditability, etc.) -->

| Actor | Description | Goals |
|-------|-------------|-------|
| [ACTOR_NAME] | [Who or what this actor represents] | [What this actor wants to achieve] |
| System | [Optional — use for system-level stakeholder requirements] | [System properties like availability, security] |

## Success Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| [METRIC_NAME] | [Quantifiable target] | [How it will be measured] |

## Use Case

### Preconditions

- [What must be true before this use case can begin]

### Main Flow

1. [Step in business terms — what happens, not how]
2. [Next step]
3. [Continue until postconditions are met]

### Alternative Flows

#### [ALTERNATIVE_NAME]

**Trigger:** [Condition that causes this alternative]

1. [Alternative step]
2. [Rejoin main flow at step N, or end differently]

### Postconditions

- [What must be true after successful completion]

## Acceptance Criteria

Requirements use format: `BUS-[CONCEPT]-[NNN]`

- [ ] BUS-[CONCEPT]-001: [Criterion — must be measurable, observable, unambiguous]
- [ ] BUS-[CONCEPT]-002: [Criterion]

### Untestable Criteria

If any criterion cannot be automatically validated, flag it:

- [ ] BUS-[CONCEPT]-NNN: [Criterion] *(untestable)*
  - **Reason:** [Why it cannot be tested automatically]
  - **Proposal:** [Measurable proxy or alternative approach]
  - **Resolution:** [How it will be verified — e.g., manual review]
