# Agents

Agents are LLM-powered actors that operate within the smaqit framework. This document defines the principles, constraints, and behaviors that all agents MUST follow.

## Unified Principles

All smaqit agents—specification and implementation—share these foundational principles:

### Prompt Interaction

Agents receive requirements from prompts in `.github/prompts/`:

- **Read prompt files**: Agents MUST read corresponding prompt files (`.github/prompts/smaqit.[layer].prompt.md` for specification agents, phase-specific prompts for implementation agents)
- **Ignore HTML comments**: Agents MUST ignore all HTML comments (`<!-- ... -->`) in prompt files to prevent example requirements from contaminating specifications
- **Interpret free-style input**: Agents consume natural language requirements without rigid structure enforcement
- **Validate sufficiency**: Agents MUST request clarification if prompt content is insufficient, using natural language guidance (e.g., "Please specify measurable success criteria" not "Missing: Success Metrics section")
- **Equivalent outcomes**: Given the same prompt set across all layers, acceptance criteria should pass/fail consistently (acknowledging LLM variance in artifact style)

See [PROMPTS](PROMPTS.md) for complete prompt architecture and input record principles.

### Template-Constrained Output
- Agents MUST produce output following their designated template
- Agents MUST NOT add sections not defined in the template
- Agents MUST NOT omit required sections from the template

### Traceable References
- Agents MUST reference their input sources explicitly
- Agents SHOULD use consistent reference format: `[LayerName](path/to/spec.md)`
- Agents MUST NOT produce output that cannot be traced to an input

### Fail-Fast on Ambiguity
- Agents request clarification when input is ambiguous
- Agents do not invent requirements not present in input
- Agents flag assumptions explicitly when clarification is unavailable
- Agents invoke assessment skill when detecting ambiguity requiring iterative analysis and planning

### Fail-Fast on Inconsistency
- Agents MUST verify coherence across all input sources before producing output
- Agents MUST stop and report when inputs contradict each other
- Agents MUST NOT proceed with output while unresolved inconsistencies exist

### Self-Validation Before Completion
- Agents MUST validate their output against completion criteria before finishing
- Agents MUST NOT declare completion if any required criterion is unmet
- Agents SHOULD iterate on output until validation passes

### Scope Boundaries

Each agent has a single responsibility defined by its layer or phase.

**Agents MUST NOT:**
- Execute work assigned to other phases
- Execute work assigned to other layers (for specification agents)
- Execute work assigned to other agents

**Boundary Enforcement:**

When user requests out-of-scope work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — State current scope and required agent for requested work
3. **Suggest next step** — Provide prompt file or agent invocation command

### Extensible Agents

All agents share foundational behaviors codified in base template principles. These foundations—bounded scope, self-validation, fail-fast on ambiguity, template-constrained output, traceable references—apply regardless of agent purpose.

Agent extensions inherit these foundations and add specialized behaviors. The base template documents what must remain invariant. Extension templates document what varies by role.

### Agent Skills

Agents invoke skills when detecting specific conditions requiring specialized handling. Skills are reusable capabilities providing structured workflows that multiple agent types share.

Skills reside in `.github/skills/` and include assessment, conflict resolution, gap detection, and other cross-cutting workflows. When agents detect ambiguity, contradictions, or complexity, they invoke the appropriate skill rather than implementing the workflow inline.

## Naming Convention

Agents follow the pattern: `smaqit.[LAYER]` for specification agents and `smaqit.[PHASE]` for implementation agents.

| Type | Pattern | Examples |
|------|---------|----------|
| Specification | `smaqit.[LAYER]` | `smaqit.business`, `smaqit.functional`, `smaqit.stack` |
| Implementation | `smaqit.[PHASE]` | `smaqit.development`, `smaqit.deployment`, `smaqit.validation` |

## Foundation Agent

The foundation agent represents what all smaqit agents must preserve regardless of their specific purpose. These guidelines apply universally and are materialized through the base agent template.

### Core Behaviors

All agents inherit these foundational behaviors from the Unified Principles:
- Template-constrained output structure
- Traceable references to input sources
- Fail-fast on ambiguity or inconsistency
- Self-validation before completion
- Bounded scope with clear boundary enforcement

### Extensibility Through Templates

Foundation behaviors materialize through shared template structures. What all agents must preserve becomes part of a base template. What differentiates agents becomes part of specialized templates that build upon the base.

This ensures consistency across agent types while enabling role-specific behaviors to emerge naturally from the template hierarchy.

## Agent Extensions

Agent extensions inherit foundation behaviors and add specialized directives for specific roles. Current extensions include specification agents that extend base with layer-specific directives and implementation agents that extend base with phase-specific execution. Future agent types extend base with their own specializations.

### Specification Agents

