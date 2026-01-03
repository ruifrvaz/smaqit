# Document Iterative Development Quick Start

**Status:** New  
**Created:** 2026-01-03  
**Revised:** 2026-01-03 (scope pruned to README update only)  
**Completed:** 2026-01-03  
**Related:** Task 014, Task 045, Task 047

## Description

Add a simple "Iterative Development" section to README.md explaining how users can leverage stateful specifications to work incrementally. This provides a practical getting-started guide without deep technical workflows.

## Context

Task 045 validated that stateful specifications infrastructure works (specs track state, CLI displays progress). However, incremental processing logic isn't implemented yet (Task 047). Before implementing incremental processing, users need simple guidance on how to work with smaqit iteratively using current capabilities.

**Scope clarification:** This task focuses on user-facing quick start guidance in README, not comprehensive workflow documentation. Detailed implementation workflows should be documented AFTER Task 047 implements incremental processing.

## Scope

### In Scope

**README.md Updates:**

Add "Iterative Development" section after "Getting Started" covering:

1. **Start small** - Begin with minimal specs, add features incrementally
2. **Track progress** - Use `smaqit status` to see what's implemented vs pending
3. **Add features** - Update prompt files, regenerate specs, implement new specs
4. **Fix failures** - Reprocess failed specs without regenerating everything
5. **Resume work** - Pick up where you left off across sessions

Keep it practical, concise, and user-friendly (2-3 paragraphs max).

### Out of Scope

- Comprehensive workflow documentation (defer until after Task 047)
- Framework file updates (agents don't need this until incremental processing exists)
- Wiki articles on implementation patterns (premature before Task 047)
- Deep technical explanations of state tracking mechanisms

## Acceptance Criteria

- [x] README.md has new "Iterative Development" section after "Getting Started"
- [x] Section explains how to start small and add features incrementally
- [x] Section mentions `smaqit status` for progress tracking
- [x] Section describes workflow for adding features (update prompt → regenerate specs → implement)
- [x] Section explains how to fix failed specs without full regeneration
- [x] Content is concise (2-3 paragraphs max) and user-friendly
- [x] No premature documentation of non-existent incremental processing logic

## Implementation Summary

Added "Iterative Development" section to README.md between "Getting Started" and "Commands" sections. Content covers:

1. **State tracking** - Specs track lifecycle (draft → implemented → deployed → validated)
2. **Working iteratively** - Four-step workflow: start minimal, track progress, add features, fix failures
3. **Current capabilities** - Manual progress tracking with `smaqit status`
4. **Future roadmap** - Automatic incremental processing in v0.6.0 (Task 047)

Section is concise (2 paragraphs), practical, and sets realistic expectations about current vs planned capabilities.

## Implementation Steps

### Step 1: Draft Content

Write clear, concise section covering:
- How stateful specs enable iterative work
- Using `smaqit status` to track progress
- Workflow for adding features incrementally
- Current capabilities vs future enhancements (Task 047)

### Step 2: Place in README

Add "Iterative Development" section after "Getting Started" and before "Commands"

### Step 3: Review

- Check tone is user-friendly and practical
- Verify no false promises about incremental processing (Task 047 not done yet)
- Ensure brevity (2-3 paragraphs max)
- Get user approval

## Notes

**Revised Scope:** Task originally proposed 3 wiki articles + 2 framework updates. Scope pruned to focus on simple README section only. Rationale:

- Incremental processing logic doesn't exist yet (Task 047)
- Documenting workflows for non-existent features creates aspirational documentation
- Users need practical getting-started guidance, not comprehensive technical workflows
- Existing wiki articles (`stateful-specifications.md`, `managing-stale-specs.md`) already adequate
- Detailed workflow documentation should follow implementation (Task 047), not precede it

**Current Capabilities (v0.5.0):**
- ✅ Specs track state via frontmatter
- ✅ CLI displays progress with `smaqit status`
- ✅ Agents update state after processing
- ❌ Agents don't check state before processing (Task 047)

**Future Capabilities (v0.6.0 - Task 047):**
- Agents will skip already-implemented specs
- Selective processing modes (incremental, force, resume)
- True incremental development without regeneration

README section should acknowledge current state while pointing toward future capabilities.

## Dependencies

**Blocks:**
- Task 047 (incremental processing implementation - benefits from user-facing guidance)

**Blocked By:**
- None (can document current capabilities independently)

**Related:**
- Task 014 (implemented stateful specs infrastructure)
- Task 045 (validated infrastructure, identified incremental gap)

## Success Metrics

- Users understand they can track progress with `smaqit status`
- Users know how to add features incrementally (update prompts → regenerate specs)
- Documentation sets realistic expectations about current vs future capabilities
- README remains concise and practical (not overwhelming new users)
