# Register PlantUML MCP Across Clients

**Status:** Completed
**Created:** 2026-08-06
**Mode:** Assisted
**Started:** 2026-08-06
**Completed:** 2026-08-06

## Description

Register the mandatory `smaqit-plantuml` MCP server in the actual project configuration read by every supported authoring client. The installer currently manages only VS Code's `.vscode/mcp.json`; Claude Code needs a root `.mcp.json`, and Codex needs a trusted `.codex/config.toml` MCP table. Generated agent declarations are tool allowlists, not host registration.

The result must retain smaqit's strict design-authoring model. Initialization and update must fail before a partial installation when an owned server name conflicts, `smaqit mcp verify` must prove all deterministic registration and transport layers, and client trust/restart/tool discovery remains an explicit host-owned precondition.

## Design Decisions

- **Three mandatory project registrations:** Manage `.vscode/mcp.json`, root `.mcp.json`, and `.codex/config.toml` for the same `smaqit-plantuml` stdio command.
- **Merge-safe strict ownership:** Preserve unrelated JSON, JSONC, and TOML content. A same-name registration that differs from smaqit's exact definition is a `DESIGN-TOOLCHAIN-UNAVAILABLE` failure; never overwrite it.
- **Byte-preserving Codex editing:** Use TOML parsing only to inspect/validate configuration. Append or remove only smaqit's exact owned table so unrelated TOML remains untouched.
- **No discovery fallback:** `smaqit mcp verify` validates configuration and wrapper transport, but cannot attest that Claude or Codex exposed tools in a live session. Specification agents stop when host discovery is absent.
- **Lifecycle parity:** Preflight, install, validate, and uninstall apply to all three configurations. Uninstall removes only exact smaqit registrations and retains unrelated configuration.

## Implementation Steps

1. Add shared Claude and Codex MCP definition and lifecycle helpers beside the VS Code helpers in `installer/design_tools.go`; add the minimal TOML parser dependency needed for semantic inspection without reserializing unrelated content.
2. Preflight all three configurations before initialization changes the target, then install all registrations during `cmdInit`; remove the obsolete claim that Codex needs no project configuration.
3. Make `smaqit validate` and `smaqit mcp verify` validate all three registration files before the existing public-wrapper MCP handshake. Update uninstall to remove each owned entry selectively.
4. Expand unit and installer smoke coverage for missing, valid, malformed, conflicting, unrelated, reinitialized, updated, and uninstalled Claude/Codex configuration, while retaining generated-agent declaration assertions.
5. Update user guidance, tool reference, test cases, and release-test documentation with client-specific configuration paths, restart/trust requirements, and the distinction between registration and host tool discovery.

## Known Issues Triage

**Triaged:** 2026-08-06
**Tools searched:** Claude Code, Codex
**Result:** Blocking

### Blocking Issues
- [#13025 Codex Desktop ignores project `.codex/config.toml` MCP server](https://github.com/openai/codex/issues/13025) — `openai/codex` — opened 2026-02-27 — bug, mcp, app, plugins
- [#21789 MCP servers from config.toml do not work anymore](https://github.com/openai/codex/issues/21789) — `openai/codex` — opened 2026-05-08 — bug, mcp, app, regression, config

### Advisory Issues
- [#75567 MCP approval for a project-scoped HTTP server never persists](https://github.com/anthropics/claude-code/issues/75567) — `anthropics/claude-code` — opened 2026-07-08 — bug, has repro, platform:macos, area:mcp
- [#36465 Codex Desktop overwrites user config.toml and removes registered MCP servers](https://github.com/openai/codex/issues/36465) — `openai/codex` — opened 2026-08-01 — bug, windows-os, mcp, app, config

### Historical (Closed)
- [#13056 Per-project MCP server configuration in `.codex/config.toml` or mcp.json](https://github.com/openai/codex/issues/13056) — `openai/codex` — closed 2026-07-25

### Unresolvable Tools
- None

Proceeding by explicit user approval on 2026-08-06. The implementation must create and validate the documented Codex project registration, retain the client-owned discovery stop condition, and keep the upstream limitations visible in user guidance.

## Acceptance Criteria

- [x] Fresh `smaqit init` and project `smaqit update` create valid registrations for `smaqit-plantuml` in `.vscode/mcp.json`, root `.mcp.json`, and `.codex/config.toml` with the intended stdio command and arguments.
- [x] Preflight rejects malformed configuration or a conflicting same-name server in any client file before a partial installation is created or changed.
- [x] Reinitialization preserves unrelated VS Code JSONC, Claude JSON, and Codex TOML content while keeping the owned registration exactly current.
- [x] `smaqit validate` and `smaqit mcp verify` fail with `DESIGN-TOOLCHAIN-UNAVAILABLE` if any mandatory registration is absent or incompatible, then `mcp verify` completes the existing wrapper handshake, tool-list, and syntax probe.
- [x] `smaqit uninstall` removes only exact smaqit MCP registrations and preserves unrelated client configuration, deleting a file only when it becomes wholly smaqit-owned and empty.
- [x] Claude and Codex design-author agents retain their expected MCP tool declarations, implementation agents retain no design-authoring MCP access, and documentation states that these declarations do not themselves register a host server.
- [x] Client documentation tells users to restart/trust Claude Code and Codex after installation and to stop authoring if their tools are still absent; it retains the documented external Codex discovery limitation without adding a fallback.
- [x] Installer unit tests, smoke tests, race coverage where supported, and release cross-builds pass after regenerated embedded assets are prepared.

## Findings

**Implementation approach:**
- Added shared lifecycle helpers for VS Code, Claude Code, and Codex MCP registrations.
- Used JSONC-aware edits for JSON files and marker-owned append/remove edits after TOML semantic inspection.

**Decisions made:**
- Registration and host discovery remain separate gates; no CLI fallback is introduced for absent authoring tools.
- Codex registration is required and uses a trusted project-local MCP table.

**Blockers encountered:**
- Upstream Codex clients can ignore valid project registrations; explicit approval retained the strict stop condition.

**Follow-up identified:**
- Recheck live Codex project-MCP discovery when the upstream client defects are resolved.

## Files to Create / Modify

| File | Action |
|------|--------|
| `installer/design_tools.go` | Modify — Claude and Codex configuration lifecycle |
| `installer/design.go` | Modify — multi-client verification gate |
| `installer/main.go` | Modify — init, validate, uninstall, and messaging wiring |
| `installer/go.mod` and `installer/go.sum` | Modify — TOML parser dependency |
| `installer/*_test.go` | Modify — configuration lifecycle unit coverage |
| `scripts/smoke-test-installer.sh` | Modify — all-client configuration lifecycle coverage |
| `README.md` and `docs/wiki/workflows/quickstart.md` | Modify — client registrations and restart guidance |
| `docs/wiki/workflows/visual-designs.md` | Modify — accurate installation and discovery contract |
| `docs/wiki/agent-tools-reference.md` | Modify — distinguish agent declarations from host registration |
| `docs/wiki/workflows/testing-smaqit.md` and `docs/test-cases/visual-design-artifacts.md` | Modify — release and acceptance expectations |

## Notes

- This task fixes the host-registration omission identified after Task 101. Task 101's runtime materialization, visual-design validation, and client-owned activation boundary remain valid.
- Claude Code project servers use root `.mcp.json`; Codex project servers use trusted `.codex/config.toml` `[mcp_servers.<name>]` tables.
