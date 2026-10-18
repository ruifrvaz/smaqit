---
name: smaqit.input-development
description: Validate and elicit development phase parameters from session context before implementation begins. Invoke automatically before starting the Development phase to confirm or default execution preferences.
metadata:
  version: "1.2.0"
---

# Development Input

Validate session context for development execution preferences before the Development phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Take any parameter the caller already passed as resolved and never ask about it again. Record the user's request verbatim (see User Request) and what started this phase (see Mode). Scan for any explicit build preferences, environment constraints, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Resolve blocking parameters** — A parameter is blocking only if it is referenced but missing, or if its absence would cause the wrong behavior. If you can ask the user directly, ask. Otherwise return `NEEDS-INPUT` (see below), listing every blocking parameter at once
4. **Output the parameter block** — Emit the Phase Parameters block (see Output). A caller passes it to the phase agent unchanged. A phase agent that applied this skill itself, because no block was passed, uses it directly

## Parameters

### Mode
**Default:** Set by what started the phase:
- `assisted` — the user started it: they typed `/smaqit.development`, or picked or named the `smaqit.development` agent. This is also the default when the phase agent applies this skill itself because no block was passed.
- `autonomous` — it started automatically: the model routed a request here on its own.
- An orchestrating skill (`smaqit.project-new`, `smaqit.feature-new`) passes the `mode` it resolved for its own run: `assisted` when the user invoked the skill (its slash command, a Codex `$` mention, or a skill-menu pick), `autonomous` when the model chose it. Use the passed mode.

**Override:** A mode the user states always wins. "Assisted" or "review each specification" resolves to `assisted`; "autonomous", "run straight through", or "no review pauses" resolves to `autonomous`.
**Elicit when:** Never — what started the phase, or the user's own words, decides it

Assisted works only from requirements the user stated, pauses with `REVIEW-REQUESTED` after each specification layer, and ends the phase by asking the user to review it before moving on. Autonomous runs straight through on a best-effort basis and stops only for `NEEDS-INPUT`.

### User Request
**Value:** The user's own words that started this phase, verbatim — never a summary, paraphrase, or expansion. Empty when the user gave none (for example a bare `/smaqit.development`). An orchestrating skill passes the request the user gave it
**Elicit when:** Never. An empty request is not blocking here: each specification layer's input gate decides whether the requirements are sufficient

### Specification Mode
**Default:** `orchestrate`. The value is `prevalidated`, with exact `spec_paths`, only when the caller passes it (e.g. `smaqit.feature-new`, or `smaqit.project-new` after its own specification phase)  
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
mode: assisted
user_request: "<the user's words, verbatim; empty if none>"
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
  - id: <parameter name, e.g. environment>
    question: "<the parameter's question above>"
    why: "<what makes it blocking>"
next: Ask the user these questions, then apply this skill again (or resume the agent) with the answers as `answers`, keyed by id.
```
