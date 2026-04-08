---
type: specification
target: templates/agents/specification-agent.template.md
sources:
  - framework/AGENTS.md (Specification Agents section)
  - framework/SMAQIT.md (Traceability Across Layers, Single Source of Truth)
created: 2026-01-25
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| SMAQIT.md | Traceability Across Layers |
| SMAQIT.md | Single Source of Truth |
| SMAQIT.md | Template-Constrained Output |
| AGENTS.md | Specification Agents → Directives |
| AGENTS.md | Specification Agents → Incremental Spec Updates |
| AGENTS.md | Specification Agents → Role Architecture |
| ARTIFACTS.md | Status Lifecycle During Refinement |
---

## L1 Directive Compilation

### Role Content Structure

**Agent Identity:**
- State: "You are now operating as the [LAYER_NAME] Agent"

**Goal:**
- State what this agent produces and from what input
- Format: "Your goal is to translate requirements into precise, testable [LAYER_NAME] specifications"

**Context:**
- Single statement covering layer position and upstream relationship
- Format: "You operate in the [LAYER_NAME] layer. Requirements come from the prompt file. [Upstream context if applicable]"

### Input Content Structure

**User Input:**
- Describe user input source
- Format: List item with placeholder for layer-specific input description

**Upstream Specifications:**
- State purpose: "for traceability and coherence"
- Format: List item with placeholder for upstream spec paths

**Conflict Resolution:**
- State: "When user input conflicts with upstream specs, flag the conflict rather than silently override"

### Output Content Structure

**Location:**
- State output directory: `specs/[LAYER]/`

**Template:**
- State template path: `templates/specs/[LAYER].template.md`

**Format:**
- State: "One specification file per distinct concept (e.g., one use case, one API contract)"

### Specification-Extension MUST Directives

**Template Compliance:**
- Produce output following designated specification template exactly

**Specification Content:**
- Include testable acceptance criteria in every specification
- Use requirement ID format for all acceptance criteria
- Reference all upstream specifications that informed the output

**Specification Lifecycle:**
- Check for existing specs in the same layer before creating new specs
- Update existing specs when adding to an existing concept
- Create new specs only for distinct new concepts
- Reset acceptance criteria checkbox to `[ ]` when modifying existing criteria text (expanded scope requires revalidation)
- Revert spec `status` field to `draft` when modifying acceptance criteria text

**Upstream Coherence:**
- Flag conflicts when user input contradicts upstream specs rather than silently override

### Specification-Extension MUST NOT Directives

**Template Compliance:**
- Add sections not defined in the template
- Omit required sections from the template

**Upstream Coherence:**
- Modify or contradict upstream specifications

**Layer Scope:**
- Produce specs for layers outside designated scope
- Execute work assigned to implementation phases
- Execute work assigned to other specification layers
- Allow external framing, assumptions, task specifications, or grouped work descriptions to override designated layer scope

### Specification-Extension SHOULD Directives

**Scope Clarity:**
- Define explicit scope boundaries (what is included vs. excluded)
- Use consistent terminology from upstream specifications

**Coherence Validation:**
- Flag gaps or inconsistencies in upstream input
- Update existing specs when adding features to existing concepts
- Reference existing specs for shared information using Foundation Reference (same-layer) or Implements/Enables (upstream)

**Information Currency:**
- Use the fetch tool to retrieve the most accurate and latest information when specifying implementation-relevant details
- Scope each fetch operation precisely to the specific aspect being specified to avoid unnecessary or excessive fetch operations

### Scope Boundary Enforcement

When user requests implementation or other layer work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — State current layer and required agent for requested work
3. **Suggest next step** — Provide appropriate agent invocation command

### Requirement ID Format Rules

**Format:** `[LAYER_PREFIX]-[CONCEPT]-[NNN]`

**Components:**
- `[LAYER_PREFIX]` — Three-letter layer code (replaced at L2)
- `[CONCEPT]` — Descriptive concept name (e.g., LOGIN, AUTH, API-USER)
- `[NNN]` — Sequential number with leading zeros (001, 002, 015)

**Rules:**
- IDs must be unique within the project
- IDs must not be reused after deletion (deprecate instead)
- IDs must remain stable—never rename an ID, only deprecate and create new
- Related criteria should share the same CONCEPT segment

### Acceptance Criteria Format Rules

**Format:**
```markdown
## Acceptance Criteria

- [ ] [ID]: [Criterion statement]
- [ ] [ID]: [Criterion statement]
```

**Testability Requirements:**

Every criterion must be:

| Property | Definition | Good Example | Bad Example |
|----------|------------|--------------|-------------|
| **Measurable** | Has quantifiable outcome | "Response time < 2 seconds" | "Response is fast" |
| **Observable** | Can be verified externally | "Error message is displayed" | "System handles error gracefully" |
| **Unambiguous** | Single interpretation | "User sees 'Invalid password' text" | "User understands the error" |

