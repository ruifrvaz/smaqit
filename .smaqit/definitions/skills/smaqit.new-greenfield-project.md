# Skill Definition: smaqit.new-greenfield-project

## Identity
- **Name:** smaqit.new-greenfield-project
- **Version:** 1.0.0
- **Description:** Orchestrate the complete SDLC from raw project input to a running production application accessible via browser. Strict-sequential — executes all phases in order with explicit human gates at phase boundaries. Invokes smaqit layer agents (business, functional, stack, infrastructure, coverage, development, deployment, validation) plus operational skills (cicd-generate, github-repo-config, provision-target, vm-bootstrap, app-deploy-rsync, deploy-verify). Re-entrant: use the pre-condition checklist to resume at any phase.

## Pre-conditions (operator must complete before starting)
All items must be checked before invoking. Re-entering at any phase requires only the items for that phase and all earlier phases.

### Always required
- [ ] Raw project assets in `assets/raw/` (code, docs, requirements)
- [ ] `gh` CLI authenticated (`gh auth login`)
- [ ] Git repository created on GitHub (public or private)

### Required before Phase 4 (Provision)
- [ ] Cloud account available (Cyso or equivalent)
- [ ] Application credential created in cloud portal
- [ ] Object Storage state bucket created
- [ ] SSH keypair generated (passphrase-free deploy key)
- [ ] SSH public key uploaded to cloud portal
- [ ] Fine-grained PAT with `variables:write` created on GitHub

### Required before Phase 7 (Domain/TLS)
- [ ] Domain purchased at registrar
- [ ] DNS A record set to VM fixed IP

## Phases

### Phase 1 — Requirements Extraction
1. Invoke `smaqit.requirements-extract`.
2. Review the flagged ambiguities with the user. Resolve any that would block specification (e.g., conflicting data model shapes, undefined score ranges).
3. **Gate:** Confirm extracted inventory is sufficient to proceed to specs.

### Phase 2 — Specification
Run each specification agent sequentially. Each agent reads the previous layer's output.
1. Run `smaqit.input-business` → invoke `@smaqit.business` agent.
2. Run `smaqit.input-functional` → invoke `@smaqit.functional` agent.
3. Run `smaqit.input-stack` → invoke `@smaqit.stack` agent.
4. Run `smaqit.input-infrastructure` → invoke `@smaqit.infrastructure` agent.
5. Run `smaqit.input-coverage` → invoke `@smaqit.coverage` agent.
6. **Gate:** All specs have `status: draft` and acceptance criteria written. User reviews and approves spec set.

### Phase 3 — Task Creation + CI/CD Generation
1. Invoke `smaqit.task-create` to create the MVP task (mode: autonomous; acceptance criteria from business spec).
2. Invoke `smaqit.infrastructure-cicd-generate` to produce all 4 GitHub Actions workflows.
3. Invoke `smaqit.infrastructure-repo-config` to set all secrets and variables.
4. Commit the generated workflows: `git add .github/workflows && git commit -m "ci: add CI/CD workflows"`.
5. **Gate:** Workflows committed. Secrets and variables verified with `gh secret list` and `gh variable list`.

### Phase 4 — Infrastructure Provisioning
1. Invoke `smaqit.infrastructure-provision-cyso` (or the appropriate `provision-target-<provider>` skill).
2. Note the `fixed_ip` output from Terraform.
3. Invoke `smaqit.infrastructure-vm-bootstrap` with the fixed IP.
4. **Gate:** SSH to VM succeeds; `/data` mounted; `.env` present; Docker group active.

### Phase 5 — Development
1. Invoke `smaqit.input-development`.
2. Invoke `@smaqit.development` agent to implement all specs with `status: draft`.
3. After implementation, invoke `smaqit.spec-status-update` to set specs to `status: implemented`.
4. **Gate:** `npm run build` passes (backend and frontend). All MVP acceptance criteria met (verified against task file).

### Phase 6 — First Deployment
1. Invoke `smaqit.infrastructure-deploy-rsync`.
2. Invoke `smaqit.infrastructure-deploy-verify`. If verification fails, stop and report — do not proceed to validation.
3. Invoke `smaqit.spec-status-update` to set infrastructure specs to `status: deployed`.
4. **Gate:** `deploy-verify` reports all checks PASS. Health endpoint returns correct SHA.

