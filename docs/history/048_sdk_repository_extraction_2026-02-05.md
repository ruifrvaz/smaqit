# SDK Repository Extraction

**Date:** 2026-02-05  
**Session Focus:** Task 075 assessment, SDK extraction to separate repository  
**Tasks Referenced:** Task 075 (Dual Release Architecture: SDK and Product)

## Session Overview

This session addressed Task 075 (Dual Release Architecture) through critical assessment, revealing deep entanglement between SDK and product code that made in-place refactoring too risky. The solution was to extract SDK to a completely separate repository (`smaqit-sdk`) while leaving the smaqit product repository untouched.

## Actions Taken

### Phase 1: Task 075 Implementation Attempt (Initial Approach)

**Goal:** Implement dual release architecture within smaqit monorepo with separate installers for SDK and product.

**Work completed:**
1. Updated README.md with dual product documentation (smaqit vs smaqit-sdk)
2. Created SDK wiki documentation (`docs/wiki/workflows/extending-smaqit.md`)
3. Updated framework principle reference in `framework/SMAQIT.md`
4. Partially updated installer files and workflows

**Issues identified:**
- User questioned unified installer approach - wanted completely separate installers
- User highlighted that SDK doesn't process specs, it compiles principles into agents
- SDK was shipping with product-specific files (layer/phase compilation rules)

### Phase 2: Critical Assessment

**User request:** "Follow instructions in session.assess.prompt.md. We need a clean separation of smaqit into a new repository."

**Assessment findings:**
1. **Deep entanglement confirmed:**
   - Level agents reference product-specific paths (business, functional, stack, etc.)
   - Installer SDK copies from parent directories (`../framework`, `../templates/agents`)
   - Framework files (LAYERS.md, PHASES.md) define smaqit-specific concepts but treated as primitives
   - Compilation files intertwined (generic rules reference layer concepts)

2. **Risk analysis:**
   - In-place extraction would break existing build system
   - Requires rewriting Level agents to be truly generic
   - Risk of contaminating both products during transition
   - Ambiguous ownership of shared files
   - Difficult rollback if extraction fails

3. **Recommendation:** Clean repository split with copy-first, then clean approach.

**User approval:** Proceed with SDK extraction to `/home/ruifrvaz/projects/smaqit-sdk/`

### Phase 3: SDK Repository Creation

**Created new repository structure:**

```
/home/ruifrvaz/projects/smaqit-sdk/
├── README.md                    # SDK-focused documentation
├── CHANGELOG.md                 # Version 0.1.0
├── LICENSE                      # Copied from smaqit
├── .gitignore                   # Build artifacts
├── install.sh                   # Installer script (improved version)
├── HANDOVER.md                  # Session continuation guide
├── framework/                   # 5 generic principle files
│   ├── SMAQIT.md
│   ├── AGENTS.md
│   ├── TEMPLATES.md
│   ├── ARTIFACTS.md
│   └── PROMPTS.md
├── templates/
│   └── agents/
│       ├── base-agent.template.md
│       ├── specification-agent.template.md
│       ├── implementation-agent.template.md
│       └── compiled/
│           ├── base.rules.md
│           ├── specification.rules.md
│           └── implementation.rules.md
├── agents/                      # Level agents (copied, need cleaning)
│   ├── smaqit.L0.agent.md
│   ├── smaqit.L1.agent.md
│   └── smaqit.L2.agent.md
├── prompts/
│   └── smaqit.new-agent.prompt.md
├── installer/
│   ├── main.go                  # Self-contained installer
│   ├── Makefile                 # No ../ references
│   └── go.mod                   # github.com/ruifrvaz/smaqit-sdk/installer
├── docs/
│   └── wiki/
│       └── extending-smaqit.md  # Moved from smaqit repo
└── .github/
    └── workflows/
        ├── release.yml          # SDK release automation
        └── test-integration.yml # Structure validation
```

**Files excluded from SDK (product-specific):**
- LAYERS.md, PHASES.md (smaqit layer/phase models)
- Layer compilation rules (business, functional, stack, infrastructure, coverage)
- Phase compilation rules (develop, deploy, validate)
- Spec templates (outputs of compiled agents)
- Prompt templates (product-specific)

**Key SDK files created:**
1. **installer/Makefile** - Self-contained build system with no parent references
2. **installer/main.go** - Generic installer (15 files: 5 framework + 3 templates + 3 rules + 3 Level agents + 1 prompt)
3. **installer/go.mod** - Module path `github.com/ruifrvaz/smaqit-sdk/installer`
4. **README.md** - SDK-focused documentation explaining generic toolkit purpose
5. **CHANGELOG.md** - Initial 0.1.0 release notes
6. **install.sh** - Improved version from smaqit/install-sdk.sh with colored output, version handling, verification
7. **.github/workflows/release.yml** - Triggers on `sdk-v*` tags
8. **.github/workflows/test-integration.yml** - Validates 15 SDK files exist, verifies no product contamination
9. **HANDOVER.md** - Comprehensive continuation guide for next session

