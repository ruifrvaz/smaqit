# Add Agent Handover Guidance

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #8 (2025-12-27)

## Description

When specification agents complete their task, they don't provide guidance on next steps in the workflow. Add handover section to agent templates and implementations to guide users through the workflow.

## Acceptance Criteria

- [x] Agent templates (Level 1) include handover section with placeholder `[PROPOSE_NEXT_STEP]`
- [x] Specification agent template updated with handover guidance structure
- [x] All 5 specification agents (Level 2) populated with workflow-aware guidance:
  - Business → "Next: Create functional specifications with `/smaqit.functional`"
  - Functional → "Next: Create stack specifications with `/smaqit.stack`"
  - Stack → "Next: Create infrastructure specifications with `/smaqit.infrastructure` OR run development phase with `/smaqit.development`"
  - Infrastructure → "Next: Create coverage specifications with `/smaqit.coverage` OR run deployment phase with `/smaqit.deployment`"
  - Coverage → "Next: Run validation phase with `/smaqit.validation`"
- [x] Handover messages are contextual (provide options when workflow branches)
- [x] Implementation agent templates also include handover guidance

## Implementation Details

### Level 1 Changes (Templates)

**Files Modified:**
- `templates/agents/specification-agent.template.md` — Added "Workflow Handover" section with `[PROPOSE_NEXT_STEP]` placeholder
- `templates/agents/implementation-agent.template.md` — Added "Workflow Handover" section with `[PROPOSE_NEXT_STEP]` placeholder

### Level 2 Changes (Agents)

**Specification Agents:**

1. **Business Agent** (`agents/smaqit.business.agent.md`)
   - Handover: "Next: Create functional specifications with `/smaqit.functional`"
   - Explains Functional layer purpose

2. **Functional Agent** (`agents/smaqit.functional.agent.md`)
   - Handover: "Next: Create stack specifications with `/smaqit.stack`"
   - Explains Stack layer purpose

3. **Stack Agent** (`agents/smaqit.stack.agent.md`)
   - Handover: Two options with recommendations
   - Option 1 (Recommended): Run Development phase (`/smaqit.development`)
   - Option 2: Continue to Infrastructure layer (`/smaqit.infrastructure`)
   - Explains phase-first workflow preference

4. **Infrastructure Agent** (`agents/smaqit.infrastructure.agent.md`)
   - Handover: Two options with recommendations
   - Option 1 (Recommended): Run Deployment phase (`/smaqit.deployment`)
   - Option 2: Continue to Coverage layer (`/smaqit.coverage`)
   - Explains phase-first workflow preference

5. **Coverage Agent** (`agents/smaqit.coverage.agent.md`)
   - Handover: "Next: Run validation phase with `/smaqit.validation`"
   - Explains Validation phase purpose

**Implementation Agents:**

6. **Development Agent** (`agents/smaqit.development.agent.md`)
   - Handover: "Next: Create infrastructure specifications with `/smaqit.infrastructure`"
   - Confirms Phase 1 complete, guides to Phase 2 start

7. **Deployment Agent** (`agents/smaqit.deployment.agent.md`)
   - Handover: "Next: Create coverage specifications with `/smaqit.coverage`"
   - Confirms Phase 2 complete, guides to Phase 3 start

8. **Validation Agent** (`agents/smaqit.validation.agent.md`)
   - Handover: "Validation Complete: The smaqit workflow cycle is complete!"
   - Provides guidance for different outcomes (all pass, some fail, low coverage)
   - Explains how to iterate on requirements

### Branching Points

Two workflow branching points with contextual guidance:

1. **After Stack specs:** Users can complete Phase 1 (recommended) OR continue spec-first
2. **After Infrastructure specs:** Users can complete Phase 2 (recommended) OR continue spec-first

Both branching points recommend phase-first workflow while acknowledging spec-first option.

### Validation

- Installer builds successfully with embedded handover guidance
- Test project initialization confirmed working
- Handover sections verified in installed agents
- Branching guidance verified in Stack and Infrastructure agents
- Completion message verified in Validation agent

## Impact

**Severity:** Medium  
**User Impact:** Users must infer next steps; reduces workflow clarity; misses opportunity for just-in-time guidance

**Resolution:** Users now receive clear, contextual guidance at every workflow transition point, with recommendations for phase-first workflow while acknowledging spec-first alternative.

## Notes

UX improvement. Agents now guide users through the workflow with just-in-time next-step instructions. Handover messages are contextual and provide options at branching points (Stack → Development OR Infrastructure, Infrastructure → Deployment OR Coverage).
