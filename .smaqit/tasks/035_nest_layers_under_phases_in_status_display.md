# Nest Layers Under Phases in Status Display

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #4 (2025-12-27)

## Description

`smaqit status` currently shows "Specification Layers" and "Phase Status" as separate sections. Restructure to nest layers under their respective phases for better progress tracking.

## Acceptance Criteria

- [x] Status display shows layers nested under phases:
  ```
  Phase 1 (Develop):
    Business:        X spec(s)
    Functional:      Y spec(s)
    Stack:           Z spec(s)
  
  Phase 2 (Deploy):
    Infrastructure:  X spec(s)
  
  Phase 3 (Validate):
    Coverage:        X spec(s)
  ```
- [x] Phase status (Not started/In progress/Completed) shown per phase
- [x] Total spec count still displayed
- [x] Maintains clear visual hierarchy

## Impact

**Severity:** Low  
**User Impact:** Reduces clarity of progress tracking; users must mentally map layers to phases

## Implementation

**Files changed:**
- `installer/main.go`:
  - Added `printPhaseStatus()` helper function to format phase completion status with optional timestamp
  - Refactored `cmdStatus()` to group layers under their respective phases
  - Maintained backwards compatibility with existing state.json files
  - Preserved all existing functionality (timestamps, next steps, total count)

**Testing:**
- ✅ Empty project (0 specs)
- ✅ Multiple specs across layers
- ✅ Phase completion with timestamps
- ✅ All phases completed
- ✅ Backwards compatibility verified

## Notes

UX improvement completed. Makes phase-first workflow more visible in CLI output. The new display format aligns with the phase definitions in PHASES.md and clarifies the layer-to-phase relationships for users.
