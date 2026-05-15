# Skills

Skills are reusable agent capabilities invoked when agents detect specific conditions during execution. They encapsulate specialized workflows that multiple agent types share, avoiding duplication across agent definitions.

**Key Principles:**

- **Skills are invocable capabilities** — Agents invoke skills when detecting conditions requiring specialized handling
- **Condition-triggered** — Agents detect patterns triggering automatic skill invocation
- **Context-aware** — Skills receive context from invoking agent and return results
- **Approval mechanisms** — Skills handle user approval or autonomous decision-making based on invocation context

**Structure:**

- Skills reside in `.github/skills/[skill-name]/`
- Each skill directory contains `SKILL.md` with YAML frontmatter and markdown instructions
- Follow agentskills.io specification format
- Optional directories: `scripts/`, `references/`, `assets/` for supporting resources

## Skills as Invocable Capabilities

Skills provide structured workflows for common agent needs. When agents detect conditions requiring specialized handling, they invoke the relevant skill rather than implementing the workflow inline.

**Skills complement agent behaviors:**

- Agents maintain bounded scope and core responsibilities
- Skills provide reusable workflows for cross-cutting concerns
- Skills avoid duplicating logic across multiple agent definitions
- Skills enable consistent handling of shared patterns

## Skill Structure

### Location

Skills live in `.github/skills/`. Each skill occupies its own subdirectory following agentskills.io specification.

**User project structure:**
```
project/
└── .github/
    └── skills/
        └── smaqit.input-business/
            ├── SKILL.md          # Required: YAML frontmatter + instructions
            ├── scripts/          # Optional: executable code
            ├── references/       # Optional: additional documentation
            └── assets/           # Optional: templates, images, data
```

### Format

Skills follow agentskills.io specification with `SKILL.md` containing YAML frontmatter and markdown instructions.

**SKILL.md frontmatter (required fields):**

```yaml
---
name: skill-name                # 1-64 chars, lowercase alphanumeric + hyphens
description: What the skill does and when to use it (1-1024 chars)
---
```

**SKILL.md frontmatter (optional fields):**

```yaml
---
name: skill-name
description: Description with keywords for agent identification
license: Apache-2.0             # License name or reference to license file
compatibility: GitHub Copilot   # Environment requirements if needed
metadata:
  author: organization          # Arbitrary key-value metadata
  version: "1.0"
allowed-tools: edit search      # Space-delimited pre-approved tools (experimental)
---
```

**SKILL.md body content:**

Markdown instructions with no format restrictions. Recommended sections include step-by-step instructions, input/output examples, and edge cases. Keep main SKILL.md under 500 lines, moving detailed reference material to separate files in `references/`.

**Progressive disclosure:**

- Metadata (~100 tokens): `name` and `description` loaded at startup for all skills
- Instructions (<5000 tokens recommended): Full SKILL.md body loaded when skill activated
- Resources (as needed): Additional files loaded only when required

## Condition Detection

Agents detect conditions through input analysis. Detection triggers skill invocation automatically when patterns match predefined criteria. The skill then executes its workflow and returns control to the invoking agent.

**Detection patterns:**

- Ambiguity in requirements (multiple interpretations possible)
- Conflicting inputs (contradictions within or across sources)
- Insufficient detail (gaps requiring assumptions)
- Complexity indicators (multi-part work requiring explicit planning)

## Skill Invocation Context

Skills receive context from the invoking agent:

- Input sources and their content
- Detected conditions triggering invocation
- Agent identity and current phase/layer
- User interaction mode (direct vs autonomous)

Skills return results or recommendations:

- Assessments and analysis
- Generated plans or proposals
- User approvals or selections
- Revised inputs or clarifications

## Handling Ambiguity

Input skills (`smaqit.input-[layer]`) detect common issues before spec generation: insufficient requirements, conflicts with upstream specs, and internal contradictions. Any ambiguity that survives input validation is handled inline by the agent:

- Surface the specific issue with enough detail for the user to act on it
- State clearly what information is missing or contradictory
- Request clarification before proceeding
- Do not invent or silently resolve ambiguous requirements

## Future Skills

Skills expand as common patterns emerge across agents. Potential future skills include:

- **Conflict resolution** — Reconcile contradictory requirements across layers
- **Gap detection** — Identify missing requirements before implementation
- **Coherence validation** — Verify cross-layer consistency
- **Decomposition** — Break large specs into manageable components
- **Traceability mapping** — Generate requirement chains across layers

New skills emerge through observed patterns, not speculation. When multiple agents duplicate logic, that logic becomes a candidate for skill extraction.
