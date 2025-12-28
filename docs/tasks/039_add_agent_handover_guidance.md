# Add Agent Handover Guidance

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #8 (2025-12-27)

## Description

When specification agents complete their task, they don't provide guidance on next steps in the workflow. Add handover section to agent templates and implementations to guide users through the workflow.

## Acceptance Criteria

- [ ] Agent templates (Level 1) include handover section with placeholder `[PROPOSE_NEXT_STEP]`
- [ ] Specification agent template updated with handover guidance structure
- [ ] All 5 specification agents (Level 2) populated with workflow-aware guidance:
  - Business → "Next: Create functional specifications with `/smaqit.functional`"
  - Functional → "Next: Create stack specifications with `/smaqit.stack`"
  - Stack → "Next: Create infrastructure specifications with `/smaqit.infrastructure` OR run development phase with `/smaqit.development`"
  - Infrastructure → "Next: Create coverage specifications with `/smaqit.coverage` OR run deployment phase with `/smaqit.deployment`"
  - Coverage → "Next: Run validation phase with `/smaqit.validation`"
- [ ] Handover messages are contextual (provide options when workflow branches)
- [ ] Implementation agent templates also include handover guidance

## Impact

**Severity:** Medium  
**User Impact:** Users must infer next steps; reduces workflow clarity; misses opportunity for just-in-time guidance

## Notes

UX improvement. Agents should guide users through the workflow, not leave them guessing.
