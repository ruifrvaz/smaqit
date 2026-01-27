# Prompts

Prompts are the user-facing interface for smaqit workflows. They capture requirements as input records and invoke agents to generate specifications.

**Key Principles:**

- **Prompts are input records** — Capture user requirements for reproducibility and auditability
- **Free-style natural language** — Users write in their own words, agents interpret
- **HTML comment examples** — `<!-- Example: ... -->` provide guidance without rigid enforcement
- **Single manifest per layer** — One prompt file accumulates all requirements for a layer (unlike specs which are one-per-concept)

**Structure:**

- YAML frontmatter: `name`, `description`, `agent`
- Requirement sections with layer-specific sub-sections
- `<!-- Example: ... -->` comments for guidance (agents MUST ignore these)
- Free-style user content in natural language

**Prompt types:**

- **Layer prompts** (5) — Capture requirements for specification layers (business, functional, stack, infrastructure, coverage)
- **Phase prompts** (3) — Trigger phase implementation agents
- **Agent creation prompts** — Specify requirements for new base agents

## Prompts as Input Records

**Prompts are versioned input records capturing user requirements at each layer.**

Filled prompts should be committed to version control alongside specs. When requirements change, users edit prompt files and regenerate specs.

## Prompt Structure

### Location

Prompts live in `.github/prompts/`. This location enables `/smaqit.[layer]` slash command invocation.

**User project structure:**
```
project/
└── .github/
    └── prompts/
        ├── smaqit.business.prompt.md
        ├── smaqit.functional.prompt.md
        ├── smaqit.stack.prompt.md
        ├── smaqit.infrastructure.prompt.md
        ├── smaqit.coverage.prompt.md
        ├── smaqit.development.prompt.md
        ├── smaqit.deployment.prompt.md
        ├── smaqit.validation.prompt.md
        └── smaqit.new-agent.prompt.md
```

### Format

**YAML Frontmatter + Free-Style Content**

Prompts use GitHub Copilot prompt format with frontmatter specifying name, description, and agent. See `templates/prompts/` for structure.

### Single Manifest per Layer

Unlike specifications (one file per concept), prompts are **single manifest files** that capture all requirements for a layer:

- **Business prompt**: All use cases, actors, goals for the project
- **Functional prompt**: All behaviors, data models, contracts for the project
- **Stack prompt**: All technology choices and rationale for the project

As projects evolve, users add requirements to existing prompts rather than creating new prompt files. This creates a consolidated input record for the entire project at each layer.

### Free-Style with Suggested Structure

Prompts are **natural language inputs**, not rigidly structured forms. Templates provide suggested structure (sections, sub-sections) but users write requirements in their own words. Agents interpret and request clarification if needed.

### Comment Convention for Examples

**Agents MUST ignore HTML comments** (`<!-- -->`). Templates include examples wrapped in `<!-- Example: ... -->` comments for user guidance only.

## Agent Interaction

### Reading Prompts

Agents read prompt files from `.github/prompts/` at the start of execution:

1. **Locate prompt**: Agent finds corresponding prompt file (e.g., Business Agent reads `smaqit.business.prompt.md`)
2. **Ignore comments**: Agent strips all HTML comments before interpretation
3. **Parse requirements**: Agent interprets free-style content per layer expectations
4. **Validate sufficiency**: Agent checks if enough information provided

### Validation Pattern

Agents apply **Fail-Fast on Ambiguity** when reading prompts:

**If prompt empty or insufficient:**
- Agent halts execution
- Agent suggests what's missing using natural language guidance
- Agent waits for user to fill prompt and re-invoke

Agents guide users naturally, not with template references or error codes.

**If prompt is filled sufficiently:**
- Agent proceeds with spec generation
- Agent uses prompt content as authoritative input

## Amendment Workflow

When requirements change, users edit prompts and regenerate specs. Prompts are the source, specs are derived. Agents always read from `.github/prompts/`.

## Prompt Types

### Specification Prompts (Layer Prompts)

Capture requirements for single specification layer:

| Prompt | Layer | Captures | Invokes |
|--------|-------|----------|---------|
| `smaqit.business.prompt.md` | Business | Use cases, actors, goals | Business Agent |
| `smaqit.functional.prompt.md` | Functional | Behaviors, data, contracts | Functional Agent |
| `smaqit.stack.prompt.md` | Stack | Technologies, tools, rationale | Stack Agent |
| `smaqit.infrastructure.prompt.md` | Infrastructure | Deployment, scaling, observability | Infrastructure Agent |
| `smaqit.coverage.prompt.md` | Coverage | Test scope, environment, thresholds | Coverage Agent |

### Implementation Prompts

Trigger single implementation agent with optional execution parameters:

| Prompt | Phase | Captures | Invokes |
|--------|-------|----------|---------|
| `smaqit.development.prompt.md` | Development | Build options, output preferences | Development Agent |
| `smaqit.deployment.prompt.md` | Deployment | Deployment target, verification | Deployment Agent |
| `smaqit.validation.prompt.md` | Validation | Execution scope, failure handling | Validation Agent |

Implementation prompts collect minimal runtime parameters (watch mode, verbosity, skip flags). Agents handle orchestration, validation, and error handling.

### Agent Creation Prompts

Define specifications for new base agents (Q&A, helper, orchestrator, custom utilities):

| Prompt | Agent Type | Usage | Invokes |
|--------|------------|-------|---------|
| `smaqit.new-agent.prompt.md` | All base agents | Interactive template for gathering specifications | Agent-L2 |

**Agent creation prompt differs from layer/phase prompts:**

| Aspect | Layer/Phase Prompts | Agent Creation Prompt |
|--------|---------------------|----------------------|
| **Location** | `.github/prompts/` | `.github/prompts/` |
| **File** | `smaqit.[layer].prompt.md` | `smaqit.new-agent.prompt.md` |
| **Pattern** | Single manifest per layer/phase | Interactive template (reusable) |
| **Content** | Free-style requirements | Structured directives (MUST/MUST NOT/SHOULD) |
| **Usage** | Filled by user, read by agent | Followed by Agent-L2, filled interactively |
| **Agent** | Layer/phase-specific agent | Agent-L2 (compiler) |
| **Output** | Specifications or artifacts | Product agents |

**Agent Creation Workflow:**

1. **User invokes Agent-L2** — Request new agent creation
2. **Agent-L2 reads template** — Loads structure from `.github/prompts/smaqit.new-agent.prompt.md`
3. **Interactive gathering** — Agent-L2 requests user input for each placeholder (name, description, tools, directives, scope)
4. **3-way merge** — L1 base template + L1 base rules + user-provided specifications → product agent
5. **Document specifications** — User inputs recorded in compilation log (template remains unchanged)
6. **Output agent** — `agents/smaqit.[agent-name].agent.md` created
7. **Use agent** — Invoke via `/smaqit.[agent-name]` in GitHub Copilot

**Directive Format:**

Agent creation prompts contain explicit MUST/MUST NOT/SHOULD statements. These directives are compiled directly into the product agent, merging with foundation directives from `templates/agents/compiled/base.rules.md`.

**See:** `.github/prompts/smaqit.new-agent.prompt.md` for complete structure and guidance.

