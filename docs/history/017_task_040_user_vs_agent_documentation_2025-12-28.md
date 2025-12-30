# Session 017: User vs Agent Documentation Distinction

**Date:** 2025-12-28  
**Task:** 040 - Document user vs agent documentation distinction  
**Previous Session:** [016_task_015_framework_embedding_2025-12-27.md](016_task_015_framework_embedding_2025-12-27.md)

## Objective

Document the principle that separates user-facing documentation from agent-facing specifications. This distinction exists implicitly (established in sessions 012 and 027) but lacks explicit documentation, leading to potential contributor confusion about where content belongs.

## Context

Task 40 originated from User Testing Report Issue #9 (2025-12-27), which identified that the framework lacks clear guidance on what content belongs in user-facing vs agent-facing documentation. Without explicit guidance:
- Contributors may pollute specs with human context
- Agent effectiveness reduced by unnecessary context in prompt
- Separation of concerns violated

## Work Done

### 1. Session Recap

Executed full session recap per `.github/prompts/session.recap.prompt.md`:
- Read all 8 framework files (SMAQIT.md, LAYERS.md, PHASES.md, TEMPLATES.md, AGENTS.md, ARTIFACTS.md, PROMPTS.md)
- Read 3 most recent history files (002, 003, 004)
- Read PLANNING.md to understand task context
- Examined existing wiki structure and style

### 2. Created Comprehensive Wiki Document

Created `docs/wiki/concepts/user-vs-agent-documentation.md` (413 lines, 42 sections):

**Major sections:**
- **Definition** — Clear statement of the principle
- **The Two Audiences** — Detailed breakdown of agent specs vs user docs with file lists and content rules
- **How It Works** — Practical examples showing the separation in practice
- **Why This Distinction Matters** — Benefits for both LLM agents and human contributors
- **Examples of Inappropriate Content** — Five detailed examples with before/after comparisons:
  1. Stakeholder names
  2. Historical context
  3. Delivery dates and business politics
  4. Extended examples as false requirements
  5. "Why" explanations in directives
- **Validation Patterns** — Questions to ask when reviewing each document type
- **Trade-offs** — Honest assessment of benefits and costs
- **How This Enables LLM Effectiveness** — Token optimization, reduced hallucination, clear success criteria
- **Relation to Other Principles** — Links to Explicit Over Implicit, Bounded Agents, Template Constraints
- **Migration Guide** — Step-by-step process for fixing violations

### 3. Updated References

- **README.md** — Added reference to new wiki concept in Documentation Structure section
- **.github/copilot-instructions.md** — Added reference in "When documenting framework concepts" section

### 4. Task Completion

- Updated `docs/tasks/040_document_user_vs_agent_documentation_distinction.md`:
  - Marked status as Completed
  - Added completion date
  - Marked all acceptance criteria as met
  - Added completion summary with files modified
- Updated `docs/tasks/PLANNING.md`:
  - Removed task 040 from Active table
  - Added task 040 to Completed table

## Key Decisions

### Document Structure Follows Existing Pattern

Followed the structure established in existing wiki concepts (bounded-agents.md, layer-independence.md, template-constraints.md):
- Definition → How It Works → Why It Matters → Trade-offs → Related
- Comprehensive examples with practical scenarios
- Links to related concepts
- Honest assessment of costs and benefits

### Five Categories of Inappropriate Content

Identified five specific categories that commonly appear inappropriately in specs:
1. **Stakeholder names** — Human identities don't belong in execution instructions
2. **Historical context** — Past decisions are rationale, not requirements
3. **Delivery dates and business politics** — Timeline/org dynamics are human context
4. **Extended examples** — Can contaminate specs if not in HTML comments
5. **"Why" explanations** — Rationale belongs in wiki, not directives

These categories provide concrete guidance for contributors.

### Emphasize LLM Effectiveness Benefits

Document explicitly explains how this separation improves LLM agent performance:
- **Token budget optimization** — More room for actual work
- **Reduced hallucination risk** — No false signals from examples/context
- **Clear success criteria** — Objective validation without subjective rationale

This makes the principle practical, not just philosophical.

### Migration Guide Included

Provided actionable 4-step process for fixing violations:
1. Identify content type
2. Extract to appropriate location
3. Simplify agent specification
4. Link between documents

This enables contributors to fix existing violations systematically.

## Files Modified

- `docs/wiki/concepts/user-vs-agent-documentation.md` (created, 413 lines)
- `README.md` (added reference)
- `.github/copilot-instructions.md` (added reference)
- `docs/tasks/PLANNING.md` (moved task 040 to completed)
- `docs/tasks/040_document_user_vs_agent_documentation_distinction.md` (marked completed with summary)
- `docs/history/017_task_040_user_vs_agent_documentation_2025-12-28.md` (this file)

## Lessons Learned

### Principle Was Already Practiced

The distinction between user and agent documentation was already established implicitly:
- Session 012: Split framework files from rationale
- Session 027: Separated framework instructions from human rationale
- Copilot instructions already had guidance (lines 110-121)

This task made explicit what was already implicit practice.

### Documentation Serves Two Masters

smaqit has a dual audience problem:
- **Agents** need crisp, actionable instructions (framework, agents, templates)
- **Humans** need context, rationale, and trade-offs (wiki, README, tasks)

Trying to serve both audiences in the same document creates noise for agents and insufficient context for humans. Separation solves both problems.

### Examples Are Dangerous

Extended examples in templates/specs risk contamination if agents don't ignore HTML comments. The wiki document provides clear guidance:
- Use generic placeholders in templates
- Wrap examples in `<!-- -->` comments
- Instruct agents to ignore HTML comments
- Provide real examples in wiki, not in agent-facing files

## Related Tasks

- **Task 027** (Completed) — Separate framework instructions from human rationale (established the pattern)
- **Task 028** (Completed) — Audit all smaqit levels for meta-rationale (cleaned up violations)
- **Task 037** (Active) — Clarify phase-first workflow in framework (similar documentation clarity task)
- **Task 039** (Active) — Add agent handover guidance (agent-facing improvement)

## Next Steps

Task 40 complete. Suggested next priorities from active tasks:
1. **Task 037** (CRITICAL) — Clarify phase-first workflow — addresses fundamental workflow misunderstanding
2. **Task 039** — Add agent handover guidance — improves user experience
3. **Task 038** — Add state.json validation — defensive validation for reliability

## Open Questions

None. All acceptance criteria met, work complete.
