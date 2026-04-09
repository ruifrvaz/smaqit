# SDK Product Separation

**Date:** 2026-01-27  
**Session Focus:** Rationalize log locations, create development binary, assess SDK shipping strategy, consolidate duplicate tasks  
**Tasks Referenced:** 075, 076 (abandoned)

## Actions Taken

### 1. Log Location Rationalization
- **Moved** `agents/logs/qa-compilation-2026-01-26.md` to `docs/logs/` for SDK development documentation
- **Updated narrative** in QA compilation log to reflect SDK development bootstrap process (manual creation vs. interactive gathering)
- **Updated Agent-L2** (4 replacements):
  - Changed all log path references from `agents/logs/` to `.smaqit/logs/`
  - Documented 7-step interactive gathering workflow
  - Updated completion criteria to reference new log location
- **Updated installer** to create `.smaqit/logs/` directory during `smaqit init`
- **Rationale:** SDK development logs belong in `docs/logs/` (meta-documentation), while user compilation logs go to `.smaqit/logs/` (product artifact)

### 2. Development Binary Creation
- **Created `smaqit-dev` binary** separate from production `smaqit`
- **Modified `installer/Makefile`:**
  - Added `DEV_BINARY_NAME = smaqit-dev`
  - Updated `build`, `install`, `uninstall` targets to use development binary name
- **Installed** `smaqit-dev` to `~/.local/bin` via `make install`
- **Updated user-testing agent** to use `smaqit-dev` instead of `../installer/dist/smaqit`
- **Rationale:** Avoid chicken-and-egg problem of developing smaqit with smaqit by having separate development binary

### 3. SDK Shipping Strategy Assessment
- **Applied session.assess.prompt.md methodology** to evaluate SDK distribution options
- **Created detailed tree structures** for both products:
  - **smaqit (app builder):** 8 agents, 8 prompts, 5 spec templates, reports directory
  - **smaqit.sdk (framework builder):** L0/L1/L2/QA agents (4), framework files (7), agent templates (14), new-agent prompt (1)
- **Iterated structure** with user to refine SDK contents:
  - Removed spec templates (app builder only)
  - Removed prompt templates (too abstract for SDK users)
  - Removed session management prompts (app builder workflow)
  - Added QA agent (development tool)
  - Added new-agent prompt (agent creation workflow)
- **Verified current installation** against proposed structure
- **Recommended Option B:** Separate `smaqit-sdk` installer with independent versioning and release workflow

### 4. Task Consolidation
- **Created Task 076** initially with detailed implementation instructions for installer separation
- **Discovered overlap** with existing Task 075 (Dual Release Architecture)
- **Performed critical assessment** comparing both tasks:
  - Task 075: Comprehensive (CI/CD, versioning, integration testing, documentation)
  - Task 076: Focused on installers only
  - Task 075: 3 agents in SDK, Task 076: 4 agents (includes QA)
- **Updated Task 075** with clarifications from Task 076:
  - Added QA agent to SDK (4 agents: L0, L1, L2, QA)
  - Added new-agent prompt explicitly
  - Added `.smaqit/logs/` directory to SDK installation
  - Added explicit exclusions for product installer (no QA, no new-agent, no logs)
  - Added detailed Makefile prepare targets with exact copy commands
  - Updated embed directives and installation structures
  - Updated SDK artifact count (25→26 files)
- **Abandoned Task 076** as duplicate, added to Abandoned section in PLANNING.md
- **Deleted** `docs/tasks/076_separate_sdk_and_app_installers.md`

### 5. Framework Documentation Fix
- **Fixed PROMPTS.md formatting:**
  - Removed duplicate pipe in table
  - Removed duplicate paragraph
  - Documented interactive agent creation pattern

## Problems Solved

### Chicken-and-Egg Development Problem
**Issue:** Developing smaqit using smaqit installer would overwrite the production binary being developed.

**Solution:** Created `smaqit-dev` binary that:
- Installs to same location as production but with different name
- Allows development work without affecting production releases
- Used by testing agent for E2E validation

### Log Location Confusion
**Issue:** SDK development logs (documenting how SDK was bootstrapped) mixed conceptually with user compilation logs (product artifacts from using SDK).

