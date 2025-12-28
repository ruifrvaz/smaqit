# Fix state.json Phase Ordering

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #2 (2025-12-27)

## Description

Fix phase order in `.smaqit/state.json` initialization. Currently generates phases in order: deploy, develop, validate. Should be: develop, deploy, validate (matching workflow sequence).

## Acceptance Criteria

- [ ] `initStateFile()` function in `installer/main.go` generates phases in correct order
- [ ] JSON output has phases ordered: develop, deploy, validate
- [ ] Existing projects with old ordering remain functional (backwards compatibility)
- [ ] New installations generate correct ordering

## Impact

**Severity:** Low  
**User Impact:** JSON readability issue; does not affect functionality but violates semantic ordering

## Notes

Simple fix in installer code. Update the JSON generation to output phases in workflow order.
