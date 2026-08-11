# Migrate to Global User-Level Installation (Learned from smaqit-extensions)

**Status:** Completed
**Created:** 2026-08-10
**Started:** 2026-08-11
**Completed:** 2026-08-11
**Mode:** Assisted

## Description

`smaqit` currently installs all agents and skills into the *project* directory on every `smaqit init` — `.github/agents/`, `.github/skills/`, `.claude/agents/`, `.claude/commands/`, `.claude/skills/`, `.codex/agents/`, `.agents/skills/` (installer/main.go:363-370, 442+). This means every repository using smaqit carries its own full copy of the framework's agents and skills, requiring re-init on every project and re-sync after every framework update.

The sibling project `smaqit-extensions` (github.com/ruifrvaz/smaqit-extensions) just completed and shipped this exact migration — Task 023, released as v1.14.0 through v1.14.2. Its CLI now installs agents and skills to global, platform-specific user-level paths (`~/.copilot/agents/`, `~/.claude/agents/`, `~/.claude/skills/`, `~/.codex/agents/`, `~/.agents/skills/` shared by Copilot+Codex) once per machine, and `init` only scaffolds per-project state (`.smaqit/` tracking, `.github/workflows/`). This task ports that architecture to `smaqit`.

**This task exists specifically to carry forward the mistakes made during that migration, not just the destination design** — see Design Decisions and Notes below for what went wrong and why, sourced directly from the smaqit-extensions session history.

## Design Decisions

