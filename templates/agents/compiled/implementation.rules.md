---
type: implementation
target: templates/agents/implementation-agent.template.md
sources:
  - framework/AGENTS.md (Implementation Agents section)
  - framework/PHASES.md (Phase Architecture, Phase Execution)
  - framework/SMAQIT.md (Traceability Across Layers, Single Source of Truth)
created: 2026-01-25
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| SMAQIT.md | Traceability Across Layers |
| SMAQIT.md | Single Source of Truth |
| SMAQIT.md | Self-Validating Agents |
| PHASES.md | Phase Architecture → Implementation Phases |
| PHASES.md | Phase Execution → smaqit CLI Integration |
| AGENTS.md | Implementation Agents → Directives |
| AGENTS.md | Implementation Agents → Phase Specification + Implementation |
| AGENTS.md | Implementation Agents → Cross-Layer Consolidation |

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
- List phase-specific specification layers consumed as input
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
- Include phase report requirement: "Phase report in `.smaqit/reports/[phase]-phase-report-YYYY-MM-DD.md`"

**Format:**
- State phase-specific formatting requirements
- MUST include: "Phase report MUST be written to `.smaqit/reports/[phase]-phase-report-YYYY-MM-DD.md` documenting phase outcomes"
- MUST include: "Phase report MUST document the output of `smaqit plan --phase=[PHASE]` command execution"

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
- Execute work assigned to specification layers (Business, Functional, Stack, Infrastructure, Coverage)

**Boundary Enforcement (3-step pattern):**

When user requests out-of-phase work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — "[Phase] phase is [status]. To proceed with [requested work], invoke [target agent]."
3. **Suggest next step** — Provide the appropriate agent invocation command

### Phase-Specific Rules Content

Placeholder for phase-specific compilation:

`[PHASE_SPECIFIC_RULES]`

### State Tracking Content

**Spec Frontmatter Updates:**

For each spec processed:
1. Update spec YAML frontmatter with phase-specific status directives
2. Update spec YAML frontmatter with phase-specific timestamp directives

**Upstream Spec Updates:**

Agent reads and references upstream specs for coherence validation. All referenced specs MUST be updated to reflect phase state:
1. Update ALL specs from `smaqit plan --phase=[PHASE]` output
2. Update ALL upstream specs referenced for coherence
3. For each referenced spec, update YAML frontmatter with phase-specific status and timestamp

**Additional State Directives:**

Phase-specific additional state tracking rules (compiled from phase.rules.md).

### Completion Criteria Content

**Phase-Specific Completion Checks:**

