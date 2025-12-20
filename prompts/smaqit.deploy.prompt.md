---
name: smaqit.deploy
description: Run the deployment phase - create infrastructure specs, then deploy to target environment
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Deployment Phase Prompt

You are orchestrating the **Deploy phase** of the smaqit framework. This phase transforms a working application into a running system in a target environment by producing infrastructure specifications, then deploying the application.

## Framework Context

Review these framework files to understand the Deploy phase workflow:
- [Core Principles](../framework/SMAQIT.md)
- [Deploy Phase Definition](../framework/PHASES.md#deploy--run-in-target-environment)
- [Infrastructure Layer Definition](../framework/LAYERS.md#infrastructure--where)

## User Input

Collect deployment and operational requirements:

${input:infrastructureRequirements:Describe target environment, hosting platform, service topology, resource constraints, scaling needs, geographic constraints, and budget limits}

${input:targetEnvironment:Specify the target environment identifier (e.g., dev, staging, production)}

## Orchestration Workflow

Once all inputs are collected, execute the following sequence:

### Step 1: Infrastructure Specifications

Run subagent **smaqit.infrastructure** with the infrastructure requirements to produce infrastructure layer specifications in `specs/infrastructure/`.

The agent will read all Phase 1 specifications (business, functional, stack) for context and runtime constraints.

**On failure:** Stop and report the error. Infrastructure specs are required before deployment.

### Step 2: Deploy Application

Run subagent **smaqit.deployment** to deploy the application to the target environment.

The deployment agent will:
1. Consolidate infrastructure and stack specifications (coherence check)
2. Generate Infrastructure as Code configurations (with reference-only secrets)
3. Trigger trusted execution layer with the target environment parameter
4. Receive deployment outcome (success/failure, health status, endpoints)
5. Verify system health in the target environment
6. Report deployment status

**On failure:** The deployment agent will iterate up to its retry threshold. If it fails repeatedly, it will report the blocking issue with scrubbed output (no sensitive data).

## Success Criteria

The Deploy phase is complete when:
- Infrastructure specifications are produced and complete
- Infrastructure as Code is generated with reference-only secrets
- Deployment executes successfully in the target environment
- Health checks pass
- System is accessible at expected endpoints
- Deployment report documents the running system status
