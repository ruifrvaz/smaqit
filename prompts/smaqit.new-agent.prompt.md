---
name: smaqit.new-agent
description: Interactive template for creating new base agents through Agent-L2
agent: smaqit.L2
---

# New Agent Creation

This prompt provides the structure for Agent-L2 to gather agent specifications interactively when creating new base agents (Q&A, helper, orchestrator, custom utilities).

Agent-L2 follows this structure and requests user input for each placeholder. User-provided specifications are documented in the compilation log, NOT stored in this file (this remains a reusable template).

## Agent Identity

### Name
[Agent name without smaqit prefix, e.g., "qa", "helper"]

<!-- Agent-L2 requests: "What is the agent name? (lowercase, single word or hyphenated)" -->

### Description
[One-line description for agent frontmatter]

<!-- Agent-L2 requests: "What is the agent description? (clear, concise, under 80 chars)" -->

### Tools
[Comma-separated list: fetch, search, read, edit, runCommands, etc.]

<!-- Agent-L2 requests: "What tools does this agent need? (comma-separated: fetch, search, read, edit, runCommands)" -->
<!-- Example: "fetch, search, read" -->

## Agent Purpose

### Goal
[What does this agent produce or accomplish?]

<!-- Agent-L2 requests: "What is this agent's goal?" -->
<!-- Example: "Fetch and answer questions about smaqit documentation" -->

### Context
[What constraints or scope define this agent?]

<!-- Agent-L2 requests: "What constraints or scope define this agent?" -->
<!-- Example: "Read-only operations, documentation focus, no code generation" -->

## Input Sources

[What does this agent read or consume?]

<!-- Agent-L2 requests: "What input sources does this agent read?" -->
<!-- Example: "User questions" -->
<!-- Example: "Wiki URLs from https://github.com/ruifrvaz/smaqit/blob/main/docs/wiki/" -->
<!-- Example: "Local framework files from framework/ directory" -->

## Output Format

[What does this agent produce?]

<!-- Agent-L2 requests: "What output format does this agent produce?" -->
<!-- Example: "Direct answers with source references" -->
<!-- Example: "Generated configuration files in JSON format" -->

## Specialized Directives

### MUST
[Agent-specific mandatory behaviors]

<!-- Agent-L2 requests: "What are the agent-specific MUST directives? (imperative statements, one per line)" -->
<!-- Example: "Fetch wiki content from GitHub when local files not available" -->
<!-- Example: "Provide source references for all answers" -->
<!-- Example: "Redirect implementation questions to appropriate agents" -->

### MUST NOT
[Agent-specific prohibitions]

<!-- Agent-L2 requests: "What are the agent-specific MUST NOT directives? (prohibitions, one per line)" -->
<!-- Example: "Generate code or implementation" -->
<!-- Example: "Create specifications or requirements" -->
<!-- Example: "Modify any files" -->

### SHOULD
[Agent-specific recommendations]

<!-- Agent-L2 requests: "What are the agent-specific SHOULD directives? (recommendations, one per line)" -->
<!-- Example: "Prefer local files over remote fetch when available" -->
<!-- Example: "Include multiple source references when relevant" -->

## Scope Boundaries

[What is explicitly out of scope? Include redirections to other agents.]

<!-- Agent-L2 requests: "What is explicitly out of scope for this agent? (include redirections)" -->
<!-- Example: "Code generation → redirect to Development agent" -->
<!-- Example: "Spec creation → redirect to layer-specific spec agents (Business, Functional, Stack, Infrastructure, Coverage)" -->
<!-- Example: "File modifications → read-only operations only" -->

## Completion Criteria

[Agent-specific validation checks beyond foundation criteria]

<!-- Agent-L2 requests: "What are the agent-specific completion criteria? (testable checks)" -->
<!-- Example: "Answer addresses user's question directly" -->
<!-- Example: "At least one source reference provided" -->
<!-- Example: "Out-of-scope questions redirected with specific agent suggestion" -->

## Failure Scenarios

[Agent-specific failure cases beyond foundation patterns - provide as table]

<!-- Agent-L2 requests: "What are the agent-specific failure scenarios? (situation | action pairs)" -->
<!-- Example table:
| Situation | Action |
|-----------|--------|
| Wiki content not found | Respond with: "Documentation not found for [topic]. Available sections: [list]" |
| Ambiguous question | Request clarification: "Did you mean [interpretation A] or [interpretation B]?" |
| Implementation question | Redirect: "This requires code generation. Please invoke Development agent." |
-->

---

## Guidelines for Agent-L2

**When gathering specifications:**
1. Request each section's content systematically
2. Validate user directives for compatibility with foundation directives (base.rules.md)
3. Ensure directives are specific and measurable
4. Verify scope boundaries are clear with appropriate redirections
5. Confirm completion criteria are testable
6. Document all user inputs in compilation log

**Specialized Directives:**
- Should be imperative statements starting with verbs
- Should be agent-specific (not duplicating base.rules.md)
- MUST directives should be compatible with foundation MUST directives
- MUST NOT directives should not conflict with foundation behaviors

**Scope Boundaries:**
- Clearly define what this agent does NOT do
- Provide explicit redirection to appropriate agents for out-of-scope work
- Maintain single responsibility principle

**Completion Criteria:**
- Define validation checks specific to this agent's output
- Make criteria testable (observable/measurable)
- Complement foundation criteria from base.rules.md

**Failure Scenarios:**
- Provide agent-specific failure cases (foundation patterns already in base.rules.md)
- Use table format: Situation | Action
- Include specific error messages and redirection patterns
