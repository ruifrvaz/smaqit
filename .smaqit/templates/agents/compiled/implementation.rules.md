---
type: implementation
target: templates/agents/implementation-agent.template.md
sources:
  - framework/AGENTS.md (Implementation Agents section)
  - framework/SMAQIT.md (Traceability Across Layers, Single Source of Truth)
created: 2026-01-25
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| SMAQIT.md | Traceability Across Layers |
| SMAQIT.md | Single Source of Truth |
| SMAQIT.md | Self-Validating Agents |
| AGENTS.md | Implementation Agents → Directives |
| AGENTS.md | Implementation Agents → Phase Specification + Implementation |
| AGENTS.md | Implementation Agents → Cross-Layer Consolidation |

---

## Placeholder Catalog

The following placeholders appear in `templates/agents/implementation-agent.template.md` in addition to those defined in `base-agent.template.md`. Agent-L2 MUST resolve all base placeholders plus these when compiling an implementation agent.

| Placeholder | Description |
|-------------|-------------|
| `[PHASE]` | Lowercase phase identifier appended to agent name in frontmatter (e.g., `development`) |
| `[PHASE_NAME]` | Title-case phase name used in descriptions (e.g., `Development`) |
| `[AGENT_NAME]` | Display name for the phase agent document heading (e.g., `Development Agent`) — used as the H1 title, overriding the base `[AGENT_TITLE]` |
| `[IMPLEMENTATION_MUST_DIRECTIVES]` | Implementation-extension MUST directives compiled from this rules file |
| `[PHASE_MUST_DIRECTIVES]` | Phase-specific MUST directives from a user-created phase compilation file (omit if none) |
| `[IMPLEMENTATION_MUST_NOT_DIRECTIVES]` | Implementation-extension MUST NOT directives compiled from this rules file |
| `[PHASE_MUST_NOT_DIRECTIVES]` | Phase-specific MUST NOT directives from a user-created phase compilation file (omit if none) |
| `[IMPLEMENTATION_SHOULD_DIRECTIVES]` | Implementation-extension SHOULD directives compiled from this rules file |
| `[PHASE_SHOULD_DIRECTIVES]` | Phase-specific SHOULD directives from a user-created phase compilation file (omit if none) |
| `[CROSS_LAYER_CONSOLIDATION_CONTENT]` | Phase-specific cross-layer coherence checking workflow |
| `[SCOPE_BOUNDARIES_CONTENT]` | Phase-specific scope boundary enforcement |
| `[PHASE_SPECIFIC_RULES_CONTENT]` | Phase-unique directives from a user-created phase compilation file (omit if none) |
| `[STATE_TRACKING_CONTENT]` | Phase-specific state update instructions; contains secondary placeholders `[STATUS_VALUE]` and `[TIMESTAMP_FIELD]` resolved at L2 with concrete product-defined values |
| `[COMPLETION_CRITERIA_CONTENT]` | Phase-specific self-validation checklist (extends base criteria) |
| `[WORKFLOW_HANDOVER_CONTENT]` | Phase-specific next-step guidance after completion |
| `[FAILURE_HANDLING_CONTENT]` | Phase-specific failure handling (extends base failure table with cross-layer conflict row) |

---

## L1 Directive Compilation

### Role Content Structure

**Agent Identity:**
- State: "You are now operating as the [PHASE_NAME] Agent"

**Goal:**
- State what this agent produces and from what input
- Format: "Your goal is to transform [upstream specifications] into [output artifacts]"

**Phase Context:**
- Single statement covering phase position in workflow and scope
- Format: "You operate in the [PHASE_NAME] phase. [Phase-specific context about workflow position and scope]"

### Input Content Structure

**Upstream Specifications:**
- List upstream specifications consumed as input for this phase
- Format: Bullet list with file paths

**User Input:**
- Describe phase-specific user-provided context or requirements
- Format: Brief description of what user may provide