- **Global paths, mirrored from smaqit-extensions exactly:**
  - `~/.agents/skills/` — shared skill tree for GitHub Copilot + Codex (identical content, one install)
  - `~/.claude/skills/` — Claude Code skills
  - `~/.copilot/agents/` — GitHub Copilot custom agents
  - `~/.claude/agents/` — Claude Code subagents
  - `~/.claude/commands/` — Claude Code slash commands
  - `~/.codex/agents/` — Codex custom agents
  - Respect `COPILOT_HOME`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME` overrides for their respective agent roots (no override needed for the shared skills path).
- **`init` becomes scaffold-only.** After migration, `smaqit init` in a project must create *only* `.smaqit/` state (planning, task tracking, whatever else is genuinely project-local) and any project-local CI/workflow files. It must **not** create `.github/agents/`, `.github/skills/`, `.claude/*`, `.codex/agents/`, or `.agents/skills/` in the project — those live globally now.
- **No user-facing `install` subcommand.** smaqit-extensions initially shipped a new `install` subcommand (with `--scope`/`--agent` flags) as the mechanism for global installation, then had to walk it back one release later because it created three competing meanings of "install" (the shell installer script, the CLI subcommand, and the legacy `init`). Go straight to the corrected design: `install.sh` (or smaqit's equivalent installer script) downloads the binary and immediately triggers global installation via a hidden/internal flag (e.g. `--install-global`), not a subcommand a user is expected to type. `smaqit init` remains the only project-facing command, and it only scaffolds.
- **Command-line default must show help, not silently act.** A related regression: an early version of the corrected smaqit-extensions CLI made bare `smaqit-extensions` (no args) scaffold the project silently. This is surprising behavior for a bare invocation — restore/confirm that `smaqit` with no arguments prints help, and require the explicit `init` verb to scaffold.
- **Global-only installation applies everywhere.** The installer script installs global payloads through hidden `--install-global`; there is no public `install` command or CI-only project-scoped fallback. Copilot setup runs that same hidden bootstrap on every ephemeral runner before it conditionally scaffolds the checked-out project.
- **Project scaffolding stays genuinely project-local.** `init` continues to create and validate specs, `docs/designs`, `.smaqit` templates/reports/runtime, managed `.gitignore` and MCP configuration, project instruction-file integration, and workflow scaffolding. It must neither create nor require platform agent/skill mirror directories.
- **Migrate owned legacy files safely.** On an existing project, `init` may remove only exact smaqit-owned agent/skill artifacts left by the legacy layout and prune empty directories; unrelated user files remain untouched.
- **Global skill self-references must be runtime-resolvable.** Generated Copilot/Codex shared skills and Claude skills must refer to their own global installation roots without embedding the build machine's home directory.
- **Existing workflow customization wins.** `init` creates the managed Copilot setup template when it is absent and preserves any existing workflow unchanged.

## Implementation Steps

1. Add testable global-path resolution and hidden global installation to `installer/main.go`. Install Copilot agents, Claude agents/commands/skills, Codex agents, and one shared Copilot/Codex skill tree to the approved global paths; preserve unrelated user files through exact-owned-file replacement and removal helpers.
2. Update `scripts/generate-agents.py` and installer staging so generated skill self-references resolve from global paths at runtime. Produce a shared Copilot/Codex skill rendering and a Claude rendering that honours `CLAUDE_CONFIG_DIR`; do not bake a build-machine home path.
3. Refactor `cmdInit`, conflict detection, `cmdValidate`, and project cleanup around project-local assets only. Retain specifications, designs, templates, PlantUML runtime/MCP registration, instructions, and workflows; remove legacy owned project mirrors safely without deleting user content.
4. Wire `install.sh` to invoke the installed binary's hidden `--install-global` flag after binary verification using a stable top-level binary path. Refactor `update` to refresh global assets using the fresh binary and re-scaffold a project only when project state exists. Align uninstall behaviour with the split project/global ownership model.
5. Update the Copilot setup source workflow and embedded template to execute hidden global installation on every runner, conditionally scaffold project state, and validate/report global payload paths. Preserve any existing workflow rather than overwriting it. Add a live cloud-agent discovery acceptance check without adding a project-scope fallback.
6. Update user and framework documentation for global installation and scaffold-only `init`, including environment overrides and any generated-path references.
7. Replace project-local smoke assertions with sandboxed-home global-install, override, legacy-migration, and scaffold-only checks. Run Go tests, installer smoke tests, and a real tagged-release `curl | bash` installation into a sandboxed `$HOME`; inspect every global directory directly before release preparation.

## Known Issues Triage

**Triaged:** 2026-08-11
**Tools searched:** GitHub Copilot CLI, Claude Code, OpenAI Codex
**Result:** Advisory

### Blocking Issues
- None.

### Advisory Issues
- [#1756 Allow external custom agents (installed from plugins) to access globally configured MCP servers](https://github.com/github/copilot-cli/issues/1756) — `github/copilot-cli` — opened 2026-03-02 — area:agents, area:plugins, area:mcp, area:installation, area:tools. The request is not a labeled regression and concerns plugin-installed agents rather than smaqit's direct global installation, but reinforces the need to verify global-agent MCP discovery.
- [#14202 Project-scoped plugins incorrectly detected as installed globally](https://github.com/anthropics/claude-code/issues/14202) — `anthropics/claude-code` — opened 2025-12-16 — bug, has repro, area:core. It concerns plugins rather than standalone agents/skills, but is relevant to project-versus-global discovery boundaries.

### Historical (Closed)
- [#31392 Global agents in ~/.claude/agents/ not recognized, only project-local agents work](https://github.com/anthropics/claude-code/issues/31392) — `anthropics/claude-code` — closed — historical, stale report. Cover current global-agent discovery in integration verification rather than assuming the historical behavior persists.

### Unresolvable Tools
- OpenAI Codex — GitHub issue API search returned HTTP 403 during triage.

## Acceptance Criteria

- [x] Agents and skills install to the global paths listed in Design Decisions, not into any project directory, after running the installer script
- [x] `smaqit init` in a project creates only `.smaqit` state and project-local CI wiring — verified by inspecting a fresh project tree and confirming no platform agent/skill mirror directories appear
- [x] No user-facing `install` subcommand exists; global installation is triggered automatically by the installer script, not a command the user is told to type
- [x] `smaqit` with no arguments prints help; `smaqit init` is the explicit, only way to scaffold a project
- [x] `COPILOT_HOME`, `CLAUDE_CONFIG_DIR`, `CODEX_HOME` overrides work for their respective agent install roots
- [x] `.github/workflows/copilot-setup-steps.yml` installs the global payload on every runner, conditionally scaffolds only project-local state, and validates/reports global agent and skill paths; no CI-only project-scoped installation mode exists
- [x] The Copilot setup workflow is created when absent and any existing workflow remains untouched
- [x] Generated Copilot/Codex and Claude skill artifacts resolve their own global roots at runtime without embedding a build-machine home directory
- [x] Existing projects lose only exact smaqit-owned legacy project-local agent/skill files; unrelated user files and non-empty directories survive migration
- [x] A real `curl | bash` install against a sandboxed `$HOME` succeeds end-to-end and populates every expected global directory
- [x] `make -C installer test` and `make -C installer smoke-test` pass with sandboxed-home assertions for global paths and environment overrides
- [x] `CHANGELOG.md` and `README.md` describe the new installation model

## Findings

**Implementation approach:**
- Split global payload installation from project scaffolding through a hidden installer-only flag and global path resolver.
- Kept specifications, designs, runtime tooling, MCP registration, instructions, and workflow scaffolding project-local.

**Decisions made:**
- Shared Copilot/Codex skills install under `~/.agents/skills`; platform configuration overrides apply only to their respective agent roots.
- Copilot setup workflow is create-if-absent so existing project workflow customization is preserved.

**Blockers encountered:**
- Release-branch push initially lacked the GitHub token workflow scope; refreshing credentials resolved it.
- Smoke-test conflict resolution temporarily lost the executable bit; CI exposed and verified the mode-only fix.

**Follow-up identified:**
- Pre-v3 self-updaters require one additional `smaqit update` or installer run to trigger the new global payload bootstrap.
- Verify global-agent discovery in live Copilot, Claude, and Codex sessions as host behavior evolves.

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/main.go` | Modify — global path resolver, split project/global install, reduced `init` |
| `install.sh` | Modify — trigger global install automatically after binary download |
| `README.md` | Modify — document global install model |
| `CHANGELOG.md` | Modify — record the change |
| `.github/workflows/copilot-setup-steps.yml` | Modify or confirm compatible — CI/ephemeral-runner design question |
| Any skill/agent files referencing project-local `.github/agents/`, `.github/skills/`, `.claude/*`, `.codex/agents/`, `.agents/skills/` | Modify — update to reflect global paths |

## Notes

**Source of this task:** ported directly from `smaqit-extensions` Task 023 ("Global User-Level Installation with Agent-Specific Adapters"), which shipped as v1.14.0, then required two immediate patch releases:

- **v1.14.1** — fixed `init` re-delegating to the full per-project install path (an earlier version of the migration aliased the deprecated `init` command straight to the new project-scoped install path, which reproduced the exact "install everything into the project" behavior the migration existed to eliminate — caught only after a real user ran the real installer, not during automated testing).
- **v1.14.2** — fixed a broken `install.sh` (`$target` variable referenced out of function scope), which meant the *first* released version of the corrected installer downloaded the binary successfully but silently never installed anything globally. Caught by manually running `curl | bash` against a sandboxed `$HOME` and inspecting the resulting directory tree — the automated smoke-test suite did not catch this because it invokes the binary directly, not through the shell installer script.

**The core lesson driving this task's emphasis on real end-to-end verification:** automated test suites (`make smoke-test`, unit tests) verify that *code paths* work when invoked directly. They do not verify that the actual user-facing entry point (the `curl | bash` command a real user runs) wires those code paths together correctly. Both smaqit-extensions patch releases were needed specifically because the shell-script integration layer was never exercised end-to-end before release. Do not repeat this — budget time in this task specifically for a real, unscripted install run before considering it done.

`smaqit.task-plan` (or this project's planning skill, if named differently) should be invoked on this task before implementation — several fields above (Implementation Steps, the CI/ephemeral-runner question) are intentionally left as discovery items rather than pre-resolved.
