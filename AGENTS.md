# AGENTS.md — smaqit Development

You are developing smaqit, a spec-driven agent orchestration framework. This file is the
canonical source of coding-agent instructions for **GitHub Copilot** and **Codex**, both of
which read `AGENTS.md` natively. Claude Code reads `CLAUDE.md`, which is a thin pointer
(`@AGENTS.md`) back to this file — do not duplicate content there.

## Kit Components

- **Framework files** (`framework/`) — LLM execution instructions
- **Wiki** (`docs/wiki/`) — Human-readable context and rationale (concepts, designs, patterns, workflows)
- **Specification templates** (`templates/specs/`) — Structure for spec documents per layer
- **Design templates** (`templates/designs/`) — PlantUML design pair templates per layer
- **Agent sources** (`agents/*.md`) — canonical agent bodies, compiled per platform
- **Agent metadata** (`.smaqit/definitions/agents/*.frontmatter.yaml`) — per-platform frontmatter and placeholder values
- **Skills** (`skills/<name>/SKILL.md`) — one shared source tree, compiled per platform
- **Installer** (`installer/`) — Go CLI that scaffolds smaqit into user projects and installs agents/skills globally

## Source vs Artifacts

**Source (this repo)** — Kit development files:
```
smaqit/
├── framework/
│   ├── SMAQIT.md             # Index + core principles
│   ├── LAYERS.md             # Layer definitions
│   ├── PHASES.md             # Phase workflows
│   ├── TEMPLATES.md          # Template structure rules
│   ├── AGENTS.md             # Agent behaviors
│   ├── ARTIFACTS.md          # Artifact rules
│   └── SKILLS.md             # Skill structure rules
├── templates/
│   ├── specs/                 # Specification templates (5)
│   ├── designs/                # PlantUML design pair templates (5)
│   ├── AGENTS.md.template      # Installed into user projects
│   └── CLAUDE.md.template      # Installed into user projects
├── agents/*.md                        # 9 canonical agent bodies (no frontmatter)
├── .smaqit/definitions/agents/*.yaml  # Per-platform frontmatter + placeholder values
├── skills/<name>/SKILL.md             # 26 shared skill sources
├── installer/main.go                  # CLI tool (Go embeds compiled output)
├── docs/
│   ├── wiki/                 # Human-readable rationale
│   ├── logs/                 # Session logs (meta)
│   └── ...
└── README.md                 # User docs
```

**Artifacts (generated)** — never committed at the repo root, rebuilt by `make -C installer prepare`:
```
installer/
├── agents-copilot/*.agent.md   # GitHub Copilot custom agents
├── agents-claude/*.md          # Claude Code subagents
├── agents-codex/*.toml         # Codex project custom agents
├── commands-claude/*.md        # Claude Code slash commands
├── skills-shared/<name>/**     # Skills for GitHub Copilot + Codex (installed to ~/.agents/skills/)
└── skills-claude/<name>/**     # Skills for Claude Code (installed to ~/.claude/skills/)
```

**Global installation (user machines)** — installed once per user by `install.sh` / `--install-global`,
never per-project:
```
~/.copilot/agents/       # GitHub Copilot custom agents
~/.claude/{agents,commands,skills}/  # Claude Code
~/.codex/agents/         # Codex project custom agents
~/.agents/skills/        # Shared skills — read by GitHub Copilot and Codex
```

**Per-project scaffolding (`smaqit init`)** — project-local only, no agents or skills:
```
user-project/
├── .smaqit/               # Task/spec state, templates, reports, PlantUML runtime
├── specs/{business,functional,stack,infrastructure,coverage}/
├── docs/designs/{business,functional,stack,infrastructure,coverage}/
├── .github/workflows/     # Copilot setup workflow + CI
├── AGENTS.md              # Project instructions (installed once, never overwritten)
└── CLAUDE.md              # Thin @AGENTS.md pointer
```

## Communication Style

### Provide step-by-step explanations

When performing work:

1. **Before taking action** — Explain what you're about to do and why
2. **During execution** — Describe each step as you perform it
3. **After completion** — Summarize what was accomplished

**Be verbose and educational** — Help the user understand the process, not just the outcome.

## Content Guidelines

### Templates, agents, and skills contain directives

**Templates** (`templates/`), **agents** (`agents/`), and **skills** (`skills/`) contain execution instructions:
- What to do (directives, rules, structure)
- How to structure output (templates, formats)
- When to execute (conditions, triggers)
- Where to find input and write output (file paths)
- Validation criteria (MUST/MUST NOT/SHOULD rules)

### Wiki files contain human context

