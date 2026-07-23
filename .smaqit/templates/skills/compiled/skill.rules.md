---
type: base-skill
target: templates/skills/base-skill.template.md
sources:
  - framework/SKILLS.md (all principles)
created: 2026-03-02
---

# Skill Rules

This file is the L1 vocabulary and compiled directives for skills. It defines the named things that exist in the skill format, their structure, loading behavior, and the directives Agent-L2 MUST follow when compiling a skill from a definition file.

## Source L0 Principles

| Source File | Section |
|-------------|---------|
| SKILLS.md | Instructions, Not Data |
| SKILLS.md | Progressive Disclosure |
| SKILLS.md | Instruction-Only Content |
| SKILLS.md | Context-Driven Input |
| SKILLS.md | Description-Driven Activation |

---

## Skill Format

Skills are expressed as YAML frontmatter followed by a markdown body.

**Required frontmatter fields:**

| Field | Constraint |
|-------|------------|
| `name` | Skill identifier. Maximum 64 characters. Lowercase letters, numbers, and hyphens only. Loaded at discovery alongside `description`. |
| `description` | Activation signal. Maximum 1024 characters. Must use imperative phrasing directed at the agent: "Use this skill when..." or "Use when...". Must describe user intent (what the user is trying to achieve) not implementation mechanics. Must state what the skill produces. Should cover indirect trigger contexts — situations where the skill is relevant even if the user doesn't use the exact domain keyword. |
| `metadata.version` | Semantic version string for the skill. |

**Body:** Markdown prose containing the skill's procedure. The body is read-only at runtime — unchanged between executions.

---

## Loading Stages

| Stage | What loads | Typical constraint |
|-------|-----------|-------------------|
| Discovery | `name` + `description` only | ~100 tokens per skill |
| Activation | Full `SKILL.md` body | < 5000 tokens recommended |
| Execution | Referenced external files | On demand |

Discovery is cheap by design — the full body never loads unless the skill is relevant. The description carries the activation decision; the body carries the execution instructions.

**Structural growth:** When a skill requires supporting reference content, that content lives in files bundled within the skill directory — including in subdirectories (`scripts/`, `references/`, `assets/`). All such files MUST be referenced directly from `SKILL.md`. Reference files longer than 100 lines MUST include a table of contents at the top. Reference chains must not be nested: a file referenced by `SKILL.md` cannot itself reference another file.

**Progressive disclosure directive:** When compiling a skill whose body would exceed 400 lines, Agent-L2 MUST extract detailed reference content into a `references/` subdirectory and link from `SKILL.md` with explicit load conditions. Example: "Read `references/api-errors.md` if the API returns a non-200 status." The main `SKILL.md` body must remain under 400 lines after extraction.

---

## Agent-Skill Interaction Model

Skills are consumed by agents in three stages that correspond to the loading stages above.

**Discovery** — The active agent reads `name` and `description` from every available skill. This happens at session start or when the agent needs to determine which skill matches the current task. No body content loads at this stage.

**Activation** — When the agent determines a skill matches, it reads the full `SKILL.md` body into context. This is a deliberate read — the agent has committed to the skill before loading its body.

**Execution** — The agent follows the skill's instructions step by step. If the skill references external files, the agent reads them on demand during execution. The skill body does not change as a result of execution.

---

## Placeholder Catalog

The following placeholders appear in `templates/skills/base-skill.template.md`. Agent-L2 MUST resolve every placeholder when compiling a skill from a definition file.

| Placeholder | Description |
|-------------|-------------|
| `[SKILL_NAME]` | Skill identifier in YAML frontmatter `name` field. Lowercase, hyphens only. |
| `[SKILL_DESCRIPTION]` | Activation signal in YAML frontmatter `description` field. Maximum 1024 characters. Must use imperative phrasing ("Use when..." or "Use this skill when..."). Must describe the user's intent and what the skill produces. Should include indirect trigger contexts. Must not describe internal implementation mechanics. |
| `[SKILL_VERSION]` | Semantic version string in `metadata.version` field. Start at `"1.0.0"`. |
| `[SKILL_TITLE]` | Display heading for the compiled skill document. |
| `[STEPS_CONTENT]` | Ordered procedure the agent follows. Each step written at the appropriate degree of freedom (see Degrees of Freedom below). |
| `[OUTPUT_CONTENT]` | Artifact produced, its path, and the subagent invoked (if any). |
| `[SCOPE_CONTENT]` | What this skill does NOT handle. Redirections to other agents or skills where applicable. |
| `[EXAMPLES_CONTENT]` | At least one concrete example: a representative triggering request and the output produced. |
| `[GOTCHAS_CONTENT]` | Environment-specific facts the agent must know before executing. Non-obvious corrections to mistakes it would make without instruction. Proactive, not reactive. If none apply, write "None identified." |
| `[COMPLETION_CONTENT]` | How the agent knows the skill is done. Checklist of verifiable conditions. |
| `[FAILURE_HANDLING_CONTENT]` | Situation/action table for failure modes. |
| `[COMPATIBILITY]` | Optional. Environment requirements for this skill (agent product, system packages, network). Omit from compiled output if not specified. |
| `[ALLOWED_TOOLS]` | Optional. Space-separated pre-approved tools (e.g., `Bash(git:*) Read`). Omit from compiled output if not specified. |

