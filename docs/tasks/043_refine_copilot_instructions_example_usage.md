# Refine Copilot Instructions Regarding Example Usage

**Status:** Not Started  
**Created:** 2025-12-28

## Description

Copilot instructions contain clear rules about example usage in smaqit framework content:
- **DO**: "Prefer abstract categories over specific examples"
- **DON'T**: "Put examples or extended explanations in template/agent files"

These rules are not being consistently respected, leading to specific examples appearing in framework files, templates, and agents where they should use abstract categories and generic placeholders instead.

Need to strengthen and clarify these rules in `.github/copilot-instructions.md` to prevent example pollution in framework artifacts.

## Acceptance Criteria

- [ ] Updated copilot instructions with strengthened guidance on example usage
- [ ] Clear distinction between allowed examples (HTML comments in prompts) and prohibited examples (framework files, templates, agents)
- [ ] Explicit rule about generic placeholders (`[PLACEHOLDER]`) vs specific examples (BUS-LOGIN-001, authentication, etc.)
- [ ] Cleanup of existing specific examples in framework files, templates, and agents
- [ ] Exception documented: HTML comment examples in prompts (like mario-hello test case) are allowed and intentional
- [ ] Validation that framework/, templates/, and agents/ directories use only abstract categories and generic placeholders

## Notes

**Current problems:**
- Framework files may contain specific technology/architecture examples that should be abstract
- Templates may have leftover specific examples instead of placeholders
- Agents may contain example pollution from copy-paste refinements

**Allowed examples:**
- HTML comments in prompts: `<!-- Example: "Mario Fan - Users who love Nintendo's Mario franchise" -->`
- Test cases: `docs/test-cases/mario-hello.md` as demonstration scenario
- Session history: Examples in history files for documentation purposes

**Prohibited examples:**
- Specific requirement IDs in framework/template/agent files (except when demonstrating format structure)
- Technology-specific examples (JWT, React, AWS, etc.) in layer definitions
- Architecture patterns (microservices, containers, etc.) in abstract guidance
- Business domain examples (login, checkout, etc.) in reusable templates

The goal is to ensure framework content remains maximally reusable across all project types without prescribing specific solutions.