Specification agents translate prompt file requirements into precise, testable specifications for a single layer.

### Role Architecture

Each specification agent's Role section MUST include:

1. **Agent identity** — Direct statement: "You are now operating as the [Layer] Agent"
2. **Goal** — What this agent produces and from what input
3. **Context** — Single statement covering layer position and upstream relationship

**Purpose:** Role section establishes agent identity and boundaries upfront, preventing scope confusion and context pollution in multi-agent workflows.

**Structure:** Agent identity + goal + context in 3-4 concise sentences maximum.

### Input
- **Prompt file**: Requirements from `.github/prompts/smaqit.[layer].prompt.md` (the primary source)
- **Context specifications**: Documents from previous layers for coherence and traceability (not requirements)

Each layer reads from its own prompt file. Upstream layers provide context for coherence, not requirements. When prompt requirements would create incoherence with existing specs, agents MUST flag the conflict rather than silently override.

### Output
- Specification documents in `specs/{layer}/`
- Documents MUST follow `templates/{layer}.template.md`

### Directives

**Specification agents MUST:**
- Produce one specification file per distinct concept (e.g., one use case, one API contract)
- Generate YAML frontmatter with required fields: `id`, `status: draft`, `created`, `prompt_version`
- Capture git commit hash of prompt file at generation time for `prompt_version` field
- Include testable acceptance criteria in every specification
- Reference context specs used for coherence and traceability
- Validate output against layer template before completion
- Check for existing specs in the same layer before creating new specs

**Specification agents MUST NOT:**
- Include implementation details (code, technology choices outside Stack layer)
- Create inconsistencies with context layer specifications
- Produce specs for layers outside their scope
- Duplicate information present in existing specs

**Specification agents SHOULD:**
- Define explicit scope boundaries (what is included vs. excluded)
- Use consistent terminology across layers
- Flag potential inconsistencies with context specs
- Update existing specs when adding to an existing concept (e.g., adding feature to existing app)
- Create new specs only for distinct new concepts (e.g., separate service/component)
- Reference existing specs for shared information using Foundation Reference (same-layer) or Implements/Enables (upstream)

### Incremental Spec Updates vs New Specs

When users add requirements that could extend existing specifications, agents decide whether to update existing specs or create new ones:

| Scenario | Action | Rationale |
|----------|--------|-----------|
| **Feature extends existing concept** | Update existing spec | Consolidates related requirements, maintains single source of truth |
| **Feature is distinct new concept** | Create new spec with Foundation Reference | Preserves separation of concerns, references shared requirements |
| **Shared infrastructure/base requirements** | Create foundation spec, reference from feature specs | Avoids conflicting sources of truth |
| **Uncertainty** | Favor updating existing spec | Prevents duplication, easier to refactor later if needed |

**Examples:**

| Requirement | Existing Spec | Decision | Foundation Reference Pattern |
|-------------|---------------|----------|------------------------------|
| Add argparse CLI to Python console app | `python-console-stack.md` exists | **Update** existing spec | N/A (same spec) |
| Add authentication service to app | `app-stack.md` exists | **Create** `auth-service-stack.md` | Reference `[STK-APP](./app-stack.md)` for base requirements |
| Add logging to existing feature | `feature-functional.md` exists | **Update** existing spec | N/A (same spec) |

### Specification Agent Mappings

| Agent | Layer | Prompt File | Context (for coherence) | Output |
|-------|-------|-------------|---------------------------|--------|
| `smaqit.business` | Business | `smaqit.business.prompt.md` | None | `specs/business/*.md` |
| `smaqit.functional` | Functional | `smaqit.functional.prompt.md` | Business specs | `specs/functional/*.md` |
| `smaqit.stack` | Stack | `smaqit.stack.prompt.md` | Business and Functional specs | `specs/stack/*.md` |
| `smaqit.infrastructure` | Infrastructure | `smaqit.infrastructure.prompt.md` | Phase 1 specs | `specs/infrastructure/*.md` |
| `smaqit.coverage` | Coverage | `smaqit.coverage.prompt.md` | All layer specs | `specs/coverage/*.md` |

## Implementation Agents

Implementation agents transform specifications into working software, deployed systems, or validated results.

### Role Architecture

Each implementation agent's Role section MUST include:

1. **Agent identity** — Direct statement: "You are now operating as the [Phase] Agent"
2. **Goal** — What this agent produces and from what input
3. **Phase context** — Single statement covering phase position in workflow and scope

**Purpose:** Role section establishes agent identity and workflow position upfront, preventing scope confusion in multi-phase execution.

**Structure:** Agent identity + goal + phase context in 3-4 concise sentences maximum.

### Phase Orchestration

Implementation agents orchestrate their entire phase including specification generation and implementation execution. Phase orchestration involves coordinating multiple agents, tracking dependencies, managing workflow state, and handling errors across agent invocations.

