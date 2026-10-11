---
name: smaqit.input-development
description: Validate and elicit development phase parameters from session context before implementation begins. Invoke automatically before starting the Development phase to confirm or default execution preferences.
metadata:
  version: "1.1.0"
---

# Development Input

Validate session context for development execution preferences before the Development phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Take any parameter the caller already passed as resolved and never ask about it again. Scan for any explicit build preferences, environment constraints, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Resolve blocking parameters** — A parameter is blocking only if it is referenced but missing, or if its absence would cause the wrong behavior. If you can ask the user directly, ask. Otherwise return `NEEDS-INPUT` (see below), listing every blocking parameter at once
4. **Output the parameter block** — Emit the Phase Parameters block (see Output). A caller passes it to the phase agent unchanged. A phase agent that applied this skill itself, because no block was passed, uses it directly

## Parameters

### Mode
**Default:** `autonomous`  
**Elicit when:** The request mentions reviewing or approving specifications before implementation but does not make clear whether every layer should pause for review  
**Question:** "Should each specification layer pause for your review and approval before implementation continues (assisted), or run straight through (autonomous)?"

A request that clearly asks to review each specification resolves to `assisted` without asking.

### Specification Mode
**Default:** `orchestrate`. The value is `prevalidated`, with exact `spec_paths`, only when the caller passes it (e.g. `smaqit.feature-new`, or `smaqit.new-greenfield-project` after its own specification phase)  
**Elicit when:** Never — the caller determines it

### Build Options
**Default:** Standard build for the target platform; all tests run; no watch mode  
**Elicit when:** User has indicated a non-standard build mode but hasn't specified it  
**Question:** "Any build-time preferences? (e.g., skip tests for fast iteration, enable watch mode, production-optimized build)"

### Output Preferences
**Default:** Standard output; errors and warnings shown  
**Elicit when:** Prior context suggests sensitivity around log verbosity or secret scrubbing  
**Question:** "How should output be displayed? (verbose, quiet/errors-only, scrub sensitive data)"

### Environment
**Default:** Local development environment as defined in Stack specs  
**Elicit when:** Multiple environments exist and the target is ambiguous  
**Question:** "Any environment-specific settings? (Docker isolation, specific runtime version, target environment)"

## Readiness Condition

No required sections — defaults are always valid. Proceed immediately unless a parameter is ambiguous and would block correct execution.

## Output

```
PHASE-PARAMETERS
phase: develop
mode: autonomous
specification_mode: orchestrate
spec_paths: []
approved_spec_paths: []
build_options: standard
output: standard
environment: local (per Stack specs)
```

The caller passes this block to the `smaqit.development` agent as its Execution Parameters. A caller adds `approved_spec_paths` and `answers` when it resumes or re-invokes the agent after a returned signal.

## NEEDS-INPUT Return

When step 3 cannot reach the user, end with this block instead of the Output block:

```
NEEDS-INPUT
source: smaqit.input-development
phase: develop
written: nothing
questions:
  - id: <parameter name, e.g. mode>
    question: "<the parameter's question above>"
    why: "<what makes it blocking>"
next: Ask the user these questions, then apply this skill again (or resume the agent) with the answers as `answers`, keyed by id.
```