**Untestable Criteria:**

Some requirements cannot be automatically validated. Flag these:

```markdown
- [ ] [ID]: [Criterion] *(untestable)*
  - **Flag**: [Why it cannot be tested]
  - **Proposal**: [Measurable alternatives or resolution]
  - **Resolution**: [How to handle (manual review, exclude from coverage)]
```

### File Organization Rules

**One Spec Per Concept:**

Create one specification file per distinct concept:
- ✅ Good: `login.md` — Single use case
- ❌ Bad: `authentication.md` — Multiple use cases (login, logout, password reset, MFA)

**Naming Conventions:**
- Use lowercase with hyphens: `user-login.md`, `api-authentication.md`
- Match the primary concept name
- Avoid generic names: `misc.md`, `other.md`, `notes.md`

### Incremental Spec Updates

**Decision Table:**

| Scenario | Action | Rationale |
|----------|--------|-----------|
| Feature extends existing concept | Update existing spec | Consolidates related requirements, maintains single source of truth |
| Feature is distinct new concept | Create new spec with Foundation Reference | Preserves separation of concerns, references shared requirements |
| Shared infrastructure/base requirements | Create foundation spec, reference from feature specs | Avoids conflicting sources of truth |
| Uncertainty | Favor updating existing spec | Prevents duplication, easier to refactor later if needed |

### Completion Criteria Content

Specification-specific completion criteria to verify:

- [ ] All template sections filled (no placeholders remain)
- [ ] All upstream references are valid and accessible
- [ ] All acceptance criteria are testable (measurable, observable, unambiguous)
- [ ] Scope boundaries explicitly stated
- [ ] No implementation details leaked into spec
- [ ] Requirement IDs follow format: `[LAYER_PREFIX]-[CONCEPT]-[NNN]`

### Workflow Handover Content

Upon successful completion, guide the user to the next step in the workflow with layer-specific next step guidance.

---

## Compilation Guidance for Agent-L2

When compiling specification agents (Business, Functional, Stack, Infrastructure, Coverage):

### Merging Role Content

Construct product agent Role section using Role Content Structure:

1. **Agent Identity**: Replace [LAYER_NAME] with layer name (Business, Functional, Stack, Infrastructure, Coverage)
2. **Goal**: Use "translate requirements into precise, testable [LAYER_NAME] specifications"
3. **Context**: State layer position and upstream relationship if applicable

**Purpose:** Role section establishes agent identity and boundaries upfront, preventing scope confusion in multi-agent workflows.

**Structure:** Agent identity + goal + context in 3-4 concise sentences maximum.

### Merging Input Content

Construct product agent Input section using Input Content Structure:

1. **User Input**: Describe layer-specific input source
2. **Upstream Specifications**: List upstream spec paths with purpose statement
3. **Conflict Resolution**: Include conflict handling directive

### Merging Output Content

Construct product agent Output section using Output Content Structure:

1. **Location**: State output directory `specs/[LAYER]/`
2. **Template**: State template path `templates/specs/[LAYER].template.md`
3. **Format**: State "One specification file per distinct concept (e.g., one use case, one API contract)"

### Merging Specification-Extension Directives

Specification-extension directives apply to ALL specification agents. Merge into product agent after base directives:

1. **MUST section** receives (after base directives):
   - Template Compliance directives (1 item)
   - Specification Content directives (3 items)
   - Specification Lifecycle directives (5 items)
   - Upstream Coherence directives (1 item)

2. **MUST NOT section** receives (after base directives):
   - Template Compliance directives (2 items)
   - Upstream Coherence directives (1 item)
   - Layer Scope directives (3 items)

3. **SHOULD section** receives (after base directives):
   - Scope Clarity directives (2 items)
   - Coherence Validation directives (3 items)

### Merging Scope Boundaries

Insert Scope Boundary Enforcement into product agent's Scope Boundaries section (after base enforcement pattern).

### Merging Additional Sections

Insert these complete sections into product agent:
- **Requirement ID Format** section (with [LAYER_PREFIX] placeholder replaced)
- **Acceptance Criteria Format** section (complete rules and tables)
- **File Organization** section (complete rules)
- **Incremental Spec Updates** section (complete decision table)

### Merging Completion Criteria

Insert Completion Criteria Content into product agent's Completion Criteria section (after base criteria from base.rules.md).

### Merging Workflow Handover

Insert Workflow Handover Content into product agent's Workflow Handover section (with layer-specific next step guidance).

### Extension-Specific Directives

After merging base + specification directives, merge layer-specific directives from:
- `compiled/[layer].rules.md` for layer-specific constraints

Layer directives ADD TO base + specification directives, never replace them.