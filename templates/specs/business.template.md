---
id: BUS-[CONCEPT]
status: draft
created: [TIMESTAMP]
---

# UC[N]-[CONCEPT]: [USE_CASE_NAME]

<!-- Use Case Identifier Format: UC[N]-[CONCEPT]
     - UC: Use Case prefix
     - [N]: Sequential number (UC1, UC2, UC3, ...)
     - [CONCEPT]: Short uppercase descriptor matching acceptance criteria concept
     - Example: UC1-LOGIN, UC2-CHECKOUT, UC3-PROFILE
     - Use Case IDs are stable identifiers - never reuse after deletion -->

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## Actors

<!-- Anyone who cares about some aspect of what is being built -->
<!-- Actor goals may express: interactive outcomes, system properties, or success criteria -->
<!-- Examples: end users, operations teams, compliance officers, client organizations, accessibility advocates -->

| Actor | Description | Goals |
|-------|-------------|-------|
| [ACTOR_NAME] | [Who or what this actor represents] | [What this actor wants to achieve or what properties they require] |

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

---

*Generated with smaqit [SMAQIT_VERSION]*
