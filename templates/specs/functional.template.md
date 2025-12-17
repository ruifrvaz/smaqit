# [CONCEPT_NAME]

## References

<!-- References establish traceability and coherence, not requirement derivation -->
<!-- Use Implements for feature specs (1:1 mapping) -->
<!-- Use Enables for foundation specs (1:many mapping) -->

### Implements

- [BUS-CONCEPT](../business/[FILENAME].md) — [Business use case this is consistent with]

### Enables

<!-- For foundation specs that serve multiple business cases -->
- [BUS-CONCEPT-A](../business/[FILENAME].md) — [Business case this foundation must be consistent with]
- [BUS-CONCEPT-B](../business/[FILENAME].md) — [Business case this foundation must be consistent with]

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## User Flow

### Overview

[Brief description of the behavioral flow]

### Steps

1. [User/system action — what happens behaviorally]
2. [Next step]
3. [Continue until flow completes]

### Error Handling

| Condition | Behavior |
|-----------|----------|
| [Error condition] | [How the system responds] |

## Data Model

### [ENTITY_NAME]

| Attribute | Type | Description | Constraints |
|-----------|------|-------------|-------------|
| [ATTRIBUTE] | [Logical type — e.g., text, number, date] | [What it represents] | [Required, unique, etc.] |

### Relationships

| From | To | Type | Description |
|------|-----|------|-------------|
| [ENTITY_A] | [ENTITY_B] | [one-to-many, etc.] | [Nature of relationship] |

## API Contract

### [OPERATION_NAME]

**Purpose:** [What this operation accomplishes]

#### Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| [FIELD] | [Logical type] | [Yes/No] | [What it represents] |

#### Response

**Success:**

| Field | Type | Description |
|-------|------|-------------|
| [FIELD] | [Logical type] | [What it represents] |

**Errors:**

| Condition | Response | Description |
|-----------|----------|-------------|
| [Error condition] | [Error identifier] | [What went wrong] |

## State Transitions

*Include if the concept involves stateful behavior. Otherwise, state "Not applicable: [reason]"*

### States

| State | Description | Entry Condition |
|-------|-------------|-----------------|
| [STATE_NAME] | [What this state represents] | [How the system enters this state] |

### Transitions

```
[INITIAL_STATE] → [EVENT] → [NEXT_STATE]
```

| From | Event | To | Guard Condition |
|------|-------|-----|-----------------|
| [STATE_A] | [EVENT] | [STATE_B] | [Condition that must be true] |

## Acceptance Criteria

Requirements use format: `FUN-[CONCEPT]-[NNN]`

- [ ] FUN-[CONCEPT]-001: [Criterion — must be measurable, observable, unambiguous]
- [ ] FUN-[CONCEPT]-002: [Criterion]

### Untestable Criteria

If any criterion cannot be automatically validated, flag it:

- [ ] FUN-[CONCEPT]-NNN: [Criterion] *(untestable)*
  - **Reason:** [Why it cannot be tested automatically]
  - **Proposal:** [Measurable proxy or alternative approach]
  - **Resolution:** [How it will be verified]