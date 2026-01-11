# Reset Checkboxes on Requirement Refinement

**Status:** Completed  
**Created:** 2026-01-11  
**Completed:** 2026-01-11  
**Priority:** Medium  
**Related:** Issue 10 from Task 059 (E2E Regression Testing)

## Description

When specification agents modify existing acceptance criteria to expand scope during incremental additions, they should reset the checkbox state from `[x]` to `[ ]` to indicate revalidation is needed for the expanded scope.

**Current Behavior:**
- Functional agent updated FUN-OUTPUT-006 and FUN-OUTPUT-013 to include Luigi character scope
- Checkboxes remained `[x]` (checked) even though requirements expanded
- This misleads developers - criteria appear satisfied but actually need revalidation

**Expected Behavior:**
- When agent modifies acceptance criterion text, checkbox resets to `[ ]`
- Development agent later updates checkbox to `[x]` when expanded criterion is satisfied
- Provides accurate per-requirement status during incremental development

## Acceptance Criteria

- [x] Business agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [x] Functional agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [x] Stack agent directive: Reset checkbox to `[ ]` when modifying existing acceptance criteria text
- [!] Validation: Re-run Luigi incremental addition test case (Deferred - requires interactive testing agent workflow)
- [!] Validation: Verify modified requirements show `[ ]` after spec update (Deferred - part of E2E testing)
- [!] Validation: Verify Development agent later updates to `[x]` after implementation (Deferred - part of E2E testing)

## Implementation Summary

### Changes Made

**Level 0 (Framework - ARTIFACTS.md):**
- Added "Checkbox Lifecycle During Refinement" section after "Acceptance Criteria State"
- Documented explicit rules:
  - Specification agents MUST reset `[x]` → `[ ]` when modifying acceptance criterion text
  - Specification agents MUST reset `[!]` → `[ ]` when modifying acceptance criterion text
  - Implementation agents later update `[ ]` → `[x]` or `[!]` after revalidation
  - Adding new criteria always starts with `[ ]`
- Included rationale: Expanded/modified requirements need revalidation; checkboxes reflect implementation status
- Provided concrete example showing checkbox lifecycle: before update (`[x]`) → after update (`[ ]`) → after reimplementation (`[x]`)

**Level 2 (Agents):**
- Updated `agents/smaqit.business.agent.md`: Added MUST directive "Reset checkbox to `[ ]` when modifying existing acceptance criteria text (expanded scope requires revalidation)"
- Updated `agents/smaqit.functional.agent.md`: Added identical MUST directive
- Updated `agents/smaqit.stack.agent.md`: Added identical MUST directive

### Validation Performed

**Build Validation:**
- ✅ Installer built successfully (version 9b0c908-dirty)
- ✅ No compilation errors
- ✅ Changes are surgical and consistent across all three agents

**Logical Validation:**
- ✅ Framework principle clearly documented in ARTIFACTS.md
- ✅ All affected agents have consistent directive
- ✅ Directive is clear and actionable ("Reset checkbox to `[ ]` when modifying existing acceptance criteria text")
- ✅ Rationale provided for understanding the "why"
- ✅ Example demonstrates the lifecycle clearly

**E2E Testing:**
- ⚠️ Deferred: Luigi incremental addition test requires interactive testing agent workflow
- ⚠️ Testing agent requires opening new workspace and manual interaction
- ⚠️ Will be validated in future E2E regression testing sessions

## Design Decisions

### Level Hierarchy Approach

**Decision:** Follow smaqit levels (Level 0 → Level 2), skip Level 1 templates.

**Rationale:**
- Level 0 (Framework) defines the principle and rules
- Level 1 (Templates) not needed - templates don't contain checkbox directives (agents execute checkbox logic)
- Level 2 (Agents) implement the directive with consistent wording

### Directive Placement

**Decision:** Added to MUST section (not SHOULD).

**Rationale:**
- Checkbox reset is not optional - misleading checkboxes cause real problems
- Severity is Medium because frontmatter `status: draft` provides fallback signal
- But per-requirement accuracy matters for developer workflow
- MUST ensures consistent behavior across all specification agents

### Consistent Wording

**Decision:** Used identical directive text across all three agents.

**Rationale:**
- Reduces cognitive load
- Ensures predictable behavior
- Makes cross-agent consistency easy to verify
- Aligns with Single Source of Truth principle (documentation in one place, referenced consistently)

## Notes

**Severity:** Medium (not High) because frontmatter `status: draft` still indicates global revalidation needed, reducing impact of misleading checkbox state.

**Impact:** Per-requirement accuracy is lost during intermediate state. Developers checking individual criteria status could be misled about what actually needs revalidation.

**Affected Agents:**
- `agents/smaqit.business.agent.md` ✅ Updated
- `agents/smaqit.functional.agent.md` ✅ Updated
- `agents/smaqit.stack.agent.md` ✅ Updated

**Not Affected:**
- Infrastructure agent (Phase 2 - no prior implementation to modify)
- Coverage agent (reads only, doesn't modify upstream specs)

## Future Work

**E2E Testing:**
- Run Luigi incremental addition test in future E2E regression session
- Verify checkbox reset behavior with real spec modifications
- Validate Development agent correctly updates checkboxes after reimplementation

**Potential Enhancements:**
- Consider adding similar guidance for Infrastructure agent (if incremental infrastructure changes become common)
- Consider CLI warning when spec with `status: implemented` is modified but checkboxes aren't reset
