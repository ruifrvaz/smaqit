# Document User vs Agent Documentation Distinction

**Status:** Not Started  
**Created:** 2025-12-28  
**Source:** User Testing Report Issue #9 (2025-12-27)

## Description

Framework lacks clear guidance on what content belongs in user-facing documentation vs agent-facing specifications. This principle exists implicitly (session 012 split instructions from rationale) but needs explicit documentation.

## Acceptance Criteria

- [ ] Create wiki entry: `docs/wiki/concepts/user-vs-agent-documentation.md`
- [ ] Document explains why the distinction matters in spec-driven development
- [ ] Clearly defines what content belongs in each category:
  - **Agent specifications** (specs, framework, agents, templates): Pure execution instructions, no meta-rationale, no human context, no explanations
  - **User documentation** (wiki, README, task files): Can include business context, stakeholder names, delivery dates, rationale, examples, explanations
- [ ] Provides examples of inappropriate content in specs:
  - Stakeholder names
  - Delivery dates
  - Business politics
  - Historical context
  - "Why" explanations
- [ ] Explains how this enables LLM agents to focus on execution without human context noise
- [ ] References relevant framework principles (separation of concerns, explicit over implicit)

## Impact

**Severity:** Medium  
**User Impact:** Without explicit guidance, contributors may pollute specs with human context; reduces agent effectiveness; violates separation of concerns

## Notes

This is a fundamental principle that should be documented for contributors and users. Helps maintain clean separation between human context and agent instructions.