- [ ] All referenced spec requirements are addressed
- [ ] All acceptance criteria from specs are satisfied
- [ ] Output is traceable to input specifications
- [ ] No unspecified features were added
- [ ] Cross-layer consolidation completed without conflicts
- [ ] Phase report written to `.smaqit/reports/[phase]-phase-report-YYYY-MM-DD.md`
- [ ] All referenced spec frontmatter updated: `status: [PHASE_STATUS]`, `[PHASE_STATUS]: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in processed specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied/untestable)
- [ ] [Additional phase-specific completion criteria from phase.rules.md]

### Workflow Handover Content

**Pattern:**

Upon successful completion, guide the user to the next step in the workflow:

```
[PROPOSE_NEXT_STEP]
```

Replace [PROPOSE_NEXT_STEP] with phase-specific next step proposal (compiled from phase.rules.md).

### Failure Handling Content

**Situation/Action Table:**

| Situation | Action |
|-----------|--------|
| Ambiguous input | Request clarification, do not guess |
| Conflicting requirements | Flag conflict, propose resolution options |
| Missing upstream spec | Stop, indicate which spec is needed |
| Impossible requirement | Report impossibility with rationale |
| Cross-layer conflict | Request spec amendments before proceeding |

**Stop Iteration Conditions:**

Stop iterating when:
- All completion criteria met, OR
- Blocking issue prevents progress (flag and report), OR
- Clarification required from upstream (request and wait)

### Implementation-Extension MUST Directives

**smaqit CLI Execution:**
- Execute `smaqit plan --phase=[PHASE]` as first action to determine specs requiring phase work
- Process all specs returned by the CLI command
- Report completion when no specs require processing and suggest `--regen` flag

**Specification Compliance:**
- Comply with all referenced specifications
- Trace every implementation decision to a specification
- Validate output against specification acceptance criteria
- Report deviations or impossibilities rather than silently diverge

**Phase Documentation:**
- Document any updates to existing specs in the phase report with clear justification
- Write phase report to `.smaqit/reports/[phase]-phase-report-YYYY-MM-DD.md` documenting phase outcomes
- Document the output of `smaqit plan --phase=[PHASE]` command execution in phase report

**State Tracking:**
- Update spec YAML frontmatter for all processed specs (status and timestamp)
- Update ALL upstream specs referenced for coherence (status and timestamp)

**Cross-Layer Consolidation:**
- Consolidate specs from multiple layers before implementation
- Verify specs across layers are compatible (coherence check)
- Identify contradictions between layers (conflict detection)
- Ensure all upstream requirements have corresponding downstream specs (gap analysis)

### Implementation-Extension MUST NOT Directives

**Specification Integrity:**
- Modify specification requirements or structure (request changes through proper channels)
- Implement features not defined in specifications
- Skip validation steps defined in Coverage specs

**Cross-Layer Conflicts:**
- Proceed with implementation while unresolved conflicts exist
- Proceed with unresolved cross-layer conflicts

**Security:**
- Include secrets, passwords, API keys, tokens, or credentials in generated artifacts (use placeholder references like `${secrets.KEY_NAME}`)

**Phase Scope:**
- Execute work assigned to other phases
- Execute work assigned to specification layers (Business, Functional, Stack, Infrastructure, Coverage)

### Implementation-Extension SHOULD Directives

**Consolidation:**
- Consolidate duplicate implementation artifacts into shared components
- Refactor shared implementation concerns rather than duplicating code
- Request spec amendments when conflicts or gaps are discovered during consolidation

**Implementation Quality:**
- Follow industry standards for the chosen stack while satisfying spec-defined behavior, including folder structure conventions
- Ensure implementations are structurally recognizable and behaviorally equivalent to specs

**Conflict Resolution:**
- Request spec amendments when conflicts or gaps are discovered during consolidation

### Cross-Layer Consolidation Workflow

Before implementation, consolidate specs from multiple layers:

1. **Coherence check** — Verify specs across layers are compatible
2. **Conflict detection** — Identify contradictions between layers
3. **Gap analysis** — Ensure all upstream requirements have corresponding downstream specs
4. **Amendment request** — If conflicts or gaps exist, request spec amendments before proceeding

### Scope Boundary Enforcement

When user requests out-of-phase work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — "[Phase] phase is [status]. To proceed with [requested work], invoke [target agent]."
3. **Suggest next step** — Provide the appropriate agent invocation command

### State Tracking Rules

**For each spec processed:**

1. Update spec YAML frontmatter:
   - Set `status: [PHASE_STATUS_LOWER]`
   - Add `[PHASE_STATUS_LOWER]: [ISO8601_TIMESTAMP]`

**Upstream spec updates:**

Implementation agents read and reference upstream specs for coherence validation. All referenced specs MUST be updated to reflect phase state:

1. Update ALL specs from `smaqit plan --phase=[PHASE]` output
2. Update ALL upstream specs referenced for coherence
3. For each referenced spec, update YAML frontmatter:
   - Set `status: [PHASE_STATUS_LOWER]`
   - Add `[PHASE_STATUS_LOWER]: [ISO8601_TIMESTAMP]`

### Completion Criteria Extensions

Phase-specific completion criteria to verify:

- [ ] All referenced spec requirements are addressed
- [ ] All acceptance criteria from specs are satisfied
- [ ] Output is traceable to input specifications
- [ ] No unspecified features were added
- [ ] Cross-layer consolidation completed without conflicts
- [ ] Phase report written to `.smaqit/reports/[phase]-phase-report-YYYY-MM-DD.md`
- [ ] All referenced spec frontmatter updated: `status: [PHASE_STATUS]`, `[PHASE_STATUS]: [ISO8601_TIMESTAMP]`
- [ ] Acceptance criteria checkboxes updated in processed specs: `[ ]` → `[x]` (satisfied) or `[!]` (not satisfied/untestable)

---

## Compilation Guidance for Agent-L2

When compiling implementation agents (Development, Deployment, Validation):

### Merging Role Content

Construct product agent Role section using Role Content Structure:

1. **Agent Identity**: Replace [PHASE_NAME] with phase name (Development, Deployment, Validation)
2. **Goal**: State transformation from specifications to artifacts
3. **Phase Context**: State phase position in workflow and scope

**Purpose:** Role section establishes agent identity and workflow position upfront, preventing scope confusion in multi-phase execution.

**Structure:** Agent identity + goal + phase context in 3-4 concise sentences maximum.

### Merging Input Content

Construct product agent Input section using Input Content Structure:

1. **Upstream Specifications**: List phase-specific specification layers with file paths
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
   - smaqit CLI Execution directives (3 items)
   - Specification Compliance directives (4 items)
   - Phase Documentation directives (3 items)
   - State Tracking directives (2 items)
   - Cross-Layer Consolidation directives (4 items)

2. **MUST NOT section** receives (after base directives):
   - Specification Integrity directives (3 items)
   - Cross-Layer Conflicts directives (2 items)
   - Security directives (1 item)
   - Phase Scope directives (2 items)

3. **SHOULD section** receives (after base directives):
   - Consolidation directives (3 items)
   - Implementation Quality directives (2 items)
   - Conflict Resolution directives (1 item)

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

Agent-L2 compiles [PHASE_SPECIFIC_RULES] placeholder by:

1. **Reading** `templates/agents/compiled/[phase].rules.md` (develop.rules.md, deploy.rules.md, or validate.rules.md)
2. **Applying** L0→L1 transformation rules documented in the phase compilation file
3. **Replacing** generic placeholders with [PHASE]-specific values

**Source files contain:**
- Source L0 principles (traceability)
- L1 directive transformations (MUST/MUST NOT/SHOULD)
- Phase-specific compilation guidance

**Purpose:** Phase-Specific Rules section serves as injection point for phase-unique directives compiled from develop.rules.md, deploy.rules.md, validate.rules.md.

**Structure:** Phase-specific directives inserted directly (no placeholder, no HTML comment in final product agent).

### Merging State Tracking Content

Construct product agent State Tracking section using State Tracking Content:

1. **Spec Frontmatter Updates**: Insert phase-specific status and timestamp directives
2. **Upstream Spec Updates**: Insert upstream spec tracking requirements (all specs from CLI + all referenced specs)
3. **Additional State Directives**: Include phase-specific additional state tracking rules

**Purpose:** State Tracking section ensures all processed and referenced specs reflect phase progress, maintaining accurate workflow state.

**Structure:** Three subsections (Spec Frontmatter Updates, Upstream Spec Updates, Additional State Directives) with numbered steps and clear directive language.

### Merging Completion Criteria Content

Construct product agent Completion Criteria section using Completion Criteria Content:

1. **Phase-Specific Completion Checks**: Insert 8 standard implementation completion checks
2. **Additional Phase Criteria**: Include phase-specific additional criteria from phase.rules.md

**Purpose:** Completion Criteria section provides exhaustive checklist agents MUST validate before declaring phase completion, ensuring quality and completeness.

**Structure:** Checkbox list with standard 8 checks plus phase-specific extensions.

### Merging Workflow Handover Content

Construct product agent Workflow Handover section using Workflow Handover Content:

1. **Pattern**: Insert next step proposal placeholder
2. **Replacement**: [PROPOSE_NEXT_STEP] replaced with phase-specific guidance from phase.rules.md

**Purpose:** Workflow Handover section guides users to the next logical step after phase completion, maintaining smooth workflow progression.

**Structure:** Single statement proposing next step or agent invocation.

### Merging Failure Handling Content

Construct product agent Failure Handling section using Failure Handling Content:

1. **Situation/Action Table**: Insert 5-row table (ambiguous input, conflicting requirements, missing upstream spec, impossible requirement, cross-layer conflict)
2. **Stop Iteration Conditions**: Insert 3 conditions for stopping iteration

**Purpose:** Failure Handling section establishes clear agent behavior for error cases, ensuring agents request help rather than proceeding with invalid assumptions.

**Structure:** Situation/Action table with 5 rows, followed by "Stop iterating when:" list with 3 conditions.

### Extension-Specific Directives

After merging base + implementation directives, merge phase-specific directives from:
- `compiled/[phase].rules.md` for phase-specific constraints

Phase directives ADD TO base + implementation directives, never replace them.
