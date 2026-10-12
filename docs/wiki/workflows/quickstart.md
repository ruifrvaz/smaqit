# Quickstart: Mario Hello

Build a "Hello, Mario!" console app from requirements to working code using smaqit's spec-driven workflow.

## Prerequisites

- smaqit installed (`curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash`). The installer puts smaqit's agents, commands, and skills in your user-level directories once (`~/.claude/`, `~/.copilot/agents/`, `~/.codex/agents/`, `~/.agents/skills/`), so every project can use them
- Node.js 22 or newer on `PATH` (all JavaScript/WASM packages are already embedded in smaqit)
- GitHub Copilot in VS Code, Claude Code, or OpenAI Codex
- A new or existing project directory

## Step 1: Initialize smaqit

```bash
mkdir mario-hello && cd mario-hello
smaqit init
```

This creates project-local files only (agents and skills were already installed for your user):
- `.smaqit/` — Templates, project state, and the bundled PlantUML runtime (`.smaqit/tools/`, added to `.gitignore`)
- `.smaqit/slots.yml` — Deployment slot bindings; every slot uses its default until you edit it
- `specs/` — Where generated specifications go
- `docs/designs/{business,functional,stack,infrastructure,coverage}/` — Canonical PlantUML Markdown/PNG design pairs
- `.vscode/mcp.json`, `.mcp.json`, `.codex/config.toml` — `smaqit-plantuml` MCP registration for VS Code, Claude Code, and Codex
- `AGENTS.md` and `CLAUDE.md` — Project instructions, including when to use smaqit; `CLAUDE.md` imports `AGENTS.md`
- `.github/workflows/copilot-setup-steps.yml` — Installs smaqit for the Copilot coding agent

Initialization and project update automatically run `smaqit mcp verify`, which proves the VS Code, Claude Code, and Codex registrations plus local transport. Restart the active client after installation: trust `smaqit-plantuml` in VS Code and Claude Code, and trust the project before Codex loads `.codex/config.toml`. A fresh authoring-agent session must still expose the two MCP tools or the specification agent stops with `DESIGN-TOOLCHAIN-UNAVAILABLE`. Re-run `smaqit mcp verify` manually to diagnose only the deterministic local layer.

## Step 2: Run the Development Phase

In GitHub Copilot chat or Claude Code, invoke the Development phase with your requirements in the same message:

```
/smaqit.development I want a simple console application that displays "Hello, Mario!" when run.

Actors:
- Mario Fan: A user who loves Nintendo's Mario franchise and wants a fun greeting

The app should:
- Display "Hello, Mario!" to standard output
- Exit successfully after displaying the message
- Use Python 3.8+ with no external dependencies
```

In Codex, ask it to spawn the `smaqit.development` agent with the same requirements. In Claude Code you can also skip the command: the project's `AGENTS.md` tells Claude to use smaqit for any request that adds or changes behavior, so asking for the app in plain language routes through the same workflow.

The agent will:
1. Generate Business, Functional, and Stack specifications in `specs/` with linked, visually reviewed design pairs in `docs/designs/`
2. Build the application code
3. Run tests to verify it works

If your requirements are unclear or incomplete, the agent asks clarifying questions before writing anything. Answer them in the chat and it continues where it left off. To review and approve each specification layer before the code is written, ask for **assisted** mode (for example, "use assisted mode, I want to review each specification").

## Expected Output

After the Development phase completes, you'll have:

```
mario-hello/
├── .smaqit/
│   └── reports/development-phase-report-YYYY-MM-DD.md
├── AGENTS.md
├── CLAUDE.md
├── docs/designs/          # design pairs for each specification
├── specs/
│   ├── business/
│   │   └── greet-mario-fan.md
│   ├── functional/
│   │   └── console-output.md
│   └── stack/
│       └── python-console.md
├── src/
│   └── main.py
├── tests/
│   └── test_main.py
└── README.md
```

Run it:

```bash
python src/main.py
# Output: Hello, Mario!
```

## What's Next?

- **Add features**: Invoke `smaqit.development` again with updated requirements, or use the `smaqit.feature-new` skill once the app is deployed
- **Deploy**: Invoke the `smaqit.deployment` agent with your infrastructure requirements
- **Validate**: Invoke the `smaqit.validation` agent with your test requirements
- **Check status**: Run `smaqit status` to see spec coverage

## Troubleshooting

**Agent asks for clarification?** Your requirements may be ambiguous. Add more detail in the chat response.

**Specs don't match expectations?** Invoke the relevant specification agent with updated requirements. In GitHub Copilot use it directly (for example, `/smaqit.business`). In Claude Code it has no slash command, so ask Claude to delegate to it by name (for example, "use the smaqit-business agent to revise the business specification").

**Build fails?** Review the Development agent's output. It will indicate what went wrong and suggest fixes.
