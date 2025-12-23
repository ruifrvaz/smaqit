# Task: Implement Installer CLI

**ID**: 023
**Status**: Completed

## Context

Implement the smaqit CLI executable for scaffolding, status checking, validation, and project management. The CLI handles quality-of-life operations while phase/layer orchestration happens via Copilot prompt files. Remove phase command stubs (`develop`, `deploy`, `validate`) as these are now handled by prompt files.

## Acceptance Criteria

### Core Commands

- [x] `smaqit init` — Scaffold smaqit project structure
- [x] `smaqit status` — Show project state and spec coverage
- [x] `smaqit validate` — Verify project structure integrity
- [x] `smaqit help` — Display available CLI and prompt commands
- [x] `smaqit uninstall` — Remove smaqit from project
- [x] `smaqit version` — Show smaqit version (already implemented)

### Init Command

**Function:** Scaffold `.smaqit/` and `.github/` directories, copy framework files, templates, and agents.

- [x] Create directory structure:
  ```
  .smaqit/
  ├── framework/           # Copy from kit: framework/*.md (7 files)
  └── templates/
      ├── specs/           # Copy from kit: templates/specs/*.template.md (5 files)
      └── prompts/         # Copy from kit: templates/prompts/*.template.md (2 files)
  
  specs/                   # At project root (moved in session 009)
  ├── business/            # Empty directory
  ├── functional/          # Empty directory
  ├── stack/               # Empty directory
  ├── infrastructure/      # Empty directory
  └── coverage/            # Empty directory
  
  .github/
  ├── agents/              # Copy from kit: agents/*.agent.md (8 files)
  └── prompts/             # Copy from kit: prompts/*.prompt.md (8 files)
  ```
- [x] Handle existing directories: error if `.smaqit/` already exists, prompt user to uninstall first
- [x] Embed smaqit version in `.smaqit/VERSION` file for compatibility checking
- [x] Print success message with next steps (run `/smaqit.develop` in Copilot chat)
- [x] Accept optional directory parameter: `smaqit init [dir]` (added in session 009)

### Status Command

**Function:** Display current project state, spec coverage, and phase completion.

- [x] Scan `specs/` for existing specification files (updated path in session 009)
- [x] Report per-layer spec count:
  ```
  Business:        3 specs
  Functional:      5 specs
  Stack:           1 spec
  Infrastructure:  0 specs
  Coverage:        0 specs
  ```
- [x] Calculate phase status based on spec presence
- [x] Display next steps with `/smaqit.*` prefix

### Validate Command

**Function:** Verify project structure integrity and spec template compliance.

- [x] Check directory structure exists and is complete (all expected folders present)
- [x] Verify framework files are present (6 files)
- [x] Validate spec files follow template structure:
  - No placeholder text remaining (e.g., `[PLACEHOLDER]`)
  - Requirement IDs use correct layer prefix
  - Acceptance Criteria section present
- [x] Report validation errors with file paths
- [x] Exit code 0 on success, non-zero on validation failure

### Help Command

**Function:** Display available commands with separate sections for CLI and Copilot prompts.

- [x] List CLI commands with descriptions
- [x] List Copilot prompts with `/smaqit.*` prefix:
  ```
  /smaqit.develop          Run develop phase
  /smaqit.deploy           Run deploy phase
  /smaqit.validate         Run validate phase
  /smaqit.business         Create business layer specifications
  /smaqit.functional       Create functional layer specifications
  /smaqit.stack            Create stack layer specifications
  /smaqit.infrastructure   Create infrastructure layer specifications
  /smaqit.coverage         Create coverage layer specifications
  ```
- [x] Include usage example and documentation link

### Uninstall Command

**Function:** Remove all smaqit files and directories from the project.

- [x] Prompt user for confirmation with clear list of what will be removed
- [x] Remove `.smaqit/` directory recursively
- [x] Remove `specs/` directory (added in session 009)
- [x] Remove `.github/agents/` directory
- [x] Remove `.github/prompts/` directory
- [x] Optionally remove `.github/` if empty after cleanup
- [x] Print confirmation message with status per directory

### Code Cleanup

- [x] Remove `cmdDevelop()` stub function
- [x] Remove `cmdDeploy()` stub function
- [x] `cmdValidate()` is now the full implementation (not a stub)
- [x] Update all next steps to use `/smaqit.*` prefix

## Implementation Notes

**Actual structure differs from initial spec:**
- Agent templates NOT installed (Level 0 only, per session 009)
- `specs/` at project root instead of `.smaqit/specs/` (per session 009)
- All prompts use `/smaqit.*` prefix (per session 009)
- Init accepts optional directory parameter for clean project creation

**Files embedded at compile time:**
- Uses `go:embed` directives for portability
- `make prepare` copies files before build
- VERSION file created with build version

**Validation is lenient:**
- Focuses on critical violations
- Warnings for malformed IDs, not errors
- Placeholder detection prevents incomplete specs
