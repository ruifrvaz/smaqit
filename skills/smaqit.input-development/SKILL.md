---
name: smaqit.input-development
description: Validate and elicit development phase parameters from session context before implementation begins. Invoke automatically before starting the Development phase to confirm or default execution preferences.
metadata:
  version: "1.0.0"
---

# Development Input

Validate session context for development execution preferences before the Development phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Scan for any explicit build preferences, environment constraints, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Elicit only when blocking** — Ask only if a parameter is referenced but missing, or if its absence would cause the wrong behavior
4. **Proceed** — Begin implementation with confirmed or defaulted parameters

## Parameters

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
