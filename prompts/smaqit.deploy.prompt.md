---
name: smaqit.deploy
description: Run the deployment phase - create infrastructure specs, then deploy to target environment
---

# Deploy Phase Prompt

This phase orchestrates Infrastructure specification agent, then invokes the Deployment implementation agent to deploy the application.

## Pre-Run Validation

Before starting, verify `.github/prompts/smaqit.infrastructure.prompt.md` has content beyond template structure.

**If prompt is empty:** Halt and provide natural language guidance about target environment, hosting platform, and service topology requirements.

## Orchestration Workflow

### Step 1: Infrastructure Specifications

Invoke `@smaqit.infrastructure` agent to produce infrastructure layer specifications in `specs/infrastructure/`.

Agent reads all Phase 1 specifications (business, functional, stack) for context and runtime constraints.

**On failure:** Stop and report error. Infrastructure specs required before deployment.

### Step 2: Deploy Application

Invoke `@smaqit.deployment` agent to deploy application to target environment.

Agent performs:
1. Consolidate infrastructure and stack specifications (coherence check)
2. Generate Infrastructure as Code configurations (reference-only secrets)
3. Trigger trusted execution layer with target environment parameter
4. Receive deployment outcome (success/failure, health status, endpoints)
5. Verify system health in target environment
6. Report deployment status

**On failure:** Deployment agent iterates up to retry threshold, then reports blocking issue with scrubbed output (no sensitive data).

## Completion Criteria

Phase complete when:
- [ ] Infrastructure specifications produced
- [ ] Infrastructure as Code generated with reference-only secrets
- [ ] Deployment executes successfully in target environment
- [ ] Health checks pass
- [ ] System accessible at expected endpoints
- [ ] Deployment report documents running system status
