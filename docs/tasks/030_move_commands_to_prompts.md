# Move session and task commands into prompts

**Status:** Completed  
**Created:** 2025-12-24  
**Completed:** 2025-12-24

## Description

Session commands (session.wrap, session.recap) and task commands (task.create, task.list, task.complete) are currently defined in `.github/copilot-instructions.md`. Moving them into GitHub Copilot prompts would:

1. Free up copilot-instructions.md space for project-specific development guidelines
2. Make commands more discoverable (users can see them in prompt panel)
3. Enable version control of command logic separately from instructions
4. Allow command reuse across projects (prompts are portable)

## Current State

**Session commands in copilot-instructions.md:**
- `session.recap` - Loads full project context for new chat
- `session.wrap` - Documents session history at end of work

**Task commands in copilot-instructions.md:**
- `task.create [title]` - Create new task file
- `task.list` - Show current active tasks
- `task.complete [id]` - Mark task as completed

## Proposed Structure

**Session prompts:**
- `.github/prompts/session.recap.prompt.md` - Context loading for new sessions
- `.github/prompts/session.wrap.prompt.md` - Session documentation at completion

**Task prompts (optional):**
- `.github/prompts/task.create.prompt.md` - Task creation workflow
- `.github/prompts/task.list.prompt.md` - Task listing logic
- `.github/prompts/task.complete.prompt.md` - Task completion workflow

**Invocation:**
Users would invoke via: `/session.recap`, `/session.wrap`, `/task.create`, etc.

## Scope

### Must Have

- [x] Create `session.recap.prompt.md` with full context loading logic
- [x] Create `session.wrap.prompt.md` with session documentation logic
- [x] Remove session commands from copilot-instructions.md
- [x] Verify prompts work via / invocation

### Should Have

- [x] Create task management prompts (create, list, complete)
- [x] Remove task commands from copilot-instructions.md
- [ ] Document new prompt-based commands in README.md (not needed per user)

### Could Have

- [ ] Create template for meta-workflow prompts (session, task management)
- [ ] Add these prompts to installer (copy to user projects)

## Acceptance Criteria

- [x] Session commands work via prompt invocation (/session.recap, /session.wrap)
- [x] Session command logic removed from copilot-instructions.md
- [x] Task commands work via prompt invocation
- [x] Task command logic removed from copilot-instructions.md
- [x] Copilot instructions file is significantly smaller
- [x] README.md documents new prompt-based command usage (not needed per user)

## Trade-offs

**Advantages:**
- Copilot instructions focused on project development rules
- Commands become discoverable (visible in prompt panel)
- Commands become portable (can copy to other projects)
- Command logic version controlled separately

**Disadvantages:**
- Slightly more verbose invocation (`/session.wrap` vs `session.wrap`)
- Commands are now "just prompts" not "special commands"
- May need education for users expecting keyword commands

## Related

- Copilot instructions: `.github/copilot-instructions.md`
- Prompt architecture: `framework/PROMPTS.md`
- Session command usage: History files in `docs/history/`

## Notes

This is a meta-workflow improvement. Session and task commands are not part of the smaqit framework itself—they're project management utilities. Moving them to prompts clarifies the separation:

- **Copilot instructions**: How to work on smaqit development
- **Framework files**: How to execute smaqit workflows
- **Prompts**: What to invoke for both smaqit workflows and meta-workflows

Consider: Should these meta-workflow prompts be included in the installer, or are they smaqit-project-specific only?
