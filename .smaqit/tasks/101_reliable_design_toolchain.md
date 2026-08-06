# Reliable Design Toolchain

**Status:** In Progress
**Created:** 2026-08-06
**Mode:** Assisted
**Started:** 2026-08-06

## Description

Make the mandatory PlantUML design toolchain reliable across normal consumer workflows. The downstream pilot showed that a valid secondary Git worktree lacks the ignored runtime, MCP configuration alone does not make tools reachable in a live client session, Claude metadata does not match its documented schema, lifecycle coupling is insufficiently visible to authors, and validation reports only one independent failure at a time.

The task preserves smaqit's strict authoring model: PlantUML MCP tooling is mandatory, and an unavailable authoring tool is a clear stop condition rather than a CLI fallback. It adds a deterministic local transport check while documenting the remaining client-owned activation check explicitly.

## Design Decisions

- **On-demand embedded runtime:** All project-local PlantUML consumers prepare the embedded, pinned runtime when it is absent. Manual copying between worktrees is not supported.
- **Cross-process-safe materialization:** Runtime extraction uses a per-bundle lock, atomic publication, revalidation after lock acquisition, and safe rollback so concurrent agents cannot publish or delete partial state.
- **Transport versus host activation:** `smaqit mcp verify` proves generated configuration and the actual `smaqit mcp plantuml` stdio transport. It cannot attest to VS Code/Claude/Codex session trust or tool discovery; specification agents MUST verify their declared tools and stop with an activation diagnostic if absent.
- **All supported clients:** Keep VS Code/Copilot, Claude Code, and Codex in scope. Correct Claude's generated `mcpServers` declaration to its documented list form; do not introduce compatibility fallbacks.
- **Two validation modes:** Explicit `smaqit design validate` and `smaqit validate` collect deterministic independent artifact errors. `smaqit plan` phase readiness remains fail-fast because it is an execution gate.
- **Lifecycle coupling is user-facing:** Active design status equals the least-advanced active linked specification status. Agents make status-only changes through `smaqit.spec-status-update`.

## Implementation Steps

1. Add a shared, read/write runtime-preparation helper in `installer/design_tools.go` that preflights Node/archive integrity, locks per bundle, materializes missing or corrupt runtime atomically, revalidates it, and preserves the narrow Git-ignore ownership contract.
2. Route `runPlantUMLMCP`, `openPlantUMLSession`, and `cmdValidate` through that helper without changing the pure read-only validation helper used by archive checks.
3. Add `smaqit mcp verify`: inspect owned configuration, launch the public `smaqit mcp plantuml` wrapper in the project root, complete MCP initialization, list the expected tools, and run a fixed syntax probe. Add help and actionable diagnostics.
4. Correct Claude author-agent generation to emit supported list-form `mcpServers`, then update generated metadata assertions.
5. Make authoring-agent instructions require MCP tool presence and stop on unavailable activation; document VS Code, Claude Code, and Codex activation/trust/session-reload checks separately from local verification.
6. State lifecycle-rank synchronization and `DESIGN-ARTIFACT-STALE` behavior beside every design template status field and in the visual-design workflow.
7. Refactor explicit design validation to collect all independent design and active-spec diagnostics in deterministic layer/path order; preserve immediate failure when a global prerequisite prevents further checks.
8. Add unit, real-Git-worktree, concurrent-materialization, wrapper-MCP, schema, lifecycle, and multi-failure regression coverage. Rebuild generated installer assets through the established Makefile path before testing.

## Known Issues Triage

**Triaged:** 2026-08-06
**Tools searched:** PlantUML MCP, Visual Studio Code, Claude Code, Codex, Model Context Protocol Go SDK
**Result:** Blocking

