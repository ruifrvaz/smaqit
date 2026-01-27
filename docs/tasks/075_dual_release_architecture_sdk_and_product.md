# Dual Release Architecture: smaqit-sdk and smaqit

**Status:** Not Started  
**Created:** 2026-01-25

## Description

Implement dual release architecture that treats smaqit-sdk (for agent development) as the primary product and smaqit (for app development) as a compiled byproduct. Both releases live in the same monorepo but have independent versioning, CI/CD workflows, and installers.

## Context

Session 046 established the SDK foundation (Level Up Architecture with 3 compilation patterns). The SDK generates product agents through compilation, making the product a true byproduct of the SDK. This relationship should be reflected in release strategy.

**Key principle:** smaqit-sdk can reach stability (v1.0) while smaqit continues evolution (v0.x).

## Acceptance Criteria

### Infrastructure
- [ ] Create `installer-sdk/` directory with SDK-specific installer
- [ ] SDK installer embeds: `framework/*.md`, `templates/agents/**/*.md`, `.github/agents/smaqit.L*.agent.md`, `agents/smaqit.qa.agent.md`, `prompts/smaqit.new-agent.prompt.md`
- [ ] SDK installer creates: `.smaqit/framework/`, `.smaqit/templates/agents/`, `.smaqit/logs/`, `.github/agents/` (Level agents + QA), `.github/prompts/` (new-agent prompt)
- [ ] Product installer embeds: `agents/*.md` (8 product agents, excludes qa), `prompts/*.md` (8 product prompts, excludes new-agent), `templates/specs/*.md`
- [ ] Product installer creates: `.smaqit/templates/specs/`, `.smaqit/reports/`, `.github/agents/` (8 product agents), `.github/prompts/` (8 product prompts), `specs/` (layer dirs)
- [ ] Product installer does NOT create `.smaqit/logs/` directory (SDK development only)
- [ ] Create `.github/workflows/release-sdk.yml` triggered on `sdk-v*` tags
- [ ] Rename `.github/workflows/release.yml` to `release-product.yml` for clarity (triggered on `v*` tags)
- [ ] Create `.github/workflows/test-sdk-integration.yml` for compilation validation
- [ ] Update `installer-sdk/Makefile` to copy SDK artifacts during `make prepare`:
  - `cp -r ../framework .`
  - `cp -r ../templates/agents templates/`
  - `cp ../.github/agents/smaqit.L*.agent.md agents/`
  - `cp ../agents/smaqit.qa.agent.md agents/`
  - `cp ../prompts/smaqit.new-agent.prompt.md prompts/`
- [ ] Update `installer/Makefile` to copy product artifacts explicitly (excluding SDK files):
  - `cp ../agents/smaqit.{business,functional,stack,infrastructure,coverage,development,deployment,validation}.agent.md agents/` (excludes qa)
  - `cp ../prompts/smaqit.{business,functional,stack,infrastructure,coverage,development,deployment,validation}.prompt.md prompts/` (excludes new-agent)
  - `cp -r ../templates/specs templates/`

### Versioning
- [ ] SDK releases use `sdk-v*.*.*` tag pattern (e.g., `sdk-v1.0.0`)
- [ ] Product releases use `v*.*.*` tag pattern (e.g., `v0.7.0`)
- [ ] Split CHANGELOG.md into two sections: "smaqit (for app development)" and "smaqit-sdk (for agent development)"
- [ ] Update comparison links at bottom of CHANGELOG to handle dual versioning

### Build Artifacts
- [ ] SDK workflow builds `smaqit-sdk` binary with embedded SDK artifacts
- [ ] Product workflow builds `smaqit` binary with embedded product artifacts
- [ ] Both workflows generate platform-specific binaries (Linux, macOS Intel/ARM, Windows)
- [ ] Both workflows generate SHA256 checksums
- [ ] Both workflows extract release notes from appropriate CHANGELOG section (SDK vs Product)
- [ ] SDK binary name: `smaqit-sdk_linux_amd64`, `smaqit-sdk_darwin_amd64`, `smaqit-sdk_darwin_arm64`, `smaqit-sdk_windows_amd64.exe`
- [ ] Product binary name: `smaqit_linux_amd64`, `smaqit_darwin_amd64`, `smaqit_darwin_arm64`, `smaqit_windows_amd64.exe`

