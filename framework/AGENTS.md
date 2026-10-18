# Agents

Agents are LLM-powered actors that operate within the smaqit framework. This document defines the principles, constraints, and behaviors that all agents MUST follow.

## Unified Principles

All smaqit agents—specification and implementation—share these foundational principles:

### Session Context

Agents receive requirements from the user, through session context:

- **Read session context**: Agents MUST read requirements from current session context (including context in compacted blocks) or open tasks
- **User-stated provenance**: A requirement counts only when the user stated it: their request (passed verbatim as `user_request`), their `answers`, their own messages and the open tasks they wrote, or files they pointed at, such as `.smaqit/requirements-extract.md`. A project, repository, or directory name, the agent's own domain knowledge, and a caller's summary, paraphrase, or expansion of the request never satisfy a required section. In `autonomous` mode, upstream specifications also count as requirement sources; in `assisted` mode, every required section traces to the user's own statements and upstream specifications serve coherence only
- **Apply input skill**: Specification agents MUST apply their layer-specific `smaqit.input-[layer]` skill as a gate that checks requirements are sufficient before generating specifications. Phase agents apply their phase-specific `smaqit.input-[phase]` skill to resolve any execution parameters the caller did not pass (see Phase Parameters). The input skill MUST be reachable by the agent through its platform's skill mechanism, preloaded or invocable. An instruction to apply a skill the agent cannot reach is a defect, not a step to improvise.
- **Interpret free-style input**: Agents consume natural language requirements without rigid structure enforcement
- **Validate sufficiency**: Agents MUST request clarification if session context is insufficient, directly when they can reach the user and through a `NEEDS-INPUT` return signal otherwise (see User Interaction). They use natural language guidance (e.g., "Please specify measurable success criteria", not "Missing: Success Metrics section")
- **Equivalent outcomes**: Given equivalent session context across all layers, acceptance criteria should pass/fail consistently (acknowledging LLM variance in artifact style)

### Template-Constrained Output
- Agents MUST produce output following their designated template
- Agents MUST NOT add sections not defined in the template
- Agents MUST NOT omit required sections from the template

### Traceable References
- Agents MUST reference their input sources explicitly
- Agents SHOULD use consistent reference format: `[LayerName](path/to/spec.md)`
- Agents MUST NOT produce output that cannot be traced to an input

### Fail-Fast on Ambiguity
- Agents request clarification when input is ambiguous, directly or through `NEEDS-INPUT` (see User Interaction)
- Agents do not invent requirements not present in input
- Agents never fill a required section with an assumption: when the user's statements leave it insufficient, they return `NEEDS-INPUT`. Clarification is always available through that signal
- Agents surface the specific ambiguity, state what information is missing or contradictory, and request clarification before proceeding

### Fail-Fast on Inconsistency
- Agents MUST verify coherence across all input sources before producing output
- Agents MUST stop and report when inputs contradict each other
- Agents MUST NOT proceed with output while unresolved inconsistencies exist

### User Interaction

Agents often run as subagents that cannot address the user directly. Questions and review requests therefore travel as structured return signals, up to whichever caller can reach the user:

- **Ask when reachable** — An agent that has a user-question capability on its platform asks the user directly
- **Return when not** — An agent without one stops and returns a signal. It does not guess, silently default a blocking value, or produce partial output
- **Write nothing on `NEEDS-INPUT`** — An agent that returns `NEEDS-INPUT` has created or modified no specification, design, or implementation artifact in that invocation, so invoking it again with the answers is always safe
- **Pass signals up unchanged** — An orchestrating agent that receives a signal from an agent it invoked, and cannot reach the user itself, returns that signal to its own caller without answering it
- **The caller owns the loop** — The caller that can reach the user asks, then resumes the returning agent with the answers. If its platform cannot resume a returned agent, it invokes the agent again with the same parameters plus `answers` and `approved_spec_paths`. It repeats until a final result arrives
- **Self-describing signals** — Every signal ends with a one-line instruction to its receiver, so a caller that did not anticipate the signal still knows how to handle it

| Signal | Returned when | Contents |
|--------|---------------|----------|
| `NEEDS-INPUT` | An input gate, or a blocking ambiguity, needs information only the user can give | Source agent or skill; layer or phase; numbered questions, each with an id, the question, and why it is needed; a statement that nothing was written |
| `REVIEW-REQUESTED` | Assisted mode, after a specification layer completes | Layer; the specification and design paths produced; iteration `n/3` |

### Self-Validation Before Completion
- Agents MUST validate their output against completion criteria before finishing
- Agents MUST NOT declare completion if any required criterion is unmet
- Agents SHOULD iterate on output until validation passes

### Mandatory Visual Design Gate

Specification agents produce canonical PlantUML/PNG pairs for their layer. Before completion, specification agents MUST:

1. invoke the PlantUML MCP syntax tool;
2. render or refresh the PNG through the smaqit design command;
3. open the PNG through the host image-reading capability;
4. apply the standard visual rubric;
5. stop on any deterministic, render, or visual failure.