**Solution:** 
- SDK development logs → `docs/logs/` (meta-documentation about smaqit development)
- User compilation logs → `.smaqit/logs/` (artifacts produced by users using smaqit.sdk)
- Clear separation of "developing the SDK" vs "using the SDK"

### Product Identity Confusion
**Issue:** Terminology "smaqit SDK" implied the installer shipped SDK, but it actually shipped app-building toolkit.

**Clarification:**
- **smaqit** = App builder toolkit (8 product agents for spec-driven development)
- **smaqit.sdk** = Framework builder toolkit (L0/L1/L2 agents for agent development)
- Current installer ships smaqit (app builder), not smaqit.sdk

### Duplicate Task Creation
**Issue:** Created Task 076 without checking for existing Task 075 covering same problem space.

**Resolution:**
- Applied session.assess.prompt.md methodology
- Performed detailed comparison
- Merged best ideas from Task 076 into Task 075
- Abandoned Task 076 as duplicate
- Established Task 075 as comprehensive single source of truth

## Decisions Made

### 1. Dual Binary Strategy for Development
**Decision:** Maintain separate `smaqit` (production) and `smaqit-dev` (development) binaries.

**Rationale:**
- Prevents overwriting production during development
- Allows testing without affecting releases
- Clear separation of development vs production use

### 2. Log Location Semantics
**Decision:** 
- `docs/logs/` = SDK development documentation (how smaqit itself was built)
- `.smaqit/logs/` = User compilation logs (artifacts from using smaqit.sdk)

**Rationale:**
- `docs/` hierarchy is meta-documentation about smaqit
- `.smaqit/` hierarchy is product artifacts for users
- Clear conceptual boundary between "developing smaqit" and "using smaqit"

### 3. QA Agent Belongs in SDK
**Decision:** Include QA agent in smaqit.sdk (framework builder toolkit), not smaqit (app builder toolkit).

**Rationale:**
- QA agent answers documentation questions about framework
- Framework builders need to query wiki, principles, design decisions
- App builders don't need meta-framework documentation access
- Aligns with "SDK = tools for building frameworks" identity

### 4. New-Agent Prompt Belongs in SDK
**Decision:** Ship new-agent prompt with smaqit.sdk only, not with smaqit.

**Rationale:**
- New-agent prompt structures agent creation workflow
- App builders use pre-built agents, don't create new ones
- Framework builders extend smaqit by creating custom agents
- Clear separation of concerns

### 5. Task 075 as Single Source of Truth
**Decision:** Consolidate all dual-release architecture work into Task 075, abandon Task 076.

**Rationale:**
- Task 075 more comprehensive (includes CI/CD, testing, versioning)
- Avoids fragmentation of implementation work
- Task 076's best ideas merged into Task 075
- Duplicate tasks create confusion and maintenance burden

### 6. SDK Contents Finalized
**Decision:** SDK ships:
- Framework files (7): L0 principles
- Agent templates (14): L1 templates + compilation rules
- Level agents (4): L0, L1, L2, QA
- New-agent prompt (1): Agent creation workflow
- Total: 26 files

**Exclusions:**
- Spec templates (app builder concern)
- Prompt templates (too abstract)
- Session management prompts (app builder workflow)
- Product agents (compiled L2 artifacts, not SDK)

**Rationale:**
- SDK provides meta-framework for building agent orchestration systems
- App builder toolkit provides concrete agents for spec-driven development
- Each product serves distinct audience with clear boundaries

## Files Modified

### Created
- `docs/logs/qa-compilation-2026-01-26.md` - Moved from `agents/logs/`, updated narrative for SDK development context
- `docs/tasks/076_separate_sdk_and_app_installers.md` - Created then deleted after consolidation into Task 075

### Modified
- `.github/agents/smaqit.L2.agent.md` - 4 replacements changing log paths, added interactive gathering workflow
- `.github/agents/smaqit.user-testing.agent.md` - Updated commands to use `smaqit-dev`
- `installer/Makefile` - Added `DEV_BINARY_NAME`, updated build/install/uninstall targets
- `installer/main.go` - Added `.smaqit/logs` directory creation
- `framework/PROMPTS.md` - Fixed table formatting, removed duplicate content, documented interactive pattern
- `docs/tasks/075_dual_release_architecture_sdk_and_product.md` - 6 updates incorporating Task 076 clarifications:
  - SDK now includes QA agent + new-agent prompt
  - Product explicitly excludes QA agent + new-agent prompt
  - Added `.smaqit/logs/` to SDK installation
  - Product does NOT create logs directory
  - Detailed Makefile prepare targets
  - Updated embed directives and installation structures
