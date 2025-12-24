# Task 030: Move Commands to Prompts

**Date:** 2025-12-24  
**Session Focus:** Convert session and task management commands into development helper prompts

## Actions Taken

### Planning Phase

Started with `/session.recap` to load full project context, then assessed task 030.

**Initial critical assessment:** Challenged the task premise, believing these would be smaqit framework prompts subject to the established architecture (prompts collect parameters, agents execute workflows). Identified 6 red flags including architectural mismatch, semantic confusion, and maintenance burden.

**User clarification:** These are development helper prompts for working on smaqit itself, not smaqit framework prompts. Much simpler scope - just move command logic from copilot-instructions.md to `.github/prompts/` for discoverability and version control.

### Implementation

**Created `.github/prompts/` directory** for development helper prompts (distinct from `prompts/` which contains smaqit framework prompts for distribution to user projects).

**Created 5 prompt files** with simple YAML frontmatter (name, description only - no agent or tools fields):

1. `session.recap.prompt.md` - Full context loading workflow (reads 8 framework files, 3 history files, PLANNING.md)
2. `session.wrap.prompt.md` - Session documentation workflow (review conversation, create history file)
3. `task.create.prompt.md` - Task creation with auto-numbering and flexible input formats
4. `task.list.prompt.md` - Show active tasks from PLANNING.md
5. `task.complete.prompt.md` - Task completion with verification of acceptance criteria

**Updated copilot-instructions.md:** Replaced "Session Commands" and "Task Commands" sections (~80 lines of detailed workflow logic) with brief "Workflow Commands" section referencing the new prompts. Copilot instructions now ~75 lines shorter and focused on development guidelines.

**Updated task files:** Marked task 030 as completed in both individual task file and PLANNING.md.

## Problems Solved

### Problem 1: Cognitive Dissonance from Task Premise

**Initial state:** Read task 030 through the lens of "smaqit framework prompts" which follow strict architecture rules established in task 029.

**Confusion:** Task description said "move commands into prompts" but recent work established that prompts collect parameters and agents execute workflows. Session/task commands ARE workflows, not parameter collection.

**Resolution:** User clarified these are **development helper prompts** for working on smaqit project, not smaqit framework prompts for user projects. Different category, different rules. No architectural conflict.

### Problem 2: Over-Analysis of Simple Task

**Initial approach:** Wrote comprehensive assessment with red flags, semantic concerns, architectural violations - appropriate for framework changes but overkill for simple utility extraction.

**Learning:** Context matters. When user says "it's much simpler," trust that assessment. Development utilities don't need framework-level architectural scrutiny.

## Decisions Made

1. **Prompt structure:** Simple YAML frontmatter (name, description only) - no agent or tools fields needed
2. **Location:** `.github/prompts/` (separate from framework `prompts/` directory)
3. **Installer scope:** Do NOT include in installer - these are smaqit-development-specific utilities
4. **Documentation:** No README updates - prompts discoverable via prompt panel
5. **Invocation:** Slash commands (`/session.recap`, `/task.create`, etc.) instead of bare keywords

## Files Modified

**Created (6 files):**
- `.github/prompts/` directory
- `.github/prompts/session.recap.prompt.md`
- `.github/prompts/session.wrap.prompt.md`
- `.github/prompts/task.create.prompt.md`
- `.github/prompts/task.list.prompt.md`
- `.github/prompts/task.complete.prompt.md`

**Modified (3 files):**
- `.github/copilot-instructions.md` - Reduced by ~75 lines
- `docs/tasks/030_move_commands_to_prompts.md` - Marked completed
- `docs/tasks/PLANNING.md` - Moved task 030 to Completed table

## Key Learnings

### 1. Context Switching Between Framework and Development

smaqit project has two distinct concerns:
- **Framework development:** Building the spec-driven orchestration kit (follows strict architecture)
- **Project development:** Working on smaqit itself (uses practical utilities)

Development helper prompts are project utilities, not framework components. They don't need to follow framework architecture rules.

### 2. Trust User Clarifications

When initial assessment identifies complexity and user says "it's much simpler," believe them. Re-anchor understanding rather than defending analysis.

### 3. Critical Assessment Has Context

Critical assessment is valuable for framework changes with wide impact. For utility extraction tasks, simpler "does this work?" assessment is appropriate.

### 4. Separation of Concerns via Directory Structure

Using `.github/prompts/` (not `prompts/`) creates clear separation:
- `prompts/` → Framework prompts (distributed to user projects)
- `.github/prompts/` → Development prompts (smaqit project only)

This physical separation reinforces the conceptual distinction.

## Next Steps

**Immediate:**
- Test new prompt invocations (`/session.recap`, `/task.create`, etc.) in next chat
- Verify copilot-instructions.md is cleaner and more focused

**Active tasks remaining:**
- Task 014: Define iterative development using smaqit
- Task 015: Investigate framework bundling at installation
- Task 022: Create GitHub Action for automated releases
- Task 025: Integrate testing agent with CI/CD

## Session Metrics

**Duration:** ~1 hour  
**Task completed:** 1 (030)  
**Files created:** 6 (1 directory + 5 prompts)  
**Files modified:** 3  
**Lines reduced from copilot-instructions:** ~75  
**Pattern established:** Development helper prompts in `.github/prompts/`
