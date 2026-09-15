# Compilation Log: smaqit.infrastructure-image-build-push-static

**Timestamp:** 2026-09-15
**Agent:** Agent-L2 (Skill Compiler)
**Pattern:** Skill Compilation (3-way merge)
**Output:** `skills/smaqit.infrastructure-image-build-push-static/SKILL.md` (ADK-source location — this
repo is smaqit's own canonical source, not a consumer project; no `.agents/skills/` or
`.claude/skills/` directories exist here)

## Sources Read

| Source | Path | Purpose |
|--------|------|---------|
| Definition file | `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md` | Primary input — identity, provenance, required-inherited-context, steps, output, scope, completion, failure handling, gotchas, allowed tools, examples |
| Base skill template | `~/.agents/smaqit-adk/templates/skills/base-skill.template.md` | Structure — YAML frontmatter, section headers, placeholder slots |
| Skill rules | `~/.agents/smaqit-adk/templates/skills/compiled/skill.rules.md` | Compilation directives — degrees of freedom, conciseness, base failure-handling pattern, placeholder catalog |
| Reference skill (compiled) | `skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md` | Structural/style reference — target shape and section order for this repo's canonical compiled skills |
| Reference skill (definition) | `.smaqit/definitions/skills/smaqit.infrastructure-deploy-k3s-app.md` | Confirmed the Provenance-and-Required-inherited-context-drop convention (present in definition, absent from compiled output) |
| Reference skill pair | `skills/smaqit.infrastructure-deploy-rsync-python-tornado/SKILL.md` + its definition | Confirmed section order (Steps, Output, Scope, Examples, Gotchas, Completion, Failure Handling, Allowed Tools) and that `Required-inherited-context` content is folded inline rather than kept as its own heading |

## Merge Summary

**3-way merge:** base-skill template structure + skill.rules.md compilation directives +
definition content.

### Placeholder Resolution

| Placeholder | Value | Source |
|-------------|-------|--------|
| `[SKILL_NAME]` | `smaqit.infrastructure-image-build-push-static` | Definition identity |
| `[SKILL_DESCRIPTION]` | "Use when containerizing a static HTML/CSS/vanilla-JS site..." (886 chars) | Definition Description section, condensed to match this repo's established description length/style (~700-900 chars, per `smaqit.infrastructure-deploy-k3s-app`/`smaqit.infrastructure-cicd-generate`) |
| `[SKILL_VERSION]` | `"1.0.0"` | Explicit instruction — first release into canonical smaqit |
| `[SKILL_TITLE]` | `Containerize a Static Site for k3s Deployment` | Derived from skill name/purpose |
| `[STEPS_CONTENT]` | 2 numbered steps with sub-bullets (Dockerfile authoring; Deployment/Service manifest authoring) | Definition Steps section, carried near-verbatim (already correctly leveled MUST-style directives) |
| `[OUTPUT_CONTENT]` | 3-item list | Definition Output section |
| `[SCOPE_CONTENT]` | 5-item list with redirections | Definition Scope section |
| `[EXAMPLES_CONTENT]` | Single concrete input/output example | Definition Examples section |
| `[GOTCHAS_CONTENT]` | 3 gotchas | Definition Gotchas section |
| `[COMPLETION_CONTENT]` | 5-item checklist | Definition Completion section |
| `[FAILURE_HANDLING_CONTENT]` | Base pattern (4 rows) + definition scenarios (4 rows) | Rules base pattern + definition Failure Handling table |
| `[COMPATIBILITY]` | Omitted | Not specified in definition |
| `[ALLOWED_TOOLS]` | `Bash(docker:*), Bash(kubectl:*), Read, Write, Edit` | Definition Allowed Tools section |

### Sections Dropped from Definition (by established repo convention)

- **Provenance** — stays only in `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md`; confirmed by diffing `smaqit.infrastructure-deploy-k3s-app`'s definition (has Provenance) against its compiled `SKILL.md` (omits it).
- **Required-inherited-context** — every fact in this section (Namespace/manifest-path convention, Ingress-not-authored-here, build-and-push-job-lives-elsewhere, DEPLOY_SHA-as-pod-env-not-file) is already restated inline within Steps or Scope, matching the precedent set by both `smaqit.infrastructure-deploy-k3s-app` and `smaqit.infrastructure-deploy-rsync-python-tornado`, whose compiled outputs likewise drop this heading and fold its content into Steps/Scope without a separate section or cross-reference footnotes.

### Section Order

