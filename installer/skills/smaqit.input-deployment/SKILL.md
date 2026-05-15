---
name: smaqit.input-deployment
description: Validate and elicit deployment phase parameters from session context before deployment begins. Invoke automatically before starting the Deployment phase to confirm or default execution preferences.
metadata:
  version: "1.0.0"
---

# Deployment Input

Validate session context for deployment execution preferences before the Deployment phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Scan for any explicit deployment target overrides, verification preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Elicit only when blocking** — Ask only if a parameter is referenced but missing, or if the deployment target cannot be resolved from Infrastructure specs
4. **Proceed** — Begin deployment with confirmed or defaulted parameters

## Parameters

### Deployment Target
**Default:** Target environment as defined in Infrastructure specs  
**Elicit when:** Multiple deployment targets exist and the active target is ambiguous  
**Question:** "Any target-specific preferences? (e.g., deploy to staging first, use blue-green strategy)"

### Verification
**Default:** Run standard post-deployment health checks as defined in Coverage specs  
**Elicit when:** User has indicated they want to skip or extend verification  
**Question:** "How should deployment be verified? (run smoke tests, skip health checks, full suite)"

### Output Preferences
**Default:** Standard output; deployment steps and results shown  
**Elicit when:** Sensitive environment where log scrubbing is needed  
**Question:** "How should output be displayed? (verbose, scrub sensitive data from logs)"

## Readiness Condition

No required sections — defaults are always valid. Proceed immediately unless the deployment target cannot be determined from Infrastructure specs.
