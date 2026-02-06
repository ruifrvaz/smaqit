# Task 080: Copilot Setup Workflow for smaqit Installation

**Status:** new  
**Priority:** High  
**Created:** 2026-02-03

## Problem

App builders using smaqit need smaqit installed and configured before GitHub Copilot coding agent starts work. Currently, users must manually install smaqit, which is error-prone and adds friction to onboarding.

GitHub Copilot supports [environment customization through setup steps](https://docs.github.com/en/copilot/how-tos/use-copilot-agents/coding-agent/customize-the-agent-environment) that run as pre-run workflows before the coding agent executes.

## Goal

Create a GitHub workflow that:
1. Runs automatically before Copilot coding agent sessions
2. Detects if smaqit is installed (`.smaqit/` directory exists)
3. If not installed, downloads and runs the smaqit installer
4. Validates the installation completed successfully
5. Provides clear feedback to the Copilot agent about smaqit availability

## Scope

**In Scope:**
- Create `.github/workflows/copilot-setup-steps.yml` workflow file
- Add detection logic for existing smaqit installation
- Download and execute smaqit installer if needed
- Validate installation success
- Handle installation errors gracefully
- Document the setup workflow in README.md
- Add workflow to installer for scaffolding
- Add workflow template to templates/workflows

**Out of Scope:**
- Modifying the installer itself (separate concern)
- Adding version update detection (can be added later)
- Custom configuration during installation (uses defaults)

## Acceptance Criteria

- [x] Workflow file created at `.github/workflows/copilot-setup-steps.yml`
- [x] Workflow detects if `.smaqit/` directory exists
- [x] If not installed, downloads smaqit installer binary
- [x] Executes `smaqit init` command
- [x] Validates installation by checking for `.smaqit/framework/SMAQIT.md`
- [x] Workflow fails gracefully with clear error message if installation fails
- [x] Documentation added to README.md about Copilot setup integration
- [x] Workflow added to installer for automatic scaffolding
- [x] Workflow template added to templates/workflows
- [ ] Tested with fresh repository (no smaqit installed)
- [ ] Tested with existing smaqit installation (should skip)

## Design Notes

### Workflow Structure

```yaml
name: Copilot Setup
on:
  workflow_dispatch:
  # Triggered automatically by Copilot coding agent
  
jobs:
  setup-smaqit:
    runs-on: ubuntu-latest
    steps:
      - name: Check for existing smaqit
        # If .smaqit/ exists, skip installation
      
      - name: Download smaqit installer
        # Get latest release from GitHub
      
      - name: Install smaqit
        # Run smaqit init
      
      - name: Validate installation
        # Check for expected files
```

### Key Considerations

1. **Idempotency** — Workflow should be safe to run multiple times
2. **Speed** — Keep setup time minimal (use cached installer if possible)
3. **Visibility** — Provide clear output for debugging
4. **Failure handling** — Don't block Copilot if smaqit can't install, but warn clearly

### Installation Detection

Check for presence of:
- `.smaqit/framework/SMAQIT.md` (primary indicator)
- `.github/agents/smaqit.business.agent.md` (secondary check)

### Version Considerations

Initial implementation installs latest release. Future enhancement could:
- Detect version mismatch and offer update
- Pin to specific version via configuration file

## Implementation Plan

1. Create workflow file with detection logic
2. Add installer download step (fetch from GitHub releases)
3. Add installation execution step
4. Add validation checks
5. Test with fresh repo
6. Test with existing installation
7. Update README.md documentation
8. Consider adding example to templates/ for user projects

## Related Tasks

- Task 023: Implement installer CLI (completed)
- Task 022: Create GitHub Action for automated releases (completed)
- Task 075: Dual Release Architecture (in progress)

## Notes

This workflow should be included in:
1. **smaqit repository** — As example/documentation
2. **Installer output** — Optionally scaffolded into user projects
3. **Templates** — Add to templates/ for user reference

The workflow enables seamless Copilot + smaqit integration, reducing setup friction from manual installation to automatic initialization.
