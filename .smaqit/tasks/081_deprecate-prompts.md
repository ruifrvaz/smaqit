# Task 081: Deprecate Prompts

**Status:** New  
**Priority:** High  
**Created:** 2026-05-09

## Description

Prompts were designed as "versioned input records" — files users fill with requirements that agents read to generate specs. In practice, agents always received requirements directly from session context and wrote directly to specs. Prompt files remained empty placeholders. The concept never materialized.

This task performs a full, no-backward-compatibility deprecation of the entire prompts feature: prompt files, the prompts framework document, prompt templates, all installer infrastructure, `prompt_version` from spec frontmatter, and every documentation and agent reference.

**Replacement for agent input:** Agents now read requirements from current session context (including context in compacted blocks) or open tasks. Assessment skill is applied when input is ambiguous or insufficient.

## Acceptance Criteria

- [ ] No `prompts/` directories exist anywhere in the repo (source, installer, test fixtures)
- [ ] No `templates/prompts/` directory exists
- [ ] `framework/PROMPTS.md` is deleted
- [ ] No agent file contains a reference to `.github/prompts/` or a prompt file as input source
- [ ] No framework file references prompt files as requirements input for layers
- [ ] `prompt_version` field is removed from spec frontmatter struct in `installer/spec.go`
- [ ] `prompt_version` field is removed from all spec template frontmatter sections
- [ ] Installer does not embed, copy, or uninstall `.github/prompts/`
- [ ] `installer/main.go` builds successfully after all changes
- [ ] Wiki files that exist solely for the prompts concept are deleted
- [ ] Remaining wiki files have prompt references cleaned up
- [ ] README quickstart no longer references prompt files
- [ ] `.github/copilot-instructions.md` Kit Components tree is updated

## Scope Note

**Do NOT touch:**
- `.smaqit/tasks/` and `.smaqit/history/` — task and history files that mention prompts are historical record, leave them
- `.github/prompts/` in the smaqit repo itself — these are session/task management prompts (session.start, task.create, release commands), a completely different category, not the layer input prompts being deprecated
- `docs/user-testing/` and `docs/logs/` — historical test reports, leave them
- `installer/test/mario-hello[-ci]/specs/` — spec content files in test fixtures, leave them

## Step-by-Step Implementation

### Step 1 — Delete Dead Files

These files/directories have no value post-deprecation. Delete them outright.

**Directories to delete entirely:**
```
prompts/                                              # 6 source prompt files
templates/prompts/                                    # specification-prompt.template.md, implementation-prompt.template.md
installer/prompts/                                    # mirrored prompt files for embedding
installer/test/mario-hello/.github/prompts/           # test fixture prompt files
installer/test/mario-hello-ci/.github/prompts/        # test fixture prompt files
```

**Files to delete:**
```
framework/PROMPTS.md
docs/wiki/concepts/prompts-as-input-records.md
docs/wiki/designs/free-style-prompts.md
docs/wiki/patterns/archiving-prompts.md
docs/wiki/patterns/prompt-evolution.md
docs/wiki/workflows/amending-requirements.md          # entirely about editing prompts to change requirements
docs/wiki/workflows/managing-stale-specs.md           # entirely about prompt_version staleness detection
docs/wiki/patterns/html-comment-convention.md         # HTML comments were for prompt template examples only
```

Use `rm -rf` for directories, `rm` for files. Verify each is gone.

### Step 2 — Update Framework Files

All framework files are mirrored in `installer/framework/`. Update the source files in `framework/` first, then copy each updated file to `installer/framework/` to keep them in sync.

#### `framework/SMAQIT.md`

- In **"Layer Independence"** principle: remove "Each layer's prompt file is the sole source of requirements for that layer." Rewrite to: layer requirements come from session context (user input in chat). Upstream layers provide context for coherence, not requirements.
- In **"Traceability Across Layers"** principle: remove "Each layer receives requirements from its prompt file." Replace with session context as the input origin.
- Delete the **"Reproducible from Input Set"** principle entirely — it is defined entirely around identical prompt sets producing equivalent behavior. This concept does not survive prompt removal.
- Remove any remaining references to `.github/prompts/` or prompt files.

#### `framework/LAYERS.md`

- Remove the opening paragraph under "Layer Independence" that says "Each layer has a dedicated prompt file in `.github/prompts/`"
- In **Layer Definitions** section: each layer's "Input" description says "User requirements from prompt file" — change to "User requirements from session context (chat input, compacted context blocks, or open tasks)"
- Remove any table rows listing prompt files per layer

#### `framework/PHASES.md`

- In the **Develop phase** Specification Agents table: the "User Input" column lists prompt files (e.g., "Prompt: smaqit.business.prompt.md"). Change to "Session context"
- Remove "Input sufficiency check for required prompt files" from Pre-Orchestration Validation section
- Remove any references to prompt files as phase inputs