**Specification Generation Coordination:**

Implementation agents coordinate specification generation when upstream artifacts are missing or regeneration is requested:
- **Artifact detection** — Agents check for required specification artifacts before beginning implementation
- **Agent invocation** — Missing specifications trigger invocation of specification agents
- **Dependency ordering** — Specification agents invoked in dependency sequence based on artifact requirements
- **Generation completion** — Specification generation completes before implementation activities begin

**Multi-Agent Coordination:**

Phase orchestration coordinates multiple agents within a single phase:
- Agents invoked as needed based on artifact availability
- Invoked agents produce outputs consumed by orchestrating agent  
- Invocation sequence respects upstream dependencies
- Each invocation tracked with context and outcome

**Progress Tracking:**

Orchestration progress tracked across activities and agent invocations:
- Each workflow activity reports start, progress, and completion
- Agent invocations logged with input context and output status
- Activity milestones visible to user during execution
- Workflow state persists across activities for traceability

**Error Context Preservation:**

Errors propagate with context across orchestration boundaries:
- Failed activities report diagnostic information with execution context
- Agent invocation failures include agent identity and input state
- Error messages provide remediation guidance
- Partial completion tracked when workflow halts mid-execution

**Phase Workflow Activities:**

Phase workflows contain distinct activities that execute in sequence:
- **Pre-orchestration validation** — Input sources and dependencies verified for readiness
- **Specification generation** — Specification agents invoked when upstream artifacts missing
- **Artifact consolidation** — Multiple specification sources merged and checked for coherence
- **Implementation generation** — Output artifacts produced from consolidated specifications
- **Execution** — Generated artifacts executed or deployed in target environment
- **Orchestration completion validation** — Outcomes verified against expected criteria

### Pre-Orchestration Validation

Validation precedes workflow execution to ensure readiness. Implementation agents verify input sufficiency, dependency availability, and execution prerequisites before beginning workflow activities.

**Input Validation:**

Input sources undergo validation before workflow begins:
- **Sufficiency check** — Required input content exists and contains necessary information
- **Format verification** — Input structure matches expected patterns
- **Completeness assessment** — All mandatory input elements present

Validation failures produce guidance describing what's missing or incorrect.

**Dependency Verification:**

Upstream artifacts and dependencies verified for accessibility:
- **Existence check** — Referenced artifacts present in expected locations
- **State verification** — Upstream artifacts in appropriate lifecycle state
- **Version consistency** — Input versions align across dependencies

Missing dependencies halt execution with clear identification of gaps.

**Execution Readiness:**

Execution environment verified before workflow activities begin:
- **Tool availability** — Required execution tools present and accessible
- **Permission verification** — Agent has necessary permissions for planned operations
- **Resource checks** — Sufficient resources available for workflow activities

Readiness failures prevent workflow initiation with actionable remediation steps.

**Validation Outcomes:**

Pre-orchestration validation produces binary outcome:
- **Pass** — All checks satisfied, workflow proceeds
- **Fail** — One or more checks failed, workflow halts with diagnostic report

Failed validations include specific remediation guidance for each failed check.

### Orchestration Completion Validation

Orchestration completion validation verifies that all activities executed successfully and produced expected outcomes. Implementation agents validate completion before declaring phase success.

**Activity Completion Verification:**

Each workflow activity verified for successful completion:
- **Completion status** — Activity reached completion state without errors
- **Output presence** — Activity produced expected output artifacts
- **Output validity** — Produced artifacts meet structural and content requirements

Incomplete activities identified with execution context and failure reason.

**Outcome Validation:**

Workflow outcomes validated against expected criteria:
- **Acceptance criteria** — Generated artifacts satisfy specified acceptance criteria
- **Behavioral correctness** — Execution outcomes match expected behavior
- **State consistency** — Artifact state reflects successful orchestration completion

Outcome validation failures include specific criteria not met and diagnostic information.

**Completion Outcomes:**

Orchestration completion validation produces status with context:
- **Success** — All activities completed, outcomes validated, phase complete
- **Partial** — Some activities completed, workflow halted mid-execution
- **Failed** — Workflow failed with error context and attempted remediation

Completion status includes detailed report of activity outcomes and validation results.

### Input
- Specification documents from relevant layers
- Existing codebase (for Development agent)
- Deployed system (for Validation agent)

### Output
- **Development**: Source code, configurations, build artifacts
- **Deployment**: Running infrastructure, deployed applications
- **Validation**: Test results, validation report with spec coverage percentage and unverified requirements

### Directives

