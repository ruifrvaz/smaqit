# Slot-Driven Deployment Orchestration

## Metadata

- **Date:** 2026-10-11
- **Session focus:** Plan, implement, and release deployment slots (variation points), making the Deployment agent the single deployment orchestrator for app and platform repositories
- **Tasks completed:** 125 (parent, released v3.7.0, PR #94), 127, 128, 129 (children)
- **Tasks created:** 127, 128, 129 (children of 125)
- **Tasks referenced:** 074, 114, 115, 122, 123 (re-scoped onto slots), 113, 126 (next candidates)

## Actions Taken

- Ran `session.start`: loaded project docs, memory, history 087, `PLANNING.md`, and the compendium, and read the code task 125 touches. Found local `main` 4 commits ahead of `origin/main`; pushed them on request (two further commits from a parallel session — task 126 creation and a compendium update — went up with them).
- Planned task 125 via `smaqit.task-plan`. Three Explore agents mapped the current deployment wiring, agent skill-invocation and installer mechanics, and the bench/principle/wiki layout. Discovery showed the task's linear 13-slot catalogue contradicted current behaviour (k3s credentials are two-stage, IaC and workflows are generated before provisioning, TLS runs after verify, image building had no slot). The user chose: a revised fixed order with `provisioning_mode` as the default-profile selector, resolution in the Go CLI, agents reading filler `SKILL.md` from a resolved path, and a parent plus three children on one branch with one release.
- Updated the task 125 file with the revised plan (five confirmed fields), started it (branch/worktree, research-map block, issue triage), and created children 127, 128, 129.
- **Child 127:** L0 principle edits in `framework/AGENTS.md`, `SKILLS.md`, `SMAQIT.md`; `installer/slots.go` (catalogue table, strict `slots.yml` parser, project-local-then-global skill lookup, `smaqit slots plan|validate`, aggregated `SLOTS-*` diagnostics); create-if-absent `.smaqit/slots.yml` skeleton preserved by update and uninstall; `metadata.smaqit-slots` on 13 default filler skills; three invalid-YAML frontmatters fixed; vault-loader `CREDENTIAL_SCOPE=access`; wiki page; 13 Go tests and smoke-test assertions.
- **Child 128:** the Deployment agent executes the slot plan (local fillers read from `skillPath`, family candidates by stack, builtins, CI-executed slots monitored); `smaqit.input-deployment` resolves the plan once; feature-new Phase 3 routes credentials and repo config through the agent; Copilot deployment agent gained `read/readFile`.
- **Child 129:** greenfield Phases 4–6 route through the agent (Phase 4 `standard`, Phase 5 new `ci-push` path, Phase 6 single-slot `tls`); added the `ci-push` path to the agent.
- Parent close-out: built `.smaqit/bench/agents/smaqit.deployment/` (stub CLI serving a fixed plan, stub fillers writing markers); refreshed re-scope notes on 074, 114, 115, 122, 123; ran a `smaqit.qa` consistency review and fixed all 3 contradictions and 7 inaccuracies; ran the bench live (both Cases passed).
- Completed task 125 Phase 1 (findings, release analysis → v3.7.0, PR #94, changelog on branch). After the user merged it, ran Phase 2: confirmed the merge, merged `origin/main` into local `main`, marked the task Completed, removed the worktree and local branch. The release workflow tagged `v3.7.0` and published the GitHub Release.

## Problems Solved

- **Agent Skills `metadata` must be strings:** triage surfaced agentskills#514; a nested `smaqit-slots` map would violate the spec. Encoded as one string of `slot=operation` pairs.
- **k3s credential ordering:** the kubeconfig only exists after onboarding, but vault-loader loaded it in the same run as other credentials. Split into `access-credentials` / `target-credentials` with an opt-in `CREDENTIAL_SCOPE=access`.
- **CI starting before secrets are synced:** the qa review found `environment-config` positioned after the first CI-executed slot on CI paths. Moved it ahead of `artifact-build`, and added a resolver refusal for any local binding that would split a CI run.
- **`ci-push` host bootstrap:** current greenfield Phase 5 never runs `vm-bootstrap` (cloud-init configures the production host), so `ci-push` defaults `host-bootstrap` to none instead of treating it as CI-executed.
- **Sparse worktree hides nested fixture paths:** the task worktree's unanchored `!.agents/skills/` pattern blocked committing bench stub skills; stored them as `stub-skills/` and moved them into place in the bench Case's `prepare`.
- **Issue triage rate limit:** the parent task's GitHub searches all returned 403 (unauthenticated rate limit); recorded as warnings, not Clear.
- **Stray build binary:** an early `go build ./...` left `installer/smaqit` in the worktree; removed before committing.

## Decisions Made

- Deployment agent is the orchestrator (no separate machine-deployment agent), resolving task 114 Finding 5.
- Slot resolution is deterministic Go CLI work; agents never re-derive it.
- Principle edits made directly, not via `smaqit.new-principle` (smaqit-adk extension migration deliberately not done).
- Path rules: `standard` defaults `environment-config` and `tls` to none; `existing-cicd-pr` refuses `iac-generate`, `ci-generate`, `provision`, `host-bootstrap`; `ci-push` defaults `iac-generate`, `ci-generate`, `host-bootstrap`, `tls` to none. Single-slot runs ignore full-sequence defaults but never refusals.
- Default `teardown` only for `provision`; `existing-owned` no longer destroys a retained VM.
- When a filler needs secret input, the agent hands the operator the exact command; secrets never pass through chat.
- The bench tests the agent's behaviour with a stub CLI; resolver determinism stays in Go tests.
- Released as MINOR v3.7.0.

## Files Modified

- `framework/AGENTS.md`, `framework/SKILLS.md`, `framework/SMAQIT.md` — slot principle, filler skills, variation points
- `installer/slots.go`, `installer/slots_test.go` (new) — catalogue, resolver, CLI, 15 tests
- `installer/main.go`, `installer/Makefile`, `scripts/smoke-test-installer.sh` — embed, init skeleton, uninstall preservation, validate hook, help, smoke assertions
- `templates/slots.yml.template` (new), `templates/AGENTS.md.template` — skeleton and pointer
- `agents/deployment.md`, `.smaqit/definitions/agents/deployment.frontmatter.yaml` — slot-plan execution, `ci-push`, Copilot read tool
- `skills/smaqit.input-deployment`, `skills/smaqit.feature-new`, `skills/smaqit.new-greenfield-project` — plan resolution and routing
- 13 `skills/smaqit.infrastructure-*/SKILL.md` — `metadata.smaqit-slots`, version bumps, 3 YAML fixes; `vault-loader/scripts/load-credentials.sh` — `CREDENTIAL_SCOPE`
- `docs/wiki/workflows/deployment-slots.md` (new), `docs/wiki/workflows/extending-smaqit.md`, `README.md`
- `.smaqit/bench/agents/smaqit.deployment/` (new), `.smaqit/bench/README.md`
- `.smaqit/references/project-research.md` — task 125, 127, 128, 129 blocks
- `CHANGELOG.md` — v3.7.0
- `.smaqit/tasks/125, 127, 128, 129`, re-scope notes on `074, 114, 115, 122, 123`, `PLANNING.md` — task lifecycle bookkeeping
- `.smaqit/compendium.md` — slot entries
- `smaqit.code-workspace` — worktree lifecycle

## Next Steps

- Downstream platform repository can bind its own fillers on v3.7.0.
- Tasks 114, 115, 122, 123 build on slots (refreshed notes in each); 122's private-mesh support is a `platform-prereqs` filler plus binding params.
- Self-contained next picks: task 126 (design-tools install-path lock), task 113.
- Unfiled follow-ups: triage helper token support and unanchored worktree sparse patterns (both smaqit-extensions); whether production VMs should run `vm-bootstrap`; an old `<!-- amendment: -->` comment in greenfield Pre-conditions may match the Phase 8 scan; installer skill-retirement pruning gap (carried over).

## Session Metrics

- Tasks completed: 4 (125 + children 127, 128, 129); tasks created: 3
- Release shipped: v3.7.0 (PR #94)
- Implementation: 5 commits on the task branch; about 40 files changed (4 new source/doc files, a new bench suite)
- Tests: 15 new Go tests; smoke test extended; bench suite 2/2 Cases passed live
- Subagents: 3 Explore (planning discovery), 1 `smaqit.qa` review (3 contradictions, 7 inaccuracies, all fixed)