#### `framework/AGENTS.md`

- Delete the entire "Reading Prompts" section (currently lines 11–19 approximately): removes "Agents receive requirements from prompts in `.github/prompts/`" and the sub-bullets
- In **Specification Agents** section: remove the "Prompt file" row from the agent input table (line ~124). Change input description to "Session context (chat input, compacted context, or open tasks)"
- In the Specification Agent output MUST list: remove `prompt_version` field generation requirement (lines ~137–138)
- In the agent registry table (lines ~181–185): remove the prompt file column
- Remove all remaining `.github/prompts/` path references
- Remove cross-reference to `[PROMPTS](PROMPTS.md)` (line ~19)

#### `framework/TEMPLATES.md`

- Remove the **"Prompt Templates"** section entirely (the section describing `templates/prompts/` structure, frontmatter format, and comment conventions — roughly lines 190–275)
- Remove the "Prompt templates" row from the templates registry table at the top
- Remove `prompt_version` from the spec frontmatter template block (line ~103) and its description (line ~111)
- Remove `<!-- Example: ... -->` HTML comment convention content — this was solely for prompt templates

After updating each file, copy to `installer/framework/`:
```bash
cp framework/SMAQIT.md installer/framework/SMAQIT.md
cp framework/LAYERS.md installer/framework/LAYERS.md
cp framework/PHASES.md installer/framework/PHASES.md
cp framework/AGENTS.md installer/framework/AGENTS.md
cp framework/TEMPLATES.md installer/framework/TEMPLATES.md
```

### Step 3 — Update Spec Templates

Each spec template has `prompt_version: [GIT_HASH]` in its YAML frontmatter block. Remove that field.

Files to edit (source and installer mirror — edit both):
```
templates/specs/business.template.md
templates/specs/functional.template.md
templates/specs/stack.template.md
templates/specs/infrastructure.template.md
templates/specs/coverage.template.md
installer/templates/specs/business.template.md
installer/templates/specs/functional.template.md
installer/templates/specs/stack.template.md
installer/templates/specs/infrastructure.template.md
installer/templates/specs/coverage.template.md
```

In each file, find the frontmatter block and remove the `prompt_version` line. Example before:
```yaml
---
id: [ID]
status: draft
created: [DATE]
prompt_version: [GIT_HASH]
---
```
After:
```yaml
---
id: [ID]
status: draft
created: [DATE]
---
```

### Step 4 — Update Agent Files

All 9 specification agents have an `## Input` section that starts with a "**Prompt File:**" subsection pointing to `.github/prompts/smaqit.[layer].prompt.md`. This must be replaced.

Files to edit (source and installer mirror — edit both):
```
agents/smaqit.business.agent.md         + installer/agents/smaqit.business.agent.md
agents/smaqit.functional.agent.md       + installer/agents/smaqit.functional.agent.md
agents/smaqit.stack.agent.md            + installer/agents/smaqit.stack.agent.md
agents/smaqit.infrastructure.agent.md   + installer/agents/smaqit.infrastructure.agent.md
agents/smaqit.coverage.agent.md         + installer/agents/smaqit.coverage.agent.md
agents/smaqit.development.agent.md      + installer/agents/smaqit.development.agent.md
agents/smaqit.deployment.agent.md       + installer/agents/smaqit.deployment.agent.md
agents/smaqit.validation.agent.md       + installer/agents/smaqit.validation.agent.md
agents/smaqit.qa.agent.md               + installer/agents/smaqit.qa.agent.md
```

**Specification agents (business, functional, stack, infrastructure, coverage):**

Replace the prompt file subsection:
```markdown
**Prompt File:** `.github/prompts/smaqit.[layer].prompt.md`

- Read requirements from prompt file
- Ignore all HTML comments (`<!-- Example: ... -->`) to prevent example pollution
- Interpret free-style natural language without rigid structure enforcement
- Validate sufficiency - if content insufficient, request clarification with natural language guidance
```

With:
```markdown
**Session Context:**

- Read requirements from current session context (including context in compacted blocks) or open tasks
- Apply assessment skill when input is ambiguous, conflicting, or insufficient
```

**Implementation agents (development, deployment, validation):**

These agents have a "Conflict Resolution" note that says "When prompt requirements conflict with upstream specs..." — change "prompt requirements" to "user requirements".

**QA agent:**

Check for and remove any prompt file references in its Input or workflow sections.

After editing each source agent, copy to the installer mirror:
```bash
for agent in business functional stack infrastructure coverage development deployment validation qa; do
  cp agents/smaqit.$agent.agent.md installer/agents/smaqit.$agent.agent.md
done
```

### Step 5 — Update Installer Go Code

#### `installer/spec.go`

