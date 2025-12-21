# smaqit Framework

Spec-driven agent orchestration. AI agents write specifications first, then implement from those specs.

## Core Principles

### Specs Before Code

**Never write implementation without a corresponding specification.**

Specifications are not documentation—they are the source of truth. Implementation agents consume specs as contracts, not guidelines. This inverts the common pattern where code comes first and docs follow.

### Traceability Across Layers

**Every output MUST trace to a prompt file.**

- Each layer receives requirements from its prompt file
- Upstream layers provide context for coherence, not requirements
- Code references specs
- Tests reference requirements

Traceability enables impact analysis: when a requirement changes, the chain of dependencies is explicit.

### Layer Independence

**Each layer's prompt file is the sole source of requirements for that layer.**

Each layer has its own prompt file where users input requirements. Upstream layers provide context for coherence, not requirements. This ensures that user intent guides every layer without false derivation chains.

### Specification Coverage

**Every requirement MUST be verified through traceable test coverage.**

Traceability enables complete specification coverage: the Coverage layer traces requirements through all upstream specs to ensure nothing is missed. Untested requirements are explicit gaps, not silent omissions.

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

### Reproducible from Input Set

**Identical input sets should produce equivalent validated behavior.**

The complete set of prompts across all layers defines a reproducible workflow. Given the same prompt set:

- **Equivalent outcomes**: Acceptance criteria pass or fail consistently
- **Traceable changes**: Modifying any prompt in the set reveals requirement changes explicitly
- **Audit trail**: Prompt sets document what was requested at each layer

## Design Philosophy

### Progressive Refinement

Each layer addresses a distinct concern:

```
Business (intent) | Functional (behavior) | Stack (tools) | Infrastructure (environment) | Coverage (verification)
```

Layers are independent but must be coherent. No layer derives requirements from another—each reads from its own prompt file. Implementation agents validate cross-layer coherence before execution.

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

## See Also

Read SMAQIT.md first for framework overview. Consult these files as needed:

| File | Purpose | When to Consult |
|------|---------|-----------------|
| [PROMPTS](PROMPTS.md) | Prompt structure, input records, agent interaction | Understanding prompt files or agent invocation |
| [LAYERS](LAYERS.md) | Five specification layers and their dependencies | Generating or validating layer specs |
| [PHASES](PHASES.md) | Three development phases and their workflows | Orchestrating multi-agent workflows |
| [TEMPLATES](TEMPLATES.md) | Template structure rules for prompts, specs, and agents | Creating or validating templates |
| [AGENTS](AGENTS.md) | Agent behaviors (actors) | Understanding agent responsibilities |
| [ARTIFACTS](ARTIFACTS.md) | Artifact rules (outputs) | Understanding spec structure and traceability |
