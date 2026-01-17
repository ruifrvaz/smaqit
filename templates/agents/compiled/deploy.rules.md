# Deployment Phase Compilation Rules

**L1 Transformation Rules:** Compile L0 principles → L1 directives for Deployment phase agent

**Target Agent:** `agents/smaqit.deployment.agent.md`

---

## Source L0 Principles

### Primary Source: ARTIFACTS.md § The Isolation Principle

> "Agents operate on references, never values. Secrets and credentials MUST remain outside the agent's context at all times—resolution happens in a trusted execution layer that returns only outcomes, never the sensitive data itself."

### Secondary Source: PHASES.md § Deploy Phase Activities

> "The Infrastructure agent produces infrastructure specifications from user deployment requirements.
>
> The Deployment agent consolidates infrastructure and stack specifications for coherence, generates Infrastructure as Code with credential references (never values), triggers a trusted execution layer that resolves secrets and performs deployment, and verifies system health in the target environment."

### Tertiary Source: ARTIFACTS.md § Implementation Artifacts by Phase

> **Deploy Phase:**
> - Infrastructure code (Terraform, etc.)
> - Deployment manifests, environment configs
> - Deployment report with health status and endpoints

---

## L1 Directive Compilation

### Output Artifacts

**L0 Source:** PHASES.md § Deploy Phase Output

**Compile to [OUTPUT_ARTIFACTS]:**
```markdown
- Infrastructure as Code (IaC) configurations with reference-only secrets
- Deployment manifests
- Environment configurations
- Running system in target environment
- Deployment report in `.smaqit/reports/deployment-phase-report-YYYY-MM-DD.md` with health status, endpoints, and scrubbed logs
```

### Phase-Specific Rules: Isolation Principle

**L0 Source:** "Agents operate on references, never values"

**Compile to MUST directives:**
- Generate Infrastructure as Code with credential references only
- Use placeholder references: `${secrets.SECRET_NAME}` (never actual values)
- Trigger trusted execution layer for deployment
- Verify system health in target environment
- Configure observability per infrastructure specs
- Scrub credentials from all logs and reports

**Compile to MUST NOT directives:**
- Include actual secrets, passwords, API keys, tokens, or credentials in generated artifacts
- Expose credentials in logs or reports
- Deploy without health verification

### Completion Criteria

**L0 Source:** PHASES.md § Deploy Phase Completion

**Compile to [ADDITIONAL_COMPLETION_CRITERIA]:**
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

---

## Version

- **Created:** 2026-01-14
- **L0 Sources:** ARTIFACTS.md (Isolation Principle), PHASES.md (Deploy Phase)
- **Compilation Target:** agents/smaqit.deployment.agent.md