Per explicit instruction and precedent (`smaqit.infrastructure-deploy-rsync-python-tornado`):
Steps → Output → Scope → Examples → Gotchas → Completion → Failure Handling → Allowed Tools. No
separate `Pre-conditions` heading was introduced — the definition's `## Steps` section (unlike
`smaqit.infrastructure-deploy-k3s-app`'s) has no discrete `### Pre-conditions` subsection to
extract, and the base template itself has no `[PRECONDITIONS_CONTENT]` placeholder; inventing one
would have meant fabricating content not present in the source.

### Degrees of Freedom Applied

| Step | Fragility | Form |
|------|-----------|------|
| Step 1 (Dockerfile authoring: base image, COPY, USER switching, health.json bake, .dockerignore) | High | Exact base image, exact USER sequencing, exact rationale for each — errors here are costly and non-obvious (admission-time failures, not local `docker run` failures) |
| Step 2 (Deployment/Service manifest authoring, `runAsUser`, `imagePullSecrets`) | High | Exact numeric UID guidance, exact error message to recognize, exact remediation command |

Both steps in this skill are high-fragility by nature (the entire skill exists to encode two
non-obvious, kubelet-admission-specific facts a generic guide would omit) — no low/medium
fragility steps were compressed further, consistent with the "MUST over-specify high-fragility
steps" directive.

### Conciseness Filter

Reviewed every sentence against the definition; no cuts were needed beyond the description
condensation above — the definition's Steps/Scope/Gotchas/Failure-Handling content was already
written at an appropriate density (project-specific paths, non-obvious sequences, domain-specific
rules only). Minor rewording: "with one addition this skill's own hard-won fact requires" →
"with one addition:" (removed self-referential narrative framing, kept the directive).

## Validation Checklist

- [x] No unresolved compile-time placeholders (`[SKILL_NAME]`, `[SKILL_DESCRIPTION]`, etc. all resolved; verified via `grep -n '\[[A-Z_]*\]'` — zero matches)
- [x] Description uses "Use when..." imperative phrasing, matching this repo's established convention (confirmed against `smaqit.infrastructure-deploy-k3s-app`, `smaqit.infrastructure-cicd-generate`, `smaqit.infrastructure-vm-bootstrap`, `smaqit.infrastructure-repo-config`)
- [x] Description under 1024 characters (886 chars, verified via YAML parse)
- [x] `metadata.version: "1.0.0"` set per explicit instruction
- [x] All required sections present (Steps, Output, Scope, Examples, Gotchas, Completion, Failure Handling, Allowed Tools)
- [x] Provenance section correctly omitted from compiled output (definition-only, per repo convention)
- [x] Optional frontmatter fields (`compatibility`, `allowed-tools` as YAML field) omitted — this repo's convention expresses allowed tools as a body section, not a frontmatter field, matching `smaqit.infrastructure-deploy-k3s-app`
- [x] Body under 400 lines (118 lines) — no progressive-disclosure extraction to `references/` needed
- [x] No nested reference chains
- [x] Degrees of freedom correctly applied per step fragility
- [x] Conciseness filter applied
- [x] No principle explanations or rationale (L0 contamination)
- [x] No template placeholders (L1 contamination)
- [x] Written to exactly one location (`skills/smaqit.infrastructure-image-build-push-static/SKILL.md`) — this repo is smaqit's own source, not a downstream consumer project, so no `.agents/skills/` or `.claude/skills/` duplication applies

## Issues / Decisions

- **`allowed-tools` as body section vs. frontmatter field:** `skill.rules.md`'s base template
  offers an optional `allowed-tools:` frontmatter field, but every canonical skill inspected in
  this repo (including the reference `smaqit.infrastructure-deploy-k3s-app`) instead expresses
  allowed tools as a trailing `## Allowed Tools` body section. Followed the repo's actual
  established convention over the template's optional frontmatter field, consistent with this
  compilation's explicit instruction to match "how other canonical skills in this same repo are
  compiled."
- **Failure Handling base-pattern rows:** the definition's own Failure Handling table did not
  include the 4 generic base rows. Cross-checked three other definition/compiled pairs
  (`smaqit.infrastructure-domain-tls`, `smaqit.infrastructure-hook-post-deploy-stamp`,
  `smaqit.infrastructure-provision-cyso`) and confirmed the current convention prepends all 4 base
  rows even when absent from the definition. Applied that convention here.
- No `[?]` annotations found in the definition file — nothing was left for user clarification.

## `[?]` Annotation Report

A full scan of `.smaqit/definitions/skills/smaqit.infrastructure-image-build-push-static.md` found
**zero** `[?]` annotations. All fields in the definition were fully specified (it documents a
skill already proven live against a real cluster across six iterations per its own Provenance
section).
