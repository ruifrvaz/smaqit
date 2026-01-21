---
layer: coverage
target: agents/smaqit.coverage.agent.md
sources:
  - framework/LAYERS.md
created: 2026-01-19
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| LAYERS.md | Coverage — What's verified? |
| LAYERS.md | Dependency Graph |
| LAYERS.md | Coverage specs MUST / MUST NOT |

---

## L1 Directive Compilation

### Layer-Specific MUST Directives

- Reference every acceptance criterion from upstream specs by ID
- Define a test case for each testable requirement
- Map: Requirement ID → Test Case → Expected Outcome
- Flag untestable requirements explicitly
- Include integration, E2E, and acceptance test definitions
- Report spec coverage (% of requirements with corresponding tests)

### Layer-Specific MUST NOT Directives

- Add acceptance criteria not present in upstream specs
- Skip upstream acceptance criteria without justification
- Modify or reinterpret upstream acceptance criteria
- Define unit tests (those are implementation details)

### Cross-Cutting Context

Coverage is a cross-cutting layer that reads from ALL upstream specifications:
- Business specs → verify user goals are testable
- Functional specs → verify behaviors are testable
- Stack specs → verify technology constraints are testable
- Infrastructure specs → verify operational requirements are testable

Every acceptance criterion across all layers must have a corresponding test case or explicit justification for being untestable.

### Upstream Context

**[UPSTREAM_CONTEXT_DESCRIPTION]:**
```
The Coverage layer ensures all requirements across all layers are testable and traceable. It receives acceptance criteria from Business, Functional, Stack, and Infrastructure layers, mapping each to verification tests.
```

**[UPSTREAM_SPEC_PATHS]:**
```
`specs/business/`, `specs/functional/`, `specs/stack/`, and `specs/infrastructure/` (source of acceptance criteria to verify)
```

### User Input Description

**[USER_INPUT_DESCRIPTION]:**
```
User test requirements describing test scope, test environment setup, integration points, and acceptance thresholds. Requirements should define how to verify that all upstream acceptance criteria are satisfied.
```

### Other Layers

**[OTHER_LAYERS]:**
```
Business, Functional, Stack, Infrastructure
```

---

## Compilation Guidance for Agent-L2

When compiling `agents/smaqit.coverage.agent.md`:

1. **Replace [LAYER]** → `coverage`
2. **Replace [LAYER_NAME]** → `Coverage`
3. **Replace [LAYER_PREFIX]** → `COV`
4. **Replace [UPSTREAM_CONTEXT_DESCRIPTION]** → upstream context text above
5. **Replace [UPSTREAM_SPEC_PATHS]** → "`specs/business/`, `specs/functional/`, `specs/stack/`, and `specs/infrastructure/` (source of acceptance criteria to verify)"
6. **Replace [USER_INPUT_DESCRIPTION]** → user input text above
7. **Replace [OTHER_LAYERS]** → "Business, Functional, Stack, Infrastructure"
8. **Insert Layer-Specific MUST directives** → Coverage-specific MUST rules
9. **Insert Layer-Specific MUST NOT directives** → Coverage-specific MUST NOT rules
10. **Insert Cross-cutting context** → Guidance about reading from all upstream layers
11. **Preserve template structure** → All sections from specification-agent.template.md