Remove the `PromptVersion` field from `SpecFrontmatter` struct:
```go
// Before:
type SpecFrontmatter struct {
    ID            string    `yaml:"id"`
    Status        string    `yaml:"status"`
    Created       time.Time `yaml:"created"`
    Implemented   time.Time `yaml:"implemented,omitempty"`
    Deployed      time.Time `yaml:"deployed,omitempty"`
    Validated     time.Time `yaml:"validated,omitempty"`
    PromptVersion string    `yaml:"prompt_version,omitempty"`
}

// After:
type SpecFrontmatter struct {
    ID          string    `yaml:"id"`
    Status      string    `yaml:"status"`
    Created     time.Time `yaml:"created"`
    Implemented time.Time `yaml:"implemented,omitempty"`
    Deployed    time.Time `yaml:"deployed,omitempty"`
    Validated   time.Time `yaml:"validated,omitempty"`
}
```

#### `installer/main.go`

**Remove the embed directive and variable:**
```go
// Remove these two lines:
//go:embed prompts/*.md
var promptFiles embed.FS
```

**Remove from conflict detection table** (around line 293):
```go
// Remove this entry from the installTargets or conflictTargets slice:
{promptFiles, "prompts", ".github/prompts", false},
```

**Remove from uninstall path list** (around line 390):
```go
// Remove this entry:
".github/prompts",
```

**Remove the prompt copy block** (around lines 414–435):
```go
// Remove:
// Copy prompt files
if err := copyEmbeddedDir(promptFiles, "prompts", ".github/prompts"); err != nil {
    fmt.Printf("Error copying prompt files: %v\n", err)
    ...
}
fmt.Println("✓ Copied prompt files")
```

**Remove from status output** (around line 510):
```go
// Remove:
fmt.Println("  • .github/prompts/")
```

**Remove from uninstall removal block** (around lines 561–565):
```go
// Remove:
if err := os.RemoveAll(filepath.Join(".github", "prompts")); err != nil && !os.IsNotExist(err) {
    fmt.Printf("Error removing .github/prompts/: %v\n", err)
}
    fmt.Println("✓ Removed .github/prompts/")
```

**Remove from conflict check list** (around line 611):
```go
// Remove:
".github/prompts",
```

Also update any help text strings (lines ~84, ~102) that mention prompt files or `.github/prompts/`.

#### Build Verification

```bash
cd installer && make build
```

Fix any compilation errors. The build MUST succeed before continuing.

### Step 6 — Update Wiki Files

#### Files to fully rewrite or heavily edit

**`docs/wiki/concepts/layer-independence.md`**  
The core concept (layers don't derive requirements from each other) survives. Remove the prompt-file-per-layer table and all references to `.github/prompts/`. Rewrite the concept around session context as the input source: each layer receives requirements from session context (user input at that layer's agent invocation), not from upstream layers.

**`docs/wiki/concepts/stateful-specifications.md`**  
Remove `prompt_version` field from the frontmatter example block and from all explanatory text about staleness detection. The lifecycle states (draft, implemented, deployed, validated) are unaffected.

**`docs/wiki/concepts/team-alignment.md`**  
Replace "X role fills the Y prompt" with "X role provides Y requirements when invoking the Y agent". Remove the parallel prompt-filling workflow.

**`docs/wiki/concepts/traceability.md`**  
Rewrite the traceability chain origin. Instead of "prompt file → spec → implementation → test", it becomes "session context → spec → implementation → test". Update any ASCII diagrams showing prompt files.

**`docs/wiki/concepts/accept-mutability.md`**  
Remove the "same prompt set" equivalence framing. Minor cleanup only.

**`docs/wiki/concepts/bounded-agents.md`**  
Line ~33: "Suggest: Provide prompt file or agent invocation command" — remove the prompt file suggestion. Line ~196: remove "Layer Independence: Each layer's prompt is the sole source of requirements" — update to session context. Minor cleanup.

**`docs/wiki/concepts/user-vs-agent-documentation.md`**  
Remove "prompt templates" from the `templates/**/*.template.md` list.

**`docs/wiki/designs/fail-fast-on-ambiguity.md`**  
Rewrite to session-context model. The core behavior (agents halt when input is insufficient) is preserved but the mechanism changes: instead of checking if `.github/prompts/smaqit.business.prompt.md` is empty, agents evaluate session context and apply the assessment skill. Remove the phase prompt check examples. Remove cross-reference to `free-style-prompts.md`.

**`docs/wiki/designs/template-constraints.md`**  
Remove the "Prompt templates" section (the `templates/prompts/` tree, requirements, format). Keep spec and agent template content.

**`docs/wiki/designs/layer-references-upstream.md`**  
Rewrite the opening premise. Instead of "requirements come FROM prompts, references point TO upstream specs", it becomes "requirements come FROM session context, references point TO upstream specs for coherence". Remove prompt-specific framing throughout.

