---
name: smaqit.input-validation
description: Validate and elicit validation phase parameters from session context before test execution begins. Invoke automatically before starting the Validation phase to confirm or default execution preferences.
metadata:
  version: "1.1.0"
---

# Validation Input

Validate session context for validation execution preferences before the Validation phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would produce the wrong test behavior.

## Steps

1. **Extract from session context** — Take any parameter the caller already passed as resolved and never ask about it again. Scan for any explicit test scope overrides, failure handling preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Resolve blocking parameters** — A parameter is blocking only if it is referenced but missing, or if test scope cannot be determined from Coverage specs. If you can ask the user directly, ask. Otherwise return `NEEDS-INPUT` (see below), listing every blocking parameter at once
4. **Output the parameter block** — Emit the Phase Parameters block (see Output). A caller passes it to the phase agent unchanged. A phase agent that applied this skill itself, because no block was passed, uses it directly

## Parameters

### Mode
**Default:** `autonomous`  
**Elicit when:** The request mentions reviewing or approving specifications before test execution but does not make clear whether every layer should pause for review  
**Question:** "Should each specification layer pause for your review and approval before test execution continues (assisted), or run straight through (autonomous)?"

A request that clearly asks to review each specification resolves to `assisted` without asking.

### Specification Mode
**Default:** `orchestrate`. The value is `prevalidated`, with exact `spec_paths`, only when the caller passes it (e.g. `smaqit.feature-new`, or `smaqit.new-greenfield-project` after its own specification phase)  
**Elicit when:** Never — the caller determines it

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

## Output

```
PHASE-PARAMETERS
phase: validate
mode: autonomous
specification_mode: orchestrate
spec_paths: []
approved_spec_paths: []
execution_scope: full suite
failure_handling: continue
output: summary
```

The caller passes this block to the `smaqit.validation` agent as its Execution Parameters. A caller adds `approved_spec_paths` and `answers` when it resumes or re-invokes the agent after a returned signal.

## NEEDS-INPUT Return

When step 3 cannot reach the user, end with this block instead of the Output block:

```
NEEDS-INPUT
source: smaqit.input-validation
phase: validate
written: nothing
questions:
  - id: <parameter name, e.g. mode>
    question: "<the parameter's question above>"
    why: "<what makes it blocking>"
next: Ask the user these questions, then apply this skill again (or resume the agent) with the answers as `answers`, keyed by id.
```