**Wiki files** (`docs/wiki/`, `README.md`, `.smaqit/tasks/`, `docs/logs/`) contain context for humans:
- Why the framework is designed this way
- Trade-offs between alternatives
- Examples with multiple scenarios
- Historical decisions and evolution
- Business context when relevant
- Extended explanations and tutorials

### When editing the installer

The installer embeds and installs these files:
- `templates/specs/*.md`, `templates/designs/*.md` → project-local `.smaqit/templates/`
- `agents/*.md` + `.smaqit/definitions/agents/*.yaml` → compiled per platform → global agent directories
- `skills/**/*.md` → compiled once for GitHub Copilot + Codex (`skills-shared/`) and once for Claude Code (`skills-claude/`) → global skill directories
- `templates/AGENTS.md.template`, `templates/CLAUDE.md.template` → project-local `AGENTS.md`/`CLAUDE.md` (installed once, appended to if the project already has its own content)

**Installer subdirectories are gitignored and must never be manually synced or committed:**
- `installer/agents-copilot/`, `installer/agents-claude/`, `installer/agents-codex/`, `installer/commands-claude/` — populated by `make prepare` from `agents/`, `.smaqit/definitions/agents/`, `commands/`
- `installer/skills-shared/`, `installer/skills-claude/` — populated by `make prepare` from `skills/`
- `installer/framework/`, `installer/templates/` — populated by `make prepare` from `framework/`, `templates/`

To update embedded files, run `make -C installer prepare`. Never copy files into these directories by hand.

### Version Sync

Keep `installer/main.go` `Version` const in sync with `SMAQIT.md` version.

**Quick commands:**

```bash
# Manual installer test
cd installer && make build && mkdir -p test && cd test
../dist/smaqit-dev init && ../dist/smaqit-dev status
cd .. && make uninstall  # Also cleans test/ and embedded files

<!-- smaqit:instructions:begin -->
# smaqit

This project uses [smaqit](https://github.com/ruifrvaz/smaqit) for spec-and-design-driven AI-assisted development. Specifications are canonical for requirements and acceptance criteria. PlantUML sources under `docs/designs/` are canonical for system structure, and their PNG projections are mandatory for specification-agent visual validation. Implementation agents consume validated specifications and PlantUML source without duplicating either.

## Scaffolding to ignore

The following paths are smaqit tooling, not part of this project's business domain. When reasoning about architecture, domain logic, or conventions, ignore them:

- `.smaqit/` — task/spec state, session history, templates
- `.github/workflows/copilot-setup-steps.yml` — Copilot coding-agent bootstrap workflow
- User-level agent and skill directories (`~/.copilot/agents/`, `~/.claude/`, `~/.codex/agents/`, `~/.agents/skills/`) are framework tooling, not project-domain source

## Working with smaqit

Specs live under `specs/{business,functional,stack,infrastructure,coverage}/`, one file per distinct concept, each tracking a lifecycle status (`draft` → `implemented` → `deployed` → `validated`).

Design pairs live under `docs/designs/{business,functional,stack,infrastructure,coverage}/`. Every active spec references at least one same-layer PlantUML Markdown/PNG pair. Specification agents MUST first confirm their declared `smaqit-plantuml` tools are available; if the client has not loaded, trusted, or exposed them, stop with `DESIGN-TOOLCHAIN-UNAVAILABLE` rather than substituting direct CLI authoring. They then validate PlantUML syntax, render a current PNG, open that PNG with an image-reading tool, and pass the visual rubric before handoff. If image content cannot be read, stop with `DESIGN-VISION-UNAVAILABLE`; reading PlantUML source is never an authoring-time visual-review fallback. `smaqit plan --phase` blocks downstream work unless that attestation remains current; implementation agents then read PlantUML source directly and do not repeat image review.

Invoke with `/` in GitHub Copilot chat or Claude Code. In Codex, ask it to spawn the corresponding named agent (for example, `smaqit.development`):

| Command | Purpose |
|---|---|
| `/smaqit.development` | Build a working app from Business/Functional/Stack specs |
| `/smaqit.deployment` | Deploy using Infrastructure specs |
| `/smaqit.validation` | Run tests against the deployed system |
| `/smaqit.qa` | Answer questions about the smaqit framework |

On GitHub Copilot, the five specification agents (`/smaqit.business`, `/smaqit.functional`, `/smaqit.stack`, `/smaqit.infrastructure`, `/smaqit.coverage`) are also directly invocable. On Claude Code they run automatically as part of the phase commands above. In Codex, phase agents spawn these named project subagents; repository skills are available through `/skills` or `$` mentions.

Run `smaqit status` for spec/design readiness and next steps, `smaqit design validate` for design gates, or `smaqit help` for the full CLI reference.
<!-- smaqit:instructions:end -->
