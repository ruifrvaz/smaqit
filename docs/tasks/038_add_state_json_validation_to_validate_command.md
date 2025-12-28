# Add state.json Validation to Validate Command

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #7 (2025-12-27)

## Description

`smaqit validate` command checks directory structure and spec files, but does not validate `.smaqit/state.json`. Add validation to catch structural issues in state file.

## Acceptance Criteria

- [ ] Validate command checks `.smaqit/state.json` file exists
- [ ] Validates JSON structure is valid (parseable)
- [ ] Verifies phase keys are present (develop, deploy, validate)
- [ ] Verifies phase keys are correctly ordered in JSON
- [ ] Validates each phase object has required "completed" boolean field
- [ ] Validates "version" field is present at root level
- [ ] Reports specific validation errors with actionable messages
- [ ] Does not fail on additional fields (forward compatibility)

## Impact

**Severity:** Medium  
**User Impact:** Corrupted or malformed state.json can cause status command failures; validate should catch structural issues

## Notes

Defensive validation. Helps users catch state file corruption early.
