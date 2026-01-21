---
layer: infrastructure
target: agents/smaqit.infrastructure.agent.md
sources:
  - framework/LAYERS.md
created: 2026-01-19
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| LAYERS.md | Infrastructure — Where? |
| LAYERS.md | Infrastructure Foundation vs Feature Specs |
| LAYERS.md | Infrastructure specs MUST / MUST NOT |

---

## L1 Directive Compilation

### Layer-Specific MUST Directives

- Define compute resources (containers, serverless, VMs)
- Specify networking topology and security boundaries
- Include observability (logging, metrics, tracing)
- Define scaling policies and resource limits
- Specify secrets management approach
- Be consistent with Phase 1 specs regarding requirements and runtime constraints (validated at implementation)
- Reference Phase 1 specs using Enables (foundation serving multiple) or direct reference (feature serving one)
- Include justification when foundation spec has no Phase 1 references

### Layer-Specific MUST NOT Directives

- Redefine business logic or functional behaviors
- Override technology choices from Stack layer
- Include application code or configurations
- Define test cases (those belong in Coverage)

### Foundation Spec Rules

Infrastructure specs fall into two categories:

| Type | Purpose | Phase 1 Reference |
|------|---------|--------------------|
| **Feature specs** | Infrastructure for a specific feature/component | 1:1 mapping (Enables) |
| **Foundation specs** | Base infrastructure enabling multiple features | 1:many mapping (Enables) |

**Foundation specs** (base networking, shared security policies, common observability configuration) are legitimate operational artifacts serving multiple application components.

**Foundation spec requirements:**
- SHOULD reference all Phase 1 specs (Business, Functional, Stack) they enable
- MUST flag absence of Phase 1 references with justification

**Warning:** Orphaned foundations (no Phase 1 references, no justification) indicate scope creep.

### Upstream Context

**[UPSTREAM_CONTEXT_DESCRIPTION]:**
```
The Infrastructure layer defines deployment and operational concerns. It receives coherence context from all Phase 1 layers (Business, Functional, Stack) to ensure infrastructure decisions align with requirements, behaviors, and technology choices.
```

**[UPSTREAM_SPEC_PATHS]:**
```
`specs/business/`, `specs/functional/`, and `specs/stack/` (for coherence and traceability)
```

### User Input Description

**[USER_INPUT_DESCRIPTION]:**
```
User deployment requirements describing environment, hosting preferences, scaling needs, and operational constraints. Requirements should specify where and how the application runs in production.
```

### Other Layers

**[OTHER_LAYERS]:**
```
Business, Functional, Stack, Coverage
```

---

## Compilation Guidance for Agent-L2

When compiling `agents/smaqit.infrastructure.agent.md`:

1. **Replace [LAYER]** → `infrastructure`
2. **Replace [LAYER_NAME]** → `Infrastructure`
3. **Replace [LAYER_PREFIX]** → `INF`
4. **Replace [UPSTREAM_CONTEXT_DESCRIPTION]** → upstream context text above
5. **Replace [UPSTREAM_SPEC_PATHS]** → "`specs/business/`, `specs/functional/`, and `specs/stack/` (for coherence and traceability)"
6. **Replace [USER_INPUT_DESCRIPTION]** → user input text above
7. **Replace [OTHER_LAYERS]** → "Business, Functional, Stack, Coverage"
8. **Insert Layer-Specific MUST directives** → Infrastructure-specific MUST rules
9. **Insert Layer-Specific MUST NOT directives** → Infrastructure-specific MUST NOT rules
10. **Insert Foundation guidance** → Foundation vs Feature spec table and rules
11. **Preserve template structure** → All sections from specification-agent.template.md
