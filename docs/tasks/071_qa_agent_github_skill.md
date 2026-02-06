# Task 071: Create Q&A Agent and GitHub Skill for Wiki Documentation

**Status:** new  
**Priority:** Medium  
**Created:** 2026-01-23

## Goal

Create two interfaces for users to query smaqit wiki documentation:
1. A product agent that ships with smaqit and uses the fetch tool to retrieve wiki content from GitHub
2. A GitHub Skill that provides the same capability through the skill interface

## Context

Users need quick access to smaqit documentation while working. Rather than manually navigating to wiki pages, they should be able to ask questions and get relevant documentation fetched automatically. This improves developer experience and reduces context switching.

## Deliverables

### Deliverable 1: Q&A Product Agent

**File:** `agents/smaqit.qa.agent.md`

**Requirements:**
- Agent name: `smaqit.qa`
- Description: "Fetch and answer questions about smaqit framework documentation"
- Tools: `["fetch", "search", "read"]`
- Capabilities:
  - Fetch wiki pages from `https://github.com/ruifrvaz/smaqit/blob/main/docs/wiki/`
  - Search local `docs/wiki/` for content
  - Read framework files from `framework/` directory
  - Answer questions by combining fetched content with local context
- Response format: Direct answers with references to source documentation
- Out-of-scope: Code generation, spec creation, implementation work

**SDK Compilation Pattern:** Use Pattern 1 (Base Agents) - 2-way merge
- **Sources:** `base-agent.template.md` + `base.rules.md`
- **No workflow extensions:** Q&A agent receives foundation directives only (9 MUST, 9 MUST NOT from base.rules.md)
- **Q&A-specific customizations:**
  - Role: Q&A agent identity, goal (fetch and answer documentation questions), context (read-only, documentation focus)
  - Input: User questions, wiki URLs, local framework files
  - Output: Answers with source references
  - Scope Boundaries: No implementation work, no spec generation, no code changes
  - Extension Directives (Q&A-specific, not from spec/impl workflows):
    - MUST fetch wiki content from GitHub when local not available
    - MUST provide source references for all answers
    - MUST redirect implementation questions to appropriate agents

**Compilation Workflow:**
1. Invoke Agent-L2 to compile Q&A agent using Pattern 1 (Base Agents)
2. Agent-L2 reads base-agent.template.md for structure
3. Agent-L2 reads base.rules.md for foundation directives
4. Agent-L2 fills template with Q&A-specific content (role, input, output, scope, extension directives)
5. Agent-L2 validates: no placeholders, self-contained, foundation directives embedded

**Frontmatter:**
```yaml
---
name: smaqit.qa
description: Fetch and answer questions about smaqit framework documentation
tools: ["fetch", "search", "read"]
---
```

### Deliverable 2: GitHub Skill

**File:** `.github/skills/smaqit-docs/SKILL.md`

**Requirements:**
- Skill name: `smaqit-docs`
- Description: "Query smaqit framework documentation from wiki and framework files"
- Same capabilities as product agent but packaged as GitHub Skill (agentskills.io format)
- Should work across any GitHub repository (not just smaqit projects)
- Fetch from public GitHub URLs: `https://raw.githubusercontent.com/ruifrvaz/smaqit/main/docs/wiki/`

**GitHub Skill Format (agentskills.io SKILL.md):**
```yaml
---
name: smaqit-docs
description: Query smaqit framework documentation from wiki and framework files
tools: ["fetch"]
---
```

## Acceptance Criteria

- [ ] `agents/smaqit.qa.agent.md` compiled using SDK Pattern 1 (Base Agents)
- [ ] Q&A agent contains foundation directives from base.rules.md (9 MUST, 9 MUST NOT)
- [ ] Q&A agent contains Q&A-specific customizations (role, input, output, scope, extension directives)
- [ ] Q&A agent has zero placeholders (validated by Agent-L2)
- [ ] `.github/skills/smaqit-docs/SKILL.md` created and follows agentskills.io specification
- [ ] Both interfaces can fetch wiki documentation from GitHub
- [ ] Both interfaces provide accurate answers with source references
- [ ] Q&A agent respects bounded agent principles (no implementation work)
- [ ] GitHub Skill works in repositories without local smaqit installation
- [ ] Documentation updated in README or wiki to mention Q&A capability
- [ ] Compilation documented (Agent-L2 invocation, sources used, validation performed)

## Implementation Notes

**SDK Pattern 1 (Base Agents):**
- Q&A agent uses 2-way merge: base-agent.template.md + base.rules.md
- No workflow extensions (specification.rules.md or implementation.rules.md)
- Foundation directives only: template-constrained output, fail-fast, self-validation, bounded scope, clarity
- Q&A-specific customizations added during compilation:
  - Role section: Agent identity (Q&A Agent), goal (fetch and answer docs), context (read-only)
  - Input section: User questions, wiki URLs, framework files
  - Output section: Answers with source references
  - Scope Boundaries: MUST NOT generate code, create specs, perform implementation work
  - Extension Directives: MUST fetch from GitHub, MUST provide references, MUST redirect out-of-scope questions

**Wiki URL Structure:**
- Base: `https://github.com/ruifrvaz/smaqit/blob/main/docs/wiki/`
- Raw content: `https://raw.githubusercontent.com/ruifrvaz/smaqit/main/docs/wiki/`
- Examples:
  - `docs/wiki/workflows/quickstart.md`
  - `docs/wiki/concepts/team-alignment.md`

**Agent Scope:**
- Read-only: fetch, search, read
- No editing, no code generation
- Focuses on documentation retrieval and comprehension
- Redirects implementation questions to appropriate agents

**GitHub Skill Scope:**
- Fetch-only (no local file access in external repos)
- Must construct proper GitHub raw content URLs
- Handle 404s gracefully when documentation doesn't exist

## Questions to Resolve

1. Should the Q&A agent also be able to read framework files directly, or only fetch from GitHub?
   - **Proposal:** Agent reads local `framework/` and `docs/wiki/` for smaqit projects, falls back to fetch for external contexts
   
2. Should the GitHub Skill be part of the installer scaffold, or only in smaqit repo?
   - **Proposal:** Keep skill in smaqit repo only (not scaffolded), Q&A agent gets scaffolded

3. What's the invocation pattern?
   - **Product agent:** `/smaqit.qa [question]` in Copilot chat
   - **GitHub Skill:** Invoked automatically when agents need documentation

4. ~~Should Q&A agent use specification agent template?~~
   - **Resolved:** Use SDK Pattern 1 (Base Agents) - 2-way merge with base template + base rules only

## Related Tasks

- Task 065: Clean Up Level 1 Templates (completed - established SDK compilation patterns)
- Task B001: Extensible Meta-Framework - Level Up Architecture (SDK foundation)
- Task 024: Create smaqit user testing agent (precedent for tool-using agents)
- Session 044: README streamlining and wiki documentation creation

## Success Metrics

- Users can ask "What are the five layers?" and get accurate answer with wiki references
- Users can ask "How do I start a new project?" and get quickstart guide
- Agent properly declines "Generate a business spec for login" (out of scope)
- GitHub Skill works from any repository

