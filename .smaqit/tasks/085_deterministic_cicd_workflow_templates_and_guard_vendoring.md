# Deterministic CI/CD Workflow Templates + Guard-Script Vendoring

**Status:** Not Started
**Mode:** Assisted
**Created:** 2026-07-20

## Description

`smaqit.infrastructure-cicd-generate` has no asset files today — it is pure prose instructing an
agent to synthesize `deploy.yml`, `provision.yml`, and `post-merge-deploy.yml` from a described
job/step structure, from scratch, on every invocation. Its own Gotchas section lists 8 known
failure modes (`GITHUB_TOKEN` collision, missing `plan-guard.sh` gate, `reload` vs
`reload-or-restart`, label vs body-sentinel trigger, `VITE_*` build-time timing, `.terraform.lock.hcl`
gitignore risk, etc.) that exist specifically *because* generation is re-derived each run instead of
copied from something already known-correct. This mirrors a pattern already solved elsewhere in
smaqit — several skills (`smaqit.project-compendium`, `smaqit.project-glossary`, others) ship real
`assets/*_TEMPLATE.md` files that get copied and filled in, rather than describing structure in
prose for an agent to reconstruct.

While assessing this, a second, previously-latent gap surfaced: `smaqit.infrastructure-provision-cyso`'s
`plan-guard.sh` and `ownership-guard.sh` (the latter added in Task 084) live only inside the smaqit
skill directory (`SMAQIT_SKILLS_DIR`) — a local/session concept. Nothing in the repository vendors
either script into a generated project's own checkout, yet `smaqit.infrastructure-cicd-generate`'s
current prose tells the agent to generate a `provision` job that "runs
`smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`" as if that path exists on a
GitHub-hosted runner. It doesn't, and nothing catches this until someone actually runs the generated
workflow in CI.

Separately, two other spots from Task 084 were implemented as agent-interpreted prose with inline
bash rather than as real, deterministic scripts: `smaqit.infrastructure-deploy-rsync`'s
`default_server`-vs-name-based nginx vhost decision (the highest-stakes judgment call in that skill
— getting it wrong breaks `nginx -t` on exactly the co-hosted VM scenario Task 084 exists to
support), and `smaqit.infrastructure-repo-config`'s mode-aware skip-if-absent logic for
`tfstate`/`cyso` Vault paths (currently a copy-pasted `if`/`else` bash block inside the markdown,
which a future edit could silently "simplify" back to an unconditional read).

This task ships all four fixes together, plus a hard final-review gate before any of it is treated
as the default path — CI/CD generation and infrastructure provisioning are high blast-radius; a bad
template or a mis-vendored script could break production deploys across every project that uses
these skills, not just the one being worked on when the bug is introduced.

## Design Decisions

- **Ship real template assets, not prose descriptions.** New `assets/` directory under
  `smaqit.infrastructure-cicd-generate` holding four template files: `deploy.yml.full.template`,
  `deploy.yml.deploy-only.template`, `provision.yml.template`, `post-merge-deploy.yml.template`.
  Every documented Gotcha (the plan-guard gate, `TF_VAR_github_token` naming, `VITE_*` build-time
  env placement, `docker compose` v2 syntax, `reload-or-restart`, PR body sentinel not label, path
  filters) gets encoded directly in the template YAML — once, correctly — instead of being
  something the agent must remember to reproduce on every generation.