**Conflict Resolution:**
- State conflict handling policy
- Standard: "When prompt requirements conflict with upstream specs, flag the conflict rather than silently override."

### Output Content Structure

**Artifacts:**
- List phase-specific output artifacts with file paths
- Include phase report requirement: "Phase report documenting phase outcomes and decisions"

**Format:**
- State phase-specific formatting requirements
- MUST include: "Phase report MUST be written to a designated reports location documenting phase outcomes"
- MUST include: "Phase report MUST document work discovery results and any deviations from specifications"

### Cross-Layer Consolidation Content

**4-Step Workflow:**
1. **Coherence check** — Verify specs across layers are compatible
2. **Conflict detection** — Identify contradictions between layers
3. **Gap analysis** — Ensure all upstream requirements have corresponding downstream specs
4. **Amendment request** — If conflicts or gaps exist, request spec amendments before proceeding

**Directive:** MUST NOT proceed with implementation while unresolved conflicts exist.

### Scope Boundaries Content

**MUST NOT Directives:**
- Execute work assigned to other phases
- Execute work assigned to specification agents

**Boundary Enforcement (3-step pattern):**

When user requests out-of-phase work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — "[Phase] phase is [status]. To proceed with [requested work], invoke [target agent]."
3. **Suggest next step** — Provide the appropriate agent invocation command

### Phase-Specific Rules Content

Placeholder for phase-specific compilation:

`[PHASE_SPECIFIC_RULES]`

### State Tracking Content

**Processed Spec Updates:**

For each spec processed:
1. Update processing state: set status to `[STATUS_VALUE]`
2. Record phase completion timestamp: set `[TIMESTAMP_FIELD]` to current timestamp

**Upstream Spec Updates:**

Agent reads and references upstream specs for coherence validation. All referenced specs MUST be updated to reflect phase state:
1. Update ALL specs within phase scope
2. Update ALL upstream specs referenced for coherence
3. For each referenced spec:
   - Set status to `[STATUS_VALUE]`
   - Set `[TIMESTAMP_FIELD]` to current timestamp

**State Field Conventions:**
- `[STATUS_VALUE]` — A product-defined string reflecting the phase outcome (e.g., `developed`, `deployed`, `validated`). Replaced at L2 with the concrete value for the phase.
- `[TIMESTAMP_FIELD]` — A product-defined field name for recording when the phase was applied (e.g., a frontmatter key, a database column, a header). Replaced at L2 with the concrete field name.

**Additional State Directives:**

Phase-specific additional state tracking rules (from user-defined domain compilation files).

### Completion Criteria Content

**Phase-Specific Completion Checks:**

- [ ] All referenced spec requirements are addressed
- [ ] All acceptance criteria from specs are satisfied
- [ ] Output is traceable to input specifications
- [ ] No unspecified features were added
- [ ] Cross-layer consolidation completed without conflicts
- [ ] Phase report written to designated reports location documenting phase outcomes
- [ ] Processing state updated for all consumed and referenced specs
- [ ] Acceptance criteria checkboxes updated in processed specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied/untestable)
- [ ] [Additional phase-specific completion criteria from user-defined domain compilation files]

### Workflow Handover Content

**Pattern:**

Upon successful completion, guide the user to the next step in the workflow:

```
[PROPOSE_NEXT_STEP]
```

Replace [PROPOSE_NEXT_STEP] with phase-specific next step proposal (compiled from phase.rules.md).

### Failure Handling Extension

Extend the base Failure Handling Pattern with this additional row:

| Situation | Action |
|-----------|--------|
| Cross-layer conflict | Request spec amendments before proceeding |

### Implementation-Extension MUST Directives

**Work Discovery:**
- Identify all specifications within the designated phase scope before starting execution
- Process all specifications identified during work discovery

**Specification Compliance:**
- Comply with all referenced specifications
- Trace every implementation decision to a specification
- Validate output against specification acceptance criteria
- Report deviations or impossibilities rather than silently diverge

