# Iterative Assessment Before Spec Generation and Implementation

**Status:** Complete  
**Created:** 2026-02-01  
**Updated:** 2026-02-07

## Description

Agents invoke assessment skill when detecting conditions requiring critical analysis and planning before execution. Assessment skill provides iterative evaluation with approval gates, preventing wasted effort on ambiguous, contradictory, or insufficient inputs.

**Approach:** Assessment as reusable agent skill (`.github/skills/assessment/`) rather than mandatory gate for all agent invocations. Agents automatically invoke when detecting specific conditions.

## Acceptance Criteria

- [x] L0 concepts documented in framework files
- [x] L1 skill template created following agentskills.io specification
- [x] L1 base skill template and compilation rules created
- [x] L1 assessment skill compilation rules created
- [x] session.assess prompt simplified to invoke assessment skill
- [x] L2 assessment skill compiled into `.github/skills/assessment/SKILL.md`
- [x] L2 all product agents updated with assessment skill directives

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

### 2026-02-06: L1 Templates and Rules Created

**Skill template structure (Agent-L1):**
- Created `templates/skills/base-skill.template.md` with generic skill structure
- Created `templates/skills/compiled/base.rules.md` with foundation skill directives
- Created `templates/skills/compiled/assessment.rules.md` with assessment-specific directives
- Updated `.github/prompts/session.assess.prompt.md` to invoke assessment skill

**L1 Content Added:**
- Base skill template with placeholders: [SKILL_NAME], [SKILL_DESCRIPTION], [SKILL_TITLE], [SKILL_BODY], [WORKFLOW_STEPS], [OUTPUT_FORMAT]
- Base compilation rules: frontmatter requirements, body content directives, progressive disclosure guidelines
- Assessment compilation rules: five-step workflow, invocation triggers, approval mechanisms, output format
- Compilation guidance for Agent-L2 merge process

**Next Steps:**
1. ~~Invoke Agent-L2 to compile skill: base template + base rules + assessment rules → `.github/skills/assessment/SKILL.md`~~ ✅ Complete
2. Invoke Agent-L2 to update all 8 product agents with assessment skill invocation directives

### 2026-02-07: L2 Assessment Skill Compiled

**Skill compilation (Agent-L2):**
- Created `.github/skills/assessment/SKILL.md` (64 lines)
- Performed 3-way merge: base template + base rules + assessment rules
- Replaced 6 placeholders with assessment-specific content
- Validated agentskills.io specification compliance
- Created compilation log in `.smaqit/logs/assessment-compilation-2026-02-07.md`

**Compiled skill characteristics:**
- Name: `assessment` (matches directory name)
- Description: 236 chars with all 6 trigger keywords ("critical assessment", "ambiguous requirements", "conflicting inputs", "insufficient detail", "complex planning", "approval gate")
- Workflow: Five-step process with detailed instructions per step
- Output: Structured format with 6 required components
- Progressive disclosure: 64 lines (well under 500 line limit)

**Validation results:**
- ✅ No placeholders remain
- ✅ Frontmatter complete (name and description required fields)
- ✅ Name constraints satisfied (lowercase, alphanumeric + hyphens)
- ✅ Description constraints satisfied (1-1024 char limit)
- ✅ All trigger keywords present in description
- ✅ Workflow steps actionable
- ✅ Output format clear

**Next step:**
Update 8 product agents with assessment skill invocation directives (blend into agent instructions, triggered by skill keywords)

**Agent updates (Agent-L2 - 2026-02-07):**
- Added assessment skill invocation directive to all 8 product agents' Failure Handling tables
- Directive: "Ambiguous, conflicting, insufficient, or complex inputs | Invoke `.github/skills/assessment/` for critical assessment"
- Uses 4 of 6 trigger keywords directly for automatic detection
- Blends seamlessly into existing failure handling workflow
- Agents updated: business, functional, stack, infrastructure, coverage, development, deployment, validation

**Implementation complete:**
All acceptance criteria met. Assessment skill is now functional and integrated across all product agents. Agents automatically invoke skill when detecting ambiguous requirements, conflicting inputs, insufficient detail, or complex planning scenarios.

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

