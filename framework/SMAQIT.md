# smaqit Framework

Spec-driven agent orchestration where specifications are split into layers and phases. Users input requirements in prompt files, AI specification agents read from these prompt files to write specifications, then implementation agents generate outputs from those specifications.

## Core Principles

### Phases as Workflow Units

**Each phase includes specifications and implementation together.**

Phases are the primary workflow units in smaqit. Users can generate all specifications first, but the recommended approach is to complete each phase (specifications + implementation) before moving to the next. This provides faster feedback and validates the system incrementally.

**Workflow Coordination:**

Implementation agents coordinate workflow activities within their phase via orchestrations. Workflows include validation of inputs, consolidation of sources, generation of artifacts, execution of outputs, and verification of outcomes. Each workflow activity reports progress and handles failures explicitly.

### Specs Before Code

**Never write implementation without a corresponding specification.**

Specifications are not documentation—they are the source of truth. Implementation agents consume specs as contracts, not guidelines. This inverts the common pattern where code comes first and docs follow.

### Traceability Across Layers

**Every output traces to its input source, creating an unbroken chain from user requirements through specifications to implementation.**

Each layer receives requirements from its prompt file. Upstream layers provide context for coherence, not requirements. Code references specs. Tests reference requirements. This creates explicit lineage for every decision and artifact.

### Layer Independence

**Each layer's prompt file is the sole source of requirements for that layer.**

Each layer has its own prompt file where users input requirements. Upstream layers provide context for coherence, not requirements. This ensures that user intent guides every layer without false derivation chains.

### Single Source of Truth

**Each piece of information exists in exactly one place.**

When information is needed in multiple contexts, reference the source rather than duplicate. Foundation specs contain shared requirements that multiple feature specs depend on. This prevents conflicting sources of truth, reduces maintenance burden, and ensures consistency across specifications.

### Complete Specification Coverage

**Every requirement receives verification through traceable test coverage.**

Complete coverage emerges from explicit traceability. When requirements flow through identifiable chains from intent to implementation to verification, gaps become visible rather than hidden. Untested requirements appear as explicit omissions in the coverage map, not silent assumptions about completeness.

### Self-Validating Agents

**Agents validate their own output before declaring completion.**

Agents are not fire-and-forget. Each agent has completion criteria and verifies them before finishing. This shifts quality assurance left—into the agent itself, not a separate review step.

### Bounded Agents

**Agents execute only their designated layer or phase.**

Unbounded agents lose accountability. Each agent has a single responsibility. Agents decline out-of-scope requests with clear redirection to the appropriate agent. This enforces separation of concerns and prevents scope creep across workflow boundaries.

### Template-Constrained Output

**Templates are cognitive scaffolds, not suggestions.**

Unpredictable structure breaks downstream consumers. Templates define the exact structure agents produce. This ensures consistent output across runs, predictable input for downstream consumers and reduced LLM variance.

### Extensible Through Templates

**Templates constrain execution and enable extension simultaneously.**

What agents share philosophically belongs in foundational templates. What makes agents distinct belongs in role-specific templates. When templates capture principles explicitly, generation becomes principled rather than ad-hoc. This enables agent extensibility while preserving framework coherence.

### Accept Mutability, Validate Behavior

**Embrace non-determinism in artifacts, enforce determinism in outcomes.**

LLMs rarely generate identical output twice. Rather than fighting this inherent variability, smaqit accepts it:

- **Mutable artifacts**: Code, configurations, and documents may vary between runs
- **Immutable behavior**: Specifications define expected outcomes, not implementation details
- **Validation over reproducibility**: Success is measured by passing acceptance criteria, not by identical output

### Reproducible from Input Set

**Identical input sets produce equivalent validated behavior.**

The complete set of prompts across all layers defines a reproducible workflow. Given the same prompt set:

- **Equivalent outcomes**: Acceptance criteria pass or fail consistently
- **Traceable changes**: Modifying any prompt in the set reveals requirement changes explicitly
- **Audit trail**: Prompt sets document what was requested at each layer

### Progressive Refinement

**Each layer addresses a distinct concern.**

Layers are independent but coherent. Business captures intent, Functional defines behavior, Stack selects tools, Infrastructure specifies environment, Coverage verifies implementation. No layer derives requirements from another—each reads from its own prompt file. Implementation agents validate cross-layer coherence before execution.

### Stateful Specifications

**Specifications track their lifecycle state through implementation phases.**

Specs are not static documents—they evolve through phases with tracked states:
- **Draft**: Spec generated, ready for implementation
- **Implemented**: Code generated, tests pass
- **Deployed**: Running in target environment
- **Validated**: All acceptance criteria verified
- **Failed**: Processing failed at some phase

### Explicit Over Implicit

**When in doubt, make it explicit.**

Ambiguity multiplies as context travels through layers and agents. Stated assumptions become verifiable constraints. Defined boundaries create clear accountability. Explicit references enable traceability. What remains implicit becomes a source of divergence and confusion.

### Fail-Fast on Ambiguity

**When input is unclear, stop and request clarification.**

Proceeding with unclear input compounds errors through every downstream layer. Invented requirements masquerade as user intent. Early clarification prevents late-stage rework and maintains the integrity of the requirement chain from source to implementation.

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

## See Also

Read SMAQIT.md for framework overview and principles. Consult these files as needed:

| File | Purpose | When to Consult |
|------|---------|-----------------|
| [PROMPTS](PROMPTS.md) | Prompt structure, input records, agent interaction | Understanding prompt files or agent invocation |
| [SKILLS](SKILLS.md) | Agent skills, invocable capabilities, condition detection | Understanding reusable agent capabilities |
| [LAYERS](LAYERS.md) | Five specification layers and their dependencies | Generating or validating layer specs |
| [PHASES](PHASES.md) | Three development phases and their workflows | Orchestrating multi-agent workflows |
| [TEMPLATES](TEMPLATES.md) | Template structure rules for prompts, specs, and agents | Creating or validating templates |
| [AGENTS](AGENTS.md) | Agent behaviors (actors) | Understanding agent responsibilities |
| [ARTIFACTS](ARTIFACTS.md) | Artifact rules (outputs) | Understanding spec structure and traceability |
