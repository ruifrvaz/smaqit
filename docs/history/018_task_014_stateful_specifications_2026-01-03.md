# Session: Task 014 - Stateful Specifications

**Date:** 2026-01-03  
**Task:** 014 - Define iterative development patterns  
**Status:** Phases 1-5 complete, testing and release pending

## Session Overview

Implemented stateful specifications system to enable incremental development in smaqit. Specs now track lifecycle state (draft → implemented → deployed → validated) via YAML frontmatter, allowing partial progress visibility and resumable workflows.

## Problem Statement

Original smaqit design assumed full regeneration on requirement changes. This created three issues:

1. **No progress visibility** - Couldn't see which specs were implemented vs pending
2. **Expensive iteration** - Adding one feature meant regenerating all specs
3. **Lost context** - No way to resume work across sessions

User insight: "smaqit is missing state concept on specs" (unlike tasks which have new/in-progress/completed).

## Solution Architecture

**State in specs (YAML frontmatter):**
```yaml
---
id: BUS-LOGIN-001
status: draft
created: 2025-01-03T14:23:00Z
prompt_version: abc123def
implemented: 2025-01-04T09:15:00Z
deployed: 2025-01-04T10:30:00Z
validated: 2025-01-04T11:45:00Z
---
```

**Aggregated phase state (state.json):**
```json
{
  "phases": {
    "develop": {
      "completed": true,
      "timestamp": "2025-01-04T12:00:00Z",
      "specs_processed": 15,
      "specs_succeeded": 14,
      "specs_failed": 1
    }
  }
}
```

**State lifecycle:**
- draft → implemented (development agent)
- implemented → deployed (deployment agent)
- deployed → validated (validation agent)
- Any state → failed (implementation agents)
- Any state → deprecated (user action)

## Implementation Phases

### Phase 1: Framework Updates ✓

Updated 5 core framework files:

1. **SMAQIT.md** - Added "Stateful Specifications" principle
2. **PHASES.md** - Added "Incremental Development" section, documented state.json schema
3. **ARTIFACTS.md** - Added comprehensive "Specification State" section with frontmatter schema and state transitions
4. **TEMPLATES.md** - Added frontmatter requirements to spec template section
5. **AGENTS.md** - Updated spec and impl agent directives for state tracking

### Phase 2: Template Updates ✓

Added YAML frontmatter to all 5 specification templates:
- business-specification.template.md
- functional-specification.template.md
- stack-specification.template.md
- infrastructure-specification.template.md
- coverage-specification.template.md

Each template now includes frontmatter with: id, status, created, prompt_version fields.

### Phase 3: Agent Updates ✓

**5 specification agents:**
- Removed redundant frontmatter directives (templates already contain structure)
- Removed example YAML blocks (directives only)
- Agents now follow templates automatically

**3 implementation agents:**
- Added structured state tracking sections (1-2-3 format)
- Development: Updates business/functional/stack specs to `status: implemented`
- Deployment: Updates infrastructure specs to `status: deployed`
- Validation: Updates coverage specs to `status: validated`
- All update state.json with specs_processed/succeeded/failed counts

**1 implementation agent template:**
- Updated with state tracking structure matching actual agents
- Uses placeholders for layer-specific behavior

### Phase 4: CLI Updates ✓

**installer/main.go changes:**

1. Extended `PhaseState` struct:
```go
type PhaseState struct {
    Completed       bool   `json:"completed"`
    Timestamp       string `json:"timestamp,omitempty"`
    SpecsProcessed  int    `json:"specs_processed,omitempty"`
    SpecsSucceeded  int    `json:"specs_succeeded,omitempty"`
    SpecsFailed     int    `json:"specs_failed,omitempty"`
}
```

2. Updated `initStateFile()` to initialize spec counts to 0

3. Modified `printPhaseStatus()` to display spec counts:
   - "✓ Complete (2025-01-04) - 15 processed, 14 succeeded, 1 failed"

### Phase 5: Wiki Documentation ✓

**New articles:**

1. **docs/wiki/concepts/stateful-specifications.md**
   - Core concept explanation
   - State lifecycle diagram
   - Frontmatter schema
   - Design rationale (why frontmatter vs external files)

2. **docs/wiki/workflows/managing-stale-specs.md**
   - Stale detection workflow (prompt_version tracking)
   - Regeneration decision criteria
   - Batch regeneration patterns
   - Future enhancement ideas

**Updated articles:**

3. **docs/wiki/patterns/prompt-evolution.md**
   - Added "State Tracking Integration" section
   - Links to stateful specs and stale management

4. **docs/wiki/workflows/amending-requirements.md**
   - Added "State Implications" section
   - State preservation rules
   - Re-implementation triggers
   - Progress tracking notes

