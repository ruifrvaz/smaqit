# Artifacts

Artifacts are the outputs produced by agents. This document establishes the rules all artifacts MUST follow.

There are two types of artifacts:
- **Specification artifacts** — Declarative documents stating what must be true
- **Implementation artifacts** — Imperative outputs stating how to make it true

---

## Specification Artifacts

Specifications are the source of truth in Spec Driven Development. They serve as contracts between layers.

### Purpose

- **Upstream → Downstream**: Each spec informs the next layer's work
- **Spec → Implementation**: Implementation agents consume specs as requirements, not suggestions
- **Spec → Validation**: Coverage specs translate requirements into verifiable tests

A specification is complete when another agent (or human) can implement or validate against it without requiring additional context.

### Requirement Identifiers

Every acceptance criterion MUST have a unique identifier for traceability.

**Format:**
```
[LAYER_PREFIX]-[CONCEPT]-[NNN]
```

| Component | Description | Examples |
|-----------|-------------|----------|
| `LAYER_PREFIX` | Three-letter layer code | `BUS`, `FUN`, `STK`, `INF`, `COV` |
| `CONCEPT` | Descriptive concept name | `LOGIN`, `AUTH`, `API-USER` |
| `NNN` | Sequential number (3 digits) | `001`, `002`, `015` |

**Examples:**

| Layer | Requirement ID | Description |
|-------|----------------|-------------|
| Business | `BUS-LOGIN-001` | User can authenticate with valid credentials |
| Functional | `FUN-AUTH-TOKEN-001` | JWT token expires after 24 hours |
| Stack | `STK-FRAMEWORK-001` | Use React 18+ for frontend |
| Infrastructure | `INF-SCALING-001` | Auto-scale at 80% CPU threshold |
| Coverage | `COV-LOGIN-001` | Test case for BUS-LOGIN-001 |

**Rules:**
- IDs MUST be unique within the project
- IDs MUST NOT be reused after deletion (mark as deprecated instead)
- IDs MUST remain stable—never rename an ID, only deprecate and create new
- Related criteria SHOULD share the same `CONCEPT` segment

### Acceptance Criteria

Acceptance criteria define testable conditions that must be satisfied.

**Format:**
```markdown
## Acceptance Criteria

- [ ] [ID]: [Criterion statement]
- [ ] [ID]: [Criterion statement]
```

**Testability Requirements:**

Every criterion MUST be:

| Property | Definition | Good Example | Bad Example |
|----------|------------|--------------|-------------|
| **Measurable** | Has quantifiable outcome | "Response time < 2 seconds" | "Response is fast" |
| **Observable** | Can be verified externally | "Error message is displayed" | "System handles error gracefully" |
| **Unambiguous** | Single interpretation | "User sees 'Invalid password' text" | "User understands the error" |

**Untestable Criteria:**

Some requirements cannot be automatically validated. These MUST be flagged:

```markdown
- [ ] BUS-UX-002: Dashboard feels modern and engaging *(untestable)*
  - **Flag**: Subjective criterion—cannot be automatically validated
  - **Proposal**: Define measurable proxies (animations, color palette, satisfaction score)
  - **Resolution**: Defer to manual UX review; exclude from automated coverage
```

Untestable criteria:
- MUST be flagged with `*(untestable)*` marker
- MUST include a proposal for measurable alternatives or resolution
- MUST NOT block spec completion

### Traceability

Specifications MUST reference their sources explicitly.

**Reference Types:**

| Type | Meaning | Use Case |
|------|---------|----------|
| **User Input** | Direct requirement from user | Primary source for layer requirements |
| **Context** | Adjacent layer spec used for coherence | Ensures cross-layer coherence |

**Cross-Layer Traceability:**

Even though requirements come from user input per layer, the Implements/Enables references create an explicit chain for:
- **Impact analysis** — When a Business spec changes, all referencing specs are identified
- **Coverage mapping** — Coverage can trace through references to ensure all requirements are verified

Layer Independence does not mean layer isolation. The reference chain preserves traceability without creating requirement derivation.

**User Input Traceability:**

Every requirement traces to user input for that layer:
- Business: stakeholder requirements
- Functional: experience requirements  
- Stack: technology preferences
- Infrastructure: deployment requirements
- Coverage: verification requirements

**Context References:**

Specs reference adjacent layers for coherence and traceability. Context references distinguish between feature and foundation specs:

| Reference Type | Meaning | Example |
|----------------|---------|---------|
| **Implements** | Feature spec with 1:1 mapping to business case | Feature spec → Single use case |
| **Enables** | Foundation spec serving multiple business cases | Shared component → Multiple use cases |

**Format:**
```markdown
## References

### Implements
<!-- Feature spec: direct 1:1 implementation -->
- [BUS-LOGIN](../business/login.md) — Implements login use case

### Enables  
<!-- Foundation spec: serves multiple business cases -->
- [BUS-CHECKOUT](../business/checkout.md) — Requires authenticated session
- [BUS-PROFILE](../business/profile.md) — Requires authenticated session
```

**Foundation specs without mapping:**

When a foundation spec precedes Business specs or serves anticipated needs:

```markdown
## References

### Enables
<!-- ⚠️ FOUNDATION WITHOUT MAPPING -->
**Justification:** [Why this foundation is needed before Business specs exist]
```

Orphaned foundations (no references, no justification) should be flagged by Coverage.

**Rules:**
- Every spec (except Business) MUST have a References section
- References MUST use relative paths within `.smaqit/specs/`
- References provide context for coherence, not requirements
- Implementation agents validate cross-layer coherence

**Traceability Matrix:**

For complex projects, maintain traceability across layers:

