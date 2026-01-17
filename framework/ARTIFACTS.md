# Artifacts

Artifacts are the outputs produced by agents. This document establishes the rules all artifacts MUST follow.

There are two types of artifacts:
- **Specification artifacts** — Declarative documents stating what must be true
- **Implementation artifacts** — Imperative outputs stating how to make it true

---

## Specification Artifacts

Specifications are the source of truth in Spec Driven Development. They serve as contracts between layers.

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

| Layer | Requirement ID Format | Description Pattern |
|-------|----------------------|---------------------|
| Business | `BUS-[CONCEPT]-001` | [Use case or actor goal description] |
| Functional | `FUN-[CONCEPT]-001` | [Behavior or data model requirement] |
| Stack | `STK-[CONCEPT]-001` | [Technology choice or tool requirement] |
| Infrastructure | `INF-[CONCEPT]-001` | [Deployment or scaling requirement] |
| Coverage | `COV-[CONCEPT]-001` | Test case for [upstream requirement ID] |

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
| **Prompt File** | Layer-specific prompt (`.github/prompts/smaqit.[layer].prompt.md`) | Primary source for layer requirements |
| **Context** | Adjacent layer spec used for coherence | Ensures cross-layer coherence |

**Cross-Layer Traceability:**

Even though requirements come from prompt files per layer, the Implements/Enables references create an explicit chain for:
- **Impact analysis** — When a Business spec changes, all referencing specs are identified
- **Coverage mapping** — Coverage can trace through references to ensure all requirements are verified

Layer Independence does not mean layer isolation. The reference chain preserves traceability without creating requirement derivation.

**Prompt File Traceability:**

Every requirement traces to the prompt file for that layer:
- Business: stakeholder requirements
- Functional: experience requirements  
- Stack: technology preferences
- Infrastructure: deployment requirements
- Coverage: test requirements (scope, environment, integration points, thresholds)

**Context References:**

Specs reference adjacent layers for coherence and traceability. Context references distinguish between feature and foundation specs:

| Reference Type | Meaning | Example |
|----------------|---------|---------|
| **Implements** | Feature spec with 1:1 mapping to upstream spec | Feature spec → Single upstream requirement |
| **Enables** | Foundation spec serving multiple upstream specs | Foundation spec → Multiple upstream requirements |
| **Foundation Reference** | Feature spec references foundation spec in same layer | Feature spec → Foundation spec for shared requirements |

**Cross-Layer Format:**
```markdown
## References

### Implements
<!-- Feature spec: direct 1:1 implementation -->
- [BUS-[CONCEPT]-NNN](../business/[filename].md) — Implements [use case description]

### Enables  
<!-- Foundation spec: serves multiple business cases -->
- [BUS-[CONCEPT]-NNN](../business/[filename].md) — Enables [use case description]
- [BUS-[CONCEPT]-NNN](../business/[filename].md) — Enables [use case description]
```

**Foundation Reference Format (for avoiding duplication):**
```markdown
## References

### Foundation Reference
<!-- Same-layer reference: feature spec extends foundation spec -->
- [STK-[FOUNDATION-CONCEPT]](./base-stack.md) — Shared requirements referenced here

### Implements
- [FUN-[CONCEPT]-NNN](../functional/feature.md) — Implements feature functionality
```

**Foundation Reference Rules:**
- Use when a feature spec extends a foundation spec in the same layer
- Foundation specs contain shared requirements that multiple feature specs depend on
- Example: Feature spec "[STK-CLI]" references foundation spec "[STK-PYTHON-BASE]" for base Python 3.8+ and development environment requirements
- Prefer updating existing spec over creating new spec with foundation reference when concept is not distinct

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
- References MUST use relative paths within `specs/`
- References provide context for coherence, not requirements
- Implementation agents validate cross-layer coherence

**Traceability Matrix:**

For complex projects, maintain traceability across layers:

| Business | Functional | Stack | Infrastructure | Coverage |
|----------|------------|-------|----------------|----------|
| BUS-[CONCEPT]-001 | FUN-[CONCEPT]-001 | STK-[CONCEPT]-001 | — | COV-[CONCEPT]-001 |

### Coverage Translation

The Coverage layer translates acceptance criteria into executable test definitions.

**Translation Example:**

Source (Functional spec):
```markdown
- [ ] FUN-[CONCEPT]-001: [Behavior description]
```

Coverage translation:
```gherkin
# COV-[CONCEPT]-001: Maps to FUN-[CONCEPT]-001
Feature: [Feature Name]
  Scenario: [Scenario description]
    Given [precondition]
    When [action]
    Then [expected outcome]
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
specs/
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

### Specification State

Specifications carry state through implementation phases via frontmatter metadata.

**Frontmatter Schema:**

```yaml
---
id: [LAYER_PREFIX]-[CONCEPT]
status: draft | implemented | deployed | validated | failed | deprecated
created: [ISO8601_TIMESTAMP]
implemented: [ISO8601_TIMESTAMP]
deployed: [ISO8601_TIMESTAMP]
validated: [ISO8601_TIMESTAMP]
prompt_version: [GIT_COMMIT_HASH]
---
```

**Required Fields:**
- `id`: Unique spec identifier (format: `BUS-LOGIN`, `FUN-AUTH`, etc.)
- `status`: Current lifecycle state
- `created`: Timestamp when spec was generated
- `prompt_version`: Git commit hash of prompt file at spec generation time

**Optional Fields (set by implementation agents):**
- `implemented`: When Development agent completed code generation
- `deployed`: When Deployment agent completed deployment
- `validated`: When Validation agent verified acceptance criteria

**State Transitions:**

| From State | To State | Triggered By | Agent |
|------------|----------|--------------|-------|
| (none) | `draft` | Spec generation | Specification agents |
| `draft` | `implemented` | Code generated, tests pass | Development agent |
| `draft` | `failed` | Code generation failed | Development agent |
| `implemented` | `deployed` | Deployment succeeded | Deployment agent |
| `implemented` | `failed` | Deployment failed | Deployment agent |
| `deployed` | `validated` | All tests passed | Validation agent |
| `deployed` | `failed` | Tests failed | Validation agent |
| Any | `deprecated` | Feature removed | Manual/Specification agents |

**Acceptance Criteria State:**

Each implementation agent updates checkboxes for specs it processes:
- `[ ]` = Not yet implemented/validated
- `[x]` = Satisfied (implementation complete or test passed)
- `[!]` = Failed, untestable, or not satisfied

Example:
```markdown
## Acceptance Criteria