### Integration Testing
- [ ] Integration workflow runs Agent-L2 to compile templates → product agents
- [ ] Test all 3 SDK compilation patterns: Base (Pattern 1), Specification (Pattern 2), Implementation (Pattern 3)
- [ ] Validate compiled agents have required sections (Role, Input, Output, Directives, Completion Criteria)
- [ ] Validate no unresolved placeholders remain (e.g., `[LAYER_MUST_DIRECTIVES]`, `[ROLE_CONTENT]`)
- [ ] Integration workflow fails if SDK changes break ability to compile future product agents
- [ ] Document compilation workflow in SDK documentation (how to manually compile product agents from SDK)

### Documentation
- [ ] Update README.md to explain dual products (smaqit-sdk for agent development, smaqit for app development)
- [ ] Add smaqit-sdk installation instructions to README: Download URL for smaqit-sdk vs smaqit binaries
- [ ] Create `install-sdk.sh` script parallel to `install.sh` for SDK installation
- [ ] Update `install.sh` to support `SMAQIT_SDK=true` environment variable for SDK installation
- [ ] Document release choreography: SDK changes → compilation → product release
- [ ] Add smaqit-sdk versioning strategy to wiki
- [ ] Create wiki page: "Extending smaqit via smaqit-sdk" explaining how to use templates and Level agents
- [ ] Document Agent-L2 compilation workflow: how to manually compile product agents from smaqit-sdk
- [ ] Update framework/SMAQIT.md "Extensible Through Templates" principle to reference smaqit-sdk

## Implementation Notes

**What goes in each installer:**

SDK installer (`installer-sdk/`):
```go
//go:embed framework/*.md
var frameworkFiles embed.FS

//go:embed templates/agents/*.md templates/agents/compiled/*.md
var agentTemplateFiles embed.FS

//go:embed agents/*.md
var sdkAgentFiles embed.FS  // L0, L1, L2, QA agents

//go:embed prompts/smaqit.new-agent.prompt.md
var newAgentPrompt embed.FS
```

Product installer (`installer/`):
```go
//go:embed templates/specs/*.md
var templateFiles embed.FS

//go:embed agents/*.md
var agentFiles embed.FS  // 8 product agents (business, functional, stack, infrastructure, coverage, development, deployment, validation)

//go:embed prompts/*.md
var promptFiles embed.FS  // 8 product prompts (excludes new-agent)
```

**What gets installed where:**

SDK installation (`smaqit-sdk init`):
```
user-project/
├── .smaqit/
│   ├── framework/           # L0 principles (7 files)
│   ├── templates/
│   │   └── agents/          # L1 templates + compilation files
│   │       ├── base-agent.template.md
│   │       ├── specification-agent.template.md
│   │       ├── implementation-agent.template.md
│   │       └── compiled/    # 11 compilation files
│   └── logs/                # Agent compilation logs (SDK development)
└── .github/
    ├── agents/
    │   ├── smaqit.L0.agent.md   # Principle curator
    │   ├── smaqit.L1.agent.md   # Template compiler
    │   ├── smaqit.L2.agent.md   # Agent compiler
    │   └── smaqit.qa.agent.md   # Documentation Q&A
    └── prompts/
        └── smaqit.new-agent.prompt.md  # Agent creation workflow
```

Product installation (`smaqit init`):
```
user-project/
├── .smaqit/
│   ├── templates/specs/     # Spec templates (5 layers)
│   └── reports/
├── .github/
│   ├── agents/              # 8 product agents (compiled L2 artifacts)
│   └── prompts/             # 8 product prompts
└── specs/                   # Empty layer directories
```

**Directory structure:**
```
smaqit/
├── installer/              # Product installer (existing)
├── installer-sdk/          # SDK installer (NEW)
├── .github/workflows/
│   ├── release-sdk.yml     # NEW: SDK releases
│   ├── release-product.yml # RENAMED: Product releases  
│   └── test-sdk-integration.yml  # NEW: Compilation tests
```

**Tag patterns:**
- SDK: `sdk-v1.0.0`, `sdk-v1.1.0`, etc.
- Product: `v0.7.0`, `v0.8.0`, etc.

**CHANGELOG structure:**
```markdown
# Changelog

## smaqit (for app development)

### [0.7.0] - 2026-01-26
...

## smaqit-sdk (for agent development)

### [sdk-1.0.0] - 2026-01-25
...
```

**Release choreography:**
1. **smaqit-sdk development** → Modify framework/, templates/, Level agents
2. **smaqit-sdk release** → Tag `sdk-v1.0.0`, CI builds `smaqit-sdk` binaries, publish GitHub release
3. **Compilation step** → Run Agent-L2 to update product agents in `agents/` directory
   - Can be manual: Invoke Agent-L2 via Copilot to compile templates → agents
   - Can be automated: CI workflow that runs Agent-L2 after smaqit-sdk release
