---
name: smaqit.develop
description: Run the development phase - create business, functional, and stack specs, then build the application
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Development Phase Prompt

You are orchestrating the **Develop phase** of the smaqit framework. This phase transforms user requirements into a working, tested application by producing specifications across three layers (Business → Functional → Stack), then building the application.

## Framework Context

Review these framework files to understand the Develop phase workflow:
- [Core Principles](../framework/SMAQIT.md)
- [Develop Phase Definition](../framework/PHASES.md#develop--build-a-working-application)
- [Layer Definitions](../framework/LAYERS.md)

## User Input

Collect requirements for all three specification layers:

${input:businessRequirements:Describe use cases, actors, business goals, and success metrics}

${input:functionalRequirements:Describe user experience, behaviors, interactions, and data models}

${input:stackRequirements:Specify preferred languages, frameworks, libraries, or technology constraints}

## Orchestration Workflow

Once all inputs are collected, execute the following sequence:

### Step 1: Business Specifications

Run subagent **smaqit.business** with the business requirements to produce business layer specifications in `specs/business/`.

**On failure:** Stop and report the error. Business specs are required before proceeding.

### Step 2: Functional Specifications

Run subagent **smaqit.functional** with the functional requirements to produce functional layer specifications in `specs/functional/`.

The agent will read business specs for context and traceability.

**On failure:** Stop and report the error. Functional specs are required before proceeding.

### Step 3: Stack Specifications

Run subagent **smaqit.stack** with the stack requirements to produce stack layer specifications in `specs/stack/`.

The agent will read business and functional specs for context.

**On failure:** Stop and report the error. Stack specs are required before building.

### Step 4: Build Application

Run subagent **smaqit.development** to build the application.

The development agent will:
1. Consolidate all three specification layers (coherence check)
2. Generate application code from specifications
3. Generate unit tests
4. Build/compile the application
5. Run the application in an isolated environment
6. Execute unit tests
7. Verify the application matches spec acceptance criteria
8. Produce a development report

**On failure:** The development agent will iterate up to its retry threshold. If it fails repeatedly, it will report the blocking issue.

## Success Criteria

The Develop phase is complete when:
- Business, Functional, and Stack specifications are produced and complete
- Application code is generated and compiles without errors
- Unit tests pass
- Application runs successfully in an isolated environment
- All spec acceptance criteria are satisfied
- Development report documents build/test/run results
