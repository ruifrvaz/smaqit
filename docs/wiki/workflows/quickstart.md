# Quickstart: Mario Hello

Build a "Hello, Mario!" console app from requirements to working code using smaqit's spec-driven workflow.

## Prerequisites

- smaqit installed (`curl -fsSL https://raw.githubusercontent.com/ruifrvaz/smaqit/main/install.sh | bash`)
- VS Code with GitHub Copilot extension
- A new or existing project directory

## Step 1: Initialize smaqit

```bash
mkdir mario-hello && cd mario-hello
smaqit init
```

This creates:
- `.smaqit/` — Framework files and templates
- `.github/agents/` — Agent definitions
- `.github/prompts/` — Prompt files for your requirements
- `specs/` — Where generated specifications go

## Step 2: Fill the Business Prompt

Open `.github/prompts/smaqit.business.prompt.md` and add your requirements:

```markdown
## Actors

Mario Fan - A user who loves Nintendo's Mario franchise and wants a fun greeting

## Use Cases

### Greet Mario Fan
The Mario Fan wants to see a personalized greeting featuring Mario.

**Success Criteria:**
- Display "Hello, Mario!" message
- Message appears in console output
- Program exits successfully after displaying message
```

## Step 3: Fill the Functional Prompt

Open `.github/prompts/smaqit.functional.prompt.md` and describe the experience:

```markdown
## Behaviors

### Console Output
When the application runs, it displays "Hello, Mario!" to standard output and exits with code 0.

## Data Models

None required for this simple application.
```

## Step 4: Fill the Stack Prompt

Open `.github/prompts/smaqit.stack.prompt.md` with your technology choice:

```markdown
## Language

Python 3.8+

## Rationale

Simple console application, Python is lightweight and universally available.

## Dependencies

None required (standard library only).
```

## Step 5: Run Development Phase

In VS Code's GitHub Copilot chat, invoke the Development agent:

```
/smaqit.development
```

The agent will:
1. Read your prompt files
2. Generate Business, Functional, and Stack specifications in `specs/`
3. Build the application code
4. Run tests to verify it works

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

- **Add features**: Edit prompt files and re-run `/smaqit.development`
- **Deploy**: Fill Infrastructure prompt and run `/smaqit.deployment`
- **Validate**: Fill Coverage prompt and run `/smaqit.validation`
- **Check status**: Run `smaqit status` to see spec coverage

## Troubleshooting

**Agent asks for clarification?** Your prompts may be ambiguous. Add more detail to the relevant prompt file.

**Specs don't match expectations?** Check that your prompt files have sufficient requirements. Agents rely on explicit input to maintain focused scope.

**Build fails?** Review the Development agent's output. It will indicate what went wrong and suggest fixes.