**Phase Documentation:**
- Write a phase report documenting phase outcomes, decisions, and deviations
- Include work discovery results in the phase report

**State Tracking:**
- Update processing state for all consumed specs after phase completion
- Update processing state for all upstream specs referenced for coherence

**Cross-Layer Consolidation:**
- Consolidate specs from multiple layers before implementation
- Verify specs across layers are compatible (coherence check)
- Identify contradictions between layers (conflict detection)
- Ensure all upstream requirements have corresponding downstream specs (gap analysis)

### Implementation-Extension MUST NOT Directives

**Specification Integrity:**
- Modify specification requirements or structure (request changes through proper channels)
- Implement features not defined in specifications
- Skip validation steps defined in coverage specifications

**Cross-Layer Conflicts:**
- Proceed with implementation while unresolved cross-layer conflicts exist

**Security:**
- Include secrets, passwords, API keys, tokens, or credentials in generated artifacts (use placeholder references like `${secrets.KEY_NAME}`)

**Phase Scope:**
- Execute work assigned to other phases
- Execute work assigned to specification agents

### Implementation-Extension SHOULD Directives

**Consolidation:**
- Consolidate duplicate implementation artifacts into shared components
- Refactor shared implementation concerns rather than duplicating code
- Request spec amendments when conflicts or gaps are discovered during consolidation

**Implementation Quality:**
- Follow industry standards for the chosen stack while satisfying spec-defined behavior, including folder structure conventions
- Ensure implementations are structurally recognizable and behaviorally equivalent to specs

### Scope Boundary Enforcement

When user requests out-of-phase work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — "[Phase] phase is [status]. To proceed with [requested work], invoke [target agent]."
3. **Suggest next step** — Provide the appropriate agent invocation command

---

## Compilation Guidance for Agent-L2

When compiling implementation agents for any workflow phase:

### Merging Role Content

Construct product agent Role section using Role Content Structure:

1. **Agent Identity**: Replace [PHASE_NAME] with the agent's phase name
2. **Goal**: State transformation from specifications to artifacts
3. **Phase Context**: State phase position in workflow and scope

**Purpose:** Role section establishes agent identity and workflow position upfront, preventing scope confusion in multi-phase execution.

**Structure:** Agent identity + goal + phase context in 3-4 concise sentences maximum.

### Merging Input Content

Construct product agent Input section using Input Content Structure:

1. **Upstream Specifications**: List upstream specifications consumed by this phase, with file paths
2. **User Input**: Describe what user may provide as phase-specific context
3. **Conflict Resolution**: Include standard conflict handling policy

**Purpose:** Input section documents all information sources the agent consumes, establishing clear data flow and conflict resolution behavior.

**Structure:** Three subsections (Upstream Specifications, User Input, Conflict Resolution) with bullet formatting for clarity.

### Merging Output Content

Construct product agent Output section using Output Content Structure:

1. **Artifacts**: List phase-specific output artifacts with file paths, including phase report
2. **Format**: State formatting requirements including phase report documentation requirements

**Purpose:** Output section specifies what the agent produces and where, establishing clear deliverables and documentation requirements.

**Structure:** Two subsections (Artifacts, Format) with phase report requirements MUST be included in both.

### Merging Implementation-Extension Directives

Implementation-extension directives apply to ALL implementation agents. Merge into product agent after base directives:

1. **MUST section** receives (after base directives):
   - Work Discovery directives (2 items)
   - Specification Compliance directives (4 items)
   - Phase Documentation directives (2 items)
   - State Tracking directives (2 items)
   - Cross-Layer Consolidation directives (4 items)

2. **MUST NOT section** receives (after base directives):
   - Specification Integrity directives (3 items)
   - Cross-Layer Conflicts directive (1 item)
   - Security directive (1 item)
   - Phase Scope directives (2 items)

3. **SHOULD section** receives (after base directives):
   - Consolidation directives (3 items)
   - Implementation Quality directives (2 items)

### Merging Cross-Layer Consolidation Content

