---
name: smaqit.input-validation
description: Validate and elicit validation phase parameters from session context before test execution begins. Invoke automatically before starting the Validation phase to confirm or default execution preferences.
metadata:
  version: "1.2.0"
---

# Validation Input

Validate session context for validation execution preferences before the Validation phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would produce the wrong test behavior.

## Steps

1. **Extract from session context** — Take any parameter the caller already passed as resolved and never ask about it again. Record the user's request verbatim (see User Request) and what started this phase (see Mode). Scan for any explicit test scope overrides, failure handling preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Resolve blocking parameters** — A parameter is blocking only if it is referenced but missing, or if test scope cannot be determined from Coverage specs. If you can ask the user directly, ask. Otherwise return `NEEDS-INPUT` (see below), listing every blocking parameter at once
4. **Output the parameter block** — Emit the Phase Parameters block (see Output). A caller passes it to the phase agent unchanged. A phase agent that applied this skill itself, because no block was passed, uses it directly

## Parameters

### Mode
**Default:** Set by what started the phase:
- `assisted` — the user started it: they typed `/smaqit.validation`, or picked or named the `smaqit.validation` agent. This is also the default when the phase agent applies this skill itself because no block was passed.
- `autonomous` — it started automatically: the model routed a request here on its own.
- An orchestrating skill (`smaqit.project-new`, `smaqit.feature-new`) passes the `mode` it resolved for its own run: `assisted` when the user invoked the skill (its slash command, a Codex `$` mention, or a skill-menu pick), `autonomous` when the model chose it. Use the passed mode.

**Override:** A mode the user states always wins. "Assisted" or "review the coverage specification" resolves to `assisted`; "autonomous", "run straight through", or "no review pauses" resolves to `autonomous`.
**Elicit when:** Never — what started the phase, or the user's own words, decides it

Assisted works only from requirements the user stated, pauses with `REVIEW-REQUESTED` after the Coverage specification, and ends the phase by asking the user to review the results. Autonomous runs straight through on a best-effort basis and stops only for `NEEDS-INPUT`.

### User Request
**Value:** The user's own words that started this phase, verbatim — never a summary, paraphrase, or expansion. Empty when the user gave none (for example a bare `/smaqit.validation`). An orchestrating skill passes the request the user gave it
**Elicit when:** Never. An empty request is not blocking here: the Coverage input gate decides whether the requirements are sufficient

### Specification Mode
**Default:** `orchestrate`. The value is `prevalidated`, with exact `spec_paths`, only when the caller passes it (e.g. `smaqit.feature-new`, or `smaqit.project-new` after its own specification phase)  
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
mode: assisted
user_request: "<the user's words, verbatim; empty if none>"
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
  - id: <parameter name, e.g. execution_scope>
    question: "<the parameter's question above>"
    why: "<what makes it blocking>"
next: Ask the user these questions, then apply this skill again (or resume the agent) with the answers as `answers`, keyed by id.
```
