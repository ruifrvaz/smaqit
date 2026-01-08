# [PHASE_NAME] Prompt Template

## Structure

```markdown
---
name: smaqit.[PHASE]
description: [Agent purpose - what this agent does]
agent: smaqit.[PHASE]
---

# [PHASE_NAME] Execution

## Parameters

[Execution parameters with suggested structure]

<!-- Example: [Guidance showing execution options] -->

[User parameters here, if any]
```

## Phase-Specific Sections

### Development Phase

```markdown
---
name: smaqit.development
description: Build application from specifications
agent: smaqit.development
---

# Development Execution

## Parameters

### Build Options
[Any build-time preferences?]

<!-- Example: "Run in watch mode for hot reload" -->
<!-- Example: "Skip unit tests for fast iteration" -->

### Output Preferences
[How should output be displayed?]

<!-- Example: "Verbose logging for debugging" -->
<!-- Example: "Quiet mode - errors only" -->

### Environment
[Any environment-specific settings?]

<!-- Example: "Use Docker for isolated build" -->
<!-- Example: "Build for production (optimized)" -->
```

### Deployment Phase

```markdown
---
name: smaqit.deployment
description: Deploy application to target environment
agent: smaqit.deployment
---

# Deployment Execution

## Parameters

### Deployment Target
[Any target-specific preferences?]

<!-- Example: "Deploy to staging environment first" -->
<!-- Example: "Use blue-green deployment strategy" -->

### Verification
[How should deployment be verified?]

<!-- Example: "Run smoke tests after deployment" -->
<!-- Example: "Skip health checks (trusted environment)" -->

### Output Preferences
[How should output be displayed?]

<!-- Example: "Verbose health check logging" -->
<!-- Example: "Scrub all sensitive data from output" -->
```

### Validation Phase

```markdown
---
name: smaqit.validation
description: Validate deployed system against specifications
agent: smaqit.validation
---

# Validation Execution

## Parameters

### Execution Scope
[Which tests should be executed?]

<!-- Example: "Run only smoke tests (quick validation)" -->
<!-- Example: "Run full test suite including performance tests" -->

### Failure Handling
[How should test failures be handled?]

<!-- Example: "Continue on failures (full report mode)" -->
<!-- Example: "Stop on first failure (fast feedback)" -->

### Output Preferences
[How should results be displayed?]

<!-- Example: "Verbose test output for debugging" -->
<!-- Example: "Summary only - list failures" -->
```