---
name: smaqit.coverage
description: Create coverage layer specifications from verification requirements
agent: smaqit.coverage
---

# Coverage Prompt

This prompt captures verification and testing requirements for your project. These requirements will be used to generate coverage specifications.

## Requirements

### Test Scope
[What types of testing are needed?]

<!-- Example: "Integration testing - verify greeting output" -->
<!-- Example: "End-to-end testing - full application execution" -->

### Performance Benchmarks
[What are the performance requirements?]

<!-- Example: "Application completes in under 2 seconds" -->
<!-- Example: "Memory usage under 10MB" -->

### Security Requirements
[What security verifications are needed?]

<!-- Example: "No user input - no injection vulnerabilities" -->
<!-- Example: "Read-only file access for greeting data" -->

### Test Environment
[Where and how should tests run?]

<!-- Example: "GitHub Actions on push to main branch" -->
<!-- Example: "Local test suite with pytest" -->

### Integration Points
[What external systems need testing?]

<!-- Example: "None - standalone application" -->

### Acceptance Thresholds
[What defines acceptable test results?]

<!-- Example: "100% of acceptance criteria must pass" -->
<!-- Example: "All integration tests green" -->
