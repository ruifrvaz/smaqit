# Iterative Assessment Before Spec Generation and Implementation

**Status:** In Progress  
**Created:** 2026-02-01  
**Updated:** 2026-02-06

## Description

Agents invoke assessment skill when detecting conditions requiring critical analysis and planning before execution. Assessment skill provides iterative evaluation with approval gates, preventing wasted effort on ambiguous, contradictory, or insufficient inputs.

**Approach:** Assessment as reusable agent skill (`.github/skills/assessment/`) rather than mandatory gate for all agent invocations. Agents automatically invoke when detecting specific conditions.

## Acceptance Criteria

- [x] L0 concepts documented in framework files
- [ ] L1 skill template created following agentskills.io specification
- [ ] L1 agent templates updated to reference assessment skill
- [ ] L2 assessment skill compiled into `.github/skills/assessment/`
- [ ] L2 all product agents updated with assessment skill directives
- [ ] session.assess prompt simplified to invoke assessment skill

## Progress

### 2026-02-06: L0 Concepts Documented

**Framework updates (Agent-L0):**
- Created `framework/SKILLS.md` with complete skills architecture
- Documented assessment skill concept, workflow, and invocation patterns
- Updated `framework/AGENTS.md` with brief skills reference pointing to SKILLS.md
- Updated `framework/SMAQIT.md` "See Also" section to include SKILLS.md
- Updated "Fail-Fast on Ambiguity" principle in AGENTS.md to reference assessment skill

**L0 Content Added (SKILLS.md):**
- Skills as invocable capabilities (principles and concepts)
- Skill structure and location mapping (`.github/skills/`)
- Condition detection patterns triggering skill invocation
- Skill invocation context (inputs and outputs)
- Assessment skill complete specification:
  - Purpose and workflow (five-step process)
  - Invocation triggers (ambiguity, conflicts, insufficient detail, complexity)
  - Approval mechanisms (user input vs autonomous operation)
  - Output format (structured assessment results)
- Future skills placeholder (conflict resolution, gap detection, etc.)

**Next Steps:**
1. Invoke Agent-L1 to create assessment skill template
2. Invoke Agent-L2 to compile skill and update product agents
3. Update session.assess prompt to reference skill

## Notes

**Trigger conditions for assessment skill:**
- Ambiguous requirements (multiple interpretations possible)
- Conflicting inputs (prompt vs upstream specs, or within prompt)
- Insufficient detail (cannot proceed without assumptions)
- Complex multi-part work (requires explicit planning)

**Assessment workflow:**
1. Question the premise
2. Check existing state
3. Identify trade-offs
4. Flag problems upfront
5. Present assessment and request direction

**Approval mechanisms:**
- User input context: Wait for explicit "proceed" approval
- Autonomous context: Select best option and continue

