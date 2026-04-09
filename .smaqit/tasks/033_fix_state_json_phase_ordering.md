# Fix state.json Phase Ordering

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2026-01-02  
**Source:** User Testing Report Issue #2 (2025-12-27)

## Description

Fix phase order in `.smaqit/state.json` initialization. Currently generates phases in order: deploy, develop, validate. Should be: develop, deploy, validate (matching workflow sequence).

## Acceptance Criteria

- [x] `initStateFile()` function in `installer/main.go` generates phases in correct order
- [x] JSON output has phases ordered: develop, deploy, validate
- [x] Existing projects with old ordering remain functional (backwards compatibility)
- [x] New installations generate correct ordering

## Solution

Changed `StateFile` structure from using `map[string]PhaseState` (unordered) to using a struct `Phases` with ordered fields:

```go
type Phases struct {
    Develop  PhaseState `json:"develop"`
    Deploy   PhaseState `json:"deploy"`
    Validate PhaseState `json:"validate"`
}
```

JSON struct field order is guaranteed by Go's encoding/json package.

## Files Modified

- `installer/main.go`:
  - Added `Phases` struct with ordered fields
  - Updated `StateFile` to use `Phases` instead of map
  - Updated `initStateFile()` to initialize struct fields
  - Removed map access validation (struct fields always exist)
  - Updated `cmdStatus()` to access struct fields directly

## Testing

Verified phase ordering in generated state.json:
```json
{
  "version": "1.0",
  "phases": {
    "develop": {"completed": false},
    "deploy": {"completed": false},
    "validate": {"completed": false}
  }
}
```
