# Task 064: Complete Level 0 Principle Cleanup

**Status:** new  
**Priority:** High  
**Created:** 2026-01-13

## Context

Task B001.1 created Agent-L0 (Principle Curator) to maintain Level 0 principle purity per the Level Up Architecture. Initial cleanup of `SMAQIT.md` successfully removed directives and implementation details, demonstrating the contamination pattern across framework files.

Assessment revealed significant contamination in remaining framework files:
- `LAYERS.md` — Contains extensive MUST/MUST NOT directives throughout
- `PHASES.md` — Contains procedural workflows and directive statements
- `AGENTS.md` — Opens with "Agents MUST follow..." (contaminated from line 1)
- `TEMPLATES.md` — Contains structural rules (may be acceptable as meta-level)
- `ARTIFACTS.md` — Contains format specifications and rules
- `PROMPTS.md` — Not yet assessed

## Problem

Framework files claim to contain "philosophy" but read like instruction manuals. This violates Level 0 principle purity and undermines the Level Up Architecture vision where:
- **L0** contains pure principles (WHY/WHAT conceptually)
- **L1** compiles principles into template directives (MUST/SHOULD/HOW structurally)
- **L2** compiles directives into product agents (executable behaviors)

## Goal

Complete Level 0 cleanup across all framework files, ensuring they contain only principles and philosophy without directives, implementation details, or procedural instructions, **while preserving all information and knowledge through transformation and relocation rather than deletion**.

No knowledge should be lost or become vague during cleanup. Directives and implementation details extracted from L0 must be:
- Documented completely with full context
- Transformed into appropriate forms (not simplified or generalized away)
- Verified for completeness against original contaminated content

## Scope

### In Scope
- Clean `LAYERS.md` — Remove directives, preserve layer philosophy
- Clean `PHASES.md` — Remove directives, preserve phase philosophy
- Clean `AGENTS.md` — Remove directives, preserve agent philosophy
- Clean `ARTIFACTS.md` — Remove format rules, preserve artifact philosophy
- Assess and clean `PROMPTS.md` if contaminated
- Document all extracted directives for L1 compilation
- Verify `TEMPLATES.md` (may be meta-level exempt)

### Out of Scope
- L1 compilation (that's Agent-L1's responsibility per Task 065)
- L2 compilation (that's Agent-L2's responsibility per Task 066)
- Creating new principles (only cleaning existing)
- Wiki documentation updates (defer to separate task if needed)

## Extracted Directives from SMAQIT.md

Reference for L1 compilation (Task 065):

### Single Source of Truth
- `MUST NOT duplicate information from existing specs`
- `Use Foundation Reference for same-layer references`
- `Use Implements/Enables for upstream references`
- `SHOULD update existing specs when extending concepts`
- `Create new specs only for distinct concepts`

### Explicit Over Implicit
- `State assumptions rather than assume shared context`
- `Define scope boundaries rather than imply them`
- `Reference sources rather than expect inference`

### Fail-Fast on Ambiguity
- `Do not invent requirements`
- `Flag assumptions explicitly`

### Quick Reference Removed
- Layer/phase mappings (likely duplicate LAYERS.md/PHASES.md)
- Agent name patterns
- File path references with consultation criteria

## Acceptance Criteria

- [ ] All framework files (`SMAQIT.md`, `LAYERS.md`, `PHASES.md`, `AGENTS.md`, `ARTIFACTS.md`, `PROMPTS.md`) contain zero MUST/SHOULD/MUST NOT directive statements in principle sections
- [ ] No file paths, commands, or technical specifics in principle descriptions
- [ ] No procedural instructions or step-by-step workflows in principles
- [ ] All extracted directives documented with target L1 template mappings
- [ ] Extracted content includes full original text (not paraphrased or summarized)
- [ ] Files read naturally as human-readable philosophy
- [ ] Zero conceptual meaning lost (principles preserved, form transformed)
- [ ] Zero knowledge lost (directives/details relocated, not deleted)
- [ ] Extraction documentation verified against original contaminated sections
- [ ] `TEMPLATES.md` assessed and either cleaned or documented as meta-level exempt
- [ ] Completion documented in session history

## Dependencies

- Task B001.1 (Agent-L0 created) — ✅ Completed
- Initial `SMAQIT.md` cleanup — ✅ Completed

## Blocks

- Task 065 (Clean Up Level 1 Templates) — Cannot compile directives until extraction complete

## Notes

### Principle vs Directive Examples

**Principle form (L0):**
- "Single Source of Truth: Each piece of information exists in exactly one place"
- "Layer Independence: Each layer receives requirements from its own prompt file"

**Directive form (L1):**
- "Agents MUST NOT duplicate information from existing specs"
- "MUST read from layer prompt file only"

**Implementation detail (L1):**
- "Read from `.github/prompts/smaqit.[layer].prompt.md`"
- "Output to `specs/[layer]/` directory"

### Files Likely Most Contaminated

1. **AGENTS.md** — Opens with directive list, likely heavily contaminated
2. **PHASES.md** — Contains workflow steps and completion criteria
3. **LAYERS.md** — Contains "MUST/MUST NOT" sections per layer
4. **ARTIFACTS.md** — Format specifications throughout

### Cleanup Strategy

Use Agent-L0 iteratively with knowledge preservation:
1. **Before cleaning:** Read entire file to understand complete context
2. **During extraction:** Capture full original text of contaminated sections
3. **Transform principle:** Rewrite as pure philosophy (WHY/WHAT conceptually)
4. **Document directive:** Record exact text, underlying principle, target L1 location
5. **Verify completeness:** Ensure no specificity lost in transformation
6. **Move to next file:** Repeat until all framework files cleaned
7. **Aggregate extractions:** Compile complete directive map for L1 handover

**Knowledge Preservation Rule:**
If a directive contains specific information (mechanisms, formats, criteria), that information MUST appear verbatim in extraction documentation. Generic summaries like "add validation rules" are insufficient—the actual rules must be preserved.
