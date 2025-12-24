# Templates

Templates define the structure that agents MUST follow when producing output. This document establishes the rules for both specification templates, agent templates and prompt templates.

## Template Types

smaqit uses three types of templates:

| Type | Location | Purpose | Produces |
|------|----------|---------|----------|
| **Specification templates** | `templates/specs/` | Structure for spec documents | `specs/**/*.md` |
| **Agent templates** | `templates/agents/` | Structure for agent definitions | `agents/*.agent.md` |
| **Prompt templates** | `templates/prompts/` | Structure for prompt files | `.github/prompts/*.prompt.md` |

## Placeholder Convention

All templates use `[PLACEHOLDER]` format (brackets, SCREAMING_CASE) for customizable values.

### Common Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `[LAYER]` | Lowercase layer name | `business` |
| `[LAYER_NAME]` | Title case layer name | `Business` |
| `[LAYER_PREFIX]` | 3-letter layer code | `BUS` |
| `[PHASE]` | Lowercase phase name | `development` |
| `[CONCEPT]` | Concept name in requirement ID | `LOGIN` |
| `[NNN]` | Sequential number (3 digits) | `001` |

### Agent Template Placeholders

**Shared placeholders:**

| Placeholder | Description |
|-------------|-------------|
| `[AGENT_DESCRIPTION]` | One-line agent purpose |
| `[UPSTREAM_SPEC_PATHS]` | Input spec paths |
| `[USER_INPUT_DESCRIPTION]` | What user input is accepted |

**Specification agent placeholders:**

| Placeholder | Description |
|-------------|-------------|
| `[LAYER]` | Lowercase layer name (e.g., `business`) |
| `[LAYER_NAME]` | Title case layer name (e.g., `Business`) |
| `[LAYER_PREFIX]` | 3-letter layer code (e.g., `BUS`) |
| `[LAYER_SPECIFIC_RULES]` | MUST/MUST NOT from LAYERS.md |

**Implementation agent placeholders:**

| Placeholder | Description |
|-------------|-------------|
| `[PHASE]` | Lowercase phase name (e.g., `development`) |
| `[PHASE_NAME]` | Title case phase name (e.g., `Development`) |
| `[AGENT_NAME]` | Agent display name (e.g., `Development Agent`) |
| `[PHASE_SPECIFIC_RULES]` | MUST/MUST NOT from PHASES.md |
| `[ROLE_DETAILS]` | Phase-specific role description |
| `[OUTPUT_ARTIFACTS]` | What artifacts are produced |
| `[OUTPUT_FORMAT]` | Format of output artifacts |
| `[ADDITIONAL_COMPLETION_CRITERIA]` | Phase-specific completion checks |

## Specification Templates

Specification templates define the structure for spec documents produced by specification agents.

### Location

```
templates/specs/
├── business.template.md
├── functional.template.md
├── stack.template.md
├── infrastructure.template.md
└── coverage.template.md
```

### Required Sections

Every specification template MUST include:

| Section | Purpose |
|---------|---------|
| Title | Concept name |
| References | Upstream spec links (except Business) |
| Scope | What's included and excluded |
| [Layer-specific content] | Varies by layer |
| Acceptance Criteria | Testable requirements with IDs |

### Compliance Rules

When producing specs from templates:

- Agents MUST use the template from `templates/specs/[LAYER].template.md`
- Agents MUST produce consistent output structure across all runs
- Agents MUST NOT add sections not defined in the template
- Agents MUST NOT omit required sections from the template
- Agents MUST NOT leave placeholder text in completed specs
- Agents MUST minimize variance in generated artifacts

### Placeholder Handling

- All placeholders MUST be replaced with actual content
- If a section is not applicable, state "Not applicable: [reason]"
- Empty sections are not permitted

## Agent Templates

Agent templates define the structure for agent definition files.

### Location

```
templates/agents/
├── specification-agent.template.md
├── implementation-agent.template.md
└── orchestrator-agent.template.md
```

### Required Sections

Every agent template MUST include:

| Section | Purpose |
|---------|---------|
| YAML Frontmatter | name, description, tools |
| Role | Agent's purpose, including core responsibilities |
| Framework Reference | Links to relevant framework files |
| Input | Upstream specs and user input |
| Output | Location, template, format |
| Directives | MUST/MUST NOT/SHOULD rules |
| Completion Criteria | Self-validation checklist |
| Failure Handling | Error response table |

### Agent Definition Format

Agent definitions use GitHub Custom Agent format:

```
---
name: smaqit.[layer]
description: [One-line description]
tools: ["read", "edit", "search"]
---

# [Layer] Agent

## Role
...

## Input
...

## Output
...
```

Note: The code fence above is for illustration only. Actual agent files start directly with the YAML frontmatter (`---`).

## Prompt Templates

Prompt templates define the structure for prompt files that serve as input records and agent invocation interface.

### Location

```
templates/prompts/
├── specification-prompt.template.md
├── implementation-prompt.template.md
└── orchestrator-prompt.template.md
```

### Required Sections

Every prompt template MUST include:

| Section | Purpose |
|---------|---------|
| YAML Frontmatter | name, description, agent |
| Purpose | What this prompt captures |
| Requirements | Sub-sections with suggested structure |
| Comment Examples | `<!-- Example: ... -->` for guidance |

### Prompt Template Format

Prompt templates use GitHub Copilot prompt format:

```markdown
---
name: smaqit.[layer]
description: [One-line description]
agent: smaqit.[layer]
---

# [Layer] Prompt

[Brief explanation]

## Requirements

[Sub-sections with suggested structure]

<!-- Example: [Guidance showing format] -->

[User fills requirements here]
```

### Free-Style with Structure

Prompts are **free-style natural language inputs**, not rigidly structured forms. Templates provide:

- **Suggested structure**: Sections and sub-sections to guide users
- **Commented examples**: `<!-- Example: ... -->` showing good formats
- **No enforcement**: Users write in their own words

Agents interpret natural language and request clarification if needed. See [PROMPTS](PROMPTS.md) for complete principles.

### Comment Convention

Templates and shipped prompts include examples wrapped in HTML comments:

```markdown
### Actors

<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->

[User writes actual actors here]
```

**Critical:** Agents MUST ignore HTML comments to prevent example requirements from contaminating generated specs.

### Single Manifest Pattern

Unlike specifications (one file per concept), prompts are **single manifest files**:

- One prompt per layer captures all requirements for that layer
- Users add features to existing prompts as projects evolve
- Prompts become consolidated input records for entire project

## Template Completeness

A template is complete when:

- [ ] All required sections are present
- [ ] Placeholders are clearly marked with `[PLACEHOLDER]` format
- [ ] Section purposes are unambiguous
- [ ] Layer-specific rules from LAYERS.md are incorporated (for spec templates)
- [ ] Comment examples use `<!-- Example: ... -->` format (for prompt templates)
