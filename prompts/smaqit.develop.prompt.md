---
name: smaqit.develop
description: Run the development phase - create business, functional, and stack specs, then build the application
tools: ["agent", "read", "edit", "search", "execute", "todo"]
---

# Develop Phase Prompt

This phase orchestrates Business, Functional, and Stack specification agents, then invokes the Development implementation agent to build the application.

## Pre-Run Validation

Before starting, verify the following prompts have content beyond template structure:
- `.github/prompts/smaqit.business.prompt.md`
- `.github/prompts/smaqit.functional.prompt.md`
- `.github/prompts/smaqit.stack.prompt.md`

**If any prompt is empty:** Halt and provide natural language guidance about what's needed for that layer.

## Orchestration Workflow

### Step 1: Business Specifications

Invoke `@smaqit.business` agent to produce business layer specifications in `specs/business/`.

**On failure:** Stop and report error. Business specs required before proceeding.

### Step 2: Functional Specifications

Invoke `@smaqit.functional` agent to produce functional layer specifications in `specs/functional/`.

Agent reads business specs for context and traceability.

**On failure:** Stop and report error. Functional specs required before proceeding.

### Step 3: Stack Specifications

Invoke `@smaqit.stack` agent to produce stack layer specifications in `specs/stack/`.

Agent reads business and functional specs for context.

**On failure:** Stop and report error. Stack specs required before building.

### Step 4: Build Application

Invoke `@smaqit.development` agent to build the application.

Agent performs:
1. Consolidate all three specification layers (coherence check)
2. Generate application code from specifications
3. Generate unit tests
4. Build/compile the application
5. Run application in isolated environment
6. Execute unit tests
7. Verify application matches spec acceptance criteria
8. Produce development report

**On failure:** Development agent iterates up to retry threshold, then reports blocking issue.

## Completion Criteria

Phase complete when:
- [ ] Business, Functional, and Stack specifications produced
- [ ] Application code generated and compiles
- [ ] Unit tests pass
- [ ] Application runs successfully in isolated environment
- [ ] All spec acceptance criteria satisfied
- [ ] Development report documents build/test/run results
