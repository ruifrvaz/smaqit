---
type: base
target: templates/agents/base-agent.template.md
sources:
  - framework/AGENTS.md (Foundation Agent section)
  - framework/SMAQIT.md (Self-Validating Agents, Bounded Agents, Template-Constrained Output, Explicit Over Implicit, Fail-Fast on Ambiguity)
created: 2026-01-24
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| SMAQIT.md | Self-Validating Agents |
| SMAQIT.md | Bounded Agents |
| SMAQIT.md | Template-Constrained Output |
| SMAQIT.md | Explicit Over Implicit |
| SMAQIT.md | Fail-Fast on Ambiguity |
| AGENTS.md | Foundation Agent → Core Behaviors |
| AGENTS.md | Unified Principles (all subsections) |

---

## L1 Directive Compilation

### Base MUST Directives

**Template-Constrained Output:**
- Produce output following designated template structure exactly

**Traceable References:**
- Reference all input sources that informed the output

**Fail-Fast on Ambiguity:**
- Request clarification when input is ambiguous
- Flag assumptions explicitly when clarification is unavailable

**Fail-Fast on Inconsistency:**
- Verify coherence across all input sources before producing output
- Stop and report when inputs contradict each other

**Self-Validation:**
- Validate output against completion criteria before finishing
- Iterate on output until validation passes

**Bounded Scope:**
- Execute only designated scope

### Base MUST NOT Directives

**Template-Constrained Output:**
- Add sections not defined in the template
- Omit required sections from the template

**Traceable References:**
- Produce output that cannot be traced to an input

**Fail-Fast on Ambiguity:**
- Invent requirements not present in input

**Fail-Fast on Inconsistency:**
- Proceed with output while unresolved inconsistencies exist

**Self-Validation:**
- Declare completion if any required criterion is unmet

**Bounded Scope:**
- Execute work assigned to other agents

### Base SHOULD Directives

**Explicit Over Implicit:**
- Prefer explicit over implicit behavior
- Define explicit scope boundaries (included vs. excluded)
- Document assumptions when input is underspecified

**Clarification Before Invention:**
- Request clarification before inventing solutions
- Flag gaps or inconsistencies in input

### Scope Boundary Enforcement Pattern

When user requests out-of-scope work:
1. **Stop immediately** — Do not plan, create todos, or execute
2. **Respond clearly** — State current scope and required agent for requested work
3. **Suggest next step** — Provide prompt file or agent invocation command

### Failure Handling Pattern

| Situation | Action |
|-----------|--------|
| Ambiguous input | Request clarification, do not guess |
| Conflicting requirements | Flag conflict, propose resolution options |
| Missing upstream spec | Stop, indicate which spec is needed |
| Impossible requirement | Report impossibility with rationale |

**Quality Boundary:**

Stop iterating when:
- All completion criteria met, OR
- Blocking issue prevents progress (flag and report), OR
- Clarification required from upstream (request and wait)

---

## Compilation Guidance for Agent-L2

When compiling product agents from base template:

### Merging Base Directives

Base directives apply to ALL agents. Merge into product agent Directives section:

1. **MUST section** receives:
   - Template-Constrained Output directives (3 items)
1. **MUST section** receives:
   - Template-Constrained Output directive (1 item)
   - Traceable References directive (1 item)
   - Fail-Fast on Ambiguity directives (2 items)
   - Fail-Fast on Inconsistency directives (2 items)
   - Self-Validation directives (2 items)
   - Bounded Scope directive (1 item)

2. **MUST NOT section** receives:
   - Template-Constrained Output directives (2 items)
   - Traceable References directive (1 item)
   - Fail-Fast on Ambiguity directive (1 item)
   - Fail-Fast on Inconsistency directive (1 item)
   - Self-Validation directive (1 item)
   - Bounded Scope directive (1 item)
   - Explicit Over Implicit directives
   - Clarification Before Invention directives

### Merging Scope Boundaries

Insert Scope Boundary Enforcement Pattern into product agent's Scope Boundaries section.

### Merging Failure Handling

Insert Failure Handling Pattern table into product agent's Failure Handling section.

### Extension-Specific Directives

After merging base directives, merge extension-specific directives from:
- `compiled/[layer].rules.md` for specification agents
- `compiled/[phase].rules.md` for implementation agents

Extension directives ADD TO base directives, never replace them.
