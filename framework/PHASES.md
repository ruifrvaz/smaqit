# Phases

Phases are the sequential stages of software development in smaqit. Each phase includes both specification generation and implementation execution. The recommended workflow is to complete each phase before moving to the next, though specifications can be generated ahead if needed.

## Overview

smaqit operates in three sequential phases:

| Phase | Name | Specification Artifacts | Implementation Artifacts |
|-------|------|------------------------|-------------------------|
| Phase 1 | Develop | Business, Functional, Stack | Code, README, Development report in `.smaqit/reports/` |
| Phase 2 | Deploy | Infrastructure | Running system, Deployment report in `.smaqit/reports/` |
| Phase 3 | Validate | Coverage | Validation report in `.smaqit/reports/` |

Each phase:
1. **Specifies** — One or more specification agents produce layer specs
2. **Consolidates** — Implementation agent verifies cross-layer coherence
3. **Implements** — Implementation agent produces and executes artifacts
4. **Verifies** — Implementation agent confirms success before phase completion

Phases are strictly sequential. Deploy cannot begin until Develop completes. Validate cannot begin until Deploy completes. This constraint is subject to revision based on real-world usage (see [SMAQIT](SMAQIT.md)).

### Implementation Phase Principles

**Status Cascade:** Implementation agents update all specifications they reference, not just specs from their target layer.

When an implementation agent processes work:
1. Identifies specs via `smaqit plan --phase=[PHASE]` (returns target layer specs)
2. References upstream specs for coherence/context
3. Updates frontmatter in ALL referenced specs to reflect phase completion

**Rationale:** If an implementation agent reads a specification for implementation context, that specification has been implemented/deployed/validated in that phase. Status must reflect reality for accurate lifecycle tracking.

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

**Pre-Orchestration Validation:**

Implementation agents perform pre-orchestration validation to verify readiness (see AGENTS.md Pre-Orchestration Validation concept). For Development phase, validation includes:

- Input sufficiency check for required prompt files
- Dependency verification for upstream artifacts
- Execution environment readiness

Validation failures halt workflow with guidance describing missing requirements or configuration issues.

**Phase Activities:**

Specification agents produce Business, Functional, and Stack layer specifications from user requirements.

The Development agent consolidates specs for coherence, generates application code and tests, builds the application, and verifies it works as specified in an isolated environment.

**Environment:** Implicit — local developer machine or agent runner (e.g., GitHub Actions runner)

**Output:** Working, tested application in isolated environment

**Failure Handling:**
- Iterate on code/test failures up to retry threshold
- Document failure reasons at each attempt
- Escalate to human review when threshold exceeded

**Phase Completion:**

The Develop phase completes when Business, Functional, and Stack specifications are produced, code is generated and compiles without errors, tests pass, and the application runs successfully with behavior matching spec acceptance criteria. Development report documents build, test, and run results. Spec frontmatter reflects implemented state with timestamps.

---

### Deploy — Run in Target Environment

The Deploy phase transforms a working application into a running system in a target environment.

**Specification Agent:**
| Agent | Layer | User Input | Context | Output |
|-------|-------|------------|---------|--------|
| `smaqit.infrastructure` | Infrastructure | Deployment requirements | Phase 1 specs | `specs/infrastructure/*.md` |

**Implementation Agent:** `smaqit.deployment`

**Pre-Orchestration Validation:**

Implementation agents perform pre-orchestration validation to verify readiness (see AGENTS.md Pre-Orchestration Validation concept). For Deployment phase, validation includes:

- Input sufficiency check for infrastructure requirements
- Dependency verification for development phase outputs
- Execution environment and credentials readiness

Validation failures halt workflow with guidance describing missing requirements or configuration issues.

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

**Phase Activities:**

The Infrastructure agent produces infrastructure specifications from user deployment requirements.

The Deployment agent consolidates infrastructure and stack specifications for coherence, generates Infrastructure as Code with credential references (never values) and triggers a trusted execution layer that resolves secrets and performs deployment. Once execution outcome is received, which may contain success/failure, health status and endpoints, the agent verifies system health in the target environment.

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

**Phase Completion:**

The Deploy phase completes when Infrastructure specifications are produced, Infrastructure as Code is generated with reference-only secrets, deployment executes successfully, health checks pass, and the system is accessible at expected endpoints. Deployment report documents health status and endpoints, and is archived in a designated location. Spec frontmatter reflects deployed state with timestamps across all referenced specs per Status Cascade principle and acceptance criteria checkboxes are updated across all referenced specs.