Construct product agent Cross-Layer Consolidation section using Cross-Layer Consolidation Content:

1. **4-Step Workflow**: Insert coherence check → conflict detection → gap analysis → amendment request
2. **Directive**: Include MUST NOT proceed directive

**Purpose:** Cross-Layer Consolidation section ensures agents validate coherence across layers before implementation, preventing inconsistent artifacts.

**Structure:** Numbered 4-step workflow with single MUST NOT directive below.

### Merging Scope Boundaries Content

Construct product agent Scope Boundaries section using Scope Boundaries Content:

1. **MUST NOT Directives**: Insert phase scope restrictions
2. **Boundary Enforcement**: Insert 3-step pattern (stop → respond → suggest)

**Purpose:** Scope Boundaries section prevents agents from executing work outside their designated phase, maintaining workflow discipline.

**Structure:** MUST NOT subsection with restrictions, Boundary Enforcement subsection with 3-step pattern.

### Merging Phase-Specific Rules Content

If the user has created domain-specific compilation files via Agent-L1, Agent-L2 compiles [PHASE_SPECIFIC_RULES] by:

1. **Reading** the user-created `templates/agents/compiled/[phase].rules.md` file
2. **Applying** L0→L1 transformation rules documented in that compilation file
3. **Replacing** generic placeholders with concrete phase-specific values

If no domain-specific rules file exists, omit the Phase-Specific Rules section from the compiled agent.

**Purpose:** Phase-Specific Rules section serves as optional injection point for phase-unique directives beyond the generic implementation extension.

**Structure:** Phase-specific directives inserted directly (no placeholder in final agent).

### Merging State Tracking Content

Construct product agent State Tracking section using State Tracking Content:

1. **Processed Spec Updates**: Insert phase-specific state and timestamp directives
2. **Upstream Spec Updates**: Insert upstream spec tracking requirements
3. **Additional State Directives**: Include phase-specific state tracking rules from user-defined domain compilation files (if any)

**Purpose:** State Tracking section ensures all processed and referenced specs reflect phase progress, maintaining accurate workflow state.

**Structure:** Numbered steps with directive language; omit Additional State Directives subsection if no domain compilation file exists.

### Merging Completion Criteria Content

Construct product agent Completion Criteria section using Completion Criteria Content:

1. **Phase-Specific Completion Checks**: Insert 9 standard implementation completion checks
2. **Additional Phase Criteria**: Include phase-specific additional criteria from user-defined domain compilation files (if any)

**Purpose:** Completion Criteria section provides exhaustive checklist agents MUST validate before declaring phase completion, ensuring quality and completeness.

**Structure:** Checkbox list with standard 9 checks plus any phase-specific extensions.

### Merging Workflow Handover Content

Construct product agent Workflow Handover section using Workflow Handover Content:

1. **Pattern**: Insert next step proposal placeholder
2. **Replacement**: [PROPOSE_NEXT_STEP] replaced with phase-specific guidance from phase.rules.md

**Purpose:** Workflow Handover section guides users to the next logical step after phase completion, maintaining smooth workflow progression.

**Structure:** Single statement proposing next step or agent invocation.

### Merging Failure Handling Content

Extend the base Failure Handling Pattern by appending the cross-layer conflict row from the Failure Handling Extension:

1. **Base table** comes from base.rules.md Failure Handling Pattern (4 rows)
2. **Append** the cross-layer conflict row (1 row)

**Purpose:** Failure Handling section establishes clear agent behavior for error cases. The implementation extension adds cross-layer conflict handling on top of the generic base patterns.

**Structure:** 5-row Situation/Action table (4 base rows + 1 extension row), followed by "Stop iterating when:" list from base.

### Extension-Specific Directives

After merging base + implementation directives, optionally merge phase-specific directives from user-created compilation files:
- `compiled/[phase].rules.md` — created via Agent-L1 for phase-specific constraints

Phase directives ADD TO base + implementation directives, never replace them.
