# Session 002: Meta Workflow and Housekeeping

**Date**: 2025-11-30
**Previous Session**: [001_initial_scaffolding_2025-11-29.md](001_initial_scaffolding_2025-11-29.md)

## Objective

Add session-based development workflow (history/tasks) and housekeeping improvements to the kit.

## Work Done

1. **Session Workflow** — Added meta documentation system:
   - `history/session.template.md` — Template for session logs
   - `tasks/task.template.md` — Template for work items
   - `tasks/planner.md` — Table tracking tasks with statuses (`new | in progress | completed`)

2. **Copilot Instructions** — Updated with:
   - Clear Source vs Artifacts distinction (kit source files vs installer-generated artifacts)
   - Session workflow commands: "recap", "wrap up", "new task"
   - Task management rules

3. **Renamed FRAMEWORK.md → SMAQIT.md** — Updated all references across:
   - `.github/copilot-instructions.md`
   - `installer/main.go`
   - `framework/SMAQIT.md`
   - `history/001_initial_scaffolding_2025-11-29.md`
   - `tasks/001_create_smaq_commands_file.md`

4. **Git Setup**:
   - Added remote origin: `https://github.com/ruifrvaz/smaqit.git`
   - Created `.gitignore` to exclude task files but track templates and planner

5. **First Task Created**:
   - `tasks/001_create_smaq_commands_file.md` — Create smaq commands for agent use (e.g., `smaq.load`)

## Decisions Made

- **History tracked, tasks not**: Session logs are public knowledge; task files are personal workflow
- **Source vs Artifacts**: Clear separation in copilot-instructions between kit development files and what installer generates
- **SMAQIT.md naming**: Framework file renamed to match the kit name

## Open Questions

- Should the smaq commands file live in `framework/` or as a separate top-level file?
- What other smaq commands beyond `smaq.load` would be useful?

## Next Session

- Work on task 001: Create smaq commands file with `smaq.load` command
- Consider implementing the installer `init` command
