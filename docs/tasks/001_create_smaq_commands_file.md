# Task: Create Copilot Prompt Files

**ID**: 001
**Status**: new

## Context

Create VS Code Copilot prompt files (`.prompt.md`) for smaqit phase and layer operations. These prompts enable users to invoke smaqit workflows via GitHub Copilot chat using `/smaqit.` commands (e.g., `/smaqit.develop`, `/smaqit.business`). Prompts are stored in `.github/prompts/` and follow VS Code prompt file conventions.

## Acceptance Criteria

### Phase Prompts

- [ ] Create `smaqit.develop.prompt.md` — Orchestrates business → functional → stack spec agents, then development agent
- [ ] Create `smaqit.deploy.prompt.md` — Orchestrates infrastructure spec agent, then deployment agent
- [ ] Create `smaqit.validate.prompt.md` — Orchestrates coverage spec agent, then validation agent

### Layer Prompts (Specification Agents)

- [ ] Create `smaqit.business.prompt.md` — Invokes business agent with interactive prompts for stakeholder goals
- [ ] Create `smaqit.functional.prompt.md` — Invokes functional agent with interactive prompts for experience requirements
- [ ] Create `smaqit.stack.prompt.md` — Invokes stack agent with interactive prompts for technology preferences
- [ ] Create `smaqit.infrastructure.prompt.md` — Invokes infrastructure agent with interactive prompts for deployment requirements
- [ ] Create `smaqit.coverage.prompt.md` — Invokes coverage agent with interactive prompts for verification requirements

### Prompt File Structure

Each prompt file MUST include:

- [ ] YAML frontmatter with: `name`, `description`, `agent` (reference to corresponding `.agent.md`), `tools`
- [ ] Interactive input variables using `${input:variableName:placeholder}` syntax for user input
- [ ] Markdown body with clear instructions for the LLM
- [ ] Reference to framework files using relative paths (e.g., `[SMAQIT.md](../../framework/SMAQIT.md)`)

### Integration

- [ ] Prompts reference corresponding agents in `.github/agents/` via `agent` frontmatter field
- [ ] Phase prompts orchestrate multiple agents in sequence
- [ ] Layer prompts pass user input to agents as context

## Notes

**VS Code Prompt File Conventions:**
- Extension: `.prompt.md`
- Location: `.github/prompts/` (workspace-scoped)
- Invocation: Type `/` in Copilot chat followed by prompt name
- Format: YAML frontmatter + Markdown body

**Agent invocation:**
GitHub Copilot manages agent invocation. Prompts reference agents via the `agent` field in frontmatter.

**User input examples:**
- Business: `${input:goals:Describe stakeholder goals and success criteria}`
- Functional: `${input:experience:Describe the user experience and behaviors}`
- Stack: `${input:tech:Specify preferred languages, frameworks, or constraints}`
- Infrastructure: `${input:deployment:Describe target environment and resource constraints}`
- Coverage: `${input:verification:Specify test scope and performance benchmarks}`
