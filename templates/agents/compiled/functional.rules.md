---
layer: functional
target: agents/smaqit.functional.agent.md
sources:
  - framework/LAYERS.md
created: 2026-01-19
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| LAYERS.md | Functional — What? |
| LAYERS.md | Functional Foundation vs Feature Specs |
| LAYERS.md | Functional specs MUST / MUST NOT |

---

## L1 Directive Compilation

### Layer-Specific MUST Directives

- Define user flows that implement business use cases
- Specify data models with attributes and relationships
- Define API contracts (inputs, outputs, error conditions)
- Include state transitions where applicable
- Reference business specs for traceability using Implements (1:1 feature) or Enables (1:many foundation)
- Include justification when foundation spec has no Business references

### Layer-Specific MUST NOT Directives

- Specify technology choices (languages, frameworks, databases)
- Include deployment or infrastructure concerns
- Define performance benchmarks (those belong in Infrastructure)
- Prescribe implementation patterns

### Foundation Spec Rules

Functional specs fall into two categories:

| Type | Purpose | Business Reference |
|------|---------|--------------------|
| **Feature specs** | Implement a specific business use case | 1:1 mapping (Implements) |
| **Foundation specs** | Enable multiple business use cases | 1:many mapping (Enables) |

**Foundation specs** (shared components, cross-cutting concerns, common contracts) are legitimate engineering artifacts serving multiple business goals.

**Foundation spec requirements:**
- SHOULD reference all Business specs they enable
- MUST flag absence of Business references with justification

**Warning:** Orphaned foundations (no Business references, no justification) indicate scope creep.

### Upstream Context

**[UPSTREAM_CONTEXT_DESCRIPTION]:**
```
The Functional layer translates business goals into behavioral specifications. It receives coherence context from the Business layer.
```

**[UPSTREAM_SPEC_PATHS]:**
```
`specs/business/` (for traceability and coherence)
```

### User Input Description

**[USER_INPUT_DESCRIPTION]:**
```
User experience requirements describing the shape of interactions, behaviors, and data flows needed to satisfy business goals. Requirements should describe what the system does, not how it's implemented or deployed.
```

### Other Layers

**[OTHER_LAYERS]:**
```
Business, Stack, Infrastructure, Coverage
```

---

## Compilation Guidance for Agent-L2

When compiling `agents/smaqit.functional.agent.md`:

1. **Replace [LAYER]** → `functional`
2. **Replace [LAYER_NAME]** → `Functional`
3. **Replace [LAYER_PREFIX]** → `FUN`
4. **Replace [UPSTREAM_CONTEXT_DESCRIPTION]** → upstream context text above
5. **Replace [UPSTREAM_SPEC_PATHS]** → "`specs/business/` (for traceability and coherence)"
6. **Replace [USER_INPUT_DESCRIPTION]** → user input text above
7. **Replace [OTHER_LAYERS]** → "Business, Stack, Infrastructure, Coverage"
8. **Insert Layer-Specific MUST directives** → Functional-specific MUST rules
9. **Insert Layer-Specific MUST NOT directives** → Functional-specific MUST NOT rules
10. **Insert Foundation guidance** → Foundation vs Feature spec table and rules
11. **Preserve template structure** → All sections from specification-agent.template.md
