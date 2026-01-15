# Development Phase Compilation Rules

**L1 Transformation Rules:** Compile L0 principles → L1 directives for Development phase agent

**Target Agent:** `agents/smaqit.development.agent.md`

---

## Source L0 Principles

### Primary Source: PHASES.md § Develop Phase Activities

> "Specification agents produce Business, Functional, and Stack layer specifications from user requirements.
>
> The Development agent consolidates specs for coherence, generates application code and tests, builds the application, and verifies it works as specified in an isolated environment."

### Secondary Source: ARTIFACTS.md § Implementation Artifacts by Phase

> **Develop Phase:**
> - Source code, tests, configurations, build files
> - README with build, test, and run instructions
> - Development report documenting build/test/run results

---

## L1 Directive Compilation

### Output Artifacts

**L0 Source:** PHASES.md § Develop Phase Output

**Compile to [OUTPUT_ARTIFACTS]:**
```markdown
- Source code (application, tests, configurations)
- Build artifacts
- README with build, test, and run instructions
- Development report in `.smaqit/reports/development-phase-report-YYYY-MM-DD.md` (build/test/run results)
```

### Phase-Specific Rules

**L0 Source:** PHASES.md § Develop Phase Activities

**Compile to MUST directives:**
- Generate application code from specifications
- Generate unit tests
- Build/compile application
- Run application in isolated environment
- Execute unit tests
- Verify behavior matches spec acceptance criteria
- Include README with build, test, and run instructions

**Compile to MUST NOT directives:**
- Deploy to production or staging environments
- Execute deployment-specific tasks
- Skip unit test execution

### Completion Criteria

**L0 Source:** PHASES.md § Develop Phase Completion

**Compile to [ADDITIONAL_COMPLETION_CRITERIA]:**
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

---

## Version

- **Created:** 2026-01-14
- **L0 Sources:** PHASES.md (Develop Phase), ARTIFACTS.md
- **Compilation Target:** agents/smaqit.development.agent.md
