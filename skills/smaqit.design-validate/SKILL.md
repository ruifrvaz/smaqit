---
name: smaqit.design-validate
description: Validate canonical smaqit PlantUML/PNG design pairs using deterministic CLI gates and mandatory authoring-time image interpretation. Use after a specification agent authors or changes a design, or whenever a user asks to validate, review, or repair files under docs/designs/.
---

# Validate Visual Designs

## Inputs

- Accept exact design Markdown paths, specification paths whose `## Design References` resolve the pairs, or all active project designs when no paths are supplied.
- Treat the same-basename Markdown and PNG as one artifact. Do not infer a substitute path when either side is missing.

## Procedure

1. Resolve each exact Markdown/PNG pair and its linked specs.
2. When authoring or repairing a design, use the `smaqit-plantuml` MCP tools to iterate, then run `smaqit design render <design.md>`.
3. Open the generated PNG with the active platform's image-reading tool. If image content is unavailable, stop with `DESIGN-VISION-UNAVAILABLE`. Do not read PlantUML source as a visual-review fallback.
4. Inspect the image for:
   - legible labels at normal viewing scale;
   - no clipping, overlap, or truncated elements;
   - correct interaction direction and sequence order;
   - coherent system, domain, component, or deployment boundaries;
   - no unexplained disconnected elements;
   - consistency with linked specifications and requirement IDs;
   - minimal complexity with no ceremonial content.
5. If the image fails, stop with `DESIGN-VISUAL-INVALID` and report the exact visual defect. The owning specification agent may correct the PlantUML and repeat from step 2; other callers request correction from that layer instead of editing the design.
6. After an authoring review passes, run `smaqit design attest <design.md>` to bind the review to the current source/image hashes.
7. Run `smaqit design validate <design.md>` for each pair, or `smaqit design validate` for the full project. Do not proceed while any gate fails.

## Output

Report every reviewed Markdown/PNG pair, visual result, CLI validation result, and any failure code. Never reproduce specification prose inside a design or PlantUML content inside a specification.
