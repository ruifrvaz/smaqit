# Task 026: Rethink Prompt Architecture and Integration

**Status:** New  
**Created:** 2025-12-20  
**Priority:** High

## Problem Statement

Current prompt files (`prompts/*.prompt.md`) overlap significantly with agent definitions (`agents/*.agent.md`), creating redundancy and architectural confusion:

**Observed Issues:**
1. Prompt files contain instructions that duplicate agent behavior
2. `${input:variableName}` syntax doesn't trigger input collection (may not be supported)
3. Prompts invoke agents but don't add clear value beyond agent invocation
4. Prompt maintenance creates double burden (update agent AND prompt)
5. Unclear what prompts provide that agents alone couldn't

**Current State:**
- 8 prompt files mirror 8 agents
- Prompts contain framework references and instructions that agents already have
- Prompts use `agent: smaqit.business` to delegate to agents
- No clear value proposition for prompts as separate artifacts

## Questions to Answer

### 1. GitHub Copilot Prompt Capabilities
- What is the actual purpose and capability of `.prompt.md` files in GitHub Copilot?
- Do they support input collection? If so, what's the correct syntax?
- What features do prompts provide that agents don't?
- Is there official documentation for prompt file format?

### 2. smaqit Integration Model
- Should smaqit use prompts at all, or just direct chat messages?
- If prompts add value, what specific value do they provide?
- How should prompts and agents divide responsibilities?
- What's the correct user experience for invoking smaqit functionality?

### 3. Architectural Patterns
- **Option A: Prompts as Input Collectors** - Prompts gather requirements, agents generate specs
- **Option B: Prompts as Orchestrators** - Prompts coordinate multi-agent workflows (phase prompts)
- **Option C: Prompts as Shortcuts** - Thin wrappers for common agent invocations with context
- **Option D: No Prompts** - Remove prompts entirely, users type requirements directly in chat
- **Option E: Agents Only** - Remove prompts, document how to interact with agents directly

## Investigation Steps

### Step 1: Research GitHub Copilot Prompts
- [ ] Search for official GitHub Copilot prompt documentation
- [ ] Test various `.prompt.md` syntax patterns in test project
- [ ] Test conversational input collection (multi-turn chat)
- [ ] Identify supported features (variables, input collection, tools, etc.)
- [ ] Validate hypothesis: prompts = "Follow instructions in {file}"
- [ ] Document actual vs assumed capabilities

### Step 2: Test Input Collection Patterns
- [ ] Test: Prompt asks questions, user responds, prompt formats for agent
- [ ] Test: Prompt provides template, user fills in, prompt validates
- [ ] Test: Prompt offers examples, user chooses/modifies
- [ ] Document which pattern works best for smaqit requirements
- [ ] Measure friction: how many interactions needed?

### Step 3: Redesign Prompt Architecture
- [ ] Rewrite one prompt (e.g., business) with input collection focus
- [ ] Remove agent instructions from prompt (keep in agent only)
- [ ] Test redesigned prompt in test project
- [ ] Compare UX: old vs new prompt design
- [ ] Decide if redesign is worthwhile or if current approach is better

### Step 3: Redesign Prompt Architecture
- [ ] Rewrite one prompt (e.g., business) with input collection focus
- [ ] Remove agent instructions from prompt (keep in agent only)
- [ ] Test redesigned prompt in test project
- [ ] Compare UX: old vs new prompt design
- [ ] Decide if redesign is worthwhile or if current approach is better

### Step 4: Analyze Value Proposition
- [ ] List what prompts currently do
- [ ] List what prompts SHOULD do (input collection model)
- [ ] List what agents currently do (should remain unchanged)
- [ ] Identify overlaps and gaps in current design
- [ ] Determine effort required to redesign 8 prompts
- [ ] Assess ROI: does input collection justify redesign cost?

### Step 5: Define Architecture Decision
Based on investigation, choose one pattern:

**Pattern A: Prompts for Input Collection (NEW HYPOTHESIS)**
- Prompts collect and format user requirements through conversation
- Prompts provide examples, templates, validation
- Agents receive clean, structured input
- Clear separation: prompts = UX, agents = logic
- Requires: Redesign all 8 prompts to focus on input collection

**Pattern B: Keep Current Design with Documentation**
- Accept that prompts invoke agents with framework context
- Document that prompts + agents work together as system
- Improve current prompts to reduce duplication
- Simpler than full redesign, maintains working state

**Pattern C: Simplify Prompts to Minimal Wrappers**
- Prompts just invoke agents with user message as context
- Remove all instructions from prompts (keep in agents)
- Prompts become thin command wrappers
- Reduces duplication without full redesign

**Pattern D: Remove Prompts Entirely (IF POSSIBLE)**
- Delete all prompt files
- Document: "Interact with agents directly in chat with requirements"
- Simplifies architecture, reduces maintenance
- Only viable if agents can work without prompt wrappers

**Pattern E: Hybrid - Phase Prompts Only**
- Remove 5 layer prompts (too simple, just invoke one agent)
- Keep 3 phase prompts (orchestrate multiple agents, add value)
- Reduces duplication while keeping orchestration capability

## Acceptance Criteria

- [ ] GitHub Copilot prompt capabilities documented with evidence
- [ ] User experience tested for both prompts and direct agent invocation
- [ ] Value proposition analysis completed (overlap/gap identification)
- [ ] Architecture decision made with clear rationale
- [ ] Implementation path defined (what changes to make)
- [ ] `.github/copilot-instructions.md` updated with prompt guidance (or removal notice)
- [ ] Task 001 reassessed (may need to be reverted or redefined)

## Dependencies

- Blocks: Clear user onboarding documentation (need to know how users should invoke smaqit)
- Blocks: Task 024 completion (testing agent needs to know whether to test prompts)

## Notes

**Evidence from Testing:**
- `/smaqit.business` invokes agent but doesn't collect input (auto-generates from context)
- `@` syntax is NOT for invoking agents (confirmed)
- Agents are invoked via prompts using `agent:` frontmatter property
- Testing checklist file exists in test project, generated by testing agent

**Key Question:**
How are agents meant to be invoked if not via prompts? Are prompts the ONLY way to invoke agents in GitHub Copilot?

**Architectural Hypothesis (2025-12-20):**
Prompts act as embedded commands that translate to: "Follow instructions in {prompt file}."

This suggests a clear separation of concerns:
- **Prompts** = Input collection + formatting layer (UX)
- **Agents** = Execution logic + spec generation (business logic)

The model:
```
User: /smaqit.business
  ↓
Copilot: "Follow instructions in smaqit.business.prompt.md"
  ↓
Prompt: Collect business requirements, format them, invoke agent
  ↓
Agent: Receive structured input, generate specs per template
```

**Current Problem:**
Prompts duplicate agent instructions instead of focusing on input collection. Need to redesign prompts to:
1. Explain what input is needed (with examples)
2. Collect requirements through natural conversation
3. Format requirements for agent consumption
4. Pass clean, structured input to agent

**Agent Role (unchanged):**
1. Receive structured input from prompt
2. Read framework files for rules
3. Read templates for structure
4. Generate specifications
5. Validate output

This creates complementary, non-overlapping responsibilities.

**Historical Context:**
- Task 001 created prompt files based on assumed capability
- No official documentation found for GitHub Copilot `.prompt.md` format
- Implementation proceeded on hypothesis, now testing reveals issues

**Impact:**
- If prompts removed: delete 8 files, update installer, update docs
- If prompts redefined: redesign 8 files, may need agent changes too
- If prompts kept as-is: need to justify redundancy, improve value prop

