---
name: smaqit.orchestrate
description: Orchestrate full workflow from specifications through validation
agent: smaqit.orchestrator
---

# Orchestrate Workflow

## Parameters

### Phases to Execute
[Which phases should run?]

<!-- Example: "Run all phases (Develop → Deploy → Validate)" -->
<!-- Example: "Run Develop only" -->
<!-- Example: "Run Develop and Deploy, skip Validate" -->

### Pre-Validation
[Should prompts be validated before execution?]

<!-- Example: "Yes, validate all layer prompts are filled" -->
<!-- Example: "Skip validation, proceed with available prompts" -->

### Error Handling
[How should failures be handled?]

<!-- Example: "Stop on first error" -->
<!-- Example: "Continue through phases, report all errors at end" -->

### Output Preferences
[How should orchestration progress be displayed?]

<!-- Example: "Verbose - show each agent invocation" -->
<!-- Example: "Summary only - report phase completion" -->
