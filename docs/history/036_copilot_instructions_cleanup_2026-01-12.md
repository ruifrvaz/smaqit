# Copilot Instructions Cleanup

**Date:** January 12, 2026  
**Session Focus:** Comprehensive cleanup of copilot-instructions.md after Level agent completion  
**Tasks Referenced:** B001, B001.3

## Session Context

After completing all three Level agents (B001.1, B001.2, B001.3), we performed a comprehensive review and cleanup of `.github/copilot-instructions.md` to eliminate redundancies and clarify agent routing instructions.

## Actions Taken

### 1. Identified Redundant Content
- Example Usage Rules section (duplicated in agent MUST NOT directives)
- Level Transition Workflow implementation details (duplicated in agent directives)
- Level-Specific Work Routing incomplete section (partially redundant with Level Agents summary)
- General guidelines in "When documenting framework concepts" (duplicated in agent directives)

### 2. Removed Redundant Sections
- **Example Usage Rules** — Deleted 32+ lines of anchor bias prevention rules now embedded in all three Level agents
- **Level Transition Workflow details** — Removed 85+ lines of implementation and refinement workflows now in agent directives
- **Level-Specific Work Routing incomplete section** — Removed partial routing table that was redundant with individual "When Editing" sections
- **General guidelines** — Removed placeholder rules, rationale guidance duplicating agent content

### 3. Simplified Agent Routing
Converted all "When Editing X" sections to consistent, concise pattern:
```markdown
### When Editing [Component Type]

**Invoke Agent-LX** (`.github/agents/smaqit.LX.agent.md`) for [description] (`path/pattern`).
```

Applied to:
- When Documenting Framework Concepts → Agent-L0
- When Editing Agent Templates → Agent-L1
- When Editing Agents → Agent-L2
- When Editing Prompt Templates → Agent-L1
- When Editing Prompt Files → Agent-L1

### 4. Clarified Terminology
- **Development agents** — Clarified as meta-agents in `.github/agents/` (4 files: L0, L1, L2, orchestrator), distinct from product development agent in `agents/`
- **Prompt files** — Clarified dual nature: L1 templates in smaqit repo (`prompts/`), become user input records after installation in user projects
- Added note: "In the smaqit repo, prompt files are L1 templates with structure and guidance comments. Users fill these with concrete requirements at their product"

### 5. Maintained Contextual Structure
Preserved "When Editing X" sections as contextual triggers rather than compiling into single table, because:
- Instructions appear at point of need when agent is working on specific component
- No lookup required — immediate routing guidance
- More valuable than generic routing table

## Problems Solved

### Terminology Confusion
**Problem:** "Development agents" conflated with product development agent  
**Resolution:** Explicitly listed `.github/agents/` contains 4 meta-agents (L0, L1, L2, orchestrator), while `agents/` contains 9 product agents

### Prompt File Misconception
**Problem:** Described prompt files as "user-facing input records"  
**Resolution:** Clarified they are "L1 templates with structure and guidance" in smaqit repo, only become user input after installation

### Example Usage Rules Duplication
**Problem:** 32+ lines of anchor bias prevention rules duplicated between copilot-instructions.md and all three Level agents  
**Resolution:** Removed from copilot-instructions.md, retained in agents where LLMs execute them

### Level Transition Workflow Duplication
**Problem:** 85+ lines of workflow details duplicated between copilot-instructions.md and agent directives  
**Resolution:** Removed from copilot-instructions.md, retained in agents as executable directives

## Decisions Made

### Keep Contextual "When Editing" Sections
**Decision:** Maintain separate "When Editing X" sections rather than compile into single routing table  
**Rationale:** Contextual triggers provide routing instruction at point of need, eliminating lookup overhead

### Remove L0 Reference Section
**Decision:** Remove "When documenting framework concepts" general guidelines, keep only agent invocation  
**Rationale:** copilot-instructions.md provides L0 routing reference for agents working on smaqit itself, not human-oriented documentation

### Simplify All Routing to Invoke Pattern
**Decision:** Convert all routing instructions to consistent "**Invoke Agent-LX**" pattern  
**Rationale:** Uniform structure makes agent routing predictable and scannable

## Files Modified

- `.github/copilot-instructions.md` — Reduced from 343 to ~185 lines
  - Removed Example Usage Rules section (lines 119-151)
  - Removed Level Transition Workflow details (lines 218-303)
  - Simplified all "When Editing" sections to agent invocations
  - Fixed prompt file description from "Input record files" to "Prompt templates with structure and guidance"
  - Removed incomplete "Level-Specific Work Routing" section
  - Added note about prompt files being L1 templates in smaqit repo

## Session Metrics

- **Duration:** ~45 minutes
- **Files Modified:** 1
- **Lines Reduced:** 158 (343 → 185, 46% reduction)
- **Sections Removed:** 3 major sections (Example Usage Rules, Level Transition Workflow, Level-Specific Work Routing)
- **Sections Simplified:** 6 "When Editing" sections
- **Terminology Clarifications:** 2 (development agents, prompt files)
- **Tasks Progressed:** B001 (all subtasks complete, parent task nearing completion)

## Next Steps

- Monitor whether simplified routing instructions are sufficient in practice
- Validate that Level agents contain all necessary directives previously in copilot-instructions.md
