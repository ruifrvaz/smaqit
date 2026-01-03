---
name: smaqit.orchestrator
description: [ORCHESTRATOR_DESCRIPTION]
tools: ['edit', 'search', 'runCommands', 'problems', 'changes', 'testFailure', 'todos', 'runSubagent', 'runTests']
---

# Orchestrator Agent

## Role

[ORCHESTRATOR_ROLE]

## Input

**User Input:**
- [ORCHESTRATOR_PROMPT_PATH]
- [REQUIRED_PROMPTS]

**Conflict Resolution:**
When orchestration parameters conflict with agent requirements, flag the conflict rather than silently override.

## Output

**Artifacts:**
- [OUTPUT_ARTIFACTS]

**Format:**
- [OUTPUT_FORMAT_RULES]

## Directives

### MUST

[ORCHESTRATOR_MUST_RULES]

### MUST NOT

[ORCHESTRATOR_MUST_NOT_RULES]

### SHOULD

[ORCHESTRATOR_SHOULD_RULES]

## Orchestration Workflow

### Pre-Run Validation

[PRE_RUN_VALIDATION_STEPS]

### Agent Invocation Sequence

[AGENT_INVOCATION_SEQUENCE]

### Error Handling

[ERROR_HANDLING_STRATEGIES]

### Completion Criteria

Before declaring workflow complete, verify:

[COMPLETION_CHECKLIST]

## Failure Handling

| Situation | Action |
|-----------|--------|
[FAILURE_SCENARIOS]

Stop iterating when:
[STOP_CONDITIONS]