| Business | Functional | Stack | Infrastructure | Coverage |
|----------|------------|-------|----------------|----------|
| BUS-LOGIN-001 | FUN-AUTH-001 | STK-JWT-001 | — | COV-LOGIN-001 |

### Coverage Translation

The Coverage layer translates acceptance criteria into executable test definitions.

**Translation Example:**

Source (Functional spec):
```markdown
- [ ] FUN-AUTH-001: User receives JWT token upon successful login
```

Coverage translation:
```gherkin
# COV-AUTH-001: Maps to FUN-AUTH-001
Feature: Authentication Token
  Scenario: Successful login returns JWT token
    Given a registered user with valid credentials
    When the user submits login request
    Then the response contains a JWT token
```

**Coverage Rules:**
- Each testable criterion MUST map to at least one test case
- Coverage IDs MUST reference their source requirement ID
- Untestable criteria MUST be listed with justification for exclusion
- Spec coverage % = (tested criteria / total testable criteria) × 100

### File Organization

**One Spec Per Concept:**

| Good | Bad |
|------|-----|
| `login.md` — Login use case | `authentication.md` — Login, logout, password reset, MFA |
| `user-registration.md` — Registration flow | `users.md` — Registration, profile, settings, deletion |

**Naming Conventions:**
- Use lowercase with hyphens: `user-login.md`, `api-authentication.md`
- Match the primary concept name
- Avoid generic names: `misc.md`, `other.md`, `notes.md`

**Directory Structure:**
```
.smaqit/specs/
├── business/
├── functional/
├── stack/
├── infrastructure/
└── coverage/
```

### Specification Completeness

A specification is complete when:

- All template sections are filled (no placeholders remain)
- All acceptance criteria have unique IDs
- All acceptance criteria are testable (or flagged as untestable)
- All upstream references are valid and accessible
- Scope boundaries are explicitly stated
- No implementation details are present (except Stack layer)

---

## Implementation Artifacts

Implementations are the imperative outputs produced by implementation agents. They satisfy spec-defined behavior while following industry standards.

### The Anchoring Principle

> "Implementations MUST comply with industry standards for their stack, while satisfying spec-defined behavior. Two compliant implementations may differ internally, but MUST be structurally recognizable and behaviorally equivalent."

### The Isolation Principle

> "Agents operate on references, never values. Secrets and credentials MUST remain outside the agent's context at all times—resolution happens in a trusted execution layer that returns only outcomes, never the sensitive data itself."

### Three Dimensions

Every implementation exists across three dimensions:

```
┌─────────────────────────────────────────────────────────────┐
│ BEHAVIOR (from Specs)                                       │
│ Invariant — MUST be identical across implementations        │
├─────────────────────────────────────────────────────────────┤
│ STRUCTURE (from Industry Standards)                         │
│ Consistent — SHOULD follow stack-specific best practices    │
├─────────────────────────────────────────────────────────────┤
│ INTERNALS (Implementation Freedom)                          │
│ Variable — MAY differ, no two implementations identical     │
└─────────────────────────────────────────────────────────────┘
```

**Behavior (Invariant):**
- Defined by specifications, MUST be satisfied exactly
- No deviation permitted—behavior is the contract

**Structure (Consistent):**
- Follows industry standards for the chosen stack
- Implementations SHOULD be recognizable to practitioners

**Internals (Variable):**
- Variable names, helper functions, internal patterns
- May vary freely between implementations

### Traceability

Implementation code SHOULD include references to specifications:

```csharp
/// <summary>
/// Authenticates user and returns JWT token.
/// Implements: FUN-AUTH-001, FUN-AUTH-002
/// </summary>
public async Task<AuthResult> Login(LoginRequest request)
```

**Rules:**
- Major components SHOULD reference the spec requirements they implement
- Traceability MUST be verifiable during validation phase

### Validation Requirements

| Dimension | Verifiable? | How |
|-----------|-------------|-----|
| Behavior | MUST | Automated tests from Coverage specs |
| Structure | SHOULD | Static analysis, architectural tests |
| Internals | NOT REQUIRED | — |

### Implementation Artifacts by Phase

**Develop Phase:**
- Source code, tests, configurations, build files
- README with build, test, and run instructions
- Development report (build/test/run results)
- MUST satisfy all spec acceptance criteria
- MUST follow stack-specific standards

**Deploy Phase → Infrastructure:**
- Infrastructure code (Terraform, etc.)
- Deployment manifests, environment configs
- MUST NOT hardcode secrets (Isolation Principle)

**Validate Phase → Reports:**
- Test results, coverage report, validation summary
- MUST map results to Coverage spec test cases
- MUST include spec coverage percentage

**Validation Report Format:**
```markdown
# Validation Report

## Summary
- Specs Covered: 47/50 (94%)
- Tests Passed: 45/47 (96%)

## Coverage Gaps
| Requirement | Reason |
|-------------|--------|
| BUS-UX-002 | Untestable: subjective criterion |

## Failures
| Test | Requirement | Result | Details |
|------|-------------|--------|---------|
| COV-AUTH-005 | FUN-AUTH-003 | FAIL | Token expiration is 48h, spec requires 24h |
```

### Implementation Completeness

An implementation is complete when:

- All referenced spec acceptance criteria are satisfied
- Stack-specific standards are followed
- Traceability to specs is documented
- No unspecified features were added
- Validation can verify behavior against specs

---

## See Also

- [SMAQIT](SMAQIT.md) — Framework overview and principles
- [LAYERS](LAYERS.md) — Layer definitions and dependencies
- [PHASES](PHASES.md) — Phase workflows and transitions
- [AGENTS](AGENTS.md) — Agent behaviors
- [TEMPLATES](TEMPLATES.md) — Template structure rules
