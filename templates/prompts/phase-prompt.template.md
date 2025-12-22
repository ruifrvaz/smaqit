# [PHASE_NAME] Prompt Template

Use this template to create phase orchestration prompts (develop, deploy, validate).

## Structure

```markdown
---
name: smaqit.[PHASE]
description: Run the [PHASE] phase - [brief workflow description]
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# [PHASE_NAME] Phase Prompt

This prompt orchestrates the **[PHASE_NAME] phase** workflow, coordinating multiple agents to [phase purpose].

## Orchestration Workflow

### Pre-Run Validation

Before starting, validate that all required prompts are filled:

- [ ] Check [upstream prompt 1] has content
- [ ] Check [upstream prompt 2] has content
- [ ] Check [upstream prompt 3] has content

If any prompts are empty or insufficient, halt and guide user to fill missing inputs.

### Execution Steps

[Sequential agent invocations with error handling]

### Completion Criteria

Phase is complete when:
- [ ] [Criterion 1]
- [ ] [Criterion 2]
- [ ] [Criterion 3]
```

## Phase-Specific Patterns

### Develop Phase

```markdown
---
name: smaqit.develop
description: Run the development phase - create business, functional, and stack specs, then build the application
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Development Phase Prompt

This prompt orchestrates the **Develop phase** workflow. It validates inputs, coordinates specification generation across three layers (Business → Functional → Stack), then builds the application.

## Pre-Run Validation

Validate all required prompts are filled:

- [ ] `.github/prompts/smaqit.business.prompt.md` has requirements
- [ ] `.github/prompts/smaqit.functional.prompt.md` has requirements
- [ ] `.github/prompts/smaqit.stack.prompt.md` has requirements

If any prompts are empty, halt and guide user: "Please fill [prompt file] with your [layer] requirements before starting development."

## Orchestration Workflow

### Step 1: Business Specifications

Invoke `@smaqit.business` agent:
- Agent reads `.github/prompts/smaqit.business.prompt.md`
- Agent generates specs in `specs/business/`
- **On failure:** Stop and report error (Business specs required)

### Step 2: Functional Specifications

Invoke `@smaqit.functional` agent:
- Agent reads `.github/prompts/smaqit.functional.prompt.md`
- Agent references business specs for context
- Agent generates specs in `specs/functional/`
- **On failure:** Stop and report error (Functional specs required)

### Step 3: Stack Specifications

Invoke `@smaqit.stack` agent:
- Agent reads `.github/prompts/smaqit.stack.prompt.md`
- Agent references business and functional specs for context
- Agent generates specs in `specs/stack/`
- **On failure:** Stop and report error (Stack specs required)

### Step 4: Build Application

Invoke `@smaqit.development` agent:
- Agent consolidates all Phase 1 specs
- Agent generates code and tests
- Agent builds and runs application
- **On failure:** Iterate on failures up to retry threshold

## Completion Criteria

Development phase is complete when:
- [ ] All layer specs produced (business, functional, stack)
- [ ] Code generated and compiles without errors
- [ ] Tests pass
- [ ] Application runs successfully in isolated environment
```

### Deploy Phase

```markdown
---
name: smaqit.deploy
description: Run the deployment phase - create infrastructure spec, then deploy to target environment
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Deployment Phase Prompt

This prompt orchestrates the **Deploy phase** workflow. It validates inputs, generates infrastructure specifications, then deploys the application to the target environment.

## Pre-Run Validation

Validate all required prompts are filled:

- [ ] `.github/prompts/smaqit.infrastructure.prompt.md` has deployment requirements

If prompt is empty, halt and guide user: "Please fill smaqit.infrastructure.prompt.md with your deployment requirements."

## Orchestration Workflow

### Step 1: Infrastructure Specifications

Invoke `@smaqit.infrastructure` agent:
- Agent reads `.github/prompts/smaqit.infrastructure.prompt.md`
- Agent references Phase 1 specs (business, functional, stack) for context
- Agent generates specs in `specs/infrastructure/`
- **On failure:** Stop and report error (Infrastructure specs required)

### Step 2: Deploy Application

Invoke `@smaqit.deployment` agent:
- Agent consolidates infrastructure + stack specs
- Agent generates Infrastructure as Code (configurations as references only)
- Agent triggers deployment to target environment
- Agent verifies system health
- **On failure:** Iterate on failures up to retry threshold

## Completion Criteria

Deployment phase is complete when:
- [ ] Infrastructure specs produced
- [ ] IaC generated with reference-only secrets
- [ ] Deployment executed successfully
- [ ] Health checks pass
- [ ] System accessible at expected endpoints
```

### Validate Phase

```markdown
---
name: smaqit.validate
description: Run the validation phase - create coverage spec, then verify deployed system
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Validation Phase Prompt

This prompt orchestrates the **Validate phase** workflow. It validates inputs, generates coverage specifications, then executes tests against the deployed system.

## Pre-Run Validation

Validate all required prompts are filled:

- [ ] `.github/prompts/smaqit.coverage.prompt.md` has verification requirements

If prompt is empty, halt and guide user: "Please fill smaqit.coverage.prompt.md with your test scope and verification requirements."

## Orchestration Workflow

### Step 1: Coverage Specifications

Invoke `@smaqit.coverage` agent:
- Agent reads `.github/prompts/smaqit.coverage.prompt.md`
- Agent references all upstream specs (business, functional, stack, infrastructure)
- Agent enumerates acceptance criteria and maps to test cases
- Agent generates specs in `specs/coverage/`
- **On failure:** Stop and report error (Coverage specs required)

### Step 2: Execute Validation

Invoke `@smaqit.validation` agent:
- Agent reads coverage specs
- Agent executes tests against deployed system
- Agent collects pass/fail results
- Agent calculates spec coverage percentage
- Agent produces validation report
- **Note:** Test failures do NOT trigger automatic retry (human decides next action)

## Completion Criteria

Validation phase is complete when:
- [ ] Coverage specs produced with all testable criteria mapped
- [ ] Tests executed against deployed system
- [ ] Validation report generated
- [ ] Spec coverage percentage calculated
- [ ] Untestable criteria documented with justification
```

## Placeholder Convention

Phase prompts typically don't have `<!-- Example: ... -->` comments since they orchestrate rather than collect requirements. However, they may include execution notes or configuration options.