### Phase 7 — Domain + TLS (conditional)
Only execute if a domain has been purchased and DNS configured (pre-condition checklist above).
1. Invoke `smaqit.infrastructure-domain-tls`.
2. **Gate:** HTTPS accessible, HTTP redirects, auto-renewal dry-run passes.
If skipped: application is accessible at `http://<fixed_ip>`. Note as open item.

### Phase 8 — Validation
1. Invoke `smaqit.input-validation`.
2. Invoke `@smaqit.validation` agent.
3. **Gate:** All validation checks pass. User signs off.

### Phase 9 — Task Completion + Release
1. Invoke `smaqit.task-complete` for the MVP task.
2. Invoke `smaqit.release-analysis` to assess changes and suggest version.
3. Invoke `smaqit.release-approval` for version confirmation.
4. Invoke `smaqit.release-prepare-files` to update CHANGELOG and version files.
5. Invoke `smaqit.release-git-local` (or `smaqit.release-git-pr` for PR-based release).
6. **Final output:** Application running at `https://<domain>/` (or `http://<fixed_ip>/` if Phase 7 was skipped), with tagged release on GitHub.

## Output
- Running production application accessible via browser
- Tagged git release with updated CHANGELOG
- All specs at `status: deployed`
- MVP task closed in PLANNING.md

## Scope
- Covers the single-app, single-environment (production) path only. Staging environments are out of scope.
- Does NOT handle database schema migrations. The current project uses SQLite with append-only schema changes.
- Does NOT cover post-MVP feature cycles. Use individual smaqit agents for iterative feature work after this skill completes.
- Phase 7 (domain/TLS) is conditional on domain purchase — a human action outside the system.

## Gotchas
- **`GITHUB_TOKEN` reserved name** — enforced in Phase 3 via `smaqit.infrastructure-cicd-generate` and `smaqit.infrastructure-repo-config`. Never set an env var named `GITHUB_TOKEN` in any workflow to a PAT.
- **PR body sentinel** — Coding Agent must include `smaqit:deploy` as a line in any PR body to trigger post-merge deployment. This must be set at PR creation time, not via label.
- **Build-time Vite vars** — `VITE_DEMO_MODE` and any `VITE_*` vars must be passed to the frontend build step, not at runtime.
- **rsync trailing slash** — `rsync backend/dist/` (trailing slash) copies contents, not the directory. Always `mkdir -p /opt/him/backend/dist` before rsyncing. Covered by `smaqit.infrastructure-deploy-rsync`.
- **Floating IP (Cyso)** — use `fixed_ip` from Terraform outputs, not the floating IP. The floating IP does not route on Cyso's flat network.
- **Re-entry** — if the skill is interrupted at any phase, the pre-condition checklist indicates which phases are complete. Resume from the first incomplete phase. All phases are idempotent: re-running provisioning on an existing VM will show a plan with no changes if state is intact.

## Completion
- [ ] Phase 1: requirements extracted, ambiguities resolved
- [ ] Phase 2: all specs drafted and approved
- [ ] Phase 3: task created, workflows generated, secrets configured
- [ ] Phase 4: VM provisioned and bootstrapped
- [ ] Phase 5: implementation complete, specs set to `implemented`
- [ ] Phase 6: deployment verified (all deploy-verify checks pass)
- [ ] Phase 7: TLS live (or documented as open item)
- [ ] Phase 8: validation complete
- [ ] Phase 9: task closed, release tagged, application accessible via browser

## Failure Handling
| Situation | Action |
|-----------|--------|
| Phase fails with a hard blocker | Stop at the gate. Report the blocker. Do not advance to the next phase. |
| Specification agent returns incomplete output | Re-run the agent with additional context or user clarification. Do not advance with incomplete specs. |
| `deploy-verify` fails | Stop. Report the failing check. Do not mark deployment as complete. |
| Phase 7 skipped (no domain) | Document as open item. Continue to Phase 8. |
| Subagent invocation fails | Report the failure; request user decision to retry, skip, or abort |

## Examples
**Input:** `assets/raw/code.txt` contains a React prototype. User invokes `/zero.to.prod`.
**Output:** All 9 phases completed. Application running at `https://himcorp.com/`, tagged as v1.0.0 on GitHub, all specs `status: deployed`, MVP task closed.
