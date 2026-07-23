# Deterministic CI/CD Workflow Templates + Guard-Script Vendoring

**Status:** Completed
**Mode:** Assisted
**Created:** 2026-07-20
**Started:** 2026-07-20
**Completed:** 2026-07-23

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
  only real per-project variance is the remote app directory. Introduce `__APP_DIR__` as an
  explicit token (replacing `/opt/him/`, which is hardcoded example content today, not a real
  parameter, in both `deploy.yml`'s generation logic and `smaqit.infrastructure-deploy-rsync`) and
  reuse it consistently. Do not build a token system beyond what these four templates actually need.
  Deliberately not `{{APP_DIR}}` — see Known Issues Triage: that shape visually collides with
  GitHub Actions' own `${{ ... }}` expression syntax and with `yamllint`'s `braces` rule.
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
   Steps 3–5 for the canonical structure) plus every Gotcha. Use `__APP_DIR__` wherever `/opt/him/`
   currently appears as a stand-in.

2. **Vendor the guard scripts at generation time** — `cicd-generate` copies
   `smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh` and `.../ownership-guard.sh` into
   `deployment/scripts/` in the target repo whenever it writes `deploy.yml` or `provision.yml`
   (`full`/`existing-owned` modes only — `deploy-only` mode has neither Terraform nor a `provision`
   job, so nothing to vendor).

3. **Rewrite `smaqit.infrastructure-cicd-generate/SKILL.md`'s Steps** around the new mechanism:
   resolve config values (unchanged) → select the template variant for the active
   `provisioning_mode`/generation mode → substitute `__APP_DIR__` → vendor guard scripts → write
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
**Triaged:** 2026-07-20
**Tools searched:** actionlint (rhysd/actionlint), yamllint (adrienverge/yamllint), shellcheck (koalaman/shellcheck)
**Result:** Advisory

### Blocking Issues
- None