4. **smaqit release** → Tag `v0.7.0`, CI builds `smaqit` binaries (embeds updated agents/), publish GitHub release

**Parallelization:** Multiple smaqit-sdk changes can accumulate before triggering product compilation + release.

**Integration testing prevents:**
- smaqit-sdk v1.1.0 changes compilation file format → Agent-L2 fails to parse → test catches before release
- smaqit-sdk v1.1.0 changes directive structure → compiled agents have missing sections → test catches before release
- smaqit-sdk v1.1.0 adds new placeholder → Agent-L2 doesn't resolve it → test catches unresolved placeholders

**Integration test workflow (`.github/workflows/test-sdk-integration.yml`):**
```yaml
# Triggered on: push to main, PR to main, smaqit-sdk release tags
# Purpose: Validate smaqit-sdk can compile future product agents

jobs:
  test-compilation:
    steps:
      - Checkout code
      - Setup environment (Go, dependencies)
      - Run Agent-L2 to compile all 3 patterns:
        * Pattern 1 (Base): base-agent.template.md + base.rules.md
        * Pattern 2 (Spec): specification-agent.template.md + base + specification.rules + layer.rules (×5 layers)
        * Pattern 3 (Impl): implementation-agent.template.md + base + implementation.rules + phase.rules (×3 phases)
      - Validate compiled outputs:
        * All required sections present (Role, Input, Output, Directives, Completion Criteria)
        * No unresolved placeholders ([...])
        * Valid frontmatter structure
        * Proper markdown syntax
      - Fail build if any validation fails
```

## Success Metrics

- smaqit-sdk reaches v1.0.0 (stable compilation architecture)
- smaqit continues v0.x evolution independently
- Clear separation visible in GitHub releases page
- Users can subscribe to smaqit-sdk or smaqit release streams independently
- CI/CD workflows run independently without blocking each other

## Technical Background

**SDK Compilation Patterns (established in Session 046):**

1. **Pattern 1: Base Agents (2-way merge)**
   - Input: `base-agent.template.md` + `base.rules.md`
   - Output: Simple agents (Q&A, helper, custom utilities)
   - Used for: Agents that don't fit specification/implementation workflows

2. **Pattern 2: Specification Agents (4-way merge)**
   - Input: `specification-agent.template.md` + `base.rules.md` + `specification.rules.md` + `{layer}.rules.md`
   - Output: Business, Functional, Stack, Infrastructure, Coverage agents
   - Used for: Agents that generate specifications

3. **Pattern 3: Implementation Agents (4-way merge)**
   - Input: `implementation-agent.template.md` + `base.rules.md` + `implementation.rules.md` + `{phase}.rules.md`
   - Output: Development, Deployment, Validation agents
   - Used for: Agents that execute phases

**Level agents:**
- **Agent-L0** (smaqit.L0.agent.md): Principle curator - maintains framework conceptual purity
- **Agent-L1** (smaqit.L1.agent.md): Template compiler - compiles L0 principles into L1 directives
- **Agent-L2** (smaqit.L2.agent.md): Agent compiler - compiles L1 templates + rules into L2 product agents

**Current product agents (compiled L2 artifacts):**
- Specification agents: Business, Functional, Stack, Infrastructure, Coverage (5)
- Implementation agents: Development, Deployment, Validation (3)
- Total: 8 product agents

**SDK artifact count:**
- Framework files: 7 (SMAQIT.md, AGENTS.md, LAYERS.md, PHASES.md, TEMPLATES.md, PROMPTS.md, ARTIFACTS.md)
- Agent templates: 3 (base, specification, implementation)
- Compilation files: 11 (base + specification + implementation + 5 layers + 3 phases)
- Level agents: 4 (L0, L1, L2, QA)
- New-agent prompt: 1 (smaqit.new-agent.prompt.md)
- Total SDK artifacts: ~26 files

**Pure template pattern (Session 046):**
- Templates contain ONLY structure and placeholders (e.g., `[ROLE_CONTENT]`, `[BASE_MUST_DIRECTIVES]`)
- Compilation files contain directive content (MUST/MUST NOT/SHOULD rules)
- Agent-L2 merges templates + compilation files → product agents
- Zero HTML comments in templates (all meta-guidance moved to compilation files)
