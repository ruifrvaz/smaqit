# Phases

Phases are the sequential stages of software development in smaqit. Each phase follows the same pattern: specification agents produce specs, then an implementation agent consolidates and executes. This document defines the three phases and their workflows.

## Overview

smaqit operates in three sequential phases:

| Phase | Name | Specification Artifacts | Implementation Artifacts |
|-------|------|------------------------|-------------------------|
| Phase 1 | Develop | Business, Functional, Stack | Code, README, Development report |
| Phase 2 | Deploy | Infrastructure | Running system |
| Phase 3 | Validate | Coverage | Validation report |

Each phase:
1. **Specifies** — One or more specification agents produce layer specs
2. **Consolidates** — Implementation agent verifies cross-layer coherence
3. **Implements** — Implementation agent produces and executes artifacts
4. **Verifies** — Implementation agent confirms success before phase completion

Phases are strictly sequential. Deploy cannot begin until Develop completes. Validate cannot begin until Deploy completes. This constraint is subject to revision based on real-world usage (see [SMAQIT](SMAQIT.md)).

## Phase Definitions

### Develop — Build a Working Application

The Develop phase transforms user requirements into a working, tested application running in an isolated environment.

**Specification Agents:**
| Agent | Layer | User Input | Context | Output |
|-------|-------|------------|---------|--------|
| `smaqit.business` | Business | Stakeholder goals | None | `specs/business/*.md` |
| `smaqit.functional` | Functional | Experience shape | Business specs | `specs/functional/*.md` |
| `smaqit.stack` | Stack | Technology preferences | Business and Functional specs | `specs/stack/*.md` |

**Implementation Agent:** `smaqit.development`

**Workflow:**
```
1. Business agent produces business specifications
2. Functional agent produces functional specifications
3. Stack agent produces stack specifications
4. Development agent:
   a. Consolidates specs (coherence check, conflict detection)
   b. Generates application code
   c. Generates unit tests
   d. Compiles/builds application
   e. Runs application in isolated environment
   f. Executes unit tests
   g. Verifies application works as specified
```

**Environment:** Implicit — local developer machine or agent runner (e.g., GitHub Actions runner)

**Output:** Working, tested application in isolated environment

**Failure Handling:**
- Iterate on code/test failures up to retry threshold
- Document failure reasons at each attempt
- Escalate to human review when threshold exceeded

**Completion Criteria:**
- [ ] All layer specs produced and complete
- [ ] Code generated and compiles without errors
- [ ] Unit tests pass
- [ ] Application runs successfully in isolated environment
- [ ] Behavior matches spec acceptance criteria
- [ ] README includes build, test, and run instructions
- [ ] Development report documents build/test/run results

---

### Deploy — Run in Target Environment

The Deploy phase transforms a working application into a running system in a target environment.

**Specification Agent:**
| Agent | Layer | User Input | Context | Output |
|-------|-------|------------|---------|--------|
| `smaqit.infrastructure` | Infrastructure | Deployment requirements | Phase 1 specs | `specs/infrastructure/*.md` |

**Implementation Agent:** `smaqit.deployment`

**User Input Required:**

| Category | Purpose |
|----------|----------|
| Target environment | Where the system will run |
| Hosting platform | Provider or infrastructure type |
| Service topology | How the application is structured for deployment |
| Resource constraints | Compute, memory, storage limits |
| Scaling requirements | How the system should handle load |
| Geographic constraints | Location or data residency requirements |
| Budget constraints | Cost limits or optimization goals |
| Integration points | Existing systems to connect with |

**Workflow:**
```
1. Infrastructure agent produces infrastructure specifications
2. Deployment agent:
   a. Consolidates specs (infrastructure + stack coherence)
   b. Generates Infrastructure as Code (configurations as references only, per Isolation Principle)
   c. Triggers trusted execution layer with environment parameter
   d. Receives outcome (success/failure, health status, endpoints)
   e. Verifies system health in target environment
```

**Trusted Execution Layer:**
The deployment agent operates on credential references, never values. Actual deployment happens in a trusted execution layer:

```
┌─────────────────────────────────────────────────────────────┐
│ Deployment Agent (no credentials in context)                │
│                                                             │
│  Generates: main.tf with ${secrets.AWS_ACCESS_KEY}          │
│  Calls: deploy(environment="staging")                       │
│                                                             │
│         ┌───────────────────────────────────────────┐       │
│         │ Trusted Execution Layer                   │       │
│         │ - Resolves ${secrets.X} from vault        │       │
│         │ - Runs: apply                             │       │
│         │ - Runs: health checks                     │       │
│         │ - Scrubs credentials from output          │       │
│         └───────────────────────────────────────────┘       │
│                                                             │
│  Receives: { status: "success", endpoint: "https://..." }   │
└─────────────────────────────────────────────────────────────┘
```

