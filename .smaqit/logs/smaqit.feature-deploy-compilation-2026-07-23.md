# Compilation Log: smaqit.feature-deploy

**Timestamp:** 2026-07-23
**Agent:** Agent-L2 (Skill Compiler)
**Pattern:** Skill Compilation (3-way merge)
**Output:** `.github/skills/smaqit.feature-deploy/SKILL.md`

## Sources Read

| Source | Path | Purpose |
|--------|------|---------|
| Definition file | `.smaqit/definitions/skills/smaqit.feature-deploy.md` | Primary input — identity, steps, output, scope, gotchas, examples, completion, failure handling |
| Base skill template | `.smaqit/templates/skills/base-skill.template.md` | Structure — YAML frontmatter, section headers, placeholder slots |
| Skill rules | `.smaqit/templates/skills/compiled/skill.rules.md` | Compilation directives — degrees of freedom, conciseness, reference structure, progressive disclosure, base failure handling pattern |
| Reference skill | `skills/smaqit.feature-new/SKILL.md` | Structural reference — sister skill sharing the same post-MVP context and `[SMAQIT_SKILLS_DIR]` convention |

## Merge Summary

**3-way merge:** template structure + compilation directives + definition content.

### Placeholder Resolution

| Placeholder | Value | Source |
|-------------|-------|--------|
| `[SKILL_NAME]` | `smaqit.feature-deploy` | Definition identity |
| `[SKILL_DESCRIPTION]` | Imperative "Use when..." | Definition identity, rewritten to imperative form, 540 chars |
| `[SKILL_VERSION]` | `"1.0.0"` | Definition identity |
| `[SKILL_TITLE]` | `Feature Deploy — Standalone Post-MVP Deployment` | Derived from skill name |
| `[STEPS_CONTENT]` | Pre-phase + Phase 1–3 with branching | Definition phases, degrees of freedom applied |
| `[OUTPUT_CONTENT]` | 4-item list | Definition output section |
| `[SCOPE_CONTENT]` | 7-item list with redirections | Definition scope section |
| `[EXAMPLES_CONTENT]` | Concrete deploy-to-production example | Definition examples section |
| `[GOTCHAS_CONTENT]` | 5 gotchas | Definition gotchas section |
| `[COMPLETION_CONTENT]` | 4-item checklist | Definition completion section |
| `[FAILURE_HANDLING_CONTENT]` | Base pattern (4 rows) + definition scenarios (7 rows) | Rules base pattern + definition failure handling |
| `[COMPATIBILITY]` | Omitted | Not specified in definition |
| `[ALLOWED_TOOLS]` | Omitted | Not specified in definition |

### Degrees of Freedom Applied

| Step | Fragility | Form |
|------|-----------|------|
| Pre-phase step 3 (provisioning_mode resolution) | High | Exact branching logic with stop conditions |
| Phase 1 step 1 (vault-loader) | Medium | Skill invocation with existing-shared branching notes |
| Phase 1 step 2 (provision-cyso) | Medium | Skill invocation with per-mode branching notes |
| Phase 1 step 3 (vm-bootstrap) | Low | Prose instruction with IP source notes |
| Phase 2 step 3 (git push) | High | Exact command in code block |
| Phase 2 step 4 (gh run watch) | High | Exact command in code block |
| Phase 2 step 6 (amendment gate) | High | Exact command with `[SMAQIT_SKILLS_DIR]` runtime placeholder |
| Phase 3 step 2 (re-scan) | High | Exact command (same as Phase 2 step 6) |
| Phase 3 step 3 (release tagging) | Medium | Conditional prose with skill chain |

### Conciseness Filter

Removed from definition:
- "standalone post-MVP deployment" repetition in steps (title covers this)
- "Extracted from `smaqit.new-greenfield-project` Phase 4 steps 2–5" context (scope section covers the relationship)
- Redundant "Invoke `/smaqit.deployment` agent" explanation of what the agent receives (redundant with the agent's own SKILL.md)

Retained:
- All branching logic (existing-owned vs existing-shared vs provision) — critical for correct execution
- All gate conditions — stop-and-flag scenarios must be explicit
- All gotchas — proven error-prone areas from definition

## Validation Checklist

- [x] No unresolved compile-time placeholders (`[SKILL_NAME]`, `[SKILL_DESCRIPTION]`, etc. all resolved)
- [x] `[SMAQIT_SKILLS_DIR]` intentionally retained as runtime placeholder (resolved by `scripts/generate-agents.py`, consistent with `smaqit.feature-new`)
- [x] Description uses imperative phrasing ("Use when...")
- [x] Description under 1024 characters
- [x] All required sections present (Steps, Output, Scope, Examples, Gotchas, Completion, Failure Handling)
- [x] Optional frontmatter fields (`compatibility`, `allowed-tools`) omitted — not specified in definition
- [x] Body under 400 lines (153 lines)
- [x] No progressive disclosure extraction needed
- [x] No nested reference chains
- [x] User directives compatible with base rules (no conflicts detected)
- [x] Degrees of freedom correctly applied per step fragility
- [x] Conciseness filter applied — no sentences an agent would infer without them
- [x] No principle explanations or rationale (L0 contamination)
- [x] No template placeholders (L1 contamination)

## Issues / Decisions

None. Definition was complete and all sections were well-specified. No `[?]` annotations found in the definition file.

## `[?]` Annotation Report

A full scan of `.smaqit/definitions/skills/smaqit.feature-deploy.md` found **zero** `[?]` annotations. All fields in the definition are fully specified.

The only unresolved token in the compiled output is `[SMAQIT_SKILLS_DIR]`, which is a **runtime placeholder** (not a `[?]`-annotated field) resolved by `scripts/generate-agents.py` during the installer build step. This is the same convention used by `smaqit.feature-new` and other cross-referencing skills in this repo.