**`docs/wiki/designs/progressive-refinement.md`**  
Remove "each layer reads from its own prompt file" as the basis for independence. Replace with "each layer's agent invocation is a distinct session context". Remove the "5 prompt files vs 1" trade-off note.

**`docs/wiki/designs/why-non-functional-requirements.md`**  
Lines ~130–134: remove the "Layer Independence means each layer receives requirements from its own prompt file" example block with the prompt examples. Replace with session context framing. Line ~284: remove the session history footnote about Layer Independence and prompt files.

**`docs/wiki/designs/hierarchical-levels.md`**  
Remove all prompt-related entries from the Level 2 architecture description (prompt files, prompt templates).

**`docs/wiki/designs/level-up-compilation.md`**  
Line ~117: remove "MUST read from [LAYER] prompt file" from the compilation example.

**`docs/wiki/patterns/validation-messages.md`**  
Rewrite to session-context model. Agents evaluate sufficiency of session context rather than checking prompt file content. The guidance pattern (natural language, not error codes) is preserved.

**`docs/wiki/workflows/quickstart.md`**  
Major rewrite. Currently Step 1–3 are "fill business prompt", "fill functional prompt", "fill stack prompt". Replace with: invoke the agent directly in chat and provide your requirements in the conversation. The agent will ask clarifying questions if needed (via assessment skill). Update the "Next Steps" section. Remove all `.github/prompts/` references.

**`docs/wiki/workflows/extending-smaqit.md`**  
Remove "prompt templates" from the templates count and tree. Remove the "Using the new-agent prompt" section. Remove `smaqit.new-agent.prompt.md` from the artifact list.

**`docs/wiki/troubleshooting.md`**  
Line ~17: "GitHub Copilot maintains session state across prompt invocations" — update wording to "agent invocations" if context is about session state carryover between agent sessions, not prompt files.

### Step 7 — Update README and Copilot Instructions

#### `README.md`

Find the quickstart section (currently: "1. Fill `.github/prompts/smaqit.business.prompt.md` with your requirements"). Rewrite to:
1. Open GitHub Copilot chat and run `/smaqit.development`
2. Describe your requirements in the conversation when the agent asks
3. Watch specs generate, then code build

Remove any other references to filling prompt files.

#### `.github/copilot-instructions.md`

In the **Kit Components** section Source tree, remove the `prompts/` entry. Remove any "Prompt templates" entries from `templates/`. Verify the Installer description no longer mentions prompt files being copied.

### Step 8 — Verify No Stray References

Run this check after all edits are complete:

```bash
cd /path/to/smaqit

# Should return no results (outside history/tasks/user-testing/logs dirs):
grep -r "\.github/prompts/smaqit\." \
  --include="*.md" --include="*.go" --include="*.yml" \
  --exclude-path="*/.smaqit/tasks/*" \
  --exclude-path="*/.smaqit/history/*" \
  --exclude-path="*/docs/user-testing/*" \
  --exclude-path="*/docs/logs/*" \
  .

# Should return no results:
grep -r "prompt_version" \
  --include="*.md" --include="*.go" \
  --exclude-path="*/.smaqit/tasks/*" \
  --exclude-path="*/.smaqit/history/*" \
  .

# Should return no results:
grep -r "smaqit\.\(business\|functional\|stack\|infrastructure\|coverage\|development\|deployment\|validation\)\.prompt\.md" \
  --include="*.md" --include="*.go" \
  --exclude-path="*/.smaqit/tasks/*" \
  --exclude-path="*/.smaqit/history/*" \
  --exclude-path="*/docs/user-testing/*" \
  .
```

Review any remaining hits and determine if they are legitimate (session management prompts in `.github/prompts/` that aren't layer input files) or stray references that need cleanup.

### Step 9 — Final Build and Smoke Test

```bash
cd installer
make build
./dist/smaqit-dev init /tmp/smaqit-test-081
ls /tmp/smaqit-test-081/.github/
# Should show: agents/  skills/  (NO prompts/)
./dist/smaqit-dev status
./dist/smaqit-dev validate
```

Verify:
- `.github/prompts/` is NOT created by `smaqit init`
- `smaqit status` runs without errors
- `smaqit validate` runs without errors

## Notes

- The `smaqit.new-agent.prompt.md` file was already functionally dead (its agent, Agent-L2, was removed in the smaqit-adk extraction). It is deleted as part of Step 1 with the rest of the prompts directory.
- `installer/framework/` files are mirrors of `framework/` files. After updating source files, copy them to installer. Do not edit installer framework files independently.
- `installer/agents/` files are mirrors of `agents/`. After updating source agent files, copy them to installer.
- The `.github/prompts/` directory in the smaqit repo (containing session.start, task.create, release prompts) is NOT affected — these are workflow command prompts, not layer input prompts.
