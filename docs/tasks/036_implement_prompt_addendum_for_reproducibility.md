# Implement Prompt Addendum for Reproducibility

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #5 (2025-12-27)

## Description

When users refine specs iteratively (e.g., "fix the stack spec"), the additional instructions are not captured in prompt files, breaking reproducibility. Implement "Addendum" section to capture all iterative refinement instructions.

**ASSESS BEFORE START**. could be that we do not want a prompt addendum but a refactor that updates the existing specs. it's best to determine which option is better

## Acceptance Criteria

- [ ] Prompt templates updated with optional `## Addendum` section
- [ ] Section documented as: "Iterative refinements and amendments (auto-generated)"
- [ ] Agent instructions updated to detect spec modification requests
- [ ] Agents append refinement instructions to prompt file with timestamp
- [ ] Addendum format: `[YYYY-MM-DD HH:MM] [refinement instruction]`
- [ ] Specification agent templates (Level 1) include addendum appending logic
- [ ] All 5 specification agents (Level 2) implement addendum behavior
- [ ] Framework documentation (PROMPTS.md) explains addendum principle

## Impact

**Severity:** High  
**User Impact:** Breaks reproducibility principle; prompt files no longer represent complete input record; regenerating specs from prompt would produce different output than current specs

## Notes

**FRAMEWORK LEVEL CHANGE**. Critical for maintaining input record completeness. Agents must detect when they're modifying existing specs vs creating new specs.
