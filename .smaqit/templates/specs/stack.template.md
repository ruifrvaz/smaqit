---
id: STK-[CONCEPT]
status: draft
created: [TIMESTAMP]
---

# [CONCEPT_NAME]

## References

<!-- References establish traceability and coherence, not requirement derivation -->
<!-- Use Implements for feature specs (1:1 mapping) -->
<!-- Use Enables for foundation specs (1:many mapping) -->
<!-- Use same-layer references when feature specs extend foundation specs -->

### Foundation Reference

<!-- Same-layer reference: use when this feature spec extends a foundation spec -->
<!-- Omit this section if this spec doesn't depend on a foundation spec in the same layer -->

- [STK-[FOUNDATION-CONCEPT]](./[FOUNDATION-FILENAME].md) — [Shared requirements referenced here]

### Implements

- [FUN-CONCEPT](../functional/[FILENAME].md) — [How this stack implements the functional behavior]

### Enables

- [FUN-CONCEPT-A](../functional/[FILENAME].md) — [How this stack enables the functional behavior]
- [FUN-CONCEPT-B](../functional/[FILENAME].md) — [How this stack enables the functional behavior]

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## Technology Stack

### Languages

| Language | Version |
|----------|---------|
| [LANGUAGE] | [VERSION] |

### Frameworks

| Framework | Version |
|-----------|---------|
| [FRAMEWORK] | [VERSION] |

### Libraries

| Library | Version | Purpose |
|---------|---------|---------|
| [LIBRARY] | [VERSION] | [What this library provides] |

### Build Tools

| Tool | Version | Purpose |
|------|---------|---------|
| [TOOL] | [VERSION] | [Build, test, lint, etc.] |

## Constraints

<!-- Team expertise, organizational standards, existing infrastructure, user preferences -->

| Constraint | Description | Impact |
|------------|-------------|--------|
| [CONSTRAINT_NAME] | [What the constraint is] | [How it influenced technology choices] |

## Acceptance Criteria

Requirements use format: `STK-[CONCEPT]-[NNN]`

- [ ] STK-[CONCEPT]-001: [Criterion — must be measurable, observable, unambiguous]
- [ ] STK-[CONCEPT]-002: [Criterion]

### Untestable Criteria

If any criterion cannot be automatically validated, flag it:

- [ ] STK-[CONCEPT]-NNN: [Criterion] *(untestable)*
  - **Reason:** [Why it cannot be tested automatically]
  - **Proposal:** [Measurable proxy or alternative approach]
  - **Resolution:** [How it will be verified]

---

*Generated with smaqit v0.8.2-beta*