A specification agent that cannot obtain visual image content MUST stop with `DESIGN-VISION-UNAVAILABLE`. It MUST NOT inspect PlantUML source as a substitute for authoring-time visual review.

Implementation agents do not repeat this visual gate. `smaqit plan --phase=[PHASE]` automatically verifies that every in-scope specification has a current, visually attested design pair before returning implementation work. After that gate passes, implementation agents consume specification Markdown for requirements and PlantUML source for canonical design structure. They MUST NOT author, render, visually review, attest, or repair designs.

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
3. **Suggest next step** — Provide agent invocation command

### Extensible Agents

All agents share foundational behaviors codified in base template principles. These foundations—bounded scope, self-validation, fail-fast on ambiguity, template-constrained output, traceable references—apply regardless of agent purpose.

Agent extensions inherit these foundations and add specialized behaviors. The base template documents what must remain invariant. Extension templates document what varies by role.

### Agent Skills

Agents invoke skills when detecting specific conditions requiring specialized handling. Skills are reusable capabilities providing structured workflows that multiple agent types share.

Skills are installed once per user in the platform's global skill directory and include input validation, conflict resolution, and other cross-cutting workflows. When agents detect ambiguity, contradictions, or complexity, they apply the appropriate skill rather than implementing the workflow inline.

An agent can follow only a skill it can reach. Each agent's platform metadata makes its own input skill available to it, for example by preloading the skill into the agent at startup. Build tooling verifies that every input skill an agent's instructions name is reachable from that agent.

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

Specification agents translate session context requirements into precise, testable specifications for a single layer.

### Role Architecture

Each specification agent's Role section MUST include:

1. **Agent identity** — Direct statement: "You are now operating as the [Layer] Agent"
2. **Goal** — What this agent produces and from what input
3. **Context** — Single statement covering layer position and upstream relationship

**Purpose:** Role section establishes agent identity and boundaries upfront, preventing scope confusion and context pollution in multi-agent workflows.

**Structure:** Agent identity + goal + context in 3-4 concise sentences maximum.

