# Self-Validating Agents

## Definition

Self-validating agents are agents that verify their own output against completion criteria before declaring completion. This shifts quality assurance left—into the agent itself, not a separate review step.

## How It Works

Every agent has a Completion Criteria section that defines what must be true before the agent can finish:

**Specification agents:**
```markdown
## Completion Criteria

- [ ] All template sections filled (no placeholders remain)
- [ ] All upstream references are valid and accessible
- [ ] All acceptance criteria are testable (measurable, observable, unambiguous)
- [ ] Scope boundaries explicitly stated
- [ ] No implementation details leaked into spec
```

**Implementation agents:**
```markdown
## Completion Criteria

- [ ] All referenced spec requirements are addressed
- [ ] All acceptance criteria from specs are satisfied
- [ ] Output is traceable to input specifications
- [ ] No unspecified features were added
- [ ] Cross-layer consolidation completed without conflicts
```

## The Validation Loop

```
1. Produce output following template
2. Check output against completion criteria
3. If criteria unmet → iterate on output (up to retry threshold)
4. If criteria met → declare completion
5. If criteria impossible → flag blocker and stop
```

Agents don't just produce output and move on—they validate their work.

## Why Self-Validation Matters

**Catches errors early:**
- Agent discovers incomplete output before user sees it
- Reduces back-and-forth with users
- Prevents downstream agents from consuming invalid input

**Enforces quality standards:**
- Agents can't skip required sections
- Can't leave placeholders in output
- Must satisfy testability requirements

**Reduces human review burden:**
- User reviews content, not structure
- User doesn't check for missing sections
- User doesn't verify reference validity

**Makes quality criteria explicit:**
- What "done" means is documented
- No ambiguity about agent responsibilities
- Consistent quality across all agents

## Example

**Business agent self-validation:**

Agent generates business spec for "User Login". Before completing, it checks:

- [ ] ✓ All template sections filled
- [ ] ✓ Actors defined (User, System)
- [ ] ✓ Success metrics included (login success rate)
- [ ] ✗ **Acceptance criteria testable** — "BUS-LOGIN-002: System feels responsive" is subjective

Agent stops, flags the issue:
```
⚠️ Criterion BUS-LOGIN-002 is not testable (subjective: "feels responsive")
Propose: Replace with measurable criterion, e.g., "Login response time < 2 seconds"
```

Agent iterates, produces testable criterion, rechecks, then completes.

## Failure Handling

Agents track iterations and stop when:

1. **Success**: All completion criteria met
2. **Blocked**: Impossible requirement or missing input
3. **Threshold**: Retry limit exceeded

**Failure table example:**

| Situation | Action |
|-----------|--------|
| Ambiguous input | Request clarification, do not guess |
| Conflicting requirements | Flag conflict, propose resolution options |
| Missing upstream spec | Stop, indicate which spec is needed |
| Impossible requirement | Report impossibility with rationale |

Agents don't silently fail—they report and request help.

## Self-Validation vs External Validation

**Self-validation (agents):**
- Structure compliance (template sections)
- Reference validity (upstream specs exist)
- Testability (acceptance criteria are measurable)
- Completeness (all required fields filled)

**External validation (humans/testing agents):**
- Content correctness (requirements match user intent)
- Behavioral correctness (implementation passes tests)
- Strategic alignment (solution fits business goals)

Self-validation handles mechanical checks. External validation handles judgment calls.

## Trade-offs

**Benefits:**
- Higher quality output
- Earlier error detection
- Reduced review burden
- Consistent standards

**Costs:**
- More compute (agents iterate)
- Slower initial output (validation takes time)
- More complex agent definitions

The cost is intentional—quality built-in is cheaper than quality inspected-later.

## Related

- [Fail-Fast on Ambiguity](../designs/fail-fast-on-ambiguity.md) — When agents stop and request clarification
- [Template Constraints](../designs/template-constraints.md) — What agents validate against
- [Accept Mutability, Validate Behavior](accept-mutability.md) — What agents validate (behavior, not style)