See [ARTIFACTS](ARTIFACTS.md) for the Isolation Principle.

**Environment:** User-specified target (dev/staging/prod)

**Output:** Running system in target environment

**Failure Handling:**
- Iterate on deployment failures up to retry threshold
- Document failure reasons (scrubbed of sensitive data)
- Escalate to human review when threshold exceeded

**Completion Criteria:**
- [ ] Infrastructure specs produced and complete
- [ ] IaC generated with reference-only secrets
- [ ] Deployment executed successfully
- [ ] Health checks pass
- [ ] System accessible at expected endpoints

---

### Validate — Verify Spec Compliance

The Validate phase verifies that the deployed system satisfies all specification requirements.

**Specification Agent:**
| Agent | Layer | Input | Output |
|-------|-------|-------|--------|
| `smaqit.coverage` | Coverage | All layer specs | `specs/coverage/*.md` |

**Implementation Agent:** `smaqit.validation`

**Workflow:**
```
1. Coverage agent:
   a. Reads all upstream specs (business, functional, stack, infrastructure)
   b. Enumerates all acceptance criteria by ID
   c. Produces test definitions (Gherkin format)
   d. Maps: Requirement ID → Test Case → Expected Outcome
   e. Flags untestable criteria

2. Validation agent:
   a. Executes tests against deployed system
   b. Collects pass/fail results per test case
   c. Calculates spec coverage percentage
   d. Produces validation report
```

**Environment:** Same target environment as Deploy phase

**Output:** Validation report containing:
- Spec coverage percentage
- Pass/fail status per requirement
- Unverified requirements with justification
- Failure details for failed tests

**Failure Handling:**
- Test failures do NOT trigger automatic retry
- Human decides next action:
  - Return to Develop (code/spec issue)
  - Return to Deploy (environment issue)
  - Investigate further
  - Accept with known issues

**Completion Criteria:**
- [ ] Coverage specs produced with all testable criteria mapped
- [ ] Tests executed against deployed system
- [ ] Validation report generated
- [ ] Spec coverage percentage calculated
- [ ] Untestable criteria documented with justification

---

## Phase Transitions

### Develop → Deploy

**Trigger:** Develop phase completion criteria met

**Prerequisites:**
- Application compiles and runs
- Unit tests pass
- Behavior verified in isolated environment

**Handoff:**
- Working application code
- Phase 1 specs (context for Infrastructure coherence)

---

### Deploy → Validate

**Trigger:** Deploy phase completion criteria met

**Prerequisites:**
- System deployed to target environment
- Health checks pass
- Endpoints accessible

**Handoff:**
- Running system endpoint(s)
- Target environment identifier

---

### Validate → Feedback

**Trigger:** Validation report generated

**Outcomes:**

| Result | Action |
|--------|--------|
| All tests pass | Cycle complete ✓ |
| Tests fail | Human decides: Develop, Deploy, or investigate |
| Low coverage | Review Coverage specs for gaps |

**Note:** Automated feedback routing is deferred to future versions. Currently, validation failures require human decision.

---

## Failure Handling

### Retry Threshold

Implementation agents iterate on failures up to a configurable threshold:

| Phase | Default Retries | Rationale |
|-------|-----------------|-----------|
| Develop | 3 | Code/test fixes typically converge quickly |
| Deploy | 2 | Infrastructure issues often need investigation |
| Validate | 0 | Failures require human analysis |

### Failure Documentation

Each failure attempt MUST document:
- What was attempted
- What failed (error message, scrubbed if sensitive)
- What was changed before retry
- Final status after threshold exceeded

### Escalation

When retry threshold is exceeded:
1. Agent stops iterating
2. Failure summary produced
3. Human review required to proceed or abort

---

## Spec Change Adaptation

When any layer spec changes, downstream phases must re-run:

| Spec Changed | Required Re-runs |
|--------------|------------------|
| Business, Functional, Stack | Develop → Deploy → Validate |
| Infrastructure | Deploy → Validate |
| Coverage | Validate |

Coverage phase always re-runs when any upstream spec changes to ensure test coverage remains current.

---

## Current Assumptions

These assumptions are explicitly stated and subject to revision per [SMAQIT](SMAQIT.md):

| Assumption | Status | Revision Trigger |
|------------|--------|------------------|
| Phases are strictly sequential | Active | Incremental deployment proves valuable |
| Validation failures require human decision | Active | Patterns emerge for automated routing |
| New features trigger new cycle | Active | Addressed in task 014 |

---

## See Also

- [SMAQIT](SMAQIT.md) — Framework overview and principles
- [LAYERS](LAYERS.md) — Layer definitions and dependencies
- [TEMPLATES](TEMPLATES.md) — Template structure rules
- [AGENTS](AGENTS.md) — Agent behaviors
- [ARTIFACTS](ARTIFACTS.md) — Artifact rules
