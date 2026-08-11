# Global Agent Installation Release

| Field | Value |
|---|---|
| Date | 2026-08-11 |
| Session focus | Global user-level SmaQit agent and skill installation |
| Tasks completed | 105 |

## Actions Taken

- Planned and completed the migration of SmaQit agents, skills, and Claude commands from project mirrors to user-level client directories.
- Split the installer’s global payload installation from project scaffolding: `smaqit init` now retains only project-local state, design tooling, MCP configuration, instructions, and the Copilot setup workflow.
- Added the installer-only `--install-global` bootstrap, invoked by the public `curl | bash` installer and by the updater; global roots honour the client-specific home-directory overrides.
- Updated generated skill paths, documentation, the Copilot setup workflow, unit tests, and smoke coverage for the global model.
- Released v3.0.0, diagnosed and corrected the installer smoke-test executable-mode regression, then released v3.0.1 to show the resolved project path during `smaqit init`.
- Performed a real tagged-release installation in a sandboxed home, tested initialization of a fresh repository, global uninstall ownership preservation, environment overrides, and the project validation/MCP checks.
- Completed Task 105, removed its merged worktree, recorded the CI diagnosis, removed the merged v3.0.1 release worktree, regenerated the VS Code workspace, and pushed the bookkeeping to `main`.

## Problems Solved

- Project-local agent and skill mirrors made framework installation per-repository rather than per-user. The global installer now owns only exact SmaQit files and preserves unrelated user content.
- The initial release PR’s smoke script lost its tracked executable mode during conflict resolution. CI detected the failure, the mode was restored, and the rerun passed.
- `smaqit init` printed `.` for a current-directory invocation. It now reports the resolved absolute project path.
- A sandboxed installation was initially mistaken for an installation into the user’s real home. Testing confirmed that setting `HOME` redirects both the shell installer and Go user-directory resolution without requiring a container.

## Decisions Made

- Use global paths for Copilot, Claude Code, and Codex payloads; share Copilot/Codex skills under `~/.agents/skills`.
- Keep project state deliberately local: specifications, designs, the managed PlantUML runtime, MCP registrations, instruction-file integration, and a create-if-absent Copilot setup workflow remain per repository.
- Preserve an existing Copilot setup workflow rather than overwrite user customization.
- Treat GitHub-hosted Copilot setup as an ephemeral global bootstrap: it installs the global payload for the runner before conditionally scaffolding the checked-out project.
- Do not pursue the pre-v3 updater bootstrap edge case in this session; older updaters may require one additional update or installer run after they first reach v3.

## Files Modified

- `.github/workflows/copilot-setup-steps.yml` — global bootstrap and validation on Copilot runners.
- `install.sh`, `installer/main.go`, `installer/update.go` — global installer lifecycle and project-scaffold split.
- `installer/main_test.go`, `scripts/smoke-test-installer.sh` — global-install, ownership, override, and regression coverage.
- `scripts/generate-agents.py` — runtime-resolvable global skill paths.
- `README.md`, `framework/SKILLS.md`, `templates/AGENTS.md.template`, `templates/workflows/copilot-setup-steps.yml`, `CHANGELOG.md` — current installation model and releases.
- `.smaqit/tasks/105_migrate_to_global_user_level_installation.md`, `.smaqit/tasks/PLANNING.md` — completed task record.
- `.smaqit/reports/diagnose-2026-08-11.md`, `smaqit.code-workspace` — CI diagnosis and release-worktree cleanup.

## Next Steps

- Confirm global-agent discovery inside live authenticated Copilot, Claude Code, and Codex sessions as host behavior evolves.
- If updating from a pre-v3 binary, run the installer or a second `smaqit update` to execute the current global-payload bootstrap.
- Consider the remaining active tasks 094, 077, 074, and 071 when selecting future work.

## Session Metrics

- Duration: one session day.
- Tasks completed: 1.
- Releases published: 2 (`v3.0.0`, `v3.0.1`).
- Global client surfaces checked: 3 (Copilot, Claude Code, Codex).
