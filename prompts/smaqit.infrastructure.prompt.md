---
name: smaqit.infrastructure
description: Create infrastructure layer specifications from deployment requirements
agent: smaqit.infrastructure
tools: ["read", "edit", "search"]
---

# Infrastructure Specification Prompt

You are creating infrastructure layer specifications for the smaqit framework. This layer defines **where** the system runs—the compute resources, networking, observability, and operational concerns for production deployment.

## Framework Context

Review these framework files to understand the Infrastructure layer requirements:
- [Core Principles](../framework/SMAQIT.md)
- [Infrastructure Layer Definition](../framework/LAYERS.md#infrastructure--where)
- [Template Structure](../framework/TEMPLATES.md)
- [Artifact Rules](../framework/ARTIFACTS.md)

## User Input

Collect deployment and operational requirements:

${input:infrastructureRequirements:Describe target environment, hosting platform, service topology, resource constraints, scaling needs, geographic constraints, and budget limits}

## Instructions

Once you have collected the infrastructure requirements:

1. Read all Phase 1 specifications (business, functional, stack) for context and runtime constraints
2. Read the infrastructure specification template from `../templates/specs/infrastructure.template.md`
3. Invoke the Infrastructure Agent (smaqit.infrastructure) to transform the requirements into structured specifications
4. The agent will create specification files in `specs/infrastructure/` following the template
5. Each specification MUST include:
   - References to Phase 1 specs for traceability
   - Compute resources and service topology
   - Networking topology and security boundaries
   - Observability setup (logging, metrics, tracing)
   - Scaling policies and resource limits
   - Secrets management approach
   - Acceptance criteria with IDs in format `INF-[CONCEPT]-[NNN]`

## Success Criteria

Infrastructure specifications are complete when:
- All compute, networking, and observability concerns are defined
- Infrastructure is consistent with stack runtime requirements
- Scaling and resource policies are clearly specified
- All acceptance criteria have unique IDs and are testable
- No business logic or functional behaviors are redefined (those belong in earlier layers)
