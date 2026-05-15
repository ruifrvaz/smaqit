---
id: COV-[CONCEPT]
status: draft
created: [TIMESTAMP]
---

# [CONCEPT_NAME]

## References

<!-- References establish traceability and coherence, not requirement derivation -->
<!-- Coverage specs read ALL upstream layers to ensure complete verification -->

### Business

- [BUS-CONCEPT](../business/[FILENAME].md) — [Business requirement being verified]

### Functional

- [FUN-CONCEPT](../functional/[FILENAME].md) — [Functional behavior being verified]

### Stack

- [STK-CONCEPT](../stack/[FILENAME].md) — [Technology constraint being verified]

### Infrastructure

- [INF-CONCEPT](../infrastructure/[FILENAME].md) — [Infrastructure requirement being verified]

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## Verification Requirements

<!-- User input: test environment, SLAs, security requirements -->

| Category | Requirement | Target |
|----------|-------------|--------|
| Test Environment | [Where tests execute] | [Environment specification] |
| Performance SLA | [Performance benchmark] | [Threshold value] |
| Security | [Security test requirement] | [Compliance target] |
| Integration | [Integration point to verify] | [Expected behavior] |

## Coverage Map

<!-- Requirement ID → Test Case ID → Expected Outcome -->

| Requirement ID | Source Spec | Test Case ID | Expected Outcome |
|----------------|-------------|--------------|------------------|
| BUS-[CONCEPT]-001 | [business/file.md] | COV-[CONCEPT]-001 | [What success looks like] |
| FUN-[CONCEPT]-001 | [functional/file.md] | COV-[CONCEPT]-002 | [What success looks like] |
| STK-[CONCEPT]-001 | [stack/file.md] | COV-[CONCEPT]-003 | [What success looks like] |
| INF-[CONCEPT]-001 | [infrastructure/file.md] | COV-[CONCEPT]-004 | [What success looks like] |

## Test Definitions

<!-- Full Gherkin scenarios for each test case -->

### Integration Tests

```gherkin
# COV-[CONCEPT]-001: Maps to [SOURCE_REQUIREMENT_ID]
Feature: [Feature Name]
  
  Scenario: [Scenario description]
    Given [precondition]
    When [action]
    Then [expected result]
```

### End-to-End Tests

```gherkin
# COV-[CONCEPT]-002: Maps to [SOURCE_REQUIREMENT_ID]
Feature: [Feature Name]
  
  Scenario: [Scenario description]
    Given [precondition]
    And [additional context]
    When [user action]
    Then [expected result]
    And [additional verification]
```

### Performance Tests

```gherkin
# COV-[CONCEPT]-003: Maps to [SOURCE_REQUIREMENT_ID]
Feature: [Performance Feature]
  
  Scenario: [Performance scenario]
    Given [load conditions]
    When [operation under test]
    Then response time is less than [threshold]
    And throughput is at least [target]
```

### Security Tests

```gherkin
# COV-[CONCEPT]-004: Maps to [SOURCE_REQUIREMENT_ID]
Feature: [Security Feature]
  
  Scenario: [Security scenario]
    Given [security context]
    When [potentially malicious action]
    Then [expected security response]
```

### Acceptance Tests

```gherkin
# COV-[CONCEPT]-005: Maps to [SOURCE_REQUIREMENT_ID]
Feature: [Business Feature]
  
  Scenario: [Acceptance scenario]
    Given [business context]
    When [user completes workflow]
    Then [business outcome is achieved]
```

## Untestable Criteria

<!-- Reference upstream flagged criteria with verification decision -->

| Requirement | Source | Reason | Verification Decision |
|-------------|--------|--------|----------------------|
| [REQUIREMENT_ID] | [layer/file.md](../layer/file.md) | [Why untestable] | [Manual review / Deferred / Proxy metric] |

## Coverage Summary

<!-- Pre-execution mapping: requirements to tests -->

| Layer | Total Requirements | Mapped to Tests | Untestable | Coverage % |
|-------|-------------------|-----------------|------------|------------|
| Business | [N] | [N] | [N] | [%] |
| Functional | [N] | [N] | [N] | [%] |
| Stack | [N] | [N] | [N] | [%] |
| Infrastructure | [N] | [N] | [N] | [%] |
| **Total** | **[N]** | **[N]** | **[N]** | **[%]** |

**Formula:** Coverage % = (Mapped to Tests) / (Total Requirements - Untestable) × 100

## Acceptance Criteria

Requirements use format: `COV-[CONCEPT]-[NNN]`

- [ ] COV-[CONCEPT]-001: [Test criterion — must be measurable, observable, unambiguous]
- [ ] COV-[CONCEPT]-002: [Test criterion]

### Untestable Criteria

If any criterion cannot be automatically validated, flag it:

- [ ] COV-[CONCEPT]-NNN: [Criterion] *(untestable)*
  - **Reason:** [Why it cannot be tested automatically]
  - **Proposal:** [Measurable proxy or alternative approach]
  - **Resolution:** [How it will be verified]

---

*Generated with smaqit [SMAQIT_VERSION]*
