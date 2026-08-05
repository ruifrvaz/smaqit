# Visual Design Artifact Validation

**Feature:** Canonical PlantUML source/PNG sidecars, producer-owned visual review, and automatic phase readiness

## Preconditions

- Build the current installer binary.
- Node.js 22+ is available.
- Use a temporary consumer project with no package-manager cache or network dependency.

## Installation

- [ ] `smaqit init` creates all five `docs/designs/<layer>/` directories and installs all five design templates.
- [ ] The versioned runtime contains the locked PlantUML MCP, resvg WASM, Noto Sans, manifest, full file hashes, notices, and licenses.
- [ ] `.vscode/mcp.json`, Claude agents, and Codex agents expose the owned MCP; unrelated configuration remains byte-identical.
- [ ] Removing Node from `PATH` makes initialization fail with `DESIGN-TOOLCHAIN-UNAVAILABLE` before creating the target.

## Valid Flow

- [ ] Create one same-layer spec/design pair with bidirectional links and real requirement IDs.
- [ ] `smaqit design render <design.md>` checks syntax and creates a valid PNG at deterministic dimensions with an opaque cream `#FFF9F0` canvas (every pixel alpha is 255).
- [ ] Re-rendering unchanged source creates identical PNG bytes.
- [ ] Validation fails until an agent opens the PNG and `smaqit design attest <design.md>` records the current hashes.
- [ ] `smaqit design validate <design.md>` and `smaqit validate` pass after attestation.

## Failure Matrix

- [ ] Extra prose/block/frontmatter key or uncontrolled profile → `DESIGN-VISUAL-INVALID`.
- [ ] Unsafe include/import or invalid PlantUML → `DESIGN-SYNTAX-INVALID`.
- [ ] Missing source, image, spec, or pair link → `DESIGN-ARTIFACT-MISSING`.
- [ ] Source/image hash or linked lifecycle mismatch → `DESIGN-ARTIFACT-STALE`.
- [ ] Missing/corrupt embedded or materialized runtime → `DESIGN-TOOLCHAIN-UNAVAILABLE`.
- [ ] Owning specification agent cannot open image content → `DESIGN-VISION-UNAVAILABLE`, with no source-reading fallback for its visual gate.
- [ ] Broken requirement or bidirectional reference, invalid PNG, or failed visual rubric → `DESIGN-VISUAL-INVALID`.

## Lifecycle and Migration

- [ ] `smaqit plan --phase=<phase>` exits nonzero without implementation paths when an in-scope specification has a missing, stale, failed, unattested, or lifecycle-behind design.
- [ ] A ready plan preserves the one-spec-path-per-line output contract.
- [ ] Generated implementation agents have no dedicated image or PlantUML MCP capability and instruct direct consumption of linked PlantUML source.

- [ ] An active existing spec without a pair fails validation and remains in path-only `smaqit plan` output.
- [ ] `smaqit status` reports the number of design-blocked active specs.
- [ ] Status-only transitions synchronize designs to the least-advanced linked spec; deprecated artifacts are excluded from mandatory coverage.
- [ ] Reinstallation repairs corrupt owned tooling; uninstall removes owned runtime/config and preserves `docs/designs` plus unrelated MCP entries.
