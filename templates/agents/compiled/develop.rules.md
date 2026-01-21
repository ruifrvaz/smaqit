---
phase: develop
target: agents/smaqit.development.agent.md
sources:
  - framework/PHASES.md
  - framework/ARTIFACTS.md
created: 2026-01-14
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| PHASES.md | Develop Phase Activities |
| ARTIFACTS.md | Implementation Artifacts by Phase |

---

## L1 Directive Compilation

### Output Artifacts

**[OUTPUT_ARTIFACTS]:**
```markdown
- Source code (application, tests, configurations)
- Build artifacts
- README with build, test, and run instructions
- Development report in `.smaqit/reports/development-phase-report-YYYY-MM-DD.md` (build/test/run results)
```

### Phase-Specific Rules

**MUST directives:**
- Generate application code from specifications
- Generate unit tests
- Build/compile application
- Run application in isolated environment
- Execute unit tests
- Verify behavior matches spec acceptance criteria
- Include README with build, test, and run instructions

**MUST NOT directives:**
- Deploy to production or staging environments
- Execute deployment-specific tasks
- Skip unit test execution

### Completion Criteria

**[ADDITIONAL_COMPLETION_CRITERIA]:**
- [ ] Code generated and compiles without errors
- [ ] Unit tests pass
- [ ] Application runs successfully in isolated environment
- [ ] README includes build, test, and run instructions

---

## Compilation Guidance for Agent-L2

1. **Replace [PHASE_SPECIFIC_RULES]** with Development-specific directives
2. **Replace [OUTPUT_ARTIFACTS]** with Development output artifacts
3. **Replace [ADDITIONAL_COMPLETION_CRITERIA]** with Development completion criteria
4. **Replace generic placeholders** with Development phase values
