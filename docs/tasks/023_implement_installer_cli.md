# Task: Implement Installer CLI

**ID**: 023
**Status**: new

## Context

Implement the smaqit CLI executable for scaffolding, status checking, validation, and project management. The CLI handles quality-of-life operations while phase/layer orchestration happens via Copilot prompt files. Remove phase command stubs (`develop`, `deploy`, `validate`) as these are now handled by prompt files.

## Acceptance Criteria

### Core Commands

- [ ] `smaqit init` — Scaffold smaqit project structure
- [ ] `smaqit status` — Show project state and spec coverage
- [ ] `smaqit validate` — Verify project structure integrity
- [ ] `smaqit help` — Display available CLI and prompt commands
- [ ] `smaqit uninstall` — Remove smaqit from project
- [ ] `smaqit version` — Show smaqit version (already implemented)

### Init Command

**Function:** Scaffold `.smaqit/` and `.github/` directories, copy framework files, templates, agents, and prompts.

- [ ] Create directory structure:
  ```
  .smaqit/
  ├── framework/           # Copy from kit: framework/*.md (6 files)
  ├── templates/
  │   ├── specs/           # Copy from kit: templates/specs/*.template.md (5 files)
  │   └── agents/          # Copy from kit: templates/agents/*.template.md (2 files)
  └── specs/
      ├── business/        # Empty directory
      ├── functional/      # Empty directory
      ├── stack/           # Empty directory
      ├── infrastructure/  # Empty directory
      └── coverage/        # Empty directory
  
  .github/
  ├── agents/              # Copy from kit: agents/*.agent.md (8 files)
  └── prompts/             # Copy from kit: prompts/*.prompt.md (8 files)
  ```
- [ ] Handle existing directories: error if `.smaqit/` already exists, prompt user to uninstall first
- [ ] Embed smaqit version in `.smaqit/VERSION` file for compatibility checking
- [ ] Print success message with next steps (run `/smaqit.develop` in Copilot chat)

### Status Command

**Function:** Display current project state, spec coverage, and phase completion.

- [ ] Scan `.smaqit/specs/` for existing specification files
- [ ] Report per-layer spec count:
  ```
  Business:        3 specs
  Functional:      5 specs
  Stack:           1 spec
  Infrastructure:  0 specs
  Coverage:        0 specs
  ```
- [ ] Calculate overall spec coverage percentage (layers with specs / total layers)
- [ ] Indicate which phases are complete based on spec presence and completion markers

### Validate Command

**Function:** Verify project structure integrity and spec template compliance.

- [ ] Check directory structure exists and is complete (all expected folders present)
- [ ] Verify framework files are present and unmodified (checksum comparison)
- [ ] Validate spec files follow template structure:
  - Required sections present
  - Requirement IDs follow `[LAYER_PREFIX]-[CONCEPT]-[NNN]` format
  - No placeholder text remaining (e.g., `[PLACEHOLDER]`)
- [ ] Report validation errors with file paths and line numbers
- [ ] Exit code 0 on success, non-zero on validation failure

### Help Command

**Function:** Display available commands with separate sections for CLI and Copilot prompts.

- [ ] List CLI commands:
  ```
  CLI Commands:
    smaqit init       Scaffold smaqit project structure
    smaqit status     Show project state and spec coverage
    smaqit validate   Verify project structure integrity
    smaqit help       Show this help message
    smaqit uninstall  Remove smaqit from project
    smaqit version    Show smaqit version
  ```
- [ ] List Copilot prompts:
  ```
  Copilot Prompts (use in GitHub Copilot chat with /):
    /develop          Run develop phase (business → functional → stack → build)
    /deploy           Run deploy phase (infrastructure → deploy)
    /validate         Run validate phase (coverage → validate)
    /business         Create business layer specifications
    /functional       Create functional layer specifications
    /stack            Create stack layer specifications
    /infrastructure   Create infrastructure layer specifications
    /coverage         Create coverage layer specifications
  ```
- [ ] Include usage example: `Type '/smaqit.develop' in GitHub Copilot chat to start developing`

### Uninstall Command

**Function:** Remove all smaqit files and directories from the project.

- [ ] Prompt user for confirmation: "This will remove .smaqit/ and .github/agents/, .github/prompts/. Continue? [y/N]"
- [ ] Remove `.smaqit/` directory recursively
- [ ] Remove `.github/agents/` directory
- [ ] Remove `.github/prompts/` directory
- [ ] Optionally remove `.github/` if empty after cleanup
- [ ] Print confirmation message

### Code Cleanup

- [ ] Remove `cmdDevelop()` stub function and case from main.go
- [ ] Remove `cmdDeploy()` stub function and case from main.go
- [ ] Remove `cmdValidate()` stub function and case from main.go
- [ ] Update `printUsage()` to reflect new command list (exclude develop/deploy/validate)

## Notes

**Installer location awareness:**
The installer needs to know where the smaqit kit files are located. Options:
1. Embed files at compile time using `go:embed` directives
2. Require installer to be run from the kit directory
3. Download files from GitHub releases

Recommend option 1 (embed) for portability.

**Version compatibility:**
Init should embed the installer version in `.smaqit/VERSION` so future validate commands can detect version mismatches.

**Validation scope:**
Template compliance validation should be lenient initially—focus on critical violations (missing sections, malformed IDs) rather than stylistic issues.
