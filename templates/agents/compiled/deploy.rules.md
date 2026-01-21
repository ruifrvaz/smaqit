---
phase: deploy
target: agents/smaqit.deployment.agent.md
sources:
  - framework/PHASES.md
  - framework/ARTIFACTS.md
created: 2026-01-14
---

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| ARTIFACTS.md | The Isolation Principle |
| PHASES.md | Deploy Phase Activities |
| ARTIFACTS.md | Implementation Artifacts by Phase |

---

## L1 Directive Compilation

### Output Artifacts

**[OUTPUT_ARTIFACTS]:**
```markdown
- Infrastructure as Code (IaC) configurations with reference-only secrets
- Deployment manifests
- Environment configurations
- Running system in target environment
- Deployment report in `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md` with health status, endpoints, and scrubbed logs
```

### Phase-Specific Rules: Isolation Principle

**MUST directives:**
- Generate Infrastructure as Code with credential references only
- Use placeholder references: `${secrets.SECRET_NAME}` (never actual values)
- Trigger trusted execution layer for deployment
- Verify system health in target environment
- Configure observability per infrastructure specs
- Scrub credentials from all logs and reports

**MUST NOT directives:**
- Include actual secrets, passwords, API keys, tokens, or credentials in generated artifacts
- Expose credentials in logs or reports
- Deploy without health verification

### Completion Criteria

**[ADDITIONAL_COMPLETION_CRITERIA]:**
- [ ] IaC generated with reference-only secrets (no actual values)
- [ ] Deployment executed successfully
- [ ] Health checks pass
- [ ] System accessible at expected endpoints
- [ ] All logs scrubbed of credentials

---

## Compilation Guidance for Agent-L2

1. **Replace [PHASE_SPECIFIC_RULES]** with Deployment-specific directives including Isolation Principle
2. **Replace [OUTPUT_ARTIFACTS]** with Deployment output artifacts
3. **Replace [ADDITIONAL_COMPLETION_CRITERIA]** with Deployment completion criteria
4. **Replace generic placeholders** with Deployment phase values
