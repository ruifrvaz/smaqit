# [LAYER_NAME] Prompt Template

Use this template to create specification layer prompts (business, functional, stack, infrastructure, coverage).

## Structure

```markdown
---
name: smaqit.[LAYER]
description: Create [LAYER] layer specifications from user requirements
agent: smaqit.[LAYER]
---

# [LAYER_NAME] Prompt

This prompt captures [LAYER] requirements for your project. These requirements will be used to generate [LAYER] specifications.

## Requirements

[Layer-specific sub-sections with suggested structure]

<!-- Example: [Guidance showing good format] -->

[User requirements here]
```

## Layer-Specific Sections

### Business Layer

```markdown
## Requirements

### Actors
[Who interacts with the system?]

<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->
<!-- Example: "System - Console application that manages state" -->

### Use Cases
[What do users want to accomplish?]

<!-- Example: "Greet Mario Fans - User runs application, system displays greeting" -->

### Success Metrics
[How do you measure success?]

<!-- Example: "Users recognize Mario catchphrases in output" -->
<!-- Example: "Application completes in under 2 seconds" -->

### Business Goals
[Why build this? What value does it provide?]

<!-- Example: "Delight Mario fans with authentic character experience" -->
```

### Functional Layer

```markdown
## Requirements

### User Experience
[What experience should users have?]

<!-- Example: "Console-based Mario greeting with visual ASCII art" -->

### Behaviors
[What should the system do?]

<!-- Example: "Display randomized Mario catchphrases" -->
<!-- Example: "Support colorful terminal output with ANSI codes" -->

### Data Models
[What data structures are needed?]

<!-- Example: "Catchphrase { text: string, frequency: number }" -->

### API Contracts
[What are the inputs and outputs?]

<!-- Example: "Input: None (no arguments)" -->
<!-- Example: "Output: Console text with ANSI color codes" -->

### State Transitions
[How does state change?]

<!-- Example: "START → LOADING → RENDERING → DISPLAYED → EXIT" -->
```

### Stack Layer

```markdown
## Requirements

### Technology Preferences
[What languages, frameworks, or tools do you prefer?]

<!-- Example: "Python with colorama library for colored terminal output" -->
<!-- Example: "Go with fatih/color package for fast single-binary execution" -->

### Constraints
[Any technology limitations or requirements?]

<!-- Example: "Must support cross-platform (Linux, macOS, Windows)" -->
<!-- Example: "No external dependencies beyond standard package managers" -->

### Build Tools
[How should the project be built?]

<!-- Example: "Use language-native build system (pip, go build, npm)" -->

### Development Environment
[What tools are needed for development?]

<!-- Example: "Any text editor, terminal for testing, Git for version control" -->
```

### Infrastructure Layer

```markdown
## Requirements

### Deployment Target
[Where will this run?]

<!-- Example: "Local developer machine or Docker container" -->

### Compute Resources
[What resources are needed?]

<!-- Example: "Minimal: < 50MB memory, < 1 second execution time" -->

### Networking
[Any network requirements?]

<!-- Example: "No network connectivity required (pure local execution)" -->

### Observability
[How should it be monitored?]

<!-- Example: "Simple stdout logs, errors to stderr, exit codes for status" -->

### Secrets Management
[Any secrets or credentials?]

<!-- Example: "No secrets required (stateless application)" -->

### Scaling
[How should it scale?]

<!-- Example: "Not applicable (single-user, local execution)" -->
```

### Coverage Layer

```markdown
## Requirements

### Test Scope
[What should be tested?]

<!-- Example: "Verify Mario-themed elements appear correctly in output" -->

### Performance Benchmarks
[Any performance requirements?]

<!-- Example: "Application starts in < 1 second" -->
<!-- Example: "Total execution time < 2 seconds" -->

### Security Requirements
[Any security concerns?]

<!-- Example: "No security concerns (local, non-networked application)" -->

### Test Environment
[Where will tests run?]

<!-- Example: "Local terminal with color support, multiple OS platforms" -->

### Integration Points
[What external systems need testing?]

<!-- Example: "None (standalone application)" -->

### Verification Requirements
[What must be verified?]

<!-- Example: "Mario catchphrase appears in output" -->
<!-- Example: "ASCII art renders correctly" -->
<!-- Example: "Application exits with code 0" -->
```