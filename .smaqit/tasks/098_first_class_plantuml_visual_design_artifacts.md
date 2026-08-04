# First-Class PlantUML Visual Design Artifacts

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-08-03
**Started:** 2026-08-03
**Completed:** 2026-08-04

## Description

Add canonical, layer-linked PlantUML design sources with generated PNG projections, bundled offline tooling, deterministic validation, and mandatory image-based authoring review across all supported agent hosts. Designs become a third first-class artifact type while remaining sidecars within the existing Business, Functional, Stack, Infrastructure, and Coverage layers; implementation agents consume the validated PlantUML source directly.

The released Go binary must embed the locked JavaScript and WebAssembly toolchain so consumer execution performs no mutable package resolution or network fetch. Node 22 or newer remains an explicit external runtime prerequisite, and initialization or execution must fail rather than fall back when the toolchain, design artifact, or image-reading capability is unavailable.

## Design Decisions

- **Layer model:** Designs are sidecars in the five existing layers, not a sixth layer.
- **Canonical representation:** PlantUML Markdown is canonical and editable; its same-basename PNG is mandatory and authoritative for authoring-time visual interpretation.
- **Artifact storage:** Both source and PNG are committed under `docs/designs/<layer>/` and referenced by owning specifications.
- **Content boundary:** Design Markdown contains only YAML frontmatter and exactly one PlantUML block; specifications contain links only and never embed diagrams.
- **Minimality:** Every active specification must be covered by at least one high-signal same-layer design, while one design may cover multiple related specifications.
- **Lifecycle coupling:** A semantic spec edit invalidates linked designs; a semantic design edit invalidates every linked active spec; a shared design cannot advance beyond its least-advanced linked spec.
- **Validation model:** Structural, PlantUML rendering, and specification-agent visual-review gates are independently mandatory before handoff; `smaqit plan --phase` automatically enforces current design readiness before exposing implementation work.
- **Consumption model:** Implementation agents read specifications for requirements and PlantUML source for design structure. They do not visually revalidate PNGs or own design repair.
- **No fallback:** Authoring-time visual validation never falls back to reading PlantUML source; unavailable image capability stops the owning specification agent with `DESIGN-VISION-UNAVAILABLE`.
- **Tool distribution:** The released Go binary embeds exact locked PlantUML MCP and SVG-to-PNG JavaScript/WASM dependencies; consumer execution never invokes npm, npx, or network package resolution.
- **Runtime prerequisite:** Node 22 or newer remains external and is checked before project mutation.
- **MCP configuration:** MCP configuration is agent-local wherever supported, with a smaqit-owned VS Code workspace entry where required.
- **Migration:** Existing projects fail validation and phase execution until active specifications have compliant design pairs; no compatibility artifacts or placeholders are generated.
- **Installer scope:** Literal remote `go install` support is out of scope; the existing single released Go binary remains the installation mechanism.
- **Capability verification:** Cross-platform model vision will be exercised through a normal consumer workflow and reported separately if host issues arise.

## Implementation Steps

1. Define the design artifact authority boundary, identifiers, metadata schema, controlled layer profiles, reference syntax, source/PNG synchronization, visual attestation, lifecycle rules, and canonical failure taxonomy in the framework.
2. Add five design templates and a mechanically parseable Design References section to all specification templates.
3. Build a deterministic compressed tool bundle containing pinned `@plantuml/mcp-js`, `@resvg/resvg-wasm`, exact transitive dependencies, conversion wrapper, manifest, checksums, licenses, and notices.
4. Embed and atomically materialize the tool bundle under `.smaqit/tools/plantuml/<bundle-version>`, with safe extraction, manifest verification, Node 22+ preflight, repair, upgrade, and uninstall ownership.
5. Install host-specific MCP integration for Copilot cloud, VS Code, Claude Code, and Codex without modifying unrelated user configuration.
6. Implement `smaqit mcp plantuml`, `smaqit design render`, and `smaqit design validate` using the official Go MCP SDK and the bundled renderer.
7. Extend `smaqit validate`, `plan`, and `status` with design parsing, graph validation, hash integrity, lifecycle readiness, strict migration diagnostics, and stable agent-facing output.
8. Update agent metadata generation and all five specification agents to author, render, visually inspect, attest, and reference design pairs with hard failure and no fallback.
9. Update Development, Deployment, Validation, greenfield, and feature workflows to consume exact PlantUML Markdown paths after the automatic `smaqit plan --phase` readiness gate.
10. Add unit, smoke, offline, corruption, reinstallation, uninstall, generated-agent, and release-target verification.
11. Update framework, quick-start, traceability, lifecycle, testing, migration, licensing, and consumer documentation.

