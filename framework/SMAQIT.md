# smaqit Framework

Spec-driven agent orchestration where specifications are split into layers and phases. **Phases are the primary workflow unit**—each phase includes both specification generation and implementation execution. Users input requirements in prompt files, AI specification agents read from these prompt files to write specifications, then implementation agents generate outputs from those specifications.

## Core Principles

### Phase-First Workflow

**Phases are the primary workflow unit. Each phase includes specifications and implementation together.**

smaqit workflows execute in phases, not specification-then-implementation:
- **Phase 1 (Develop)**: Generate Business, Functional, and Stack specs → implement → working application
- **Phase 2 (Deploy)**: Generate Infrastructure spec → deploy → running system
- **Phase 3 (Validate)**: Generate Coverage spec → validate → validation report

Users CAN generate all specifications first (spec-first approach), but SHOULD complete phases sequentially (phase-first approach) for faster feedback and iterative validation.

### Specs Before Code

**Never write implementation without a corresponding specification.**

Specifications are not documentation—they are the source of truth. Implementation agents consume specs as contracts, not guidelines. This inverts the common pattern where code comes first and docs follow.

### Traceability Across Layers

**Every output MUST trace to a prompt file.**

- Each layer receives requirements from its prompt file
- Upstream layers provide context for coherence, not requirements
- Code references specs
- Tests reference requirements

### Layer Independence

**Each layer's prompt file is the sole source of requirements for that layer.**

Each layer has its own prompt file where users input requirements. Upstream layers provide context for coherence, not requirements. This ensures that user intent guides every layer without false derivation chains.

### Specification Coverage

**Every requirement MUST be verified through traceable test coverage.**

Traceability enables complete specification coverage: the Coverage layer traces requirements through all upstream specs to ensure nothing is missed. Untested requirements are explicit gaps, not silent omissions.

### Self-Validating Agents

**Agents validate their own output before declaring completion.**

Agents are not fire-and-forget. Each agent has completion criteria and MUST verify them before finishing. This shifts quality assurance left—into the agent itself, not a separate review step.

### Bounded Agents

**Agents execute only their designated layer or phase.**

Each agent has a single responsibility. Agents decline out-of-scope requests with clear redirection to the appropriate agent. This enforces separation of concerns and prevents scope creep across workflow boundaries.

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

### Reproducible from Input Set

**Identical input sets should produce equivalent validated behavior.**

The complete set of prompts across all layers defines a reproducible workflow. Given the same prompt set:

- **Equivalent outcomes**: Acceptance criteria pass or fail consistently
- **Traceable changes**: Modifying any prompt in the set reveals requirement changes explicitly
- **Audit trail**: Prompt sets document what was requested at each layer

### Progressive Refinement

**Each layer addresses a distinct concern.**

Layers are independent but must be coherent:
- Business (intent) | Functional (behavior) | Stack (tools) | Infrastructure (environment) | Coverage (verification)
- No layer derives requirements from another—each reads from its own prompt file
- Implementation agents validate cross-layer coherence before execution

### Explicit Over Implicit

**When in doubt, make it explicit.**

- State assumptions rather than assume shared context
- Define scope boundaries rather than imply them
- Reference sources rather than expect inference

### Fail-Fast on Ambiguity

**When input is unclear, stop and request clarification.**

- Do not invent requirements
- Flag assumptions explicitly

## Quick Reference

### Workflow Approaches

smaqit supports two approaches, with phase-first recommended:

| Approach | Workflow | Feedback Cycle | Recommended For |
|----------|----------|----------------|-----------------|
| **Phase-First** (recommended) | Complete each phase (specs + implementation) before next phase | Fast (per phase) | Most projects, iterative development |
| **Spec-First** (optional) | Generate all 5 specs, then implement in phases | Slower (at end) | Upfront design requirements, regulatory compliance |

**Phase-First Workflow:**
1. **Develop Phase**: Business spec → Functional spec → Stack spec → Development agent → working application
2. **Deploy Phase**: Infrastructure spec → Deployment agent → running system
3. **Validate Phase**: Coverage spec → Validation agent → validation report

**Spec-First Workflow:**
1. Generate all specs: Business → Functional → Stack → Infrastructure → Coverage
2. Execute phases: Develop → Deploy → Validate

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
│   └── templates/        # Spec and prompt templates
├── .github/
│   ├── agents/           # Agent definitions
│   └── prompts/          # Prompt files (user input)
└── specs/
    ├── business/
    ├── functional/
    ├── stack/
    ├── infrastructure/
    └── coverage/
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
