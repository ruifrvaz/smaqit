# Document User vs Agent Documentation Distinction

**Status:** Completed  
**Created:** 2025-12-28  
**Completed:** 2025-12-28  
**Source:** User Testing Report Issue #9 (2025-12-27)

## Description

Framework lacks clear guidance on what content belongs in user-facing documentation vs agent-facing specifications. This principle exists implicitly (session 012 split instructions from rationale) but needs explicit documentation.

## Acceptance Criteria

- [x] Create wiki entry: `docs/wiki/concepts/user-vs-agent-documentation.md`
- [x] Document explains why the distinction matters in spec-driven development
- [x] Clearly defines what content belongs in each category:
  - **Agent specifications** (specs, framework, agents, templates): Pure execution instructions, no meta-rationale, no human context, no explanations
  - **User documentation** (wiki, README, task files): Can include business context, stakeholder names, delivery dates, rationale, examples, explanations
- [x] Provides examples of inappropriate content in specs:
  - Stakeholder names
  - Delivery dates
  - Business politics
  - Historical context
  - "Why" explanations
- [x] Explains how this enables LLM agents to focus on execution without human context noise
- [x] References relevant framework principles (separation of concerns, explicit over implicit)

## Impact

**Severity:** Medium  
**User Impact:** Without explicit guidance, contributors may pollute specs with human context; reduces agent effectiveness; violates separation of concerns

## Notes

This is a fundamental principle that should be documented for contributors and users. Helps maintain clean separation between human context and agent instructions.

## Completion Summary

**Date:** 2025-12-28

**Work Completed:**

1. Created comprehensive wiki document at `docs/wiki/concepts/user-vs-agent-documentation.md`
2. Document includes:
   - Clear definition and purpose
   - Detailed audience breakdown (agent specs vs user docs)
   - "How It Works" with practical examples
   - "Why This Distinction Matters" for both agents and humans
   - Five detailed examples of inappropriate content with before/after comparisons
   - "How This Enables LLM Effectiveness" section
   - Validation patterns for both document types
   - Trade-offs analysis
   - Relation to other principles
   - Migration guide for fixing violations
3. Updated README.md to reference new wiki concept
4. Updated copilot-instructions.md to reference new wiki concept
5. Moved task 040 from Active to Completed in PLANNING.md

**Files Modified:**
- `docs/wiki/concepts/user-vs-agent-documentation.md` (created)
- `README.md` (added reference)
- `.github/copilot-instructions.md` (added reference)
- `docs/tasks/PLANNING.md` (moved to completed)
- `docs/tasks/040_document_user_vs_agent_documentation_distinction.md` (marked completed)

All acceptance criteria met.
