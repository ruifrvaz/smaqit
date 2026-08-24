# Artifacts

Artifacts are the outputs produced by agents. This document establishes the rules all artifacts MUST follow.

There are three types of artifacts:
- **Specification artifacts** — Declarative documents stating what must be true
- **Design artifacts** — Canonical PlantUML structure plus a generated PNG visual projection
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
| **Context** | Adjacent layer spec used for coherence | Ensures cross-layer coherence |

**Cross-Layer Traceability:**

The Implements/Enables references create an explicit chain for:
- **Impact analysis** — When a Business spec changes, all referencing specs are identified
- **Coverage mapping** — Coverage can trace through references to ensure all requirements are verified

Layer Independence does not mean layer isolation. The reference chain preserves traceability without creating requirement derivation.

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
---
```

**Required Fields:**
- `id`: Unique spec identifier (format: `BUS-LOGIN`, `FUN-AUTH`, etc.)
- `status`: Current lifecycle state
- `created`: Timestamp when spec was generated
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

**Status Lifecycle During Refinement:**

When specification agents modify existing specifications, the spec status returns to draft state to signal need for revalidation through implementation phases.

Modified specifications carry reset checkboxes (granular requirement tracking) and draft status (overall validation state). Both indicators signal that previous validation no longer reflects current specification content.

**Status Transitions on Modification:**

| Previous Status | After Modification | Rationale |
|----------------|-------------------|-----------|
| `implemented` | `draft` | Code implements previous version, not modified spec |
| `deployed` | `draft` | Deployed system reflects previous spec, not current |
| `validated` | `draft` | Tests validated previous requirements, not modified ones |
| `failed` | `draft` | Previous failure may not apply to modified spec |

**Relationship to Checkbox Resets:**

Status reversion and checkbox resets operate together during specification refinement:
- **Checkboxes** track granular requirement satisfaction (per acceptance criterion)
- **Status** tracks overall validation state (spec lifecycle position)
- **Both reset** when specifications change to prevent stale validation indicators

Modified specs enter draft state regardless of how far they progressed previously. Revalidation proceeds through phases: draft → implemented → deployed → validated.

**Spec Modification Source:**

Specification modifications originate from session context changes, not manual spec editing. When requirements evolve, users invoke specification agents with updated requirements in the session context.

When requirements evolve:
1. Users invoke specification agents with updated requirements
2. Agents regenerate specifications from updated session context
3. Modified specs enter draft state with reset checkboxes
4. Specs proceed through revalidation phases

Manual spec editing bypasses the session context and breaks traceability. Session context serves as the authoritative source for all specification content.

**State Aggregation:**

The CLI aggregates phase status by scanning spec frontmatter. Run `smaqit status` to view:

```
Develop: 18 implemented, 2 failed
Deploy: 15 deployed, 3 draft
Validate: 12 validated, 5 draft
```

Implementation agents update individual spec frontmatter. The CLI reads all specs and calculates aggregate counts.

---

## Design Artifacts

Designs are canonical, layer-scoped sidecars for specifications. They contain no requirements or explanatory prose. They express only relationships, boundaries, interaction order, states, component realization, deployment topology, or requirement-trace structure.

### File Pair

Every design is a same-basename pair:

```text
docs/designs/<layer>/<design-id>.md
docs/designs/<layer>/<design-id>.png
```

The Markdown file contains YAML frontmatter followed by exactly one fenced `plantuml` block. That block MUST begin with a `title` directive set to the design's own `id`, so a rendered PNG identifies itself without external context; the block MUST NOT otherwise contain prose, a table, a reference section, a second block, HTML, or an embedded image. The PNG is generated from that PlantUML block with a deterministic opaque cream `#FFF9F0` canvas and is the mandatory representation for specification-agent visual validation; implementation agents consume the PlantUML source after readiness passes.

### Design Identifier

Design IDs use `DSG-[LAYER_PREFIX]-[CONCEPT]-[VIEW]`, where the layer prefix is one of `BUS`, `FUN`, `STK`, `INF`, or `COV`. IDs are globally unique, stable, and never reused after deprecation. Filenames are the lowercase design ID.