### Input
- **Session context**: Requirements the user stated: `user_request` (the user's own words, verbatim), user input in chat, compacted context blocks, open tasks, and files the user pointed at (see Session Context: User-stated provenance)
- **Mode**: The `mode` the caller passed (`assisted` when none was passed). It decides whether upstream specifications may serve as requirement sources
- **Context specifications**: Documents from previous layers for coherence and traceability (not requirements)
- **Input skill**: Each specification agent applies `smaqit.input-[layer]` as a validation gate before spec generation begins. When the gate finds gaps that the agent cannot resolve with the user directly, the agent returns `NEEDS-INPUT` and writes no specification
- **Answers**: When invoked again with `answers` to its own `NEEDS-INPUT`, the agent treats them as user-stated requirements and re-runs the gate before generating

Each layer reads requirements from its session context. Upstream layers provide context for coherence, not requirements. When user requirements would create incoherence with existing specs, agents MUST flag the conflict rather than silently override.

### Output
- Specification documents in `specs/{layer}/`
- Documents MUST follow `templates/{layer}.template.md`

### Directives

**Specification agents MUST:**
- Produce one specification file per distinct concept (e.g., one use case, one API contract)
- Generate YAML frontmatter with required fields: `id`, `status: draft`, `created`
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

| Agent | Layer | Context (for coherence) | Output |
|-------|-------|---------------------------|--------|
| `smaqit.business` | Business | None | `specs/business/*.md` |
| `smaqit.functional` | Functional | Business specs | `specs/functional/*.md` |
| `smaqit.stack` | Stack | Business and Functional specs | `specs/stack/*.md` |
| `smaqit.infrastructure` | Infrastructure | Phase 1 specs | `specs/infrastructure/*.md` |
| `smaqit.coverage` | Coverage | All layer specs | `specs/coverage/*.md` |

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

Implementation agents always orchestrate specification generation as the primary first phase activity. Orchestration is not conditional on missing artifacts — it is the primary path:
- **Orchestration-first** — Spec generation is always step 1; each layer is checked and generated if missing, draft, or failed, unless its paths are listed in `approved_spec_paths`
- **Deterministic routing** — The specification-agent invocation sequence is hardcoded per phase, not derived at runtime; status checks on spec files determine which layers to skip. Implementation steps that legitimately vary per project are variation points (slots), resolved deterministically as described below — never improvised at runtime
- **Scoped context passing** — Each spec agent receives user requirements from session context plus content of upstream specs already generated in the current sequence; full accumulated phase context is not passed
- **Dependency ordering** — Business → Functional → Stack for Development; Infrastructure for Deployment; Coverage for Validation
- **Generation completion** — All required spec layers complete before implementation begins
- **Iteration caps** — Spec generation retries cap at 3 per layer; on cap reached, unresolved issues are surfaced to the user before proceeding

**Implementation Variation Points (Slots):**

Where a phase's implementation steps legitimately differ between projects — an app repository versus a platform repository, a VM versus a Namespace-scoped cluster — the phase declares a fixed, ordered sequence of slots instead of hardcoding one wiring per case:
- **Fixed order** — The phase defines the slot catalogue and its order; a project never reorders, adds, or removes slots, it only binds or skips them
- **Slot contract** — Each slot defines its purpose, inputs, handoff, and completion signal; any filler bound to it must satisfy that contract
- **Binding resolution** — A project binds a slot to a filler skill (and operation) in project-owned configuration; an unbound slot uses the default filler of the phase's selected profile; an explicit `none` skips it
- **Deterministic resolution** — Bindings are resolved and validated by deterministic tooling before execution, not by agent judgment: a binding to a skill that is missing or does not declare the slot is refused
- **Path refusals** — An execution path may refuse bindings for slots it must not perform, preserving that path's prohibitions
- **Recorded resolution** — The resolved plan (slot, source, filler, operation, executor) is recorded in the phase report
- **Invariants bind every filler** — Phase invariants (guarded IaC plan/apply, no out-of-band changes to IaC-managed resources, credential references only) apply to every filler, default or project-owned; the executing agent stops and reports when a filler's instructions would break one, rather than following them
- **Single-slot runs** — Day-two operations run as single-slot invocations of the same agent under the same invariants; an execution path's full-sequence defaults do not apply to an explicitly requested slot, but its refusals do
- **No split CI runs** — Slots a CI pipeline performs run as one pipeline run; resolution refuses a locally executed binding positioned between them

**Phase Parameters:**

A phase agent invoked as a subagent does not see its caller's conversation, so callers pass its execution parameters explicitly. The caller is a command, an orchestrating skill, or the user. It resolves the parameters by applying the phase input skill wherever it can reach the user:

| Parameter | Phases | Meaning |
|-----------|--------|---------|
| `mode` | All | `assisted` or `autonomous`; the default depends on what started the phase (see Execution Modes) |
| `user_request` | All | The user's own words that started the phase, verbatim; never a summary or expansion. Passed unchanged to every specification agent |
| `specification_mode` | All | `orchestrate` (default) or `prevalidated` |
| `spec_paths` | All | Exact specification paths handed off under `prevalidated` |
| `approved_spec_paths` | All | Specification paths the user already approved in this run; treated as current even while `draft` |
| `deployment_path`, `provisioning_mode`, `environment`, `slot_plan` | Deploy | As resolved by `smaqit.input-deployment` |
| `answers` | All | Answers to a returned `NEEDS-INPUT` or `REVIEW-REQUESTED`, keyed by question id |

- **Defaults when absent** — A phase agent invoked without a parameter block applies the defaults its phase input skill defines, including `mode: assisted`, because the user started it directly. It returns `NEEDS-INPUT` only when a blocking parameter is ambiguous
- **Passed means resolved** — A parameter the caller passed is never elicited or resolved again by the agent

**Execution Modes:**

Phase agents support two execution modes, selected by the `mode` parameter. The default depends on what started the phase, and a mode the user states always wins:

- **Assisted** is the default when the user started the phase: they typed the phase command, or picked or named the phase agent. A phase agent invoked without a parameter block assumes this.
- **Autonomous** is the default when the phase started automatically: the model routed a request to it on its own, or an orchestrating skill (`smaqit.new-greenfield-project`, `smaqit.feature-new`) ran it.

- **Autonomous** — Proceed through all workflow steps without user breaks, on a best-effort basis; spec agents invoked in sequence and implementation begins immediately after consolidation. Stop only for `NEEDS-INPUT` and human operations (for example a PR merge or a credential prompt)
- **Assisted** (maker-checker) — Specification agents receive `mode: assisted` and work only from requirements the user stated. After each spec agent (maker) completes, the user (checker) reviews its output. An agent that can reach the user presents the output directly; otherwise it returns `REVIEW-REQUESTED`. Approval adds the reviewed paths to `approved_spec_paths`, and the phase continues with the next layer. Feedback invokes the spec agent again with the feedback as `answers`. The loop repeats until the output is approved or the iteration cap is reached (max 3 per layer); on cap, surface unresolved issues and proceed. When the phase completes, its final output asks the user to review it and names the next phase's entry point

**Multi-Agent Coordination:**

Phase orchestration coordinates multiple agents within a single phase:
- Agents invoked as needed based on artifact availability
- Invoked agents produce outputs consumed by orchestrating agent  
- Return signals from invoked agents are handled as described in User Interaction
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

**Context Sufficiency:**

Session context verified to contain sufficient requirements for specification generation:
- **Requirements presence** — Session context contains project goals or feature requirements that are present and actionable
- **Conflict check** — No unresolvable conflicts exist in provided requirements

Insufficient context halts execution with guidance describing what requirements are missing.

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
- Stop and return a `DESIGN-*` failure to the owning specification agent when the automatic plan readiness gate exits nonzero
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
- Record implementation choices the specs leave open (for example file layout) in the phase report; a gap in a spec's requirements or acceptance criteria is returned to the owning specification agent, never assumed
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
- Clarification is required from upstream (request it directly or through `NEEDS-INPUT`, and wait for the answers)

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
