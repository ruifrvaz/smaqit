# Layers

Layers define the progressive refinement of specifications from intent to verification. Each layer answers a specific question and adds precision while maintaining traceability to upstream layers.

## Layer Order

Layers MUST be worked through in order within each phase:

**Phase 1 (Develop):** Business → Functional → Stack

**Phase 2 (Deploy):** Infrastructure (reads all Phase 1 specs)

**Phase 3 (Validate):** Coverage (reads all specs)

Each layer depends on its predecessor(s):
- **Phase 1 layers** (Business through Stack): linear, each depends on immediate upstream layer
- **Infrastructure** (Phase 2): cross-cutting, depends on all Phase 1 specs + user input
- **Coverage** (Phase 3): cross-cutting, depends on all four upstream layers

## Layer Definitions

### Business — Why?

The Business layer captures the intent, value, and goals of what is being built.

**Purpose:** Define use cases, actors, and measurable outcomes that justify the work.

**Upstream:** User input (natural language requirements)

**Downstream:** Functional layer

**Directives:**

**Business specs MUST:**
- Identify all actors and their goals
- Define measurable success metrics for each use case
- Include preconditions and postconditions
- Describe main and alternative flows in business terms

**Business specs MUST NOT:**
- Mention specific technologies, frameworks, or libraries
- Include implementation details or technical solutions
- Define data structures or API contracts
- Reference deployment or infrastructure concerns

---

### Functional — What?

The Functional layer defines the behaviors, contracts, and data models required to fulfill business goals.

**Purpose:** Translate business use cases into precise behavioral specifications.

**Upstream:** Business specs

**Downstream:** Stack layer, Coverage layer

**Directives:**

**Functional specs MUST:**
- Define user flows that implement business use cases
- Specify data models with attributes and relationships
- Define API contracts (inputs, outputs, error conditions)
- Include state transitions where applicable
- Reference originating business specs

**Functional specs MUST NOT:**
- Specify technology choices (languages, frameworks, databases)
- Include deployment or infrastructure concerns
- Define performance benchmarks (those belong in Infrastructure)
- Prescribe implementation patterns

---

### Stack — With what?

The Stack layer selects and justifies the technologies used to implement functional requirements.

**Purpose:** Choose languages, frameworks, libraries, and tools that satisfy functional needs.

**Upstream:** Functional specs

**Downstream:** Infrastructure layer, Development agent

**Directives:**

**Stack specs MUST:**
- Justify each technology choice against functional requirements
- Define language versions and framework versions
- Specify libraries and their purposes
- Include build tools and development environment setup
- Reference functional specs that drove each choice

**Stack specs MUST NOT:**
- Define deployment topology or infrastructure
- Include compute, networking, or scaling decisions
- Specify cloud providers or hosting platforms
- Contradict functional requirements

---

### Infrastructure — Where?

The Infrastructure layer defines where and how the application runs in production.

**Purpose:** Specify compute, networking, observability, and operational concerns.

**Upstream:** Phase 1 specs (Business, Functional, Stack) + user input

**Downstream:** Deployment agent, Coverage layer

**Directives:**

**Infrastructure specs MUST:**
- Define compute resources (containers, serverless, VMs)
- Specify networking topology and security boundaries
- Include observability (logging, metrics, tracing)
- Define scaling policies and resource limits
- Specify secrets management approach
- Reference Phase 1 specs for requirements and runtime constraints
- Verify coherence across all input specs before producing output

**Infrastructure specs MUST NOT:**
- Redefine business logic or functional behaviors
- Override technology choices from Stack layer
- Include application code or configurations
- Define test cases (those belong in Coverage)

---

### Coverage — What's verified?

The Coverage layer ensures all upstream requirements are testable and traceable. It is cross-cutting, reading from all four upstream layers.

**Purpose:** Enumerate every acceptance criterion and map it to a verification test.

**Upstream:** Business, Functional, Stack, and Infrastructure specs (all layers)

**Downstream:** Validation agent

**Directives:**

**Coverage specs MUST:**
- Reference every acceptance criterion from upstream specs by ID
- Define a test case for each testable requirement
- Map: Requirement ID → Test Case → Expected Outcome
- Flag untestable requirements explicitly
- Include integration, E2E, and acceptance test definitions
- Report spec coverage (% of requirements with corresponding tests)

**Coverage specs MUST NOT:**
- Add requirements not present in upstream specs
- Modify or reinterpret upstream acceptance criteria
- Skip requirements without explicit justification
- Define unit tests (those are implementation details)

## Dependency Graph

```
                    ┌─────────────────────────────────────┐
                    │              Coverage               │
                    │          (What's verified?)         │
                    └─────────────────────────────────────┘
                        ↑         ↑         ↑         ↑
          ┌─────────────┘         │         │         └─────────────┐
          │                       │         │                       │
          │                       │         │                       │
    ┌─────┴─────┐           ┌─────┴─────┐   │                 ┌─────┴─────┐
    │  Business │           │ Functional│   │                 │  Infra    │
    │   (Why?)  │──────────→│  (What?)  │───┼─────────╮       │ (Where?)  │
    └─────┬─────┘           └─────┬─────┘   │         │       └───────────┘
          │                       │         │         │             ↑
          │                       │         │         │             │
          │                 ┌─────┴─────┐   │         │             │
          │                 │   Stack   │───┘         │             │
          │                 │(With what)│─────────────┼─────────────┘
          │                 └───────────┘             │
          │                                           │
          └───────────────────────────────────────────┘
```

**Phase 1 (Develop):** Business → Functional → Stack (linear)

**Phase 2 (Deploy):** Infrastructure reads all Phase 1 specs (cross-cutting)

**Phase 3 (Validate):** Coverage reads all specs including Infrastructure (cross-cutting)

## See Also

- [SMAQIT](SMAQIT.md) — Framework overview and principles
- [PHASES](PHASES.md) — Phase workflows and transitions
- [TEMPLATES](TEMPLATES.md) — Template structure rules
- [AGENTS](AGENTS.md) — Agent behaviors
- [ARTIFACTS](ARTIFACTS.md) — Artifact rules