### Frontmatter

Required metadata:

- `id`, `status`, and `created` follow specification frontmatter conventions.
- `layer` is one of the five canonical layers.
- `diagram_type` is allowed by that layer's controlled design profile.
- `notation` MUST equal `plantuml`.
- `specifications` lists one or more normalized relative paths to active same-layer specs.
- `requirements` lists requirement IDs that exist in those specifications.
- `source_sha256` hashes the PlantUML source normalized to LF with exactly one final newline.
- `image_sha256` hashes the raw PNG bytes.
- `visual_validation` records `status`, `validated_at`, and the exact source/image hashes reviewed by the authoring agent.

### Reference Rules

Every active spec MUST contain `## Design References` with links to the canonical Markdown and PNG. One high-signal design may serve several related specs. The design's `specifications` metadata and every spec link MUST agree bidirectionally. Specs contain links only; designs contain metadata paths only.

All reference paths MUST stay within the project, resolve without symlink/path traversal escape, and match the declared layer. Coverage designs visualize existing traceability but MUST NOT create requirements or duplicate the Coverage Map.

A Functional spec's `## Design References` section MAY additionally link its paired Design Sequence Diagram alongside its own-layer `system-sequence` design — the established two-link convention. That companion link does not count toward the spec's own-layer design-pair requirement, but it is not a layer mismatch either.

### Validation Gates

1. **Structural:** schema, ID, layer/profile, required `title` directive matching the design's own `id`, one-block/no-prose content otherwise, safe paths, bidirectional references, requirement existence, PNG signature/dimensions, opaque canvas, hashes, lifecycle, and minimum coverage.
2. **PlantUML:** syntax check, SVG render, and SVG-to-PNG conversion through the shipped pinned toolchain.
3. **Visual:** the owning specification agent opens the PNG and verifies legibility, clipping, direction/order, boundaries, disconnected elements, coherence, and excessive complexity.

Failure codes are stable: `DESIGN-TOOLCHAIN-UNAVAILABLE`, `DESIGN-ARTIFACT-MISSING`, `DESIGN-ARTIFACT-STALE`, `DESIGN-SYNTAX-INVALID`, `DESIGN-VISION-UNAVAILABLE`, and `DESIGN-VISUAL-INVALID`. There is no source-reading fallback for the authoring-time visual gate. At the downstream boundary, `smaqit plan --phase` verifies the current hash-bound attestation before implementation agents consume PlantUML source.

### Design Lifecycle

A semantic PlantUML edit resets the design and every linked active spec to `draft`. A semantic spec edit resets its linked designs to `draft`. A shared design cannot advance beyond its least-advanced linked spec. Hash disagreement makes a pair stale and phase-ineligible regardless of frontmatter status. Deprecated specs and designs are excluded from mandatory coverage but their IDs remain reserved.

Existing projects with active specifications and no compliant design pairs fail validation with migration diagnostics. Smaqit does not generate placeholders or waive gates.

---

## Design Sequence Diagrams

Design Sequence Diagrams are Phase 1/Development-owned output, not a Design Artifact sidecar. Where a `system-sequence` design documents a use case's external, black-box contract, a Design Sequence Diagram documents how the objects actually written to satisfy it collaborate internally — the standard next step in the classic analysis-design progression (Larman, *Applying UML and Patterns*): use-case → system sequence diagram (external contract) → design sequence diagram (internal realization).

### Ownership

Generated by `smaqit.development`, after code and tests pass and before Phase 1 completes — never by a specification agent. This does not relax the Bounded Agents principle's design-authorship boundary: that boundary governs Design Artifacts specifically, and a Design Sequence Diagram was never a Design Artifact.

### File Pair and Storage

Same file-pair shape as a Design Artifact, but its own sibling tree — never inside a specification layer's directory:

```text
docs/designs/design-sequence/<design-id>.md
docs/designs/design-sequence/<design-id>.png
```

