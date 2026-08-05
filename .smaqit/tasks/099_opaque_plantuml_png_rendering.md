# Opaque PlantUML PNG Rendering

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-08-05
**Started:** 2026-08-05
**Completed:** 2026-08-06

## Description

Ensure the bundled PlantUML SVG-to-PNG renderer emits deterministic, opaque cream `#FFF9F0` PNGs. The v2.0.0 consumer dry run showed that transparent RGBA output can composite to black in an agent image reader, hiding black labels and making mandatory visual review unreliable.

The correction must ship through a new versioned embedded runtime bundle so new and reinitialized projects use the fixed renderer. This is a forward-only fix: previously generated v2.0.0 transparent PNGs are not migrated, repaired, or invalidated.

## Design Decisions

- **PNG format:** Retain lossless PNG; JPEG is not an acceptable substitute for diagrams.
- **Global opacity:** Set an opaque cream `#FFF9F0` background in the Resvg renderer rather than requiring diagram authors to add styling directives.
- **Bundle upgrade:** Bump the versioned design-tool runtime identifier so reinitialization materializes the corrected bundled renderer instead of reusing the valid v2.0.0 runtime.
- **Forward-only scope:** Do not add migration, repair, or retroactive alpha validation for already-generated transparent design PNGs.
- **Release impact:** Publish the completed correction as patch release v2.0.1.

## Implementation Steps

1. Set the Resvg renderer background to opaque cream `#FFF9F0` while retaining deterministic PNG rendering.
2. Bump the embedded design-runtime bundle identifier and regenerate the bundle manifest/archive so consumers receive the updated renderer.
3. Update smoke-test runtime-path assertions for the new bundle identifier.
4. Add regression coverage that decodes the real rendered PNG and verifies every pixel has fully opaque alpha, retaining the byte-for-byte determinism assertion.
5. Update the visual design artifact contract and consumer test case to require an opaque PNG canvas.
6. Run installer unit tests, installer smoke test, cross-platform release builds, and an isolated consumer install/render visual check.
7. Prepare the v2.0.1 patch release after task completion.

## Known Issues Triage

**Triaged:** 2026-08-05
**Tools searched:** PlantUML, resvg-js
**Result:** Advisory

### Blocking Issues
- None.

### Advisory Issues
- [#619 Sprites do not generate with transparency](https://github.com/plantuml/plantuml/issues/619) — `plantuml/plantuml` — opened 2021-08-06 — tangential transparency behavior; does not affect the renderer-level opaque-canvas fix.

### Historical (Closed)
- [#1151 SVG server background not transparent](https://github.com/plantuml/plantuml/issues/1151) — `plantuml/plantuml` — closed 2022-10-06

### Unresolvable Tools
- None.

## Acceptance Criteria

- [x] The bundled SVG-to-PNG renderer produces PNGs with a deterministic opaque cream `#FFF9F0` canvas; every decoded pixel alpha value is 255.
- [x] Identical PlantUML source renders byte-identical PNG output.
- [x] A new or reinitialized project materializes the updated versioned runtime bundle and uses the opaque renderer.
- [x] Installer unit tests include full-alpha regression coverage over the real MCP-to-SVG-to-PNG path.
- [x] Installer smoke coverage recognizes the new runtime bundle identifier and passes with the released renderer.
- [x] Framework guidance and the visual-design test case specify opaque PNG rendering as a mandatory artifact property.
- [x] Installer tests, smoke test, all release-target builds, and an isolated consumer install/render visual check pass.
- [x] No migration, repair, or retroactive validation behavior is added for existing v2.0.0 transparent PNG artifacts.

## Findings

**Implementation approach:**
- Added a global Resvg background and versioned the embedded PlantUML runtime so reinitialization installs the corrected renderer.
- Decoded the rendered PNG in installer tests to assert the exact canvas color and full alpha coverage.

**Decisions made:**
- Used `#FFF9F0` for a soft cream canvas while preserving high contrast for PlantUML labels.
- Kept the correction forward-only; existing v2.0.0 PNG artifacts are unchanged.

**Blockers encountered:**
- The lifecycle resolver rejected completed legacy `B001` parent metadata during unrelated owner completion scans.
- Resolver compatibility now accepts legacy `BNNN` parent IDs alongside current `NNN` IDs.

**Follow-up identified:**
- Prepare and publish the planned v2.0.1 patch release.

## Files to Create / Modify

| File | Action |
|------|--------|
| `tools/plantuml/render-png.mjs` | Modify |
| `installer/design_tools.go` | Modify |
| `scripts/prepare-design-tools.mjs` | Modify |
| `installer/design_test.go` | Modify |
| `scripts/smoke-test-installer.sh` | Modify |
| `framework/ARTIFACTS.md` | Modify |
| `docs/test-cases/visual-design-artifacts.md` | Modify |

## Notes

The v2.0.0 dry run at `/home/ruifrvaz/projects/temp/smaqit-v2.0.0-dry-run/consumer` confirmed successful installation and structural validation, but its rendered PNG contained transparent and partially transparent pixels that hid labels in the image reader. The renderer-level fix applies to every diagram and avoids requiring author-specific style conventions.