---

## L1 Directive Compilation

### Compilation Directives

**Instructions-Not-Data:**
- Write skill body as a procedure — gathering steps, validation rules, compilation guidance only
- MUST NOT write user requirements or gathered input into the skill body
- MUST NOT write execution records or state into the skill body

**Description-Driven Activation:**
- MUST write `[SKILL_DESCRIPTION]` using imperative phrasing directed at the agent: "Use this skill when..." or "Use when..."
- MUST describe the user's intent and what the skill produces — not internal implementation mechanics
- SHOULD cover indirect trigger contexts where the skill is relevant even if the user doesn't use the exact domain keyword
- MUST NOT use first person ("I can...") or declarative third-person opening ("Creates a...", "Packages a...")
- MUST write the description with enough precision to distinguish the skill from adjacent skills

**Degrees of Freedom:**
- For each step in `[STEPS_CONTENT]`, match instruction form to step fragility:
  - High fragility (exact sequence required, errors are costly): write literal commands or exact instructions
  - Medium fragility (preferred pattern exists, variation acceptable): write template or pseudocode
  - Low fragility (many valid approaches, context determines best path): write prose guidance
- MUST NOT over-specify low-fragility steps — assume Claude can determine the best approach
- MUST over-specify high-fragility steps — do not leave critical sequences to interpretation

**Conciseness:**
- MUST NOT explain concepts Claude can derive from the task or from general knowledge
- MUST include only: project-specific paths and constraints, non-obvious sequences, domain-specific rules
- Each sentence in `[STEPS_CONTENT]` must justify its token cost — if Claude would do the right thing without it, omit it

**Reference Structure:**
- External files referenced from `SKILL.md` may be anywhere within the skill directory, including subdirectories
- Reference files longer than 100 lines MUST include a table of contents at the top
- MUST NOT create nested reference chains (SKILL.md → file-a.md → file-b.md is forbidden)

### Base Failure Handling Pattern

Insert the following as the foundation of `[FAILURE_HANDLING_CONTENT]`:

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |

---

## Compilation Guidance for Agent-L2

When compiling a skill from a definition file:

1. **Read definition file** (`.smaqit/definitions/skills/[name].md`) for skill specifications
2. **Confirm definition is complete** — all required sections present (identity, steps, output, scope, completion, failure handling). If any section is missing, stop and request it before proceeding.
3. **Read base skill template** (`templates/skills/base-skill.template.md`) for structure
4. **Read this rules file** for directives and placeholder catalog
5. **Compile:**
   - Fill `[SKILL_NAME]` from definition identity
   - Fill `[SKILL_DESCRIPTION]` from definition identity — imperative phrasing, user intent + output; if not explicitly provided, infer from skill name and project context
   - Fill `[SKILL_VERSION]` from definition identity (default `"1.0.0"` if not specified)
   - Fill `[SKILL_TITLE]` as a readable title derived from the skill name
   - Fill `[STEPS_CONTENT]` from definition steps — apply degrees of freedom per step fragility
   - Fill `[OUTPUT_CONTENT]` from definition output section
   - Fill `[SCOPE_CONTENT]` from definition scope section
   - Fill `[EXAMPLES_CONTENT]` from definition examples — at least one concrete triggering request and the output produced
   - Fill `[GOTCHAS_CONTENT]` from definition gotchas — proactive environment-specific facts; write "None identified." if none apply
   - Fill `[COMPLETION_CONTENT]` from definition completion section
   - Fill `[FAILURE_HANDLING_CONTENT]` with base failure handling pattern + definition failure scenarios
   - Include `compatibility: [COMPATIBILITY]` frontmatter field ONLY if the definition specifies a value; omit entirely otherwise
   - Include `allowed-tools: [ALLOWED_TOOLS]` frontmatter field ONLY if the definition specifies a value; omit entirely otherwise
6. **Apply progressive disclosure** — if the compiled skill body would exceed 400 lines, extract detailed reference content into a `references/` subdirectory and link from `SKILL.md` with explicit load conditions. The main `SKILL.md` body must remain under 400 lines after extraction.
7. **Apply conciseness filter** — review every sentence in the compiled body against the conciseness directive before writing
8. **Write output** to `skills/[SKILL_NAME]/SKILL.md`
9. **Validate:** No unresolved placeholders remain; description uses imperative phrasing; all sections present; optional frontmatter fields omitted when not needed
