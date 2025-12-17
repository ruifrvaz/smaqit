# smaqit Framework

Spec-driven agent orchestration. AI agents write specifications first, then implement from those specs.

## Core Principles

### Specs Before Code

**Never write implementation without a corresponding specification.**

Specifications are not documentation—they are the source of truth. Implementation agents consume specs as contracts, not guidelines. This inverts the common pattern where code comes first and docs follow.

### Traceability Across Layers

**Every output MUST trace to a user input.**

- Each layer receives requirements directly from user input
- Upstream layers provide context for consistency validation, not requirements
- Code references specs
- Tests reference requirements

Traceability enables impact analysis: when a requirement changes, the chain of dependencies is explicit.

### Layer Independence

**Layers are standalone manifests that can be selected independently.**

Each layer receives its own user input:
- Business: stakeholder goals and use cases
- Functional: experience shape and behaviors  
- Stack: technology preferences and constraints
- Infrastructure: deployment requirements
- Coverage: verification requirements

Upstream layers exist to ensure consistency across the application, not to dictate downstream requirements. Implementation agents validate that all layers form a consistent whole before proceeding.

### Self-Validating Agents

**Agents validate their own output before declaring completion.**

Agents are not fire-and-forget. Each agent has completion criteria and MUST verify them before finishing. This shifts quality assurance left—into the agent itself, not a separate review step.

### Template-Constrained Output

**Templates are cognitive scaffolds, not suggestions.**

Templates define the exact structure agents MUST produce. This ensures:
- Consistent output across runs
- Predictable input for downstream consumers
- Reduced LLM variance

### Accept Mutability, Validate Behavior

**Embrace non-determinism in artifacts, enforce determinism in outcomes.**

LLMs rarely generate identical output twice. Rather than fighting this inherent variability, smaqit accepts it:

- **Mutable artifacts**: Code, configurations, and documents may vary between runs
- **Immutable behavior**: Specifications define expected outcomes, not implementation details
- **Validation over reproducibility**: Success is measured by passing acceptance criteria, not by identical output

The specification is the invariant. The implementation is the variable.

## Iteration Through Experimentation

smaqit is designed to evolve through use. The current framework represents a minimal viable structure—sufficient to test the spec-driven hypothesis, but expected to adapt as real constraints emerge.

Complexity is added only when proven necessary by real project experience.

### Evidence Over Theory

Framework changes require evidence from actual application:

| Change Type | Required Evidence |
|-------------|-------------------|
| New layer | Multiple projects demonstrate a missing concern |
| New phase | Existing phases cannot accommodate a workflow |
| Structural change | Current structure blocks common patterns |

Anticipated edge cases do not justify framework changes. Observed constraints do.

### Explicit Assumptions

The framework operates under assumptions that may be revised:

| Assumption | Status | Revision Trigger |
|------------|--------|------------------|
| Layers are strictly linear | Active | Projects require multi-layer dependencies |
| Phases are sequential | Active | Parallel workflows prove necessary |
| Amendments are sufficient for conflicts | Active | Amendment overhead becomes prohibitive |
| Coverage reads all layers | Active | Subset coverage proves sufficient |

When an assumption is challenged by evidence, it becomes a candidate for revision.

### Amendment Protocol

When applying smaqit reveals limitations:

1. **Document** — Record the constraint encountered with context
2. **Propose** — Suggest framework amendment with rationale
3. **Test** — Apply amendment to subsequent projects
4. **Formalize** — If amendment improves outcomes, integrate into framework

## Design Philosophy

### Progressive Refinement

Each layer addresses a distinct concern:

```
Business (intent) | Functional (behavior) | Stack (tools) | Infrastructure (environment) | Coverage (verification)
```

Layers are independent but must be consistent. No layer derives requirements from another—each receives user input directly. Implementation agents validate cross-layer consistency before execution.

### Explicit Over Implicit

When in doubt, make it explicit:
- State assumptions rather than assume shared context
- Define scope boundaries rather than imply them
- Reference sources rather than expect inference

LLMs benefit from explicit context. Humans benefit from explicit documentation.

### Fail-Fast on Ambiguity

When input is unclear:
- Stop and request clarification
- Do not invent requirements
- Flag assumptions explicitly

The cost of clarification is lower than the cost of rework from incorrect assumptions.

## Framework Files

| File | Purpose |
|------|---------|
| [LAYERS](LAYERS.md) | Five specification layers and their dependencies |
| [PHASES](PHASES.md) | Three development phases and their workflows |
| [TEMPLATES](TEMPLATES.md) | Template structure rules for specs and agents |
| [AGENTS](AGENTS.md) | Agent behaviors (actors) |
| [ARTIFACTS](ARTIFACTS.md) | Artifact rules (outputs) |

## Quick Reference

### Layers

| Layer | Question | Purpose |
|-------|----------|---------|
| Business | Why? | Use cases, actors, goals |
| Functional | What? | Behaviors, contracts, flows |
| Stack | With what? | Languages, frameworks, libraries |
| Infrastructure | Where? | Compute, networking, observability |
| Coverage | Verified? | Integration, E2E, acceptance testing |

### Phases

| Phase | Spec Agents | Impl Agent | Output |
|-------|-------------|------------|--------|
| Develop | Business → Functional → Stack | Development | Working application |
| Deploy | Infrastructure | Deployment | Running system |
| Validate | Coverage | Validation | Validation report |

## File Locations (in smaqit-enabled projects)

```
project/
├── .smaqit/
│   ├── framework/        # Framework files (this directory)
│   ├── specs/
│   │   ├── business/
│   │   ├── functional/
│   │   ├── stack/
│   │   ├── infrastructure/
│   │   └── coverage/
│   └── templates/        # Layer templates
└── .github/
    └── agents/           # Agent definitions
```
