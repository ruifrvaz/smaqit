# smaqit User Testing Report

**Date:** YYYY-MM-DD  
**Tester:** [Agent/Human]  
**smaqit Version:** [Version from installer/main.go]  
**Test Case:** [Test feature name]

---

## Test Information

| Property | Value |
|----------|-------|
| Operating System | [OS name and version] |
| Go Version | [go version output] |
| smaqit Version | [Version] |
| Test Duration | [Start time - End time] |
| Test Project Path | [Absolute path to test project] |

---

## Standardized Checklist

### Environment Setup
- [ ] Go toolchain available
- [ ] smaqit source repository accessible
- [ ] Test directory created

### Installer Build
- [ ] `make prepare` executed successfully
- [ ] `make build` executed successfully
- [ ] Binary created in `dist/`
- [ ] Binary is executable

### Project Initialization
- [ ] Test project directory created
- [ ] `smaqit init` executed successfully
- [ ] `.smaqit/` directory created with framework files
- [ ] `.github/agents/` directory created with 8 agent files
- [ ] `.github/prompts/` directory created with 8 prompt files
- [ ] `specs/` directory structure created (5 subdirectories)

### Business Layer Specification
- [ ] `/smaqit.business` prompt invoked with test requirements
- [ ] Business agent executed successfully
- [ ] `specs/business/*.md` files created
- [ ] Business spec contains acceptance criteria with `BUS-` IDs

### Functional Layer Specification
- [ ] `/smaqit.functional` prompt invoked with test requirements
- [ ] Functional agent executed successfully
- [ ] `specs/functional/*.md` files created
- [ ] Functional spec contains acceptance criteria with `FUN-` IDs
- [ ] Functional spec references Business specs

### Stack Layer Specification
- [ ] `/smaqit.stack` prompt invoked with test requirements
- [ ] Stack agent executed successfully
- [ ] `specs/stack/*.md` files created
- [ ] Stack spec contains acceptance criteria with `STK-` IDs
- [ ] Stack spec references Business and Functional specs

### Infrastructure Layer Specification
- [ ] `/smaqit.infrastructure` prompt invoked with test requirements
- [ ] Infrastructure agent executed successfully
- [ ] `specs/infrastructure/*.md` files created
- [ ] Infrastructure spec contains acceptance criteria with `INF-` IDs
- [ ] Infrastructure spec references Phase 1 specs

### Coverage Layer Specification
- [ ] `/smaqit.coverage` prompt invoked with test requirements
- [ ] Coverage agent executed successfully
- [ ] `specs/coverage/*.md` files created
- [ ] Coverage spec contains acceptance criteria with `COV-` IDs
- [ ] Coverage spec references all upstream specs

### Cleanup
- [ ] Test project removed
- [ ] No residual artifacts left in smaqit source directory

---

## Execution Log

Timestamped record of all steps executed during the test.

```
[HH:MM:SS] Step: Description
[HH:MM:SS] Output: Command output or observation
```

**Example:**
```
[10:00:00] Started test execution
[10:00:05] Built installer: smaqit version dev
[10:00:10] Created test project: /tmp/smaqit-test-mario-hello
[10:00:15] Initialized smaqit in test project
[10:00:20] Invoked Business agent with Mario requirements
[10:00:45] Business spec created: specs/business/mario-greeting.md
...
```

---

## Painpoints Identified

Document any issues, friction, or unexpected behavior encountered during testing.

### Blockers
Critical issues that prevented test completion:
- **Issue:** [Description]
- **Location:** [Where it occurred]
- **Impact:** [What couldn't be completed]

### Issues
Non-critical problems that affected user experience:
- **Issue:** [Description]
- **Context:** [When/where it occurred]
- **Severity:** [High/Medium/Low]

### UX Friction
Areas where the workflow felt awkward or confusing:
- **Observation:** [Description]
- **Suggestion:** [Potential improvement]

### Performance
Timing or resource concerns:
- **Observation:** [Description]
- **Impact:** [User experience effect]

---

## Recommendations

Actionable suggestions for improving smaqit based on this test run.

1. **[Category]:** [Recommendation]
   - **Rationale:** [Why this matters]
   - **Priority:** [High/Medium/Low]

---

## Overall Result

**Status:** ✅ PASS | ❌ FAIL

**Summary:** [One paragraph summary of test outcome]

**Key Findings:**
- [Finding 1]
- [Finding 2]
- [Finding 3]

**Next Steps:**
- [Action item 1]
- [Action item 2]

---

## Notes

Any additional context, observations, or metadata relevant to this test run.
