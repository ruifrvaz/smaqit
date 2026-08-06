# Canonical Visual Designs

Smaqit treats a small set of high-signal visual designs as mandatory specification sidecars. Designs do not form a sixth layer and do not duplicate specification prose.

## Contract

Each active spec links at least one same-layer pair:

```text
docs/designs/<layer>/<design-id>.md
docs/designs/<layer>/<design-id>.png
```

The Markdown file is canonical for model structure and contains only required YAML frontmatter plus exactly one `plantuml` fence. The PNG is the generated projection used for specification-agent visual validation and human review. Specifications contain only links in `## Design References`; design metadata contains normalized spec paths and requirement IDs.

Default profiles are `use-case` (Business), `system-sequence` (Functional), `component` (Stack), `deployment` (Infrastructure), and `requirement-trace` (Coverage). Additional Functional domain/context/state views are justified only when they materially clarify the model. Avoid ceremonial diagrams.

## Authoring Gate

1. Start from `.smaqit/templates/designs/<layer>.template.md` and link the pair from its specs. Its active `status` must equal the least-advanced linked specification; use `smaqit.spec-status-update` for status-only changes or validation returns `DESIGN-ARTIFACT-STALE`.
2. Use the `smaqit-plantuml` MCP tools to check syntax and iterate.
3. Run `smaqit design render docs/designs/<layer>/<id>.md` to produce the current PNG and record hashes.
4. The owning specification agent opens the PNG with its image-reading tool and checks legibility, clipping, direction/order, boundaries, disconnected elements, coherence, and excessive complexity.
5. Correct and rerender until the image passes, then run `smaqit design attest <file>` and `smaqit design validate <file>`.

Image capability is mandatory for specification agents. If it is unavailable, stop with `DESIGN-VISION-UNAVAILABLE`; PlantUML source reading is not an authoring-time visual-review fallback.

After installation, run `smaqit mcp verify` to prove the generated configuration and local stdio transport. This does not prove an interactive client has exposed the tools: open the project in VS Code and use **MCP: List Servers** to trust/start `smaqit-plantuml`; in Claude Code and Codex, start a fresh session and confirm the specification agent receives its two declared tools. If they are absent, stop authoring with `DESIGN-TOOLCHAIN-UNAVAILABLE`; do not substitute direct CLI calls.

At implementation handoff, `smaqit plan --phase=<phase>` exits nonzero unless every in-scope design pair retains a current visual attestation. Implementation agents do not reopen PNGs: they read specification Markdown for requirements and PlantUML Markdown for canonical design structure.

## Strict Migration

After updating and re-running `smaqit init`, an existing project with active specs and incomplete design pairs fails `smaqit validate`, `smaqit plan`, and affected phase completion checks. Migrate each active spec by authoring the smallest useful same-layer pair, adding bidirectional references, rendering, visually reviewing, attesting, and validating it. Smaqit does not create placeholders or waive the gate.

## Installation and Ownership

The released Go binary embeds exact npm-lock-resolved dependencies for `@plantuml/mcp-js`, `@resvg/resvg-wasm`, and Noto Sans. Initialization first verifies the archive and Node.js 22+, then materializes the versioned runtime under `.smaqit/tools/plantuml/`. It adds the narrow `.smaqit/tools/` rule to the project root `.gitignore`, preserving user rules so the generated runtime is not committed; canonical `docs/designs/` Markdown and PNG artifacts remain versioned. It configures the owned `smaqit-plantuml` server in `.vscode/mcp.json`, while Claude and Codex agents carry project-local MCP declarations.

Reinstallation repairs smaqit-owned runtime/configuration and preserves unrelated MCP entries. Uninstall removes owned tooling and configuration but preserves `docs/designs/`.