## Known Issues Triage
**Triaged:** 2026-08-03
**Tools searched:** PlantUML, resvg-js, MCP Go SDK, Node.js, VS Code, Claude Code
**Result:** Advisory

### Blocking Issues
- None.

### Advisory Issues
- [#855 Potential race condition in (*mcp.ClientSession).Close()](https://github.com/modelcontextprotocol/go-sdk/issues/855) — `modelcontextprotocol/go-sdk` — opened 2026-03-24 — needs investigation, P2
- [#78820 Resuming a session with a pending background-task notification drops all configured MCP servers](https://github.com/anthropics/claude-code/issues/78820) — `anthropics/claude-code` — opened 2026-07-18 — bug, has repro, platform:macos, area:mcp, area:agent-sdk
- [#34751 "Request too large (max 20MB)" error on small files (99KB PNG)](https://github.com/anthropics/claude-code/issues/34751) — `anthropics/claude-code` — opened 2026-03-15 — bug, platform:linux, area:core
- [#51770 Sub-agent docs omit `mcpServers` support for main-thread `--agent` sessions](https://github.com/anthropics/claude-code/issues/51770) — `anthropics/claude-code` — opened 2026-04-22 — enhancement, area:mcp, area:agents, area:docs, stale

### Historical (Closed)
- [#271589 [tools sets] support qualified names](https://github.com/microsoft/vscode/issues/271589) — `microsoft/vscode` — closed 2026-06-08
- [#53865 Subagent tools field does not support MCP wildcards or server-prefix shortcuts](https://github.com/anthropics/claude-code/issues/53865) — `anthropics/claude-code` — closed 2026-05-30

### Unresolvable Tools
- None.

## Acceptance Criteria

- [x] Initialization creates five layer-specific design directories and installs five design templates.
- [x] Canonical design Markdown accepts only required YAML frontmatter and exactly one fenced PlantUML diagram, with no prose or additional blocks.
- [x] Every active specification has at least one valid same-layer design reference, and spec/design references agree bidirectionally without orphan or duplicate design IDs.
- [x] PlantUML MCP, the SVG-to-PNG converter, and every JavaScript/WASM transitive dependency are exactly pinned, embedded in the released Go binary, and accompanied by required license notices.
- [x] Consumer initialization and execution perform no npm, npx, or network package resolution.
- [x] Initialization checks Node 22+ and tool integrity before project mutation and exits nonzero without partial design-tool installation when either check fails.
- [x] `smaqit design render` deterministically validates PlantUML syntax, renders SVG, converts it to PNG, and records normalized source and raw-image SHA-256 values.
- [x] `smaqit design validate` and `smaqit validate` reject invalid metadata, uncontrolled profiles, extra content, unsafe includes, broken references, missing requirements, invalid PNGs, stale hashes, and incompatible lifecycle state.
- [x] `smaqit plan --phase` exits nonzero before emitting implementation work when an in-scope design is missing, stale, failed, or lifecycle-behind; ready plans preserve the existing one-spec-path-per-line output, and `smaqit status` reports design debt.
- [x] All five specification agents use the PlantUML MCP, generate the PNG, read it visually, apply the standard visual rubric, and record attestation against the current hashes before completion.
- [x] Development, Deployment, and Validation consume validated PlantUML source directly after the automatic phase-readiness gate and never author, render, visually review, attest, or repair designs.
- [x] Missing image capability stops the owning specification agent with `DESIGN-VISION-UNAVAILABLE`; no authoring instruction permits PlantUML-source reading as a visual-review fallback.
- [x] Installed Copilot, Claude, Codex, and VS Code integration exposes MCP and image tools to specification agents only while preserving unrelated configuration through initialization, reinstallation, and uninstall.
- [x] Existing projects with active specifications but incomplete designs receive strict actionable migration failures and cannot proceed through affected phases.
- [x] Reinstallation repairs or upgrades only smaqit-owned design tooling, and uninstall removes owned tooling/configuration while preserving `docs/designs` and unrelated configuration.
- [x] Unit and smoke tests cover valid flows, every canonical failure code, offline execution, corrupt bundles, stale images, lifecycle propagation, configuration ownership, and all supported release targets.
- [x] Framework, workflow, migration, testing, licensing, and consumer documentation describe the implemented contract without retaining Mermaid-specific design guidance.

## Findings

**Implementation approach:**
- Added strict PlantUML sidecar parsing, rendering, attestation, graph validation, lifecycle validation, phase gates, and design-debt reporting to the shipped Go CLI.
- Embedded an exactly pinned JavaScript/WASM rendering bundle with integrity checks, license notices, offline materialization, repair, and uninstall ownership.
- Added layer templates, generated-agent capabilities, host integrations, smoke coverage, release builds, and consumer documentation.

**Decisions made:**
- Kept designs canonical within the five existing specification layers and limited each file to metadata plus one PlantUML block.
- Assigned image-based authoring review exclusively to specification agents; implementation agents consume validated specification and PlantUML source after the automatic phase gate.
- Enforced a single fail-fast path with Node 22 or newer as the only external prerequisite and no renderer, vision, migration, or package-resolution fallback.

**Blockers encountered:**
- The lifecycle resolver could not parse an unrelated legacy task's descriptive `Parent` metadata; Task 098 ownership and absence of child tasks were verified directly from the registered worktree and task references.

**Follow-up identified:**
- Exercise the released installer through normal consumer workflows on each supported host and report any platform-specific MCP or model-vision issue.

## Files to Create / Modify

| File | Action |
|------|--------|
| `framework/{SMAQIT,ARTIFACTS,LAYERS,TEMPLATES,AGENTS,PHASES}.md` | Modify |
| `templates/specs/*.template.md` | Modify |
| `templates/designs/*.template.md` | Create |
| `agents/{business,functional,stack,infrastructure,coverage,development,deployment,validation}.md` | Modify |
| `.smaqit/definitions/agents/*.frontmatter.yaml` | Modify |
| `scripts/generate-agents.py` | Modify |
| `installer/main.go` | Modify |
| `installer/spec.go` | Modify |
| `installer/go.mod` and `installer/go.sum` | Modify |
| `installer/Makefile` and `installer/build.bat` | Modify |
| `installer/main_test.go` and new focused Go test files | Modify / Create |
| `scripts/smoke-test-installer.sh` | Modify |
| `.github/workflows/{copilot-setup-steps,installer-smoke-test,post-merge-release}.yml` | Modify |
| `skills/smaqit.spec-status-update/SKILL.md` | Modify |
| `skills/smaqit.new-greenfield-project/SKILL.md` | Modify |
| `skills/smaqit.feature-new/SKILL.md` | Modify |
| `skills/smaqit.design-validate/` | Create |
| `README.md`, `docs/wiki/`, and `docs/test-cases/` | Modify |

## Notes

The official PlantUML JavaScript MCP currently renders SVG only, so the embedded resvg WASM step is part of the mandatory rendering pipeline. Specification-agent visual-review evidence is stored with the design metadata and bound to source/image hashes. Downstream phases trust that attestation through the automatic plan gate and consume PlantUML source without repeating image review.