### Phase 4: Smaqit Repository Cleanup

**User request:** "Let's also move extending-smaqit.md wiki into the smaqit-sdk wiki"

**Actions:**
1. Copied `docs/wiki/workflows/extending-smaqit.md` to smaqit-sdk (SDK-specific content)
2. Documented move in smaqit-sdk/HANDOVER.md

**User request:** "Only two files left to revert, see what needs to be reverted to pre sdk attempt"

**Analysis:**
- `installer/Makefile` - All changes were SDK-related (explicit file copying, removed framework copy)
- `install.sh` - All changes were SDK-related (PRODUCT/TAG_PREFIX/BINARY_PREFIX variables, SDK filtering)

**Decision:** Revert both files completely using `git restore`

**User request:** "Let's abandon the sdk task"

**Actions:**
1. Moved Task 075 from Active to Abandoned in `docs/tasks/PLANNING.md`
2. Added abandonment reason: "Requires deep refactoring with high risk to existing codebase. SDK extracted to separate repository (smaqit-sdk) for clean development without contaminating product."
3. Updated `docs/tasks/075_dual_release_architecture_sdk_and_product.md` with:
   - Status: Abandoned
   - Abandonment date: 2026-02-05
   - Detailed reasoning
   - Link to new smaqit-sdk repository

## Problems Solved

### Problem 1: Monorepo Contamination Risk

**Issue:** SDK and product code deeply entangled - attempting separation within same repo would break existing functionality.

**Solution:** Complete repository split. SDK development happens in clean slate at `/home/ruifrvaz/projects/smaqit-sdk/` with zero impact on smaqit product.

### Problem 2: Product-Specific Content in SDK

**Issue:** SDK was shipping with smaqit layer/phase rules, forcing smaqit's five-layer model on all SDK users.

**Solution:** SDK ships only generic primitives (5 framework files, 3 templates, 3 compilation rules, 3 Level agents). Product-specific layer/phase rules remain in smaqit repo.

### Problem 3: Install Script Quality Difference

**Issue:** Initial smaqit-sdk/install.sh was simple but lacked features compared to smaqit/install-sdk.sh (no colored output, no version modes, no verification).

**Solution:** Copied better install-sdk.sh to smaqit-sdk, updated REPO variable to point to new repository.

## Decisions Made

### Decision 1: Separate Repository vs Monorepo

**Choice:** Extract SDK to completely separate repository.

**Rationale:**
- Zero risk to existing smaqit product
- Clean slate for SDK without legacy contamination
- Independent versioning and release cycles
- Clear ownership boundaries
- Parallel development without blocking
- Easier rollback if needed

**Trade-offs accepted:**
- Loses shared Git history
- Requires separate PRs for cross-cutting changes
- Different repositories to maintain

### Decision 2: SDK Scope - Generic Primitives Only

**Choice:** SDK ships only framework primitives, excludes all smaqit product specifics.

**Rationale:**
- SDK should be truly generic (usable for non-smaqit agents)
- smaqit becomes proof-of-concept using SDK
- Users can create custom layer models without smaqit constraints

**SDK includes (15 files):**
- Framework files (5): SMAQIT, AGENTS, TEMPLATES, ARTIFACTS, PROMPTS
- Agent templates (3): base, specification, implementation
- Compilation rules (3): base, specification, implementation
- Level agents (3): L0, L1, L2
- new-agent prompt (1)

**SDK excludes:**
- LAYERS.md, PHASES.md (smaqit-specific)
- Layer/phase compilation rules (8 files)
- Spec templates (product outputs)
- Prompt templates (product-specific)

### Decision 3: SDK Version - Start at 0.1.0

**Choice:** Initial SDK release as `sdk-v0.1.0` (private repo).

**Rationale:**
- Early/experimental stage
- Level agents need cleaning (still contain product references)
- Not yet tested or validated
- Room for iteration before public release

### Decision 4: Smaqit Repository Status - Unchanged

**Choice:** Leave smaqit repository completely untouched, revert all Task 075 changes.

**Rationale:**
- Preserve working product
- Task 075 abandoned in favor of separate repo approach
- All SDK work continues in smaqit-sdk
- No risk to existing users

## Files Modified

### smaqit Repository (This Session)