### Blocking Issues
- [#30922 MCP Tools Not Exposed to Agent in Codex Desktop, Only Resources Are Available](https://github.com/openai/codex/issues/30922) — `openai/codex` — opened 2026-07-02 — bug, mcp, tool-calls, app

Proceeding by explicit user approval on 2026-08-06. The implementation must retain an in-code reference to this client-owned limitation and fail authoring clearly when declared MCP tools remain unavailable.

### Advisory Issues
- [#324993 MCP tools and vscode.lm.registerTool tools silently dropped in Agent mode on Remote SSH host](https://github.com/microsoft/vscode/issues/324993) — `microsoft/vscode` — opened 2026-07-08 — the Remote SSH scope does not match the pilot, but it confirms client-side MCP tool exposure remains host-owned.
- [#3426 Claude Code fails to expose MCP tools to AI sessions when running a local Playwright MCP server](https://github.com/anthropics/claude-code/issues/3426) — `anthropics/claude-code` — opened 2025-07-13 — a different server, but the same local-MCP exposure category.

### Historical (Closed)
- [#13898 Custom Subagents Cannot Access Project-Scoped MCP Servers (Hallucinate Instead)](https://github.com/anthropics/claude-code/issues/13898) — `anthropics/claude-code` — closed 2026-05-23
- [#395 Claude Desktop Not Able To Detect The Tool](https://github.com/modelcontextprotocol/go-sdk/issues/395) — `modelcontextprotocol/go-sdk` — closed 2025-09-05

### Unresolvable Tools
- None

## Acceptance Criteria

- [ ] A Git worktree with no `.smaqit/tools/` can run `smaqit design render`, `smaqit design validate`, `smaqit validate`, and `smaqit mcp plantuml` without manual runtime copying when Node 22+ is available.
- [ ] Missing Node, a corrupt archive, or failed extraction leaves no usable partial runtime, lock, temporary directory, or backup artifact and reports `DESIGN-TOOLCHAIN-UNAVAILABLE`.
- [ ] Concurrent runtime preparation for one project leaves exactly one valid versioned bundle and all callers either succeed or receive a recoverable, actionable error.
- [ ] `smaqit mcp verify` validates configuration and the public wrapper's MCP handshake, required tool list, and syntax call using the installed binary.
- [ ] Claude design-author metadata uses documented list-form `mcpServers`; Copilot and Codex author-only declarations remain valid and implementation agents continue to receive neither MCP authoring tools nor image-review responsibility.
- [ ] Consumer guidance distinguishes local transport verification from client-owned activation and requires authoring agents to stop if their MCP tools are unavailable.
- [ ] Every source design template and authoring workflow states lifecycle-rank synchronization and the stale-artifact consequence.
- [ ] One explicit design-validation run reports all independent invalid designs and active unpaired specs in stable order, while global runtime/session prerequisite failures remain fail-fast.
- [ ] `smaqit plan` phase readiness retains its existing fail-fast behavior.
- [ ] Installer unit tests, smoke tests, race/concurrency coverage where supported, and release cross-builds pass after regenerated embedded assets are prepared.

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
| `installer/design_tools.go` | Modify — runtime preparation/locking and static MCP readiness |
| `installer/design.go` | Modify — wrapper transport verification and aggregate validation |
| `installer/main.go` | Modify — command dispatch, help, and validation integration |
| `scripts/generate-agents.py` | Modify — Claude MCP metadata schema |
| `installer/design_test.go` | Modify — runtime, wrapper, concurrency, and aggregate-validation tests |
| `installer/spec_test.go` | Modify — retain fail-fast phase-gate regression coverage |
| `scripts/smoke-test-installer.sh` | Modify — installed-binary, worktree, and generated metadata checks |
| `templates/designs/*.template.md` | Modify — lifecycle synchronization guidance |
| `templates/AGENTS.md.template` | Modify — strict authoring-tool activation behavior |
| `docs/wiki/workflows/visual-designs.md` | Modify — verification and lifecycle workflow |
| `docs/wiki/agent-tools-reference.md` | Modify — declaration versus host-activation boundary |
| `docs/test-cases/visual-design-artifacts.md` | Modify — deterministic multi-failure expectation |
| `README.md` and `docs/wiki/workflows/quickstart.md` | Modify — post-install MCP verification guidance |

## Notes

- Source generated installer assets are prepared through `make -C installer prepare`; do not hand-edit generated `installer/templates`, agent, skill, or runtime staging directories.
- Diagnosis evidence: `.smaqit/reports/diagnose-2026-08-06.md`.
