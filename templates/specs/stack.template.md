---
id: STK-[CONCEPT]
status: draft
created: [TIMESTAMP]
prompt_version: [GIT_HASH]
---

# [CONCEPT_NAME]

## References

<!-- References establish traceability and coherence, not requirement derivation -->
<!-- Stack specs typically enable multiple Functional specs -->

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

*Generated with smaqit [SMAQIT_VERSION]*