### Design Identifier

IDs use `DSG-DSD-[CONCEPT]-DESIGN-SEQUENCE`, following the same `DSG-[PREFIX]-[CONCEPT]-[VIEW]` shape as Design Artifact IDs, with layer prefix `DSD`.

### Frontmatter

Same required fields as a Design Artifact (`id`, `status`, `created`, `layer: design-sequence`, `diagram_type: design-sequence`, `notation: plantuml`, `specifications`, `requirements`, `source_sha256`, `image_sha256`, `visual_validation`), plus one field unique to this category:

- `realizes` — the `id` of the `system-sequence` design this diagram realizes. Required for `design-sequence`, forbidden for every other layer.

`specifications` points at the same Functional spec the paired `system-sequence` design links; `requirements` reuses the same `FUN-*` requirement IDs.

### Footbox Suppression

A Design Sequence Diagram's PlantUML source MUST include `hide footbox`. PlantUML otherwise duplicates every declared participant box at the bottom of the render — unlike `system-sequence`, a design-sequence diagram has no full black-box profile (it legitimately declares multiple internal collaborators), so this is checked as its own standalone structural requirement rather than folded into a broader profile. Missing it reports `DESIGN-VISUAL-INVALID: design-sequence diagrams must include \`hide footbox\``.

### Grounding and Completeness

Two checks run inside `smaqit design attest` before it will stamp a passing attestation — attestation is earned, not merely ordered correctly by the caller:

1. **Grounding.** Every message in the diagram carries a `' impl: <path>:<line>` PlantUML comment citing the real code it represents. Citations are checked to exist (file present, line within range) — existence-only, not semantic verification that the cited code does what the message claims. A diagram with no citations, or any citation that doesn't resolve, fails attestation.
2. **Completeness.** Every operation the paired `system-sequence` design promises must have a matching operation label somewhere in this diagram. A diagram missing a promised operation fails attestation, naming the gap.

Both are source-level heuristic scans over the PlantUML text — not a full PlantUML parser and not semantic verification of correctness — the same lint-style tradeoff already accepted for `system-sequence`'s own structural validation. An author (human or agent) could still mislabel or omit a citation to dodge the check; the goal is catching honest drift, not adversarial-proofing the format.

A diagram is a complement to code review, not a substitute for it — especially for security-sensitive or edge-case-heavy code, where a diagram is necessarily a lossy abstraction.

### Existence Enforcement

The phase-readiness gate behind `smaqit plan --phase=develop|deploy|validate` cannot enforce this artifact — it is scoped to specs still in the current incremental cycle (`draft`/`failed`), and a Design Sequence Diagram structurally cannot exist before a spec passes that point. Enforcement instead lives in `smaqit design validate`'s general sweep: for every Functional spec at `status: implemented` or beyond, each linked `system-sequence` design MUST have a companion Design Sequence Diagram in the same `## Design References` section whose `realizes` field names it, and that diagram MUST itself pass validation. A still-`draft` Functional spec is exempt — the requirement applies only once implementation has actually happened. A missing or invalid grounding diagram reports `DESIGN-ARTIFACT-MISSING`.

This is a breaking change with no grandfathering: an existing project's `functional` specs that reached `implemented` before this check existed, without a Design Sequence Diagram, fail `smaqit design validate` immediately on adoption. Fix forward by authoring the missing diagram — there is no legacy exemption, version gate, or downgrade-to-warning path.

### Design Sequence Lifecycle

Unlike a Design Artifact, a Design Sequence Diagram's own status is never forced to equal its linked specification's current lifecycle rank, and a semantic edit never resets that specification to `draft`. It is generated once its Functional spec is already `implemented`; nothing about later Deploy or Validate phases changes the collaboration it documents, so re-deriving or re-attesting it is never required by those phases advancing.

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
- One Design Sequence Diagram per implemented Functional spec (see "Design Sequence Diagrams" above — its own artifact category, not an Implementation Artifact)

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
