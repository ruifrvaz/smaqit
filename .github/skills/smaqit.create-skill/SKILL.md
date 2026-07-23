---
name: smaqit.create-skill
description: Use when the user wants to create, define, build, or package a new skill — including when they ask to turn a workflow into a reusable command, wrap domain knowledge into a slash-command, or describe a repeatable procedure they want Copilot to follow. Gathers name and purpose, infers a complete specification, writes a definition file, and invokes smaqit.L2 to compile a SKILL.md file.
metadata:
  version: "2.0.0"
---

# Create Skill

## Steps

### 1. Gather

Ask the user for the skill **name** in a single message (lowercase, hyphens allowed, e.g., `my-review`). The description will be inferred from the name and scanned context.

### 2. Scan

Before writing anything, read:
- All existing files in `.github/skills/` — for patterns and conventions already used in this project
- Project README — for domain and conventions
- Any project manifests that describe workflows or user-facing operations

Also extract any relevant detail the user has already provided in the conversation.

### 3. Infer and write definition file

Using the name and scanned context, infer a complete skill specification. Do not ask further questions.

Write the inferred specification to `.smaqit/definitions/skills/[name].md`. Create the directory if it does not exist.

For any field where the correct value is genuinely ambiguous, suffix the value with `[?]` and add a brief inline note.

The definition file must cover:
- Name and description
- Steps (what the skill does, in sequence)
- Output (what the skill produces)
- Scope (what is out of scope)
- Completion criteria (testable, checkbox-style)
- Failure handling (likely failure modes and responses)
- Gotchas — environment-specific facts the agent must know before executing; non-obvious corrections to mistakes it would make without being told. Distinct from failure handling (which is reactive). Include any project conventions, unexpected API behaviors, naming quirks, or platform constraints that apply to this skill.
- Examples — at least one concrete example of what triggers the skill and what it produces. Input: a representative user request. Output: the artifact or response produced.
- Allowed tools (optional) — if the skill requires specific tools to run (e.g., git, bash scripts), list them as `allowed-tools` values using the format `Bash(git:*)`, `Read`, etc.
- Compatibility (optional) — if the skill has environment requirements (specific agent product, system packages, network access), note them here.

### 4. Compile

Invoke `smaqit.L2` as a subagent with:
> "Compile the skill definition at `.smaqit/definitions/skills/[name].md`. Write the compiled skill to `.github/skills/[name]/SKILL.md`. After compilation, list any fields annotated with `[?]` and suggest a resolution for each. If the compiled skill body would exceed 400 lines, move detailed reference content to a `references/` subdirectory and link from SKILL.md with explicit load conditions ("Read references/[file].md if [condition]"). The main SKILL.md body must remain under 400 lines after extraction."

### 5. Report

After L2 completes, report to the user:
- Path of the compiled skill file
- Any `[?]`-annotated items and L2's suggested resolutions
- How to adjust: edit `.smaqit/definitions/skills/[name].md` and re-invoke `/smaqit.create-skill`, or switch to `smaqit.L2` directly

## Output

- `.smaqit/definitions/skills/[name].md` — inferred specification (scaffolding)
- `.github/skills/[name]/SKILL.md` — compiled skill file (source of truth)

## Scope

Does not create agents, framework files, or templates.

## Completion

- [ ] Name obtained from user
- [ ] Repository scanned for context
- [ ] Definition file written to `.smaqit/definitions/skills/[name].md`
- [ ] `smaqit.L2` invoked and compilation completed
- [ ] Compiled skill exists at `.github/skills/[name]/SKILL.md`

## Failure Handling

| Situation | Action |
|-----------|--------|
| Name not provided | Request before proceeding |
| `.smaqit/templates/` not present | Inform the user that ADK templates are required — run `smaqit-adk lite` in this repository first |
| Output artifact already exists | Report the conflict; do not overwrite without user confirmation |
| L2 invocation fails | Report the error and include the path to the definition file so the user can inspect or correct it |