- `docs/tasks/PLANNING.md` - Removed Task 076 from Active, added to Abandoned section

### Deleted
- `agents/logs/qa-compilation-2026-01-26.md` - Moved to `docs/logs/`
- `docs/tasks/076_separate_sdk_and_app_installers.md` - Abandoned as duplicate

## Next Steps

### Immediate (Task 075 Implementation)
1. **Create `installer-sdk/` directory** with SDK-specific installer structure
2. **Implement SDK installer** with proper embed directives (framework, agent templates, L0/L1/L2/QA agents, new-agent prompt)
3. **Update product installer Makefile** to explicitly copy only product files (exclude QA, new-agent)
4. **Remove `.smaqit/logs/` creation** from product installer
5. **Create SDK release workflow** (`.github/workflows/release-sdk.yml`)
6. **Create SDK integration testing workflow** (`.github/workflows/test-sdk-integration.yml`)
7. **Split CHANGELOG.md** into SDK and Product sections
8. **Update README.md** with dual-product documentation

### Documentation
- Create wiki page: "Extending smaqit via smaqit.sdk"
- Document Agent-L2 compilation workflow
- Document release choreography (SDK → compilation → product)
- Create `install-sdk.sh` script

### Validation
- Test `smaqit init` produces correct 8-agent installation
- Test `smaqit-sdk init` produces correct SDK installation
- Verify E2E workflow with development binary
- Run integration tests for compilation patterns

## Session Metrics

- **Duration:** ~2 hours
- **Tasks Completed:** 0 (Task 075 prepared but not implemented)
- **Tasks Created:** 1 (Task 076, later abandoned)
- **Tasks Abandoned:** 1 (Task 076 as duplicate)
- **Files Created:** 2 (1 log file, 1 task file later deleted)
- **Files Modified:** 7 (agents, installer files, framework docs, planning)
- **Files Deleted:** 2 (old log location, duplicate task)
- **Key Decisions:** 6 (development binary, log semantics, QA/new-agent placement, SDK contents)
- **Architecture Assessment:** 1 comprehensive evaluation following session.assess.prompt.md methodology
- **Build System Changes:** Development binary creation and installation

## Technical Context

### Binary Naming Strategy
- **smaqit** - Production release binary (v0.6.2+)
- **smaqit-dev** - Development binary (local builds)
- **smaqit-sdk** - SDK installer binary (future, Task 075)
- **smaqit-sdk-dev** - SDK development binary (future, Task 075)

### Installation Structures
**Current (smaqit app builder):**
```
.github/agents/     [9 files: 8 product + qa - will be 8 after Task 075]
.github/prompts/    [9 files: 8 product + new-agent - will be 8 after Task 075]
.smaqit/templates/  [5 spec templates]
.smaqit/reports/    [validation reports]
.smaqit/logs/       [WILL BE REMOVED in Task 075]
specs/              [layer directories]
```

**Future (smaqit.sdk framework builder - Task 075):**
```
.github/agents/     [4 files: L0, L1, L2, QA]
.github/prompts/    [1 file: new-agent]
.smaqit/framework/  [7 files: L0 principles]
.smaqit/templates/  [14 files: L1 templates + compilation rules]
.smaqit/logs/       [compilation logs]
```

### Compilation Chain (SDK → Product)
1. **L0 (framework/)** - Principles and concepts
2. **L1 (templates/agents/)** - Templates + compilation rules
3. **L2 (agents/)** - Compiled product agents
4. **Agent-L2** - Compiler that transforms L1→L2

### Assessment Methodology Applied
Followed session.assess.prompt.md workflow:
1. ✅ Questioned premise (is new task necessary?)
2. ✅ Checked existing state (read Task 075)
3. ✅ Identified trade-offs (comprehensive vs. focused)
4. ✅ Flagged problems (duplicate task with overlap)
5. ✅ Presented assessment (detailed comparison table)
6. ✅ Merged best ideas into canonical task
