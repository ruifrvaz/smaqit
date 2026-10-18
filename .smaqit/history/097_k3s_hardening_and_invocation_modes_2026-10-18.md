# k3s Hardening and Invocation Modes

## Metadata

- **Date:** 2026-10-18 to 2026-10-19
- **Session focus:**
  - harden the generated existing-k3s CI path after the first live private-mesh deploy;
  - ship its follow-up documentation;
  - upgrade a downstream platform repository;
  - make the orchestrator skills default by how they were invoked, under a new name;
  - make global install remove retired skills.
- **Tasks completed:**
  - 136 (v3.16.0, PR #105);
  - 141 (v3.18.0, PR #109).
- **Tasks created:**
  - 137 (scripts-only Vault credential slots, split from 136);
  - 139 (abandoned);
  - 141;
  - smaqit-extensions 044.
- **Releases:**
  - v3.16.0 (#105);
  - v3.17.2 (#108, task 136's late failure-handling rows);
  - v3.18.0 (#109).

  v3.15.0 (task 135) and v3.17.0/v3.17.1 (tasks 138 and 140) shipped from parallel sessions in between.

## Actions Taken

**Task 136: existing-k3s fixes from the first live private-mesh deploy**
- **Session start:** loaded the context and found task 135 half-started. Another session took it over (PR #104, v3.15.0).
- **`task.plan 136`:** three discovery agents mapped the CI templates, the credential fillers, and the probe client and spec. You settled four decisions:
  - handoff-and-verify for a scripts-only Vault policy;
  - verifying from the CI run;
  - a spec-controlled `Deployer Exec` opt-out;
  - splitting gap 6 into task 137.

  The task file was rewritten field by field.
- **Built:**
  - **Dispatcher:** `post-merge-deploy.yml` gets `--repo`, and the base ref arrives through `env:` (all modes).
  - **Secrets:** every secret reaches its step through `env:`. The kubeconfig and the VM SSH key are written under `umask 077` and removed by an `always()` step.
  - **Pull Secret:** a dockerconfigjson pull Secret, passed through a file descriptor, rendered from a parsed `Container Registry` row.
  - **Exec checks:** `namespace-guard.sh --deployer-exec` (using `can-i --subresource=exec`) and a new `exec-refused.sh` (a real exec must be refused as Forbidden).
  - **Probe client:** `python3`.
  - **Concurrency:** one deploy per machine on the `deploy` job.
- **Tested:** step tests run under `bash -eo pipefail` with stub `kubectl` and `gh`. A contract test covers every workflow template, and a test runs the python-only probe.
- **Verified:** seven deliberate breaks, each caught by a test; `actionlint` was clean.
- **Released:** Phase 1 opened PR #105, and the CHANGELOG conflict with #104 was resolved. Both PRs had also bumped `cicd-generate` to 3.1.0 without a textual conflict, so this release used 3.2.0. Phase 2 cleaned up after the merge.

**Trap documentation (task 139, then v3.17.2)**
- You asked whether the `can-i` TYPE/NAME trap and the `bash -e` trap still exist. They do, as kubectl and GitHub Actions behaviour, but smaqit's own checks are immune.
- Task 139 (a static guardrail) was created, then abandoned on your call as too brittle. Failure-handling rows went into both skills instead.
- The rows were pushed to the task 136 branch after you had merged #105, so they missed the release. They shipped separately as v3.17.2 (PR #108), after resolving a conflict with #107.

**Downstream platform repository upgrade (v3.17.2)**
- **Ran `smaqit update`,** following that repository's own upgrade checklist:
  - deleted the re-created `copilot-setup-steps.yml`, as that repository had decided earlier;
  - confirmed its `CLAUDE.md` holds only the managed block.
- **Re-vendored the plan guard.** Its vendored `plan-guard.sh` was still the v3.13.0 copy and was brought to v3.17.2.
- **Corrected its task 052's plan forward.** Then opened a chore PR (#169) with the organisation token, loaded through `vault-gh-token.sh`.

**Task 141: orchestrator mode by invocation, rename, retired-skill pruning**
- **Spotted in review:** the downstream PR showed that `smaqit.feature-new` and the greenfield skill were always autonomous, even when typed as slash commands.
- **`task.plan`:** two discovery agents. You chose:
  - an install manifest plus a seed list for old installs;
  - renaming every mention that means the skill;
  - explicit invocation syntax only.
- **Built:**
  - Phase 0 mode resolution in both skills, with one rule stated across the input skills, agents, `AGENTS.md` template, framework, README, wiki and compendium;
  - the rename with `git mv`;
  - `installer/global_manifest.go`: a per-destination `.smaqit-installed.json` manifest, pruning, a name-matched seed derived from git history and release tags, and manifest-aware uninstall;
  - Go tests and a smoke-test upgrade case.
- **Researched before approval:** the Claude Code docs, checked directly, confirm `UserPromptExpansion` fires for a typed slash command and `PreToolUse` on the `Skill` tool fires for a model invocation. Both would need hooks in user settings or a plugin, and you declined that follow-up.
- **Renamed again in review:** you renamed the skill again, to `smaqit.project-new`, to match `smaqit.feature-new`. It was renamed again on the same PR, and the guard test now also rejects the interim name.
- **Released as v3.18.0 (PR #109).** `smaqit update` on this machine removed `smaqit.new-greenfield-project` from both skill directories and wrote manifests to every destination. All 24 smaqit-extensions skills were kept.

**Other**
- **Your rule saved to memory:** no backwards compatibility, always fix forward. It is the rule for decisions you hand back.
- **smaqit-extensions task 044 created:** `smaqit.project-init`'s description overlaps with `smaqit.project-new`.
- **Session recaps:** produced twice.

## Problems Solved

- **The post-merge dispatcher never dispatched** (it ran with no checkout). Fixed with `--repo`, and proven by a test that reproduces the live failure.
- **Secrets were interpolated into scripts or passed on argv.** Now they are passed only through `env:`, and a contract test across every template enforces it.
- **The `can-i pods/exec` TYPE/NAME trap.** smaqit now runs real exec checks, and documents the trap for projects' own steps.
- **A skill version collided without a textual conflict** (both PRs bumped `cicd-generate` to 3.1.0, and 3.1.0 had already shipped). Caught during review and bumped forward twice.
- **Commits pushed after a merge.** The failure rows missed #105, were recovered as their own release, and the abandon note was corrected.
- **Retired skills stayed installed forever** (the long-standing pruning gap). Closed by the install manifest and the seed list.
- **Concurrent sessions claimed task numbers.** Task 138 was claimed mid-creation, so the new task was renumbered 139, and every release version was checked against open PR claims.
- **Shell and tool issues:**
  - `gh pr view --json body` fails on GitHub's retired Projects (classic) API, so the REST API was used instead;
  - a stray `go build` binary in `installer/` was removed twice;
  - the first push to `main` was blocked by the permission check and succeeded later.

## Decisions Made

- **Release policy:** breaking changes ship as MINOR with an action-required note, and no backwards compatibility (no alias for the renamed skill).
- **Exec checks** fail by default; `Deployer Exec: allowed` turns them into a warning, and an unclear answer fails closed under both settings.
- **Orchestrator mode** comes from explicit invocation syntax only. Naming a skill in plain words counts as model-started. Phase agents keep "picked or named", because Codex starts agents by name.
- **Mode detection stays inference-based.** You declined the hooks follow-up, so don't raise it again.
- **The pruning manifest lives in each destination**, not in a central config file. The seed list holds only `smaqit.`-prefixed names that a v3 release installed globally.
- **Downstream:** re-vendor guards in the upgrade PR itself, and correct dependent task plans forward.
- **The trap guardrail** became documentation (failure rows), not a static checker.

## Files Modified

**smaqit (shipped through PRs #105, #108 and #109):**
- `installer/infra_spec.go`, `installer/global_manifest.go` (new), `installer/main.go`, `installer/slots.go`.
- Tests: `infra_spec_test.go`, `k3s_render_test.go`, `k3s_scripts_test.go`, `skill_scripts_test.go`, `cicd_templates_test.go`, `design_test.go`, `global_manifest_test.go` (new), `slots_test.go`.
- `installer/testdata/k3s-deploy/` (regenerated, plus a `private-registry` case).
- `skills/smaqit.infrastructure-cicd-generate/`: SKILL.md, the five workflow templates, `render-k3s-deploy.sh`.
- `skills/smaqit.infrastructure-deploy-k3s-app/`: SKILL.md, `probe-job.yaml.template`, `namespace-guard.sh`, `exec-refused.sh` (new).
- `skills/smaqit.project-new/` (renamed from `smaqit.new-greenfield-project`), `skills/smaqit.feature-new`, and the three phase input skills.
- The deploy-rsync, deploy-rsync-python-tornado, provision-cyso (including `ownership-guard.sh`), deploy-k3s-app and request-k3s-onboarding skills.
- `agents/{development,deployment,validation,infrastructure}.md`, `framework/AGENTS.md`.
- `templates/AGENTS.md.template`, `templates/specs/infrastructure.template.md`.
- `README.md`, `docs/wiki/{agent-tools-reference.md,designs/fail-fast-on-ambiguity.md,workflows/deployment-slots.md}`, `.smaqit/compendium.md`.
- `.smaqit/definitions/skills/*`, `scripts/smoke-test-installer.sh`, `CHANGELOG.md`.

**smaqit task records:** tasks 136, 137, 139 and 141, `PLANNING.md`, `.smaqit/references/project-research.md` (task 136 and 141 blocks), `smaqit.code-workspace`.

**smaqit-extensions:** task 044 and `PLANNING.md`.

**Downstream platform repository** (PR #169): five spec templates, `AGENTS.md`'s managed section, `.github/scripts/plan-guard.sh`, its task 052 and `PLANNING.md`.

## Next Steps

- **Downstream platform repository:** merge PR #169, then run its upgrade again to v3.18.0 (the corrected mode rule, `project-new`). Its task 052 (the `allow_delete` workflow input) is unblocked.
- **Task 137:** credential slots under a scripts-only Vault policy. It's high priority and fully designed, and its dependency (136) has merged.
- **smaqit-extensions task 044:** the `project-init` description fix.
- **Carried over:** task 126 (the design tools' runtime lock on the install path), task 113; one-off Job template and VM double-deploy follow-ups from 136.

## Session Metrics

- **Duration:** two days (2026-10-18 to 2026-10-19).
- **Tasks:** 2 completed (136, 141), 1 abandoned (139), 3 created in smaqit (137, 139, 141), 1 created in smaqit-extensions (044).
- **Releases:** 3 from this session (v3.16.0, v3.17.2, v3.18.0), plus 1 downstream chore PR (#169).
- **Discovery agents:**
  - 5 Explore agents, across both plans and the mode assessment;
  - 1 docs-research agent.
- **Tests:**
  - **Deliberate breaks:** 7 (task 136), 3 (task 141) and 1 (the interim-name guard), each caught by a test.
  - **New test functions:** about 15 (k3s steps, dispatcher, template contract, python3 probe, manifest pruning and seed, guard).
- **Merge conflicts resolved on release PRs:** 2 (#105 and #108). There was also 1 skill-version collision that merged without a textual conflict.
