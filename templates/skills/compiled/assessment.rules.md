---
target: templates/skills/assessment/SKILL.md
sources:
  - framework/SKILLS.md (Assessment Skill section)
  - .github/prompts/session.assess.prompt.md
created: 2026-02-06
---

# Assessment Skill Rules

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| framework/SKILLS.md | Assessment Skill > Purpose |
| framework/SKILLS.md | Assessment Skill > Workflow |
| framework/SKILLS.md | Assessment Skill > Invocation Triggers |
| framework/SKILLS.md | Assessment Skill > Approval Mechanisms |
| framework/SKILLS.md | Assessment Skill > Output Format |

## L1 Directive Compilation

### Frontmatter

MUST set name to: `assessment`

MUST set description to include these trigger keywords:
- "critical assessment"
- "ambiguous requirements"
- "conflicting inputs"
- "insufficient detail"
- "complex planning"
- "approval gate"

### Purpose

MUST state that assessment prevents wasted execution on poor-quality inputs.

MUST list assessment capabilities:
- Critical evaluation of input quality and completeness
- Identification of assumptions, gaps, and conflicts
- Generation of execution plan with explicit steps
- User review and approval mechanism
- Iterative refinement based on feedback

### Workflow

MUST document five-step workflow:

1. **Question the premise**
   - Evaluate necessity
   - Check for duplication
   - Identify hidden assumptions
   - Question maintenance burden

2. **Check existing state**
   - Read relevant files first
   - Search for similar patterns
   - Verify problem exists as described
   - Empirically verify without guessing

3. **Identify trade-offs**
   - List downsides of proposed approach
   - Compare alternatives
   - Analyze cost-benefit
   - Consider token/complexity/maintenance costs

4. **Flag problems upfront**
   - State flaws clearly before proceeding
   - Surface redundancy or conflicts
   - Identify better approaches
   - Question if simpler solutions suffice

5. **Present assessment and request direction**
   - Summarize findings with rationale
   - Present alternatives with analysis
   - Request confirmation before proceeding
   - Challenge incomplete framing

### Stop and Explain Risks

MUST define risk scenarios requiring explicit halt and explanation:
- Modifies user configuration files (dotfiles, shell configs)
- Violates security best practices
- Breaks established conventions without clear justification
- Could affect system stability or user experience negatively
- Duplicates existing functionality in another location

MUST stop immediately and explain risks explicitly before proceeding with high-impact changes.

### Invocation Triggers

MUST detect and invoke assessment when:
- Ambiguous requirements: multiple valid interpretations exist
- Conflicting inputs: contradictions within or across sources
- Insufficient detail: cannot proceed without assumptions
- Complex multi-part work: requires explicit planning for coordination

### Approval Mechanisms

MUST adapt approval to invocation context:

**User input context:**
- Wait for explicit approval before proceeding
- Accept "proceed" to continue
- Accept "revise" to modify requirements
- Handle clarification that triggers reassessment

**Autonomous context:**
- Evaluate available options
- Select most appropriate based on context
- Continue execution without waiting
- Log assessment results for user review

### Output Format

MUST return structured results including:
- Summary of premise evaluation
- Current state findings (verified, not assumed)
- Trade-off analysis with recommendations
- Flagged problems requiring attention
- Proposed execution plan with explicit steps
- Approval status (user confirmed or autonomous selection)

## Compilation Guidance for Agent-L2

When compiling assessment skill:

1. Merge base-skill.template.md structure with base.rules.md directives
2. Apply assessment.rules.md specific directives
3. Replace [SKILL_NAME] with `assessment`
4. Replace [SKILL_DESCRIPTION] with description containing all trigger keywords
5. Replace [SKILL_TITLE] with `Assessment`
6. Replace [SKILL_BODY] with purpose statement
7. Replace [WORKFLOW_STEPS] with five-step numbered workflow (detailed instructions for each step)
8. Replace [OUTPUT_FORMAT] with structured output specification (all required fields)
9. Validate: frontmatter complete, workflow steps actionable, output format clear
10. Create directory .github/skills/assessment/ if not exists
11. Write compiled SKILL.md to .github/skills/assessment/SKILL.md
