# Layers

Layers are independent specification manifests that together form a coherent application. Each layer answers a specific question and receives requirements from its own prompt file. Upstream layers provide context for coherence and traceability, not requirements.

## Layer Independence

**Each layer's prompt file is the sole source of requirements for that layer.**

Each layer:
- Receives requirements from its prompt file (`.github/prompts/smaqit.[layer].prompt.md`)
- Can be selected or swapped independently
- Must be coherent with adjacent layers
- Does not derive requirements from upstream layers

## Upstream References

Layers reference upstream specifications for these purposes:

| Purpose | Description |
|---------|-------------|
| **Coherence** | Implementation agents consolidate specs from multiple layers before execution. References ensure specs are compatible. |
| **Traceability** | Coverage maps requirements through all layers to ensure nothing is missed. |

Coherence validation happens at the end of each phase, where the implementation agent consolidates the required specs before execution.

## Layer Order

Layers are worked through in order within each phase:

**Phase 1 (Develop):** Business → Functional → Stack

**Phase 2 (Deploy):** Infrastructure (reads all Phase 1 specs for context)

**Phase 3 (Validate):** Coverage (reads all specs)

The order provides context accumulation, not requirement derivation:
- **Phase 1 layers** (Business through Stack): each provides cumulative context for subsequent layers
- **Infrastructure** (Phase 2): uses all Phase 1 specs as coherence context
- **Coverage** (Phase 3): validates against all layers

## Layer Definitions

### Business — Why?

The Business layer captures the intent, value, and goals of what is being built.

**Purpose:** Define use cases, actors, and measurable outcomes that justify the work.

**Input:** User requirements (stakeholder goals, use cases, success criteria)

**Context:** None (Business is the first layer)

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

**System Actor:**

When stakeholders have requirements about system properties (availability, auditability, accessibility), use the **System** actor:

| Actor | Description | Goals |
|-------|-------------|-------|
| System | The application as a whole | [System-level properties stakeholders require] |

System actor specs remain business-level (stakeholder-driven) and do not prescribe technical solutions.

---

### Functional — What?

The Functional layer defines the behaviors, contracts, and data models required to fulfill business goals.

**Purpose:** Translate user experience requirements into precise behavioral specifications.

**Input:** User experience requirements (experience shape, behaviors, interactions)

**Context:** Business specs (for coherence and traceability)

**Directives:**

**Functional specs MUST:**
- Define user flows that implement business use cases
- Specify data models with attributes and relationships
- Define API contracts (inputs, outputs, error conditions)
- Include state transitions where applicable
- Reference business specs for traceability using Implements (1:1 feature) or Enables (1:many foundation)
- Include justification when foundation spec has no Business references

**Functional specs MUST NOT:**
- Specify technology choices (languages, frameworks, databases)
- Include deployment or infrastructure concerns
- Define performance benchmarks (those belong in Infrastructure)
- Prescribe implementation patterns

**Foundation vs Feature Specs:**

Functional specs come in two categories:

| Type | Purpose | Business Reference |
|------|---------|--------------------|
| **Feature specs** | Implement a specific business use case | 1:1 mapping (Implements) |
| **Foundation specs** | Enable multiple business use cases | 1:many mapping (Enables) |

Foundation specs (shared components, cross-cutting concerns, common contracts) are legitimate engineering artifacts that serve multiple business goals.

**Foundation spec rules:**
- SHOULD reference all Business specs they enable
- MUST flag absence of Business references with justification

**Note:** Orphaned foundations (no Business references, no justification) indicate scope creep.

---

### Stack — With what?

The Stack layer selects and justifies the technologies used to implement functional requirements.

**Purpose:** Choose languages, frameworks, libraries, and tools that can deliver the specified behaviors.

**Input:** User technology preferences (languages, frameworks, constraints, team expertise)

**Context:** Business and Functional specs (for coherence and traceability)

**Directives:**

