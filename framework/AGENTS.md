# Agents

Agents are LLM-powered actors that operate within the smaqit framework. This document defines the principles, constraints, and behaviors that all agents MUST follow.

## Unified Principles

All smaqit agents—specification and implementation—share these foundational principles:

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

Agents follow the pattern: `smaqit.[LAYER]` for specification agents and `smaqit.[PHASE]` for implementation agents.

| Type | Pattern | Examples |
|------|---------|----------|
| Specification | `smaqit.[LAYER]` | `smaqit.business`, `smaqit.functional`, `smaqit.stack` |
| Implementation | `smaqit.[PHASE]` | `smaqit.development`, `smaqit.deployment`, `smaqit.validation` |

## Specification Agents

Specification agents define *what* to build. They produce specification documents that serve as contracts for implementation agents.

### Purpose
Translate upstream inputs into precise, testable specifications for a single layer.

### Input
- **Upstream specifications**: Documents from previous layer(s) in the layer order
- **User input**: Direct requirements relevant to the agent's layer (e.g., technology preferences for Stack, deployment constraints for Infrastructure)

The Business agent is the primary entry point for user input, but all specification agents MAY receive user input relevant to their layer. When user input conflicts with upstream specs, agents MUST flag the conflict rather than silently override.

### Output
- Specification documents in `.smaqit/specs/{layer}/`
- Documents MUST follow `templates/{layer}.template.md`

### Directives

**Specification agents MUST:**
- Produce one specification file per distinct concept (e.g., one use case, one API contract)
- Include testable acceptance criteria in every specification
- Reference all upstream specs that informed the output
- Validate output against layer template before completion

**Specification agents MUST NOT:**
- Include implementation details (code, technology choices outside Stack layer)
- Modify or contradict upstream specifications
- Produce specs for layers outside their scope

**Specification agents SHOULD:**
- Define explicit scope boundaries (what is included vs. excluded)
- Use consistent terminology from upstream specs
- Flag gaps or inconsistencies in upstream input

### Specification Agent Mappings

| Agent | Layer | Input | Output |
|-------|-------|-------|--------|
| `smaqit.business` | Business | User description | `specs/business/*.md` |
| `smaqit.functional` | Functional | Business specs | `specs/functional/*.md` |
| `smaqit.stack` | Stack | Functional specs | `specs/stack/*.md` |
| `smaqit.infrastructure` | Infrastructure | Phase 1 specs + user input | `specs/infrastructure/*.md` |
| `smaqit.coverage` | Coverage | All layer specs | `specs/coverage/*.md` |

## Implementation Agents

Implementation agents build *how* it works. They consume specifications and produce executable artifacts.

### Purpose
Transform specifications into working software, deployed systems, or validated results.

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

## See Also

- [SMAQIT](SMAQIT.md) — Framework overview and principles
- [LAYERS](LAYERS.md) — Layer definitions and dependencies
- [PHASES](PHASES.md) — Phase workflows and transitions
- [TEMPLATES](TEMPLATES.md) — Template structure rules
- [ARTIFACTS](ARTIFACTS.md) — Artifact rules