## Key Decisions

### 1. Frontmatter over External State Files

**Chosen:** YAML frontmatter in specs (like tasks)

**Rejected:** Separate JSON state files per spec

**Rationale:**
- Colocation: State lives with content
- Simplicity: No database or state file management
- Git-friendly: Text-based, mergeable, diffable
- Self-contained: Specs carry their own state

**Trade-off accepted:** Slightly longer spec files

### 2. Aggregate to state.json for CLI Efficiency

**Chosen:** Phase-level counts in single state.json

**Rationale:**
- CLI can read one file for status command
- Quick overview without scanning all specs
- Unified phase completion tracking

**Implementation detail:** Agents update both spec frontmatter AND state.json

### 3. Prompt Version Tracking

**Chosen:** Git commit hash in `prompt_version` field

**Rationale:**
- Detects when prompt evolved beyond spec
- Enables stale spec detection
- User decides whether to regenerate (not automated)

**Deferred:** Automated stale detection tooling (future enhancement)

### 4. Templates Contain Structure, Agents Follow

**Principle reinforcement:** Level 1 templates define frontmatter structure, Level 2 agents don't duplicate this in directives.

**Action taken:** Removed redundant frontmatter generation directives from specification agents.

## Refinement Iterations

Multiple quality passes ensured clean implementation:

1. **Removed examples** - Agents had example YAML/JSON blocks, replaced with pure directives
2. **Template alignment** - Updated implementation agent template to match actual agents
3. **State tracking structure** - Restructured from verbose nested sections to clean 1-2-3 numbered steps
4. **Installer initialization** - Removed "create state.json" directives (installer handles it)

## Files Modified

**Framework (5 files):**
- framework/SMAQIT.md
- framework/PHASES.md
- framework/ARTIFACTS.md
- framework/TEMPLATES.md
- framework/AGENTS.md

**Templates (6 files):**
- templates/specs/business-specification.template.md
- templates/specs/functional-specification.template.md
- templates/specs/stack-specification.template.md
- templates/specs/infrastructure-specification.template.md
- templates/specs/coverage-specification.template.md
- templates/agents/implementation-agent.template.md

**Agents (8 files):**
- agents/smaqit.business.agent.md
- agents/smaqit.functional.agent.md
- agents/smaqit.stack.agent.md
- agents/smaqit.infrastructure.agent.md
- agents/smaqit.coverage.agent.md
- agents/smaqit.development.agent.md
- agents/smaqit.deployment.agent.md
- agents/smaqit.validation.agent.md

**CLI (1 file):**
- installer/main.go

**Wiki (4 files):**
- docs/wiki/concepts/stateful-specifications.md (new)
- docs/wiki/workflows/managing-stale-specs.md (new)
- docs/wiki/patterns/prompt-evolution.md (updated)
- docs/wiki/workflows/amending-requirements.md (updated)

**Total: 24 files modified/created**

## Remaining Work

### Phase 6: Testing (Not Started)

1. Build installer: `cd installer && make build`
2. Test init in clean directory
3. Verify templates have frontmatter
4. Verify status command displays spec counts
5. Validate state.json structure

### Phase 7: Release (Not Started)

1. Update CHANGELOG.md for v0.5.0
2. Update README.md with stateful specs explanation
3. Mark Task 014 as completed in docs/tasks/PLANNING.md
4. Create installer release build

## Next Session

**Immediate actions:**
1. Execute Phase 6 testing workflow
2. Fix any issues discovered
3. Complete Phase 7 release preparation
4. Close Task 014

**Testing commands:**
```bash
cd installer && make build
mkdir -p test && cd test
../dist/smaqit init
../dist/smaqit status
# Verify state.json structure
cat .smaqit/state.json
# Verify template frontmatter
cat .smaqit/templates/business-specification.template.md | head -20
```

## Session Metrics

- **Duration:** ~2 hours
- **Tasks completed:** Task 014 phases 1-5 (of 7)
- **Files created:** 2 wiki articles
- **Files modified:** 22 existing files
- **Code patterns introduced:** YAML frontmatter state tracking, state.json aggregation, 3-field spec counts
- **Documentation:** 4 wiki articles covering concepts and workflows
- **Target release:** v0.5.0

## Lessons Learned

1. **Critical assessment first** - User challenged initial assessment, leading to better solution
2. **Level hierarchy matters** - Templates define structure, agents follow (don't duplicate)
3. **Iterative refinement** - Multiple passes needed to clean up examples and align levels
4. **State colocation** - Embedding state in specs (like tasks) proved simpler than external files
5. **User responsibility** - Framework enables stale detection but leaves decisions to users