---

### Validate — Verify Spec Compliance

The Validate phase verifies that the deployed system satisfies all specification requirements.

**Specification Agent:**
| Agent | Layer | Input | Output |
|-------|-------|-------|--------|
| `smaqit.coverage` | Coverage | All layer specs | `specs/coverage/*.md` |

**Implementation Agent:** `smaqit.validation`

**Pre-Orchestration Validation:**

Implementation agents perform pre-orchestration validation to verify readiness (see AGENTS.md Pre-Orchestration Validation concept). For Validation phase, validation includes:

- Input sufficiency check for test requirements
- Dependency verification for deployed system accessibility
- Test execution environment readiness

Validation failures halt workflow with guidance describing missing requirements or configuration issues.

**Phase Activities:**

The Coverage agent translates acceptance criteria from all upstream specs into executable test definitions, mapping each requirement to expected outcomes and flagging criteria that cannot be automatically verified.

The Validation agent generates test artifacts that can run independently of agent execution, executes those tests against the deployed system, and produces a validation report documenting coverage and results.

**Environment:** Same target environment as Deploy phase

**Output:**

Test artifacts that exist independently and can execute in any environment with the appropriate runtime. These include test implementations, framework configuration, test utilities, and CI/CD integration.

Validation report documenting spec coverage percentage, pass/fail status per requirement, unverified requirements with justification, and failure details.

**Failure Handling:**
- Test failures do NOT trigger automatic retry
- Human decides next action:
  - Return to Develop (code/spec issue)
  - Return to Deploy (environment issue)
  - Investigate further
  - Accept with known issues

**Phase Completion:**

The Validate phase completes when Coverage specs exist with all testable criteria mapped, test artifacts have been generated and executed against the deployed system, and the validation report documents coverage percentage and results. Spec frontmatter reflects validated state with timestamps and acceptance criteria checkboxes are updated across all referenced specs.

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

## Acceptance Criteria Checkboxes

Each implementation agent updates checkboxes in the specs it processes as part of its self-validation process.

**Checkbox Responsibility by Phase:**

| Phase | Agent | Updates Checkboxes In | Rationale |
|-------|-------|----------------------|-----------|
| Develop | Development | Business, Functional, Stack specs | Agent implements these requirements and confirms satisfaction |
| Deploy | Deployment | Infrastructure specs | Agent deploys to environment and confirms infrastructure requirements met |

**Checkbox States:**

- `[ ]` — Not yet implemented/validated
- `[x]` — Satisfied (implementation complete or test passed)
- `[!]` — Failed, untestable, or not satisfied

**Self-Validation Principle:**

Checkbox updates are part of the implementation agent's self-validation process, confirming that requirements were addressed during execution. This creates an audit trail showing which phase satisfied which requirements.

**Note:** Checkbox updates are implementation tracking, not specification modification. They reflect work done, not changes to requirements.

---

## Incremental Development

smaqit supports incremental workflows where specs are added and implemented iteratively.

**Spec State Tracking:**

Each spec carries state through phases via frontmatter:
- Draft → Implemented → Deployed → Validated

**Determining Work:**

Implementation agents run `smaqit plan --phase=[PHASE]` to get paths to specs needing processing:

| Mode | Command | Processes |
|------|---------|-----------|
| Incremental | `smaqit plan --phase=develop` | Only specs with `status: draft` or `status: failed` |
| Regeneration | `smaqit plan --phase=develop --regen` | All specs regardless of status |

**Adding Features:**

```
1. User adds requirements to prompt file
2. Spec agent generates new specs (status: draft)
3. Implementation agent runs `smaqit plan --phase=develop`
4. CLI returns only new draft specs (existing implemented specs skipped)
5. Agent processes returned paths
6. Tests validate new + existing functionality
```

**Checking Status:**

```bash
# View aggregate phase status
smaqit status

# Shows per-phase spec counts:
Develop: 18 implemented, 2 failed
Deploy: 15 deployed, 3 draft
Validate: 12 validated, 5 draft
```

CLI aggregates status by scanning all spec frontmatter. No centralized state file.

---

## Current Assumptions

These assumptions are explicitly stated and subject to revision per [SMAQIT](SMAQIT.md):

| Assumption | Status | Revision Trigger |
|------------|--------|------------------|
| Phases are strictly sequential | Active | Incremental deployment proves valuable |
| Validation failures require human decision | Active | Patterns emerge for automated routing |
