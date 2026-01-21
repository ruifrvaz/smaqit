---
layer: business
target: agents/smaqit.business.agent.md
sources:
  - framework/LAYERS.md
created: 2026-01-19
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| LAYERS.md | Business — Why? |
| LAYERS.md | Business Actors |
| LAYERS.md | Business Separation Principle |
| LAYERS.md | Business specs MUST / MUST NOT |

---

## L1 Directive Compilation

### Layer-Specific MUST Directives

- Express requirements as user goals and needs
- Use language accessible to non-technical stakeholders
- Define acceptance criteria from user perspective
- Capture the "why" behind each requirement
- Capture actor diversity: interactive participants AND non-functional requirement stakeholders

### Layer-Specific MUST NOT Directives

- Mention specific technologies, frameworks, or libraries
- Include implementation details or technical solutions
- Define data structures or API contracts
- Reference deployment or infrastructure concerns
- Describe HOW features work (behaviors and mechanisms belong in Functional layer)
- Reference technical artifacts (console, terminal, screen, database, API, server, client, encoding)
- Include technical error handling or fallback mechanisms

### Upstream Context

**[UPSTREAM_CONTEXT_DESCRIPTION]:**
```
The Business layer is the first specification layer. It receives requirements directly from user input with no upstream specifications.
```

**[UPSTREAM_SPEC_PATHS]:**
```
None (Business is the first layer)
```

### User Input Description

**[USER_INPUT_DESCRIPTION]:**
```
User goals, needs, constraints, and desired outcomes. Requirements should describe what users want to achieve and why it matters, without prescribing how it will be implemented.
```

### Other Layers

**[OTHER_LAYERS]:**
```
Functional, Stack, Infrastructure, Coverage
```

---

## Compilation Guidance for Agent-L2

When compiling `agents/smaqit.business.agent.md`:

1. **Replace [LAYER]** → `business`
2. **Replace [LAYER_NAME]** → `Business`
3. **Replace [LAYER_PREFIX]** → `BUS`
4. **Replace [UPSTREAM_CONTEXT_DESCRIPTION]** → upstream context text above
5. **Replace [UPSTREAM_SPEC_PATHS]** → "None (Business is the first layer)"
6. **Replace [USER_INPUT_DESCRIPTION]** → user input text above
7. **Replace [OTHER_LAYERS]** → "Functional, Stack, Infrastructure, Coverage"
8. **Insert Layer-Specific MUST directives** → Business-specific MUST rules
9. **Insert Layer-Specific MUST NOT directives** → Business-specific MUST NOT rules
10. **Preserve template structure** → All sections from specification-agent.template.md