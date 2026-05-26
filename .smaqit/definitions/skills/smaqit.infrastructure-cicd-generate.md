# Skill Definition: smaqit.infrastructure-cicd-generate

## Identity
- **Name:** smaqit.infrastructure-cicd-generate
- **Version:** 1.0.0
- **Description:** Generate the canonical 4-workflow GitHub Actions CI/CD set for a Node.js + React application deployed to a VM via rsync and Docker Compose. Produces: deploy.yml (provision + deploy jobs), provision.yml (PR-only plan preview), post-merge-deploy.yml (body sentinel trigger), copilot-setup-steps.yml (Coding Agent repo access). Applies all project-learned conventions: job sequencing, GITHUB_TOKEN avoidance, PR body sentinel, path filters.

## Steps
1. **Read existing workflows** (if any) from `.github/workflows/` to understand the current state. If workflows already exist and are substantially complete, offer to diff rather than overwrite.
2. **Resolve configuration values** from Infrastructure specs and `copilot-instructions.md`:
   - VM host variable reference (default: `${{ secrets.VM_HOST }}`)
   - SSH key secret name (default: `VM_SSH_KEY`)
   - Terraform backend secret names (default: `TF_BACKEND_ACCESS_KEY`, `TF_BACKEND_SECRET_KEY`)
   - Cloud credential secret names (default: `TF_VAR_APP_CREDENTIAL_ID`, `TF_VAR_APP_CREDENTIAL_SECRET`)
   - GitHub token secret name (default: `GH_TERRAFORM_TOKEN`)
   - Terraform path filter (default: `deployment/terraform/**`)
3. **Generate `deploy.yml`** — must include:
   - Two sequential jobs: `provision` (terraform apply) and `deploy` (rsync + compose restart), with `needs: [provision]` on the deploy job.
   - Provision job env MUST use `TF_VAR_github_token: ${{ secrets.GH_TERRAFORM_TOKEN }}` — NOT `GITHUB_TOKEN` (reserved name; runner overwrites it with installation token before steps execute).
   - Deploy job steps: checkout, setup Node.js, install deps, build backend, build frontend, rsync backend dist, rsync frontend dist, install node_modules on VM via Docker, write deploy stamps, docker compose restart, nginx reload.
   - `VITE_DEMO_MODE: ${{ vars.VITE_DEMO_MODE }}` env var on the frontend build step (baked at build time).
4. **Generate `provision.yml`** — must include:
   - Trigger: `pull_request` with paths filter `deployment/terraform/**`
   - Single job: `plan` — runs `terraform plan` only, NO `terraform apply`
   - Posts plan output as a PR comment
   - Same env vars as provision job in `deploy.yml` (including `TF_VAR_github_token`)
5. **Generate `post-merge-deploy.yml`** — must include:
   - Trigger: `pull_request` event, `types: [closed]`, condition: `if: github.event.pull_request.merged == true && contains(github.event.pull_request.body, 'smaqit:deploy')`
   - Action: trigger `deploy.yml` via `workflow_dispatch`
   - NOTE: Do NOT use a label trigger. GitHub Apps tokens (used by Copilot Coding Agent) lack `issues:write` and cannot apply labels. PR body sentinel is writable at PR creation time by the creating agent.
6. **Generate `copilot-setup-steps.yml`** — must include:
   - `actions/checkout@v4` with explicit `token: ${{ secrets.GITHUB_TOKEN }}` to allow Copilot Coding Agent to access the private repository during setup steps.
   - Keep this file minimal — strip any steps that are not strictly required for Coding Agent access.
7. **Write all 4 files to `.github/workflows/`**.
8. Report: list of files written with a brief description of each.

## Output
- `.github/workflows/deploy.yml`
- `.github/workflows/provision.yml`
- `.github/workflows/post-merge-deploy.yml`
- `.github/workflows/copilot-setup-steps.yml`

## Scope
- Does NOT set up GitHub Secrets or Variables. Use `smaqit.infrastructure-repo-config` for that.
- Does NOT provision infrastructure. Terraform runs inside the generated workflow.
- Does NOT handle matrix builds, monorepos with multiple apps, or container registry pushes.

## Gotchas
- **`GITHUB_TOKEN` is reserved** — GitHub Actions runner injects its own short-lived installation token under `GITHUB_TOKEN` before any job step executes. If a workflow sets `env.GITHUB_TOKEN: ${{ secrets.ANY_PAT }}`, the runner's value overwrites it. Use `TF_VAR_github_token` (non-reserved; Terraform auto-maps `TF_VAR_*` → `var.*`).
- **PR body sentinel vs label** — Copilot Coding Agent cannot apply labels (requires `issues:write`). `post-merge-deploy.yml` must use `contains(github.event.pull_request.body, 'smaqit:deploy')`, not a label check.
- **`provision.yml` must also have `TF_VAR_github_token`** — an easy omission. If only `deploy.yml` has it, PR preview plans will hang at an interactive prompt waiting for the GitHub provider token.
- **`copilot-setup-steps.yml` must include `actions/checkout@v4`** — the Coding Agent needs repo access during setup steps for private repositories. Without checkout, the agent cannot read the codebase.
- **Build time VITE values** — `VITE_DEMO_MODE` and any other `VITE_*` env vars must be passed to the `npm run build` step, not the deploy step. Vite bakes them in at compile time.
- **provider lock file** — if `.terraform.lock.hcl` is committed, `terraform init` in CI will use the locked versions. Do NOT add it to `.gitignore`. Removing it from gitignore is a common fix if CI fails with provider version errors.

## Completion
- [ ] Existing workflows read (or confirmed absent)
- [ ] Configuration values resolved
- [ ] `deploy.yml` written (provision + deploy jobs, sequential, no GITHUB_TOKEN collision)
- [ ] `provision.yml` written (PR-only, plan only, path filter, TF_VAR_github_token present)
- [ ] `post-merge-deploy.yml` written (body sentinel, no label dependency)
- [ ] `copilot-setup-steps.yml` written (checkout included, minimal)
- [ ] Report delivered

## Failure Handling
| Situation | Action |
|-----------|--------|
| Workflows already exist and differ | Show a diff summary; ask whether to overwrite or merge |
| Infrastructure spec not available | Use defaults for all VM and Terraform references; note which values need to be updated post-generation |
| Configuration value ambiguous | Apply the documented default; annotate the generated file with a `# TODO: verify` comment |

## Examples
**Input:** New project with no CI/CD workflows. Operator invokes `/cicd.generate`.
**Output:** 4 files written to `.github/workflows/`. `deploy.yml` has provision→deploy sequential jobs; `provision.yml` PR-only plan; `post-merge-deploy.yml` uses body sentinel; `copilot-setup-steps.yml` has checkout.
