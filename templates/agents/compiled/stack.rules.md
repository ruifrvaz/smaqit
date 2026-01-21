---
layer: stack
target: agents/smaqit.stack.agent.md
sources:
  - framework/LAYERS.md
created: 2026-01-19
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| LAYERS.md | Stack — With what? |
| LAYERS.md | Stack Foundation vs Feature Specs |
| LAYERS.md | Stack specs MUST / MUST NOT |

---

## L1 Directive Compilation

### Layer-Specific MUST Directives

- Document technology choices with rationale
- Define language versions and framework versions
- Specify libraries and their purposes
- Include build tools and development environment setup
- Be consistent with Functional specs (validated at implementation)
- Reference Functional specs using Enables (foundation serving multiple) or direct reference (feature serving one)
- Include justification when foundation spec has no Functional references

### Layer-Specific MUST NOT Directives

- Include code examples, implementation patterns, or architecture code blocks
- Define deployment topology or infrastructure
- Include compute, networking, or scaling decisions
- Specify cloud providers or hosting platforms
- Contradict functional requirements

### Foundation Spec Rules

Stack specs fall into two categories:

| Type | Purpose | Functional Reference |
|------|---------|--------------------|
| **Feature specs** | Technology choices for a specific feature | 1:1 mapping (Enables) |
| **Foundation specs** | Base technologies enabling multiple features | 1:many mapping (Enables) |

**Foundation specs** (base language environments, shared build tools, common dependencies) are legitimate engineering artifacts serving multiple functional requirements.

**Foundation spec requirements:**
- SHOULD reference all Functional specs they enable
- MUST flag absence of Functional references with justification

**Warning:** Orphaned foundations (no Functional references, no justification) indicate scope creep.

### Upstream Context

**[UPSTREAM_CONTEXT_DESCRIPTION]:**
```
The Stack layer selects technologies to implement functional behaviors. It receives coherence context from both Business and Functional layers.
```

**[UPSTREAM_SPEC_PATHS]:**
```
`specs/business/` and `specs/functional/` (for coherence and traceability)
```

### User Input Description

**[USER_INPUT_DESCRIPTION]:**
```
User technology preferences describing languages, frameworks, libraries, constraints, and team expertise. Requirements should justify technology selections based on functional needs and team context.
```

### Other Layers

**[OTHER_LAYERS]:**
```
Business, Functional, Infrastructure, Coverage
```

---

## Compilation Guidance for Agent-L2

When compiling `agents/smaqit.stack.agent.md`:

1. **Replace [LAYER]** → `stack`
2. **Replace [LAYER_NAME]** → `Stack`
3. **Replace [LAYER_PREFIX]** → `STK`
4. **Replace [UPSTREAM_CONTEXT_DESCRIPTION]** → upstream context text above
5. **Replace [UPSTREAM_SPEC_PATHS]** → "`specs/business/` and `specs/functional/` (for coherence and traceability)"
6. **Replace [USER_INPUT_DESCRIPTION]** → user input text above
7. **Replace [OTHER_LAYERS]** → "Business, Functional, Infrastructure, Coverage"
8. **Insert Layer-Specific MUST directives** → Stack-specific MUST rules
9. **Insert Layer-Specific MUST NOT directives** → Stack-specific MUST NOT rules
10. **Insert Foundation guidance** → Foundation vs Feature spec table and rules
11. **Preserve template structure** → All sections from specification-agent.template.md