- **Minimal token substitution**, not a general templating engine. The target stack is already
  narrow and fixed (Node.js + React + Docker Compose + nginx, fixed secret/variable names), so the
  only real per-project variance is the remote app directory. Introduce `{{APP_DIR}}` as an
  explicit token (replacing `/opt/him/`, which is hardcoded example content today, not a real
  parameter, in both `deploy.yml`'s generation logic and `smaqit.infrastructure-deploy-rsync`) and
  reuse it consistently. Do not build a token system beyond what these four templates actually need.
- **Vendor `plan-guard.sh` and `ownership-guard.sh` into the target repo at generation time**, into
  a new `deployment/scripts/` directory (sibling to `deployment/terraform/`). `cicd-generate` copies
  both scripts from the skill directory into the target repo when it writes `deploy.yml`/
  `provision.yml`, so the generated workflow's `run:` steps reference a path that actually exists in
  the target repo's own git history — not a path that only exists on the machine/session that ran
  the skill.
- **Add `ownership-guard.sh` to the generated `provision` job too**, as defense-in-depth alongside
  `plan-guard.sh` — consistent with Task 084's own "defense in depth, not instead of the gate"
  principle. It was previously only documented as Step 0 of a *direct/manual* `provision-cyso`
  invocation; CI running an ungated `provision` job is exactly the kind of bypass that guard exists
  to catch.
- **The skill directory's copies of `plan-guard.sh`/`ownership-guard.sh` remain the canonical
  source.** The vendored copy in a target repo's `deployment/scripts/` is regenerated (overwritten)
  the next time `cicd-generate` runs — document this explicitly in both skills so nobody is
  surprised two copies exist, or edits the vendored one expecting it to persist.
- **Deterministic scripts for the other two Task 084 judgment calls:**
  - `smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` — inspects
    `/etc/nginx/sites-enabled/` on the target VM itself and writes the correctly-flagged conf
    (`default_server` only if nothing else is enabled; name-based otherwise), removing the decision
    from agent-interpreted prose entirely.
  - `smaqit.infrastructure-repo-config/scripts/sync-secrets.sh` — performs the full Vault → GitHub
    secrets/variables sync in one script, with the `tfstate`/`cyso` skip-if-absent logic as
    structural `if` branches in the script rather than markdown the agent has to preserve correctly
    on every read.
- **Hard final-review gate, regardless of mode.** Before this task can move to Completed: every new
  YAML template is validated (`actionlint` + `yamllint`, or a documented manual structural review if
  neither tool is available in the environment), every new/changed shell script passes `bash -n`
  (and `shellcheck` if available), and — the actual gate — the new templates are used to generate a
  full workflow set against a real or realistic target project and the output is reviewed and
  explicitly approved by the user before this becomes the default generation path. This step cannot
  be skipped even in Autonomous mode.

## Implementation Steps

1. **Create the four workflow templates** under `smaqit.infrastructure-cicd-generate/assets/`,
   encoding the full job/step structure currently described in prose (see the skill's existing
   Steps 3–5 for the canonical structure) plus every Gotcha. Use `{{APP_DIR}}` wherever `/opt/him/`
   currently appears as a stand-in.

2. **Vendor the guard scripts at generation time** — `cicd-generate` copies
   `smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` and `.../ownership-guard.sh` into
   `deployment/scripts/` in the target repo whenever it writes `deploy.yml` or `provision.yml`
   (`full`/`existing-owned` modes only — `deploy-only` mode has neither Terraform nor a `provision`
   job, so nothing to vendor).

3. **Rewrite `smaqit.infrastructure-cicd-generate/SKILL.md`'s Steps** around the new mechanism:
   resolve config values (unchanged) → select the template variant for the active
   `provisioning_mode`/generation mode → substitute `{{APP_DIR}}` → vendor guard scripts → write
   files → report. Condense the existing Gotchas section into maintainer-facing documentation of
   *why* the templates are built the way they are (kept for humans editing the templates later, no
   longer instructions the agent must reconstruct from memory).

4. **Update `smaqit.infrastructure-provision-cyso/SKILL.md`** to note that the vendored copies of
   `plan-guard.sh`/`ownership-guard.sh` under a target repo's `deployment/scripts/` are generated
   output, not hand-maintained — the skill directory's copies are canonical, and re-running
   `cicd-generate` overwrites the vendored ones.

5. **Add `scripts/write-vhost.sh` to `smaqit.infrastructure-deploy-rsync`**; update Step 6 to invoke
   it instead of describing the `sites-enabled/` inspection and decision in prose.

6. **Add `scripts/sync-secrets.sh` to `smaqit.infrastructure-repo-config`**; update Steps 3–5 to
   invoke it instead of the inline bash-in-markdown conditional.

7. **Validate everything added:** `bash -n` on every new/changed script (`shellcheck` too, if
   available); `actionlint`/`yamllint` on every new template (or a documented manual review if
   those tools aren't available in this environment) — do not skip this step or declare it
   "N/A" without checking tool availability first.

8. **Final review gate** — generate a complete workflow set (all four templates, both generation
   modes) against a real or realistic target project, present the generated output to the user, and
   get explicit sign-off before treating this as the default path for
   `smaqit.infrastructure-cicd-generate`. This step is mandatory in both Assisted and Autonomous
   mode — do not auto-complete this task without it.

9. Bump `metadata.version` on every skill touched.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] `smaqit.infrastructure-cicd-generate` ships 4 real template assets (not prose-only
      instructions) covering `full`/`deploy-only` `deploy.yml`, `provision.yml`, and
      `post-merge-deploy.yml`
- [ ] Generated `deploy.yml`/`provision.yml` reference `deployment/scripts/plan-guard.sh` (and, for
      the `provision` job, `ownership-guard.sh`) at a path that is actually vendored into the target
      repo by this same generation step — no dangling reference to a path that only exists in the
      smaqit skill directory
- [ ] `smaqit.infrastructure-deploy-rsync`'s `default_server`-vs-name-based decision is made by
      `scripts/write-vhost.sh`, not agent-interpreted prose
- [ ] `smaqit.infrastructure-repo-config`'s `tfstate`/`cyso` skip-if-absent logic is enforced by
      `scripts/sync-secrets.sh`, not an inline bash-in-markdown conditional
- [ ] `{{APP_DIR}}` replaces the hardcoded `/opt/him/` example content in every place touched by
      this task
- [ ] Every new script passes `bash -n` (and `shellcheck`, if available); every new template is
      validated with `actionlint`/`yamllint` or an explicitly documented manual review
- [ ] **Final review gate exercised and explicitly approved by the user** — a full generated
      workflow set was produced against a real or realistic target project and reviewed before this
      task is marked Completed. Not satisfied by unit-level script/template checks alone.

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.full.template` | New |
| `skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.deploy-only.template` | New |
| `skills/smaqit.infrastructure-cicd-generate/assets/provision.yml.template` | New |
| `skills/smaqit.infrastructure-cicd-generate/assets/post-merge-deploy.yml.template` | New |
| `skills/smaqit.infrastructure-cicd-generate/SKILL.md` | Modify — steps rewritten around template selection + substitution + guard vendoring |
| `skills/smaqit.infrastructure-provision-cyso/SKILL.md` | Modify — document vendored-vs-canonical guard script relationship |
| `skills/smaqit.infrastructure-deploy-rsync/scripts/write-vhost.sh` | New |
| `skills/smaqit.infrastructure-deploy-rsync/SKILL.md` | Modify — Step 6 invokes `write-vhost.sh` |
| `skills/smaqit.infrastructure-repo-config/scripts/sync-secrets.sh` | New |
| `skills/smaqit.infrastructure-repo-config/SKILL.md` | Modify — Steps 3–5 invoke `sync-secrets.sh` |

## Notes

- Originates from a post-Task-084 assessment conversation: first, whether `smaqit.infrastructure-deploy-rsync`
  and `smaqit.infrastructure-repo-config`'s agent-interpreted judgment calls should have been scripts
  (yes, for the two identified here); then, whether workflow generation itself should be
  deterministic (yes) — which surfaced the guard-script vendoring gap as a genuine, previously
  unnoticed bug in Task 084's own design, not just a style preference.
- `triage: skip` is deliberately **not** set — `actionlint`/`yamllint`/`shellcheck` are real
  third-party tools worth a triage pass at `task-start` time.
