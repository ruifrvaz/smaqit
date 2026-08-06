# Quickstart: Mario Hello

Build a "Hello, Mario!" console app from requirements to working code using smaqit's spec-driven workflow.

## Prerequisites

- smaqit installed (`curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash`)
- Node.js 22 or newer on `PATH` (all JavaScript/WASM packages are already embedded in smaqit)
- GitHub Copilot in VS Code, Claude Code, or OpenAI Codex
- A new or existing project directory

## Step 1: Initialize smaqit

```bash
mkdir mario-hello && cd mario-hello
smaqit init
smaqit mcp verify
```

This creates:
- `.smaqit/` — Framework files and templates
- `.github/agents/` — Agent definitions
- `.github/skills/` — Reusable agent skills
- `.claude/agents/`, `.claude/commands/`, `.claude/skills/` — Claude Code integration
- `.codex/agents/`, `.agents/skills/` — Codex project agents and repository skills
- `specs/` — Where generated specifications go
- `docs/designs/{business,functional,stack,infrastructure,coverage}/` — Canonical PlantUML Markdown/PNG design pairs

Before authoring designs, let the client load and trust the project-local `smaqit-plantuml` server. `smaqit mcp verify` proves its local transport; a fresh agent session must still expose the two MCP tools or the specification agent stops with `DESIGN-TOOLCHAIN-UNAVAILABLE`.

## Step 2: Run the Development Phase

In GitHub Copilot chat or Claude Code, invoke the Development agent:

```
/smaqit.development
```

In Codex, ask it to spawn the `smaqit.development` project agent.

The agent will ask you for your requirements. Describe what you want to build in the chat:

```
I want a simple console application that displays "Hello, Mario!" when run.

Actors:
- Mario Fan: A user who loves Nintendo's Mario franchise and wants a fun greeting

The app should:
- Display "Hello, Mario!" to standard output
- Exit successfully after displaying the message
- Use Python 3.8+ with no external dependencies
```

The agent will:
1. Generate Business, Functional, and Stack specifications in `specs/` with linked, visually reviewed design pairs in `docs/designs/`
2. Build the application code
3. Run tests to verify it works

If your requirements are unclear or incomplete, the agent will ask clarifying questions before proceeding.

## Expected Output

After the Development phase completes, you'll have:

```
mario-hello/
├── .smaqit/
├── .github/
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

- **Add features**: Invoke the `smaqit.development` agent again with updated requirements in chat
- **Deploy**: Invoke the `smaqit.deployment` agent with your infrastructure requirements
- **Validate**: Invoke the `smaqit.validation` agent with your test requirements
- **Check status**: Run `smaqit status` to see spec coverage

## Troubleshooting

**Agent asks for clarification?** Your requirements may be ambiguous. Add more detail in the chat response.

**Specs don't match expectations?** Invoke the relevant specification agent directly (for example, `smaqit.business`) with updated requirements.

**Build fails?** Review the Development agent's output. It will indicate what went wrong and suggest fixes.
