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
- Agents MUST request clarification when input is ambiguous
- Agents MUST NOT invent requirements not present in input
- Agents SHOULD flag assumptions explicitly when clarification is unavailable

### Fail-Fast on Inconsistency
- Agents MUST verify coherence across all input sources before producing output
- Agents MUST stop and report when inputs contradict each other
- Agents MUST NOT proceed with output while unresolved inconsistencies exist

### Self-Validation Before Completion
- Agents MUST validate their output against completion criteria before finishing
- Agents MUST NOT declare completion if any required criterion is unmet
- Agents SHOULD iterate on output until validation passes

## Naming Convention

Agents follow the pattern: `smaqit.[LAYER]` for specification agents, `smaqit.[PHASE]` for implementation agents, and `smaqit.orchestrator` for the orchestration agent.

| Type | Pattern | Examples |
|------|---------|----------|
| Specification | `smaqit.[LAYER]` | `smaqit.business`, `smaqit.functional`, `smaqit.stack` |
| Implementation | `smaqit.[PHASE]` | `smaqit.development`, `smaqit.deployment`, `smaqit.validation` |
| Orchestrator | `smaqit.orchestrator` | `smaqit.orchestrator` |

## Specification Agents

Specification agents translate prompt file requirements into precise, testable specifications for a single layer.

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
- Include testable acceptance criteria in every specification
- Reference context specs used for coherence and traceability
- Validate output against layer template before completion

**Specification agents MUST NOT:**
- Include implementation details (code, technology choices outside Stack layer)
- Create inconsistencies with context layer specifications
- Produce specs for layers outside their scope

**Specification agents SHOULD:**
- Define explicit scope boundaries (what is included vs. excluded)
- Use consistent terminology across layers
- Flag potential inconsistencies with context specs

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
- Comply with all referenced specifications
- Trace every implementation decision to a specification
- Validate output against specification acceptance criteria
- Report deviations or impossibilities rather than silently diverge

**Implementation agents MUST NOT:**
- Modify specifications (request changes through proper channels)
- Implement features not defined in specifications
- Skip validation steps defined in Coverage specs

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
| Specification | `read`, `edit`, `search` | Produce documents only |
| Implementation | `execute`, `read`, `edit`, `search`, `todo` | Build, run, test applications |

**Tool descriptions:**

| Tool | Purpose |
|------|---------|
| `execute` | Run terminal commands (build, test, deploy) |
| `read` | Read files and specifications |
| `edit` | Create and modify files |
| `search` | Search codebase and specifications |
| `todo` | Track multi-step task progress |

Agents MUST NOT proceed with implementation while unresolved conflicts exist.

### Implementation Agent Mappings

| Agent | Phase | Input | Output |
|-------|-------|-------|--------|
| `smaqit.development` | Develop | Business + Functional + Stack specs | Code |
| `smaqit.deployment` | Deploy | Code + Infrastructure specs | Running system |
| `smaqit.validation` | Validate | Deployed system + Coverage specs | Validation report |

## Orchestrator Agent

The orchestrator agent coordinates full workflow execution from specifications through validation.

### Input
- **Orchestrator prompt**: `prompts/smaqit.orchestrate.prompt.md` — User preferences for workflow execution
- **All prompts**: Layer prompts (5) and implementation prompts (3) — Required for pre-run validation

### Output
- **Orchestration report**: Documents agent invocations, phase outcomes, errors
- **Workflow status**: Complete/partial/failed with detailed execution log

### Directives

**Orchestrator agent MUST:**
- Execute pre-run validation before starting workflow (if requested)
- Invoke agents in correct dependency order: 5 spec agents → 3 implementation agents
- Verify each phase completion before proceeding to next phase
- Report all errors with context (phase, agent, input state)
- Respect user error handling preferences (stop on error vs continue)
- Validate workflow completion criteria before declaring success

**Orchestrator agent MUST NOT:**
- Skip required phases without user approval
- Proceed with missing upstream specifications
- Silently ignore phase failures
- Modify agent execution order to bypass dependencies
- Bypass pre-run validation when user requested it

**Orchestrator agent SHOULD:**
- Provide progress updates during long-running workflows
- Report estimated time remaining for multi-phase execution
- Suggest recovery actions when phases fail
- Document lessons learned for workflow optimization

### Tooling

Orchestrator agent requires the `agent` tool to invoke other agents:

| Tool | Purpose |
|------|----------|
| `agent` | Invoke specification and implementation agents |
| `execute` | Run validation commands |
| `read` | Read prompts and orchestration parameters |
| `edit` | Create orchestration reports |
| `search` | Locate prompt files and verify completeness |
| `todo` | Track multi-phase workflow progress |

### Orchestrator Agent Mapping

| Agent | Purpose | Input | Output |
|-------|---------|-------|--------|
| `smaqit.orchestrator` | Coordinate workflow | Orchestrator prompt + all layer/implementation prompts | Orchestration report + workflow status |

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