**Modified:**
1. `README.md` - Added dual product documentation (later reverted via git reset)
2. `docs/wiki/workflows/extending-smaqit.md` - Created SDK extension guide (later moved to smaqit-sdk)
3. `framework/SMAQIT.md` - Updated "Extensible Through Templates" principle (later reverted)
4. `installer/Makefile` - Updated for product-only artifacts (reverted via git restore)
5. `install.sh` - Added SDK filtering logic (reverted via git restore)
6. `docs/tasks/PLANNING.md` - Moved Task 075 to Abandoned section
7. `docs/tasks/075_dual_release_architecture_sdk_and_product.md` - Updated with abandonment status and reasoning

**Final state:** Repository clean except for Task 075 abandonment documentation.

### smaqit-sdk Repository (Created)

**Created 18 files total:**
1. `README.md` - SDK documentation
2. `CHANGELOG.md` - Initial 0.1.0 release
3. `LICENSE` - Copied from smaqit
4. `.gitignore` - Build artifacts
5. `install.sh` - Improved installer script
6. `HANDOVER.md` - Session continuation guide
7. `framework/SMAQIT.md` - Core principles
8. `framework/AGENTS.md` - Agent concepts
9. `framework/TEMPLATES.md` - Template structure
10. `framework/ARTIFACTS.md` - Artifact patterns
11. `framework/PROMPTS.md` - Prompt architecture
12. `templates/agents/base-agent.template.md` - Generic base template
13. `templates/agents/specification-agent.template.md` - Spec agent template
14. `templates/agents/implementation-agent.template.md` - Impl agent template
15. `templates/agents/compiled/base.rules.md` - Base compilation rules
16. `templates/agents/compiled/specification.rules.md` - Spec compilation rules
17. `templates/agents/compiled/implementation.rules.md` - Impl compilation rules
18. `agents/smaqit.L0.agent.md` - Level 0 agent (needs cleaning)
19. `agents/smaqit.L1.agent.md` - Level 1 agent (needs cleaning)
20. `agents/smaqit.L2.agent.md` - Level 2 agent (needs cleaning)
21. `prompts/smaqit.new-agent.prompt.md` - Agent creation prompt
22. `installer/main.go` - Self-contained installer
23. `installer/Makefile` - Self-contained build system
24. `installer/go.mod` - Go module definition
25. `docs/wiki/extending-smaqit.md` - SDK extension guide (moved from smaqit)
26. `.github/workflows/release.yml` - Release automation
27. `.github/workflows/test-integration.yml` - Structure validation

## Next Steps

### For smaqit Repository (This Repo)

1. **No immediate work required** - Repository is stable and clean
2. Continue normal product development
3. Eventually may import SDK as Go module dependency (future consideration)

### For smaqit-sdk Repository (New Repo)

**Critical next steps documented in `/home/ruifrvaz/projects/smaqit-sdk/HANDOVER.md`:**

1. **Clean Level agents** (REQUIRED) - Remove all smaqit product references:
   - Remove business/functional/stack/infrastructure/coverage examples
   - Remove develop/deploy/validate references
   - Remove LAYERS.md, PHASES.md references
   - Replace with generic `[DOMAIN]`, `[LAYER]`, `[PHASE]` placeholders

2. **Test build:**
   ```bash
   cd /home/ruifrvaz/projects/smaqit-sdk/installer
   make clean build test
   ```

3. **Initialize Git repository:**
   ```bash
   cd /home/ruifrvaz/projects/smaqit-sdk
   git init
   git add .
   git commit -m "Initial SDK extraction"
   gh repo create ruifrvaz/smaqit-sdk --private --source=. --push
   ```

4. **Tag release** (after successful test):
   ```bash
   git tag sdk-v0.1.0
   git push origin sdk-v0.1.0
   ```

**Success criteria for SDK:**
- ✅ SDK builds successfully
- ✅ SDK installs 15 files (no product-specific files)
- ✅ Level agents contain no smaqit product references
- ✅ GitHub repo created and pushed
- ✅ Release workflow produces binaries
- ✅ Integration test passes

## Session Metrics

- **Duration:** ~4 hours
- **Tasks completed:** Task 075 assessed and abandoned
- **Files created:** 27 (in smaqit-sdk repository)
- **Files modified:** 7 (in smaqit repository, most reverted)
- **Repositories created:** 1 (smaqit-sdk)
- **Key decisions:** 4 major architectural decisions
- **Lines of code:** ~2,000 (SDK infrastructure)

## Key Takeaways

1. **Critical assessment prevents wasted effort** - Initial implementation attempt revealed fundamental architectural issues that would have been expensive to fix later.

2. **Repository split is sometimes the right answer** - When entanglement is deep enough, separation is cleaner than refactoring.

3. **SDK as separate product enables true generality** - Extracting SDK forces it to be truly generic rather than smaqit-specific with generic aspirations.

4. **Zero risk to existing product** - Clean extraction approach means smaqit continues stable while SDK develops independently.

5. **Handover documentation is critical** - Comprehensive HANDOVER.md in smaqit-sdk enables seamless session continuation without context loss.
