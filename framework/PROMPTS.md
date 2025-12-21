# Prompts

Prompts are the user-facing interface for smaqit workflows. They capture requirements as input records and invoke agents to generate specifications.

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
        ├── smaqit.develop.prompt.md
        ├── smaqit.deploy.prompt.md
        └── smaqit.validate.prompt.md
```

### Format

**YAML Frontmatter + Free-Style Content**

Prompts use GitHub Copilot prompt format with frontmatter specifying name, description, agent, and tools. See `templates/prompts/` for structure.

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

### Pre-Run Validation

Phase agents (Development, Deployment, Validation) perform **pre-flight checks** before orchestration:

**Development agent checks:**
- Business prompt has content
- Functional prompt has content
- Stack prompt has content

**Deployment agent checks:**
- Infrastructure prompt has content

**Validation agent checks:**
- Coverage prompt has content

If any upstream prompt is empty or insufficient, phase agent halts and guides user to fill missing prompts before proceeding.

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
| `smaqit.coverage.prompt.md` | Coverage | Test scope, verification requirements | Coverage Agent |

### Phase Prompts (Orchestration Prompts)

Coordinate multiple agents for workflow execution:

| Prompt | Phase | Orchestrates | Validates |
|--------|-------|--------------|-----------|
| `smaqit.develop.prompt.md` | Develop | Business → Functional → Stack → Development | All 3 layer prompts filled |
| `smaqit.deploy.prompt.md` | Deploy | Infrastructure → Deployment | Infrastructure prompt filled |
| `smaqit.validate.prompt.md` | Validate | Coverage → Validation | Coverage prompt filled |

Phase prompts may collect orchestration parameters (e.g., "Run in watch mode") but primarily validate and coordinate.