**Stack specs MUST:**
- Document technology choices with rationale
- Define language versions and framework versions
- Specify libraries and their purposes
- Include build tools and development environment setup
- Be consistent with Functional specs (validated at implementation)
- Reference Functional specs using Enables (foundation serving multiple) or direct reference (feature serving one)
- Include justification when foundation spec has no Functional references

**Stack specs MUST NOT:**
- Include code examples, implementation patterns, or architecture code blocks
- Define deployment topology or infrastructure
- Include compute, networking, or scaling decisions
- Specify cloud providers or hosting platforms
- Contradict functional requirements

**Foundation vs Feature Specs:**

Stack specs come in two categories:

| Type | Purpose | Functional Reference |
|------|---------|--------------------|
| **Feature specs** | Technology choices for a specific feature | 1:1 mapping (Enables) |
| **Foundation specs** | Base technologies enabling multiple features | 1:many mapping (Enables) |

Foundation specs (base language environments, shared build tools, common dependencies) are legitimate engineering artifacts that serve multiple functional requirements.

**Foundation spec rules:**
- SHOULD reference all Functional specs they enable
- MUST flag absence of Functional references with justification

**Note:** Orphaned foundations (no Functional references, no justification) indicate scope creep.

---

### Infrastructure — Where?

The Infrastructure layer defines where and how the application runs in production.

**Purpose:** Specify compute, networking, observability, and operational concerns.

**Input:** User deployment requirements (environment, hosting, scaling, constraints)

**Context:** Phase 1 specs (Business, Functional, Stack) for coherence and traceability

**Directives:**

**Infrastructure specs MUST:**
- Define compute resources (containers, serverless, VMs)
- Specify networking topology and security boundaries
- Include observability (logging, metrics, tracing)
- Define scaling policies and resource limits
- Specify secrets management approach
- Be consistent with Phase 1 specs regarding requirements and runtime constraints (validated at implementation)
- Reference Phase 1 specs using Enables (foundation serving multiple) or direct reference (feature serving one)
- Include justification when foundation spec has no Phase 1 references

**Infrastructure specs MUST NOT:**
- Redefine business logic or functional behaviors
- Override technology choices from Stack layer
- Include application code or configurations
- Define test cases (those belong in Coverage)

**Foundation vs Feature Specs:**

Infrastructure specs come in two categories:

| Type | Purpose | Phase 1 Reference |
|------|---------|--------------------|
| **Feature specs** | Infrastructure for a specific feature/component | 1:1 mapping (Enables) |
| **Foundation specs** | Base infrastructure enabling multiple features | 1:many mapping (Enables) |

Foundation specs (base networking, shared security policies, common observability configuration) are legitimate operational artifacts that serve multiple application components.

**Foundation spec rules:**
- SHOULD reference all Phase 1 specs (Business, Functional, Stack) they enable
- MUST flag absence of Phase 1 references with justification

**Note:** Orphaned foundations (no Phase 1 references, no justification) indicate scope creep.

---

### Coverage — What's verified?

The Coverage layer ensures all requirements are testable and traceable. It reads from all upstream layers for traceability and coherence.

**Purpose:** Enumerate every acceptance criterion and map it to a verification test.

**Input:** User test requirements (test scope, test environment, integration points, acceptance thresholds)

**Context:** All layer specs (Business, Functional, Stack, Infrastructure) — source of upstream acceptance criteria to verify

**Directives:**

**Coverage specs MUST:**
- Reference every acceptance criterion from upstream specs by ID
- Define a test case for each testable requirement
- Map: Requirement ID → Test Case → Expected Outcome
- Flag untestable requirements explicitly
- Include integration, E2E, and acceptance test definitions
- Report spec coverage (% of requirements with corresponding tests)

**Coverage specs MUST NOT:**
- Add acceptance criteria not present in upstream specs
- Skip upstream acceptance criteria without justification
- Modify or reinterpret upstream acceptance criteria
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