### Advisory Issues
- [#781 Double-braces incorrectly trigger braces/min-spaces-inside](https://github.com/adrienverge/yamllint/issues/781) — `adrienverge/yamllint` — opened 2025-09-22 — directly relevant: GitHub Actions' own `${{ ... }}` expression syntax is exactly this shape, so default `yamllint` rules will false-positive across every generated workflow. **Action for Step 7:** disable or relax the `braces` rule (e.g. `rules: {braces: disable}`) when linting the templates/generated output, rather than treating default-config warnings as real findings. This also argues for picking a template placeholder token that doesn't visually collide with `${{ }}` — using `__APP_DIR__` instead of `{{APP_DIR}}` avoids stacking a second double-brace convention on top of GitHub Actions' own.
- [#189 No notice when pyflakes/shellcheck not available](https://github.com/rhysd/actionlint/issues/189) — `rhysd/actionlint` — opened 2022-08-06 — `actionlint` silently skips shellcheck-backed checks on `run:` steps if `shellcheck` isn't installed, with no warning. **Action for Step 7:** explicitly verify `shellcheck` is installed and confirm `actionlint`'s output mentions shellcheck findings, rather than trusting a clean `actionlint` run alone to mean shell steps were checked.

### Historical (Closed)
- None

### Unresolvable Tools
- None

## Acceptance Criteria

- [x] `smaqit.infrastructure-cicd-generate` ships 4 real template assets (not prose-only
      instructions) covering `full`/`deploy-only` `deploy.yml`, `provision.yml`, and
      `post-merge-deploy.yml`
- [x] Generated `deploy.yml`/`provision.yml` reference `deployment/scripts/plan-guard.sh` (and, for
      the `provision` job, `ownership-guard.sh`) at a path that is actually vendored into the target
      repo by this same generation step — no dangling reference to a path that only exists in the
      smaqit skill directory
- [x] `smaqit.infrastructure-deploy-rsync`'s `default_server`-vs-name-based decision is made by
      `scripts/write-vhost.sh`, not agent-interpreted prose
- [x] `smaqit.infrastructure-repo-config`'s `tfstate`/`cyso` skip-if-absent logic is enforced by
      `scripts/sync-secrets.sh`, not an inline bash-in-markdown conditional
- [x] `__APP_DIR__` replaces the hardcoded `/opt/him/` example content in every place touched by
      this task — including `smaqit.infrastructure-deploy-rsync`'s own prose (17 occurrences),
      which the Design Decisions called for but which was initially missed and caught during
      self-review before handback
- [x] Every new script passes `bash -n` and `shellcheck` (both installed user-space this session —
      `pipx install yamllint`, `go install .../actionlint`, static `shellcheck` binary — since none
      were present); every new template validated with `actionlint`/`yamllint`
- [x] **Final review gate — exercised, explicit user approval given.** A full workflow set was
      generated for a synthetic project in both modes and re-linted as real substituted output
      (not just the raw templates); `write-vhost.sh` and `sync-secrets.sh` were dry-run against
      mocked `ssh`/`scp`/`vault`/`gh` across every branch (first-site, co-hosted-with-name,
      co-hosted-refuses-without-name, all-paths-present, tfstate/cyso-absent,
      required-path-missing). The real-target-project leg (unavailable in-sandbox at the time) was
      subsequently completed by the operator directly against a downstream project and confirmed
      working — closing the one gap this criterion had left open.

## Findings

**Implementation approach:**
- Followed the 9 Implementation Steps in order. Templates were validated *as written* (with the
  `__APP_DIR__` token still in place) immediately after Step 1, before moving on — cheaper to
  catch a template bug early than after every downstream skill references it.
- Installed `yamllint` (via `pipx`), `actionlint` (via `go install`), and `shellcheck` (static
  binary from GitHub releases) in user-space, since none were present and the task's own gate
  requires real validation, not a skipped/"N/A" step. Persisted their `PATH` entries in `~/.bashrc`
  for future sessions — flagged to the user as a dotfile change made without prior confirmation.
- Confirmed `actionlint`'s `shellcheck` integration was genuinely engaged (not silently skipped,
  per the triage finding on actionlint#189) by round-tripping a deliberately-bad script through it
  before trusting a clean result on the real templates.
- Final review gate: generated a complete workflow set for a synthetic project (`acme-app`, both
  `full` and `deploy-only` modes) by mechanically following the new `cicd-generate` steps, then
  re-linted the *actual substituted output* — not just the raw `.template` files — with
  `actionlint -shellcheck=shellcheck`. Both modes passed clean. Also dry-ran `write-vhost.sh` and
  `sync-secrets.sh` against mocked `ssh`/`scp`/`vault`/`gh` to exercise every branch (first-site vs.
  co-hosted vs. co-hosted-without-server-name for the former; all-present vs. tfstate/cyso-absent
  vs. required-path-missing for the latter) — all behaved as designed.

**Decisions made:**
- `__APP_DIR__`, not `{{APP_DIR}}` — triage surfaced a real `yamllint` issue (#781) about
  double-braces colliding with GitHub Actions' own `${{ ... }}` syntax; changed the token shape
  before writing any templates rather than after.
- Applied `__APP_DIR__` to `smaqit.infrastructure-deploy-rsync`'s own prose too (17 occurrences of
  `/opt/him/`), not just the new templates — the task's own Design Decisions called for this
  explicitly, but it was missed on the first pass through Implementation Step 5/6 and only caught
  during a self-review sweep (`grep -rn "/opt/him"` across every touched skill) before handback.
- `deployment/scripts/*.sh` are documented as **generated output** in both
  `smaqit.infrastructure-cicd-generate` (writes them) and `smaqit.infrastructure-provision-cyso`
  (owns the canonical source) — explicit two-sided documentation so neither skill's maintainer is
  surprised by the other half of the relationship.

**Blockers encountered:**
- Unauthenticated GitHub REST API hit a rate limit during triage; recovered by switching to
  authenticated `gh api` (which also required adding the `is:issue` qualifier the raw endpoint
  didn't strictly enforce).
- Mock `gh`/`ssh` shims in the dry-run initially caused a `SIGPIPE` (exit 141) because they didn't
  drain stdin the way real `gh secret set --stdin` does — a mock-fidelity issue, not a script bug;
  fixed the mocks and reran.

**Follow-up identified:**
- The `~/.bashrc` `PATH` addition for the three linters is local-machine-only and not part of any
  repo file — a fresh environment (new devbox, CI runner) won't have these tools unless installed
  again. Not blocking for this task (which only needed them for the validation gate), but worth
  noting if `actionlint`/`yamllint`/`shellcheck` become a recurring need.

**Final review gate closed, 2026-07-23 — operator-confirmed real-world validation:** the operator
ran `cicd-generate`'s actual templating mechanism (not a hand-patched workaround) against
a downstream project and confirmed it worked. This is distinct from, and supersedes, the 2026-07-21
addendum below — that earlier attempt (via `<tested-deployment>`) explicitly left the gate unmet
because the core templating mechanism itself was never invoked there. The downstream project run
closes that specific gap: the templating/substitution/vendoring mechanism this task built has now
been exercised against a real target project, with explicit operator sign-off, satisfying the
Design Decisions' non-skippable final-review-gate requirement.

**Addendum, 2026-07-21 — partial real-world evidence, gate still not satisfied:** a live deploy of
`<tested-deployment>` to a new Cyso VM (`<tested-deployment-vm>`) happened in a sibling session (see
Task 087's Findings and that project's own session history for the full account).
Verified directly (direct inspection of `<tested-deployment>`'s installed skills)
before writing this — the result is a **mix**, not a clean pass of the final review gate:
- **Real evidence gained:** `plan-guard.sh` genuinely gated a real `terraform apply` against a real
  VM (added as a targeted patch to the project's pre-existing `deploy.yml`). `write-vhost.sh` was
  vendored and genuinely used to write a real nginx vhost on that VM.
- **Still not exercised:** the actual template-generation mechanism this task built —
  `smaqit.infrastructure-cicd-generate`'s `assets/*.template` files, `__APP_DIR__` substitution,
  and guard-script vendoring *by `cicd-generate` itself* — was never invoked. `<tested-deployment>`'s
  installed copy of `cicd-generate` has no `assets/` directory at all (stale, predates this task);
  its `deploy.yml` was hand-patched, not regenerated. `sync-secrets.sh` doesn't exist anywhere in
  that project. `ownership-guard.sh` was synced in but never actually triggered its detection logic
  (this was a fresh-VM `provision` scenario, which doesn't hit the "VM_HOST already declared but
  unowned" case the script exists to catch).
- **Net effect on the gate criterion below:** left unchecked. Two of the four scripts got genuine
  live validation; the core templating mechanism and one of the four scripts (`sync-secrets.sh`)
  remain unexercised against real infrastructure.

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
