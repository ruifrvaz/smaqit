---
name: smaqit.input-validation
description: Validate and elicit validation phase parameters from session context before test execution begins. Invoke automatically before starting the Validation phase to confirm or default execution preferences.
metadata:
  version: "1.0.0"
---

# Validation Input

Validate session context for validation execution preferences before the Validation phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would produce the wrong test behavior.

## Steps

1. **Extract from session context** — Scan for any explicit test scope overrides, failure handling preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Elicit only when blocking** — Ask only if a parameter is referenced but missing, or if test scope cannot be determined from Coverage specs
4. **Proceed** — Begin validation with confirmed or defaulted parameters

## Parameters

### Execution Scope
**Default:** Full test suite as defined in Coverage specs  
**Elicit when:** User has indicated partial execution but hasn't specified which subset  
**Question:** "Which tests should be executed? (full suite, smoke tests only, specific test group)"

### Failure Handling
**Default:** Continue on failures; produce full report  
**Elicit when:** User has indicated fast-feedback mode but hasn't specified it  
**Question:** "How should test failures be handled? (continue for full report, stop on first failure)"

### Output Preferences
**Default:** Summary with failure details; verbose output on failure  
**Elicit when:** Debugging context where full verbose output is needed throughout  
**Question:** "How should results be displayed? (summary, verbose, failures only)"

## Readiness Condition

No required sections — defaults are always valid. Proceed immediately unless test scope cannot be determined from Coverage specs.
