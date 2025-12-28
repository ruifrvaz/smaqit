# Nest Layers Under Phases in Status Display

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #4 (2025-12-27)

## Description

`smaqit status` currently shows "Specification Layers" and "Phase Status" as separate sections. Restructure to nest layers under their respective phases for better progress tracking.

## Acceptance Criteria

- [ ] Status display shows layers nested under phases:
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
- [ ] Phase status (Not started/In progress/Completed) shown per phase
- [ ] Total spec count still displayed
- [ ] Maintains clear visual hierarchy

## Impact

**Severity:** Low  
**User Impact:** Reduces clarity of progress tracking; users must mentally map layers to phases

## Notes

UX improvement. Makes phase-first workflow more visible in CLI output.
