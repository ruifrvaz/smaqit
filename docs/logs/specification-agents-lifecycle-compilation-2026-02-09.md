# Specification Agents Lifecycle Directives Compilation

**Date:** 2026-02-09  
**Agent:** Agent-L2 (Agent Compiler)  
**Task:** Task 079 - Spec Agents Revert Status to Draft on Modification  
**Compilation Type:** 4-way merge (base + specification-extension + layer-specific)

## Compilation Summary

Compiled specification lifecycle directives from L1 templates into all 5 L2 specification agents:
- Checkbox reset directive (Task 060 - backfill for Infrastructure and Coverage)
- Status reversion directive (Task 079 - new for all agents)

## Sources Read

### L1 Templates

**Primary Source:**
- `templates/agents/compiled/specification.rules.md`
  - Line 76: Checkbox reset directive
  - Line 77: Status reversion directive

**Foundation Source:**
- `templates/agents/compiled/base.rules.md`
  - Base MUST/MUST NOT directives (shared by all agents)

**Layer-Specific Sources:**
- `templates/agents/compiled/business.rules.md` (Business-specific directives)
- `templates/agents/compiled/functional.rules.md` (Functional-specific directives)
- `templates/agents/compiled/stack.rules.md` (Stack-specific directives)
- `templates/agents/compiled/infrastructure.rules.md` (Infrastructure-specific directives)
- `templates/agents/compiled/coverage.rules.md` (Coverage-specific directives)

### L0 Framework References

Referenced principles (traced through L1 compilation files):
- `framework/ARTIFACTS.md` - Checkbox Lifecycle During Refinement (lines 304-325)
- `framework/ARTIFACTS.md` - Status Lifecycle During Refinement (lines 327-346)

## Merge Process

### Directive Format (from L1)

**Checkbox Reset:**
```
Reset acceptance criteria checkbox to `[ ]` when modifying existing criteria text (expanded scope requires revalidation)
```

**Status Reversion:**
```
Revert spec `status` field to `draft` when modifying acceptance criteria text
```

### Compilation to L2 (Product Agents)

Both directives merged into MUST section of each agent:
- Inserted after "Validate output against completion criteria before finishing"
- Grouped together (lifecycle behaviors)
- Full verbosity maintained from L1

## Agents Updated

### 1. Business Agent

**File:** `agents/smaqit.business.agent.md`  
**Changes:**
- ✅ Status reversion directive added (new)
- ℹ️ Checkbox reset already present

**Location:** MUST section, after validation directive

### 2. Functional Agent

**File:** `agents/smaqit.functional.agent.md`  
**Changes:**
- ✅ Status reversion directive added (new)
- ℹ️ Checkbox reset already present

**Location:** MUST section, after validation directive

### 3. Stack Agent

**File:** `agents/smaqit.stack.agent.md`  
**Changes:**
- ✅ Status reversion directive added (new)
- ℹ️ Checkbox reset already present

**Location:** MUST section, after validation directive

### 4. Infrastructure Agent

**File:** `agents/smaqit.infrastructure.agent.md`  
**Changes:**
- ✅ Checkbox reset directive added (backfill from Task 060)
- ✅ Status reversion directive added (new)

**Location:** MUST section, after validation directive

**Note:** Infrastructure agent was missing checkbox reset directive from Task 060. Added both directives in this compilation.

### 5. Coverage Agent

**File:** `agents/smaqit.coverage.agent.md`  
**Changes:**
- ✅ Checkbox reset directive added (backfill from Task 060)
- ✅ Status reversion directive added (new)

**Location:** MUST section, after validation directive

**Note:** Coverage agent was missing checkbox reset directive from Task 060. Added both directives in this compilation.

## Validation Checklist

- [x] All 5 specification agents updated
- [x] Checkbox reset directive present in all agents
- [x] Status reversion directive present in all agents
- [x] Directives inserted in consistent location (after validation)
- [x] Directives grouped together (lifecycle behaviors)
- [x] Full verbosity maintained from L1
- [x] No placeholders remain ([LAYER], [CONCEPT], etc.)
- [x] Behavior consistent across all specification agents
- [x] Task 060 regression fixed (Infrastructure and Coverage backfilled)

## Compilation Notes

### Opportunistic Cleanup

This compilation addressed Task 060 regression:
- Infrastructure and Coverage agents were missing checkbox reset directive
- Root cause: Likely compiled before Task 060 was implemented
- Resolution: Backfilled both agents with complete specification lifecycle directives

This approach aligns with Level contamination cleanup guidance in `.github/copilot-instructions.md`:
> "All Level agents (L0, L1, L2) are authorized and encouraged to perform opportunistic cleanup during regular sessions"

### Directive Relationship

Checkbox reset and status reversion operate together:
- **Checkboxes** track granular requirement satisfaction (per acceptance criterion)
- **Status** tracks overall validation state (spec lifecycle position)
- **Both reset** when specifications change to prevent stale validation indicators

This complementary behavior creates consistency in specification lifecycle tracking.

## Next Steps

### Task 079 Completion

- [x] L0: Framework documents status reversion principle (ARTIFACTS.md)
- [x] L1: Directive added to specification.rules.md
- [x] L2: Directive compiled to all 5 specification agents
- [x] Compilation log created

**Task 079 Status:** Ready for completion

### Recommended Follow-Up

1. **Test behavior**: Invoke specification agents to verify directives execute correctly
2. **Update task file**: Mark Task 079 as complete with completion date
3. **Session documentation**: Record compilation in session history

## Compilation Metadata

**Compiler:** Agent-L2 (Level 2 Agent Compiler)  
**Compilation Date:** 2026-02-09  
**Agents Compiled:** 5 (Business, Functional, Stack, Infrastructure, Coverage)  
**Directives Added:** 7 total (5 status reversion + 2 checkbox reset backfills)  
**Template Sources:** 7 files (1 specification + 1 base + 5 layer-specific)  
**Framework Sources:** 1 file (ARTIFACTS.md)