**Implementation agents MUST:**
- Determine which specs to process using `smaqit plan --phase=[PHASE]` (outputs spec file paths, one per line)
- Process only specs with `status: draft` or `status: failed` by default
- Support regeneration mode via `--regen` flag to process all specs regardless of status
- Report completion when no specs require processing and suggest `--regen` flag if appropriate
- Comply with all referenced specifications
- Trace every implementation decision to a specification
- Validate output against specification acceptance criteria
- Report deviations or impossibilities rather than silently diverge
- Update spec frontmatter status and timestamps during processing

**Frontmatter tracking:**

| Agent | Updates Spec Frontmatter |
|-------|--------------------------|
| Development | `status: implemented` or `failed`<br>`implemented: [ISO8601_TIMESTAMP]` |
| Deployment | `status: deployed` or `failed`<br>`deployed: [ISO8601_TIMESTAMP]` |
| Validation | `status: validated` or `failed`<br>`validated: [ISO8601_TIMESTAMP]`<br>Update checkboxes: `[ ]` → `[x]` or `[!]` |

The CLI aggregates phase status by scanning spec frontmatter. Agents only update individual spec files.

**Implementation agents MUST NOT:**
- Modify specifications (request changes through proper channels)
- Implement features not defined in specifications
- Skip validation steps defined in Coverage specs
- Write state updates before all completion criteria are satisfied

**Implementation agents SHOULD:**
- Prefer explicit over implicit behavior
- Document assumptions when specs are underspecified
- Request spec clarification before inventing solutions

### Cross-Layer Consolidation
Implementation agents receive specs from multiple layers and MUST consolidate them before implementation:

1. **Coherence check**: Verify specs across layers are compatible
2. **Conflict detection**: Identify contradictions between layers
3. **Gap analysis**: Ensure all upstream requirements have corresponding downstream specs
4. **Amendment request**: If conflicts or gaps exist, request spec amendments before proceeding

### Tooling

Implementation agents require execution capabilities that specification agents do not need.

| Agent Type | Tools | Rationale |
|------------|-------|-----------|
| Specification | `edit`, `search`, `usages`, `fetch`, `todos` | Produce documents only |
| Implementation | `edit`, `search`, `runCommands`, `problems`, `changes`, `testFailure`, `todos`, `runTests` | Build, run, test applications |

**Tool descriptions:**

| Tool | Purpose |
|------|---------|
| `edit` | Create and modify files |
| `search` | Search codebase and specifications |
| `usages` | Find code references and usages |
| `fetch` | Fetch web content |
| `todos` | Track multi-step task progress |
| `runCommands` | Run terminal commands (build, test, deploy) |
| `problems` | Get compilation and lint errors |
| `changes` | Get git diffs and file changes |
| `testFailure` | Get test failure information |
| `runTests` | Execute unit tests |
| `runSubagent` | Invoke other agents

Agents MUST NOT proceed with implementation while unresolved conflicts exist.

### Implementation Agent Mappings

| Agent | Phase | Input | Output |
|-------|-------|-------|--------|
| `smaqit.development` | Develop | Business + Functional + Stack specs | Code |
| `smaqit.deployment` | Deploy | Code + Infrastructure specs | Running system |
| `smaqit.validation` | Validate | Deployed system + Coverage specs | Validation report |

## Validation

All agents perform self-validation before declaring completion. This section defines the validation requirements.

### Self-Validation Loop

```
1. Produce output following template
2. Check output against completion criteria
3. If criteria unmet → iterate on output
4. If criteria met → declare completion
5. If criteria impossible → flag blocker and stop
```

### Completion Criteria

Agents MUST verify these conditions before completing:

**For Specification Agents:**
- [ ] All template sections are filled (no placeholders remain)
- [ ] All upstream references are valid and accessible
- [ ] All acceptance criteria are testable (measurable, observable)
- [ ] Scope boundaries are explicitly stated
- [ ] No implementation details leaked into spec

**For Implementation Agents:**
- [ ] All referenced spec requirements are addressed
- [ ] All acceptance criteria from specs are satisfied
- [ ] Output is traceable to input specifications
- [ ] No unspecified features were added
- [ ] Cross-layer consolidation completed without conflicts (Development, Deployment)
- [ ] Spec coverage % reported with unverified requirements identified (Validation)

### Quality Boundary

Agents MUST stop iterating when:
- All completion criteria are met, OR
- A blocking issue prevents progress (flag and report), OR
- Clarification is required from upstream (request and wait)

Agents MUST NOT:
- Iterate indefinitely without progress
- Lower quality standards to force completion
- Invent solutions to bypass blockers

### Failure Modes

When an agent cannot complete:

| Situation | Action |
|-----------|--------|
| Ambiguous input | Request clarification, do not guess |
| Conflicting requirements | Flag conflict, propose resolution options |
| Missing upstream spec | Stop, indicate which spec is needed |
| Impossible requirement | Report impossibility with rationale |