- [x] BUS-LOGIN-001: User can authenticate with valid credentials
- [x] BUS-LOGIN-002: Invalid credentials show error message
- [!] BUS-LOGIN-003: Password complexity enforced (FAILED: regex bug)
```

**Checkbox Lifecycle During Refinement:**

When specification agents modify existing acceptance criteria text (expanding scope, changing requirements), they MUST reset checkbox state to `[ ]` to indicate revalidation is needed.

**Rules:**
- Specification agents MUST reset `[x]` → `[ ]` when modifying acceptance criterion text
- Specification agents MUST reset `[!]` → `[ ]` when modifying acceptance criterion text
- Implementation agents later update `[ ]` → `[x]` or `[!]` after revalidation
- Adding new criteria always starts with `[ ]` (new, not yet validated)

**Rationale:** Expanded or modified requirements need revalidation. Checkboxes reflect implementation status, not specification intent. When the requirement changes, the checkbox must reset to prevent misleading developers about what needs revalidation.

**Example:**
```markdown
# Before spec update (requirement was implemented)
- [x] FUN-OUTPUT-006: Generate output containing Mario character

# After spec update expanding scope (checkbox reset by specification agent)
- [ ] FUN-OUTPUT-006: Generate output containing Mario and Luigi characters

# After implementation (checkbox updated by Development agent)
- [x] FUN-OUTPUT-006: Generate output containing Mario and Luigi characters
```

**Stale Specs:**

Specs become stale when content changes after implementation. Detection is **user responsibility**.

**State Aggregation:**

The CLI aggregates phase status by scanning spec frontmatter. Run `smaqit status` to view:

```
Develop: 18 implemented, 2 failed
Deploy: 15 deployed, 3 draft
Validate: 12 validated, 5 draft
```

Implementation agents update individual spec frontmatter. The CLI reads all specs and calculates aggregate counts.

---

## Implementation Artifacts

Implementations are the imperative outputs produced by implementation agents. They satisfy spec-defined behavior while following industry standards.

### The Anchoring Principle

> "Implementations MUST comply with industry standards for their stack, while satisfying spec-defined behavior. Two compliant implementations may differ internally, but MUST be structurally recognizable and behaviorally equivalent."

### The Isolation Principle

> "Agents operate on references, never values. Secrets and credentials MUST remain outside the agent's context at all times—resolution happens in a trusted execution layer that returns only outcomes, never the sensitive data itself."

### The Test Independence Principle

> "Test artifacts exist independently of agent execution. Tests can run in any environment with the appropriate runtime, enabling continuous integration, local developer workflows, and automated verification outside the validation phase."

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
/// [Method description].
/// Implements: [REQ-ID-001], [REQ-ID-002]
/// </summary>
public async Task<Result> MethodName(Request request)
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
- Development report in `.smaqit/reports/development-phase-report-YYYY-MM-DD.md` (build/test/run results)

**Deploy Phase:**
- Infrastructure code (Terraform, etc.)
- Deployment manifests, environment configs
- Deployment report in `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md` with health status and endpoints

**Validate Phase:**
- **Test artifacts (executable, committable):**
  - Test files in `tests/` directory (e.g., `tests/test_*.py`)
  - Test framework configuration (e.g., `pytest.ini`, `unittest.cfg`)
  - Test fixtures and utilities (e.g., `tests/conftest.py`)
  - CI/CD workflow configuration (e.g., `.github/workflows/validation.yml`)
- **Validation report** in `.smaqit/reports/validation-phase-report-YYYY-MM-DD.md` with:
  - Test results mapped to Coverage spec test cases
  - Spec coverage percentage

**Phase State Tracking:**

Implementation agents update spec frontmatter. CLI aggregates status across all specs.

Frontmatter example:

```yaml
---
id: BUS-LOGIN-001
status: validated
created: 2025-12-26T10:00:00Z
implemented: 2025-12-26T10:30:00Z
deployed: 2025-12-26T11:00:00Z
validated: 2025-12-26T11:30:00Z
prompt_version: abc123
---
```

Agents use atomic writes (temp file + rename) to prevent corruption. The `smaqit status` command reads this file to display project state.

**Validation Report Format:**
```markdown
# Validation Report

## Summary
- Specs Covered: 47/50 (94%)
- Tests Passed: 45/47 (96%)

## Coverage Gaps
| Requirement | Reason |
|-------------|--------|
| [REQ-ID] | Untestable: [reason] |

## Failures
| Test | Requirement | Result | Details |
|------|-------------|--------|---------|
| [TEST-ID] | [REQ-ID] | FAIL | [Failure description] |
```

### Implementation Completeness

An implementation is complete when:

- All referenced spec acceptance criteria are satisfied
- Stack-specific standards are followed
- Traceability to specs is documented
- No unspecified features were added
- Validation can verify behavior against specs
