# Register PlantUML MCP Across Clients

**Status:** Not Started
**Created:** 2026-08-06

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

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] Fresh `smaqit init` and project `smaqit update` create valid registrations for `smaqit-plantuml` in `.vscode/mcp.json`, root `.mcp.json`, and `.codex/config.toml` with the intended stdio command and arguments.
- [ ] Preflight rejects malformed configuration or a conflicting same-name server in any client file before a partial installation is created or changed.
- [ ] Reinitialization preserves unrelated VS Code JSONC, Claude JSON, and Codex TOML content while keeping the owned registration exactly current.
- [ ] `smaqit validate` and `smaqit mcp verify` fail with `DESIGN-TOOLCHAIN-UNAVAILABLE` if any mandatory registration is absent or incompatible, then `mcp verify` completes the existing wrapper handshake, tool-list, and syntax probe.
- [ ] `smaqit uninstall` removes only exact smaqit MCP registrations and preserves unrelated client configuration, deleting a file only when it becomes wholly smaqit-owned and empty.
- [ ] Claude and Codex design-author agents retain their expected MCP tool declarations, implementation agents retain no design-authoring MCP access, and documentation states that these declarations do not themselves register a host server.
- [ ] Client documentation tells users to restart/trust Claude Code and Codex after installation and to stop authoring if their tools are still absent; it retains the documented external Codex discovery limitation without adding a fallback.
- [ ] Installer unit tests, smoke tests, race coverage where supported, and release cross-builds pass after regenerated embedded assets are prepared.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

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
