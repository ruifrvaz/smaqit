---
version: "1.0.0"
---

# Project Research Map
**Project:** smaqit
**Refreshed:** 2026-10-25

| Tool | Section | URL |
|------|---------|-----|
| Go | Documentation | https://go.dev/doc/ |
| Go | Getting started | https://go.dev/doc/tutorial/getting-started |
| Model Context Protocol | Specification | https://modelcontextprotocol.io/docs/2026-07-28/getting-started/intro |
| Model Context Protocol Go SDK | Package reference | https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk |
| go-toml | Package reference | https://pkg.go.dev/github.com/pelletier/go-toml/v2 |
| hujson | Package reference | https://pkg.go.dev/github.com/tailscale/hujson |
| YAML v3 | Package reference | https://pkg.go.dev/gopkg.in/yaml.v3 |
| Node.js | Documentation | https://nodejs.org/docs/latest/api/ |
| PlantUML MCP | Documentation | https://plantuml.com/ |
| PlantUML MCP | Repository | https://github.com/plantuml/plantuml |
| PlantUML MCP | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@plantuml/mcp-js |
| Resvg WASM | Repository | https://github.com/thx/resvg-js |
| Resvg WASM | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@resvg/resvg-wasm |
| Noto Sans | Documentation | https://fontsource.org/fonts/noto-sans |
| Noto Sans | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@fontsource/noto-sans |
| Git | Documentation | https://git-scm.com/doc |
| GitHub Actions | Documentation | https://docs.github.com/en/actions |
| VS Code | Documentation | https://code.visualstudio.com/docs |
| Claude Code | MCP configuration | https://code.claude.com/docs/en/mcp |
| github-copilot-sdk | Package reference | https://pypi.org/project/github-copilot-sdk/ |

## Task 107 — Merge Copilot/Codex Skills; Canonicalize AGENTS.md

No task-layer tools; this task is an internal refactor of installer assets and generated instructions.

## Task 109 — Phase Design-Readiness Gate Scans All Active Specs, Not Just the Touched Feature

No task-layer tools; this task is an internal refactor of `installer/spec.go` (Go standard library only, already covered by the project-layer Go entry).

## Task 111 — Design Sequence Diagram Has No Deterministic Enforcement, and a Layer-Mismatch Bug Rejects the Established Convention That Would Link One

No task-layer tools; this task is an internal fix to `installer/design.go` (Go standard library only, already covered by the project-layer Go entry).

## Task 110 — Vault Loader: Wrong Project-Slug Derivation, and Non-Interactive Runs Silently Write Placeholder Secrets

| Tool | Section | URL |
|------|---------|-----|
| HashiCorp Vault | KV Secrets Engine v2 | https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2 |
| HashiCorp Vault | CLI: vault kv | https://developer.hashicorp.com/vault/docs/commands/kv |
| HashiCorp Vault | Dev Server Mode | https://developer.hashicorp.com/vault/docs/concepts/dev-server |
| OpenAI Codex CLI | Documentation | https://github.com/openai/codex |

## Task 112 — Require Identifying Title Directive in Design Artifacts

No task-layer tools; this task adds a `title` directive to PlantUML design templates and enforces it in `installer/design.go` (Go standard library only, already covered by the project-layer Go entry).

## Task 116 — Contribute a k3s App-Onboarding Skill Definition

**Context fingerprint:** sha256:fbc9c1f1ad1e4dbfd2007da2031512997d7b37aaa31ce99252395131ef0fe6e9
**Refreshed:** 2026-09-12

| Tool | Section | URL |
|------|---------|-----|
| Kubernetes | RBAC | https://kubernetes.io/docs/reference/access-authn-authz/rbac/ |
| Kubernetes | Pod Security Admission | https://kubernetes.io/docs/concepts/security/pod-security-admission/ |
| Kubernetes | Network Policies | https://kubernetes.io/docs/concepts/services-networking/network-policies/ |
| Kubernetes | Resource Quotas | https://kubernetes.io/docs/concepts/policy/resource-quotas/ |
| Kubernetes | Limit Ranges | https://kubernetes.io/docs/concepts/policy/limit-range/ |
| Kubernetes | Service Accounts | https://kubernetes.io/docs/concepts/security/service-accounts/ |
| k3s | Documentation | https://docs.k3s.io/ |
| GitHub Actions | Manually running a workflow (workflow_dispatch) | https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |

## Task 117 — k3s App-Deployment Skill With Routing and CI/CD

**Context fingerprint:** sha256:de23ba8d30aa552f06449f35cbfa314993122af3338e75be066906ed4baf97b1
**Refreshed:** 2026-09-12

| Tool | Section | URL |
|------|---------|-----|
| Kubernetes | Pod Security Admission | https://kubernetes.io/docs/concepts/security/pod-security-admission/ |
| Kubernetes | Network Policies | https://kubernetes.io/docs/concepts/services-networking/network-policies/ |
| Kubernetes | Resource Quotas | https://kubernetes.io/docs/concepts/policy/resource-quotas/ |
| Kubernetes | RBAC | https://kubernetes.io/docs/reference/access-authn-authz/rbac/ |
| Kubernetes | Ingress | https://kubernetes.io/docs/concepts/services-networking/ingress/ |
| k3s | Documentation | https://docs.k3s.io/ |
| cert-manager | Documentation | https://cert-manager.io/docs/ |
| Traefik | Kubernetes Ingress provider | https://doc.traefik.io/traefik/providers/kubernetes-ingress/ |
| kubectl | Reference | https://kubernetes.io/docs/reference/kubectl/ |
| GitHub Actions | Environments | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| GitHub Actions | workflow_dispatch | https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow |

## Task 118 — Request k3s App Onboarding via Platform-Repo PR

**Context fingerprint:** sha256:b5ec7f4b31544d5b4144244783812ff3533ddac316dab252d309513692409940
**Refreshed:** 2026-09-12

| Tool | Section | URL |
|------|---------|-----|
| GitHub CLI | pr create | https://cli.github.com/manual/gh_pr_create |
| GitHub CLI | pr view | https://cli.github.com/manual/gh_pr_view |
| GitHub REST API | Pulls | https://docs.github.com/en/rest/pulls/pulls |
| Git | Documentation | https://git-scm.com/doc |

## Task 119 — Fix Registry-Write Semantics, Deployment Sizing, and Lifecycle Gaps in the k3s App-Deployment Skills

No new task-layer tools; reuses GitHub CLI/API and Kubernetes resource-management references
already verified for Tasks 117/118, plus one addition (GitHub REST API repository-contents, used
for the read-before-append fix in Finding 1).

| Tool | Section | URL |
|------|---------|-----|
| GitHub CLI | pr create | https://cli.github.com/manual/gh_pr_create |
| GitHub CLI | pr view | https://cli.github.com/manual/gh_pr_view |
| GitHub REST API | Repository contents | https://docs.github.com/en/rest/repos/contents |
| Kubernetes | Resource Quotas | https://kubernetes.io/docs/concepts/policy/resource-quotas/ |
| Kubernetes | Limit Ranges | https://kubernetes.io/docs/concepts/policy/limit-range/ |

## Task 120 — Surface and Fail Gracefully on Missing existing-k3s Infrastructure-Spec Fields

No new task-layer tools; this task only edits smaqit's own skill/agent instruction files
(`smaqit.infrastructure-request-k3s-onboarding`, `smaqit.input-deployment`,
`agents/infrastructure.md`, `smaqit.feature-new`) to name required spec fields and add a
Failure Handling row — it introduces no new external technology, script, or API call. Reuses
the Kubernetes/cert-manager references already verified for Tasks 116/117, since the fields in
scope (ingress class, `ClusterIssuer` name, quota/limit numbers) map to those same concepts.

| Tool | Section | URL |
|------|---------|-----|
| Kubernetes | Ingress | https://kubernetes.io/docs/concepts/services-networking/ingress/ |
| Kubernetes | Resource Quotas | https://kubernetes.io/docs/concepts/policy/resource-quotas/ |
| Kubernetes | Limit Ranges | https://kubernetes.io/docs/concepts/policy/limit-range/ |
| cert-manager | Documentation | https://cert-manager.io/docs/ |

## Task 121 — Harden the k3s App-Deployment Skill Family: Image Build-Push, Lint, and Vault Conventions

**Context fingerprint:** sha256:83769f1bec4a454aba9ccebb413e92276348759089fb6fcb77d1af8fcbb63573
**Refreshed:** 2026-09-15

| Tool | Section | URL |
|------|---------|-----|
| Docker | Dockerfile reference | https://docs.docker.com/reference/dockerfile/ |
| Docker | docker build CLI reference | https://docs.docker.com/reference/cli/docker/buildx/build/ |
| Docker | docker push CLI reference | https://docs.docker.com/reference/cli/docker/image/push/ |
| Kubernetes | Pod Security Admission | https://kubernetes.io/docs/concepts/security/pod-security-admission/ |
| Kubernetes | Resource Quotas | https://kubernetes.io/docs/concepts/policy/resource-quotas/ |
| Kubernetes | Limit Ranges | https://kubernetes.io/docs/concepts/policy/limit-range/ |
| k3s | Documentation | https://docs.k3s.io/ |
| GitHub Container Registry | Working with the Container registry | https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry |
| GitHub Packages | Configuring package access control and visibility | https://docs.github.com/en/packages/learn-github-packages/configuring-a-packages-access-control-and-visibility |
| GitHub Actions | Manually running a workflow (workflow_dispatch) | https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| HashiCorp Vault | KV Secrets Engine v2 | https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2 |
| HashiCorp Vault | CLI: vault kv | https://developer.hashicorp.com/vault/docs/commands/kv |

## Task 125 — Deployment Variation Points: Slot-Driven Deployment Orchestration With Project-Bound Fillers

**Context fingerprint:** sha256:aa378b49beb8f8d0854ded0161e0ce307cc1841110348f31e2d9f575e02d450d
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Subagents | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Agent Skills | https://code.claude.com/docs/en/skills |
| Claude Code | Memory (CLAUDE.md imports) | https://code.claude.com/docs/en/memory |
| GitHub Copilot | Custom agents | https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/create-custom-agents |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| Codex CLI | Repository | https://github.com/openai/codex |
| Codex CLI | AGENTS.md | https://learn.chatgpt.com/docs/agent-configuration/agents-md |
| Codex CLI | Skills | https://learn.chatgpt.com/docs/build-skills |
| Codex CLI | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |
| Agent Skills | Specification | https://agentskills.io/specification |
| smaqit-adk | Repository | https://github.com/ruifrvaz/smaqit-adk |
| Go | Documentation | https://go.dev/doc/ |
| Go | flag package | https://pkg.go.dev/flag |
| Go | embed package | https://pkg.go.dev/embed |
| gopkg.in/yaml.v3 | Package reference | https://pkg.go.dev/gopkg.in/yaml.v3 |
| gopkg.in/yaml.v3 | Node API | https://pkg.go.dev/gopkg.in/yaml.v3#Node |

## Task 127 — Deployment Slots A: Principle, Catalogue, Schema, and `smaqit slots plan` Resolver

**Context fingerprint:** sha256:0fa54aacb01ca2a206a1a63bbd13f4fd83d1c0bf5f3877df77bdb5d4c4bc78f5
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Go | Documentation | https://go.dev/doc/ |
| Go | flag package | https://pkg.go.dev/flag |
| Go | embed package | https://pkg.go.dev/embed |
| gopkg.in/yaml.v3 | Package reference | https://pkg.go.dev/gopkg.in/yaml.v3 |
| gopkg.in/yaml.v3 | Node API | https://pkg.go.dev/gopkg.in/yaml.v3#Node |
| Agent Skills | Specification | https://agentskills.io/specification |

## Task 128 — Deployment Slots B: Deployment Agent Executes the Slot Plan; feature-new Routed Through It

**Context fingerprint:** sha256:e113ce6defa27f412c550542c23ce111714b92ced41a8d1011c31a44b4ed6156
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Subagents | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Agent Skills | https://code.claude.com/docs/en/skills |
| GitHub Copilot | Custom agents | https://docs.github.com/en/copilot/how-tos/copilot-on-github/customize-copilot/customize-cloud-agent/create-custom-agents |
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| Codex CLI | Repository | https://github.com/openai/codex |
| Codex CLI | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |
| Codex CLI | Skills | https://learn.chatgpt.com/docs/build-skills |
| Agent Skills | Specification | https://agentskills.io/specification |

## Task 129 — Deployment Slots C: Route new-greenfield-project Phases 4–6 Through the Deployment Agent

**Context fingerprint:** sha256:93a690e86cb91c39d1d92914d3b3992481bbe294e5c2930880c9aafa3bffaab8
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| GitHub Actions | Workflow syntax | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| GitHub CLI | gh run watch | https://cli.github.com/manual/gh_run_watch |
| GitHub CLI | gh run list | https://cli.github.com/manual/gh_run_list |
| Terraform | plan command | https://developer.hashicorp.com/terraform/cli/commands/plan |
| Terraform | destroy command | https://developer.hashicorp.com/terraform/cli/commands/destroy |
| Kubernetes | Deployments | https://kubernetes.io/docs/concepts/workloads/controllers/deployment/ |
| k3s | Documentation | https://docs.k3s.io/ |

## Task 131 — Claude Code Agents Can't Invoke Their Input Skills or Ask the User Questions

**Context fingerprint:** sha256:0baced4c6086fffc935e8d512fcdef14f0b15cfde79df99fcbd675505c8b2731
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Subagents (skills preload, resume) | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Agent Skills | https://code.claude.com/docs/en/skills |
| Claude Code | Slash commands | https://code.claude.com/docs/en/slash-commands |
| Claude Code | Tools reference | https://code.claude.com/docs/en/tools-reference |
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| Codex CLI | Repository | https://github.com/openai/codex |
| Codex CLI | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |
| Codex CLI | Skills | https://learn.chatgpt.com/docs/build-skills |

## Task 114 — Refactor Infrastructure Skills for the Machine-Monorepo Pattern (Learnings from a Downstream Project's Infrastructure Task 001)

**Context fingerprint:** sha256:1d86aae9989446a4591e5f14be1e20be40957606b83180c230a835a14cc366fa
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |
| Bash | Bash builtins (read) | https://www.gnu.org/software/bash/manual/html_node/Bash-Builtins.html |
| Go | Documentation | https://go.dev/doc/ |
| Go | testing package | https://pkg.go.dev/testing |
| HashiCorp Vault | KV Secrets Engine v2 | https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2 |
| HashiCorp Vault | CLI: vault kv | https://developer.hashicorp.com/vault/docs/commands/kv |
| HashiCorp Vault | CLI: vault kv patch | https://developer.hashicorp.com/vault/docs/commands/kv/patch |
| GitHub CLI | gh secret set | https://cli.github.com/manual/gh_secret_set |
| GitHub CLI | gh variable set | https://cli.github.com/manual/gh_variable_set |
| GitHub CLI | gh api | https://cli.github.com/manual/gh_api |
| GitHub REST API | Deployment environments (create or update) | https://docs.github.com/en/rest/deployments/environments |
| GitHub REST API | Permissions required for fine-grained PATs | https://docs.github.com/en/rest/authentication/permissions-required-for-fine-grained-personal-access-tokens |
| GitHub | Managing personal access tokens | https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| GitHub Actions | Workflow syntax (env, jobs.<id>.env) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| Terraform | S3 backend | https://developer.hashicorp.com/terraform/language/backend/s3 |
| Terraform | lifecycle meta-argument (prevent_destroy) | https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle |
| Terraform | OpenStack provider: blockstorage_volume_v3 | https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/blockstorage_volume_v3 |
| OpenStack | CLI: volume set (retype) | https://docs.openstack.org/python-openstackclient/latest/cli/command-objects/volume/v3/index.html |
| OpenStack | Cinder volume retype | https://docs.openstack.org/cinder/latest/admin/volume-migration.html |
| OpenStack | CLI: volume type | https://docs.openstack.org/python-openstackclient/latest/cli/command-objects/volume/v3/index.html |

## Task 130 — Make smaqit Self-Invoking: Directive Instructions, Block Refresh, Trigger-Form Descriptions

**Context fingerprint:** sha256:717c2b031595e4e46835a069aab2bf9ae7726c0671bae8a038e6fe0262f131f8
**Refreshed:** 2026-10-11

| Tool | Section | URL |
|------|---------|-----|
| Go | Package testing | https://pkg.go.dev/testing |
| Go | Package strings | https://pkg.go.dev/strings |
| Python | Documentation | https://docs.python.org/3/ |
| Python | re module | https://docs.python.org/3/library/re.html |
| PyYAML | Documentation | https://pyyaml.org/wiki/PyYAMLDocumentation |
| PyYAML | Package reference | https://pypi.org/project/PyYAML/ |
| Claude Code | Agent Skills (description budget, when_to_use) | https://code.claude.com/docs/en/skills |
| Claude Code | Subagents (proactive delegation) | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Memory (CLAUDE.md imports, AGENTS.md) | https://code.claude.com/docs/en/memory |
| Claude Code | Settings (skill listing budget) | https://code.claude.com/docs/en/settings |
| Claude Code | Features overview (context cost) | https://code.claude.com/docs/en/features-overview |

## Task 133 — Copilot and Codex Entry-Agent Descriptions in Trigger Form

**Context fingerprint:** sha256:c047a36c5b8ac6a4643635c9745115a82766fe4aeb4ba5df0817b02f9d3e83f2
**Refreshed:** 2026-10-12

| Tool | Section | URL |
|------|---------|-----|
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| Codex CLI | Repository | https://github.com/openai/codex |
| Codex CLI | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |

## Task 132 — Claude Code Headless E2E Test: Install Into a Scratch Folder and Build a Site Through smaqit

**Context fingerprint:** sha256:cdfcc03d2b7522701ab1c85b55035e82917824ec0a07cd2314275a1333e70802

**Refreshed:** 2026-10-12

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Headless mode (claude -p) | https://code.claude.com/docs/en/headless |
| Claude Code | CLI reference | https://code.claude.com/docs/en/cli-reference |
| Claude Code | Authentication (setup-token, CLAUDE_CONFIG_DIR) | https://code.claude.com/docs/en/authentication |
| Claude Code | Sessions and resume | https://code.claude.com/docs/en/sessions |
| Claude Code | MCP (project servers) | https://code.claude.com/docs/en/mcp |
| Claude Code | Subagents | https://code.claude.com/docs/en/sub-agents |
| Node.js | Test runner (node --test) | https://nodejs.org/api/test.html |
| Node.js | Documentation | https://nodejs.org/docs/latest-v22.x/api/ |
| Go | Documentation | https://go.dev/doc/ |

## Task 115 — Platform Repositories: Guarded Replacement, Cyso Facts, and a Pattern Guide

**Context fingerprint:** sha256:5654b6aa8a7070271ec0e8e04714beaef698a29ab5ebd12b0b998cec1b477e7d
**Refreshed:** 2026-10-12

| Tool | Section | URL |
|------|---------|-----|
| Terraform | Command: plan (-replace, -out) | https://developer.hashicorp.com/terraform/cli/commands/plan |
| Terraform | Command: apply (saved plan) | https://developer.hashicorp.com/terraform/cli/commands/apply |
| Terraform | Command: show -json | https://developer.hashicorp.com/terraform/cli/commands/show |
| Terraform | JSON output format (resource_changes actions) | https://developer.hashicorp.com/terraform/internals/json-format |
| Terraform | S3 backend | https://developer.hashicorp.com/terraform/language/backend/s3 |
| Terraform | lifecycle meta-argument (ignore_changes, precondition) | https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle |
| Terraform | base64gzip function | https://developer.hashicorp.com/terraform/language/functions/base64gzip |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |
| jq | Manual | https://jqlang.org/manual/ |
| OpenStack | Nova user data (metadata service) | https://docs.openstack.org/nova/latest/user/metadata.html |
| OpenStack | Terraform provider: compute_instance_v2 | https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/compute_instance_v2 |
| GitHub Actions | Workflow syntax | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| cloud-init | write_files (encoding gz+b64) | https://docs.cloud-init.io/en/latest/reference/modules.html |
| cloud-init | Documentation | https://docs.cloud-init.io/en/latest/ |

## Task 134 — Apply the Guarded Plan in the Generated deploy.yml (full Mode)

**Context fingerprint:** sha256:bae7925cfc274ab03c4751b312fd1ace4b1768539f7256aedb5981304d75ed03
**Refreshed:** 2026-10-12

| Tool | Section | URL |
|------|---------|-----|
| Terraform | Command: apply (saved plan) | https://developer.hashicorp.com/terraform/cli/commands/apply |
| Terraform | Command: plan (-out) | https://developer.hashicorp.com/terraform/cli/commands/plan |
| GitHub Actions | Workflow syntax (working-directory) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |

## Task 123 — existing-k3s Skill Family: Stale Text, Cross-Skill Drift, Raw Vault Usage, and an Un-Upstreamed vault-loader Patch

**Context fingerprint:** sha256:362aa251ed6d8c28f25baa74c14c41aa62f99938991410792e8f3b1cb78cac1c
**Refreshed:** 2026-10-12

| Tool | Section | URL |
|------|---------|-----|
| HashiCorp Vault | KV Secrets Engine v2 | https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2 |
| HashiCorp Vault | CLI: vault kv | https://developer.hashicorp.com/vault/docs/commands/kv |
| HashiCorp Vault | CLI: vault secrets enable | https://developer.hashicorp.com/vault/docs/commands/secrets/enable |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |
| GitHub CLI | gh secret set | https://cli.github.com/manual/gh_secret_set |
| GitHub CLI | gh variable set | https://cli.github.com/manual/gh_variable_set |
| GitHub CLI | gh api | https://cli.github.com/manual/gh_api |
| GitHub Actions | Workflow syntax (workflow_dispatch inputs, environment) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Using environments for deployment | https://docs.github.com/en/actions/how-tos/deploy/configure-and-manage-deployments/manage-environments |
| Kubernetes | RBAC authorization (pods/log subresource) | https://kubernetes.io/docs/reference/access-authn-authz/rbac/ |
| Kubernetes | kubectl logs | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_logs/ |
| Docker | docker build --build-arg | https://docs.docker.com/reference/cli/docker/buildx/build/ |
| Docker | Dockerfile ARG | https://docs.docker.com/reference/dockerfile/ |
| Terraform | Environment variables (TF_VAR_) | https://developer.hashicorp.com/terraform/cli/config/environment-variables |
| Terraform | S3 backend | https://developer.hashicorp.com/terraform/language/backend/s3 |

## Task 122 — existing-k3s: Support Private (Mesh-Only) Cluster Targets Alongside Public Ones

**Context fingerprint:** sha256:1ab3bf9382ea0be1eef806b973b3f626a842a0a251fb0bf2d44821a5e0925f42
**Refreshed:** 2026-10-16

| Tool | Section | URL |
|------|---------|-----|
| k3s | Documentation | https://docs.k3s.io/ |
| k3s | Networking services (Traefik, ServiceLB) | https://docs.k3s.io/networking/networking-services |
| Kubernetes | Jobs | https://kubernetes.io/docs/concepts/workloads/controllers/job/ |
| Kubernetes | TTL-after-finished controller | https://kubernetes.io/docs/concepts/workloads/controllers/ttlafterfinished/ |
| Kubernetes | DNS for Services and Pods | https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/ |
| Kubernetes | Ingress (TLS) | https://kubernetes.io/docs/concepts/services-networking/ingress/ |
| Kubernetes | Persistent Volumes (claims) | https://kubernetes.io/docs/concepts/storage/persistent-volumes/ |
| Kubernetes | kubectl wait | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_wait/ |
| Kubernetes | kubectl auth can-i | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_auth/kubectl_auth_can-i/ |
| Kubernetes | Pod Security Standards (restricted) | https://kubernetes.io/docs/concepts/security/pod-security-standards/ |
| Traefik | Kubernetes Ingress provider | https://doc.traefik.io/traefik/providers/kubernetes-ingress/ |
| Traefik | TLS (default certificate, TLS stores) | https://doc.traefik.io/traefik/https/tls/ |
| cert-manager | Securing Ingress resources (ingress-shim) | https://cert-manager.io/docs/usage/ingress/ |
| cert-manager | DNS01 challenge | https://cert-manager.io/docs/configuration/acme/dns01/ |
| NetBird | Documentation | https://docs.netbird.io/ |
| NetBird | Setup keys | https://docs.netbird.io/manage/peers/register-machines-using-setup-keys |
| GitHub Actions | Workflow syntax | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Expressions (fromJSON) | https://docs.github.com/en/actions/reference/workflows-and-actions/expressions |
| GitHub Actions | Contexts (secrets, vars, env) | https://docs.github.com/en/actions/reference/workflows-and-actions/contexts |
| GitHub Actions | Events that trigger workflows (workflow_dispatch) | https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows |
| HashiCorp Vault | KV Secrets Engine v2 | https://developer.hashicorp.com/vault/docs/secrets/kv/kv-v2 |
| HashiCorp Vault | CLI: vault kv | https://developer.hashicorp.com/vault/docs/commands/kv |
| jq | Manual | https://jqlang.org/manual/ |
| Go | Package testing (golden files, flags) | https://pkg.go.dev/testing |
| Go | YAML v3 | https://pkg.go.dev/gopkg.in/yaml.v3 |

## Task 135 — plan-guard.sh: `--allow-delete <address>` for a Reviewed Pure Destroy

**Context fingerprint:** sha256:fc69da7972ce721d046fc8cd26c7be285452017c9b2edd7962161e11a342e179
**Refreshed:** 2026-10-18

| Tool | Section | URL |
|------|---------|-----|
| Terraform | JSON output format (resource_changes actions) | https://developer.hashicorp.com/terraform/internals/json-format |
| Terraform | Command: plan (-replace, -detailed-exitcode, -out) | https://developer.hashicorp.com/terraform/cli/commands/plan |
| Terraform | Command: show -json | https://developer.hashicorp.com/terraform/cli/commands/show |
| Terraform | Resource addressing | https://developer.hashicorp.com/terraform/cli/state/resource-addressing |
| jq | Manual | https://jqlang.org/manual/ |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |
| Bash | The Set builtin (-e, -u, pipefail) | https://www.gnu.org/software/bash/manual/html_node/The-Set-Builtin.html |
| Go | Package testing | https://pkg.go.dev/testing |
| Go | Package os/exec | https://pkg.go.dev/os/exec |

## Task 136 — existing-k3s: Fixes From the First Live Private-Mesh Deploy

**Context fingerprint:** sha256:97ebcc4a4907c2710981da023bf5267ba9c3d66f11635cfaeb153df42c581a0b
**Refreshed:** 2026-10-18

| Tool | Section | URL |
|------|---------|-----|
| GitHub Actions | Workflow syntax (shell defaults, bash -e, concurrency) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Control workflow concurrency | https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency |
| GitHub Actions | Using secrets in workflows | https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets |
| GitHub Actions | Security hardening (script injection) | https://docs.github.com/en/actions/reference/security/secure-use |
| GitHub Actions | Store information in variables (RUNNER_TEMP) | https://docs.github.com/en/actions/reference/workflows-and-actions/variables |
| GitHub CLI | gh workflow run | https://cli.github.com/manual/gh_workflow_run |
| GitHub CLI | gh environment variables (GH_REPO, GH_TOKEN) | https://cli.github.com/manual/gh_help_environment |
| Kubernetes (k3s) | k3s Documentation | https://docs.k3s.io/ |
| Kubernetes (k3s) | Pull an image from a private registry | https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/ |
| Kubernetes (k3s) | Jobs | https://kubernetes.io/docs/concepts/workloads/controllers/job/ |
| Kubernetes (k3s) | RBAC authorization (subresources) | https://kubernetes.io/docs/reference/access-authn-authz/rbac/ |
| kubectl | kubectl auth can-i | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_auth/kubectl_auth_can-i/ |
| kubectl | kubectl exec | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_exec/ |
| kubectl | kubectl create secret generic | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_create/kubectl_create_secret_generic/ |
| Docker | Docker config.json (credential format) | https://docs.docker.com/reference/cli/docker/#configuration-files |
| jq | Manual | https://jqlang.org/manual/ |

## Task 138 — Phase Mode Defaults by Trigger (Commands Assisted, Skills Autonomous) and Requirements Provenance

**Context fingerprint:** sha256:cf3ff3a9a37912339f822e1c638491400742fd30b4c7daf767dbb840a133ba5d
**Refreshed:** 2026-10-18

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Agent Skills (model invocation, when_to_use) | https://code.claude.com/docs/en/skills |
| Claude Code | Slash commands | https://code.claude.com/docs/en/slash-commands |
| Claude Code | Subagents (proactive delegation) | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Headless mode (claude -p) | https://code.claude.com/docs/en/headless |
| Claude Code | Memory (CLAUDE.md imports, AGENTS.md) | https://code.claude.com/docs/en/memory |
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| OpenAI Codex | Repository | https://github.com/openai/codex |
| OpenAI Codex | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |
| OpenAI Codex | AGENTS.md | https://learn.chatgpt.com/docs/agent-configuration/agents-md |

## Task 140 — plan-guard BLOCKED Guidance: Stop and Report, Never Take an Allow-Flag Address From the Guard's Output

**Context fingerprint:** sha256:f905124be2b8b66d0cc100d0ead626f21fe6311a44643eebfd65bc00fc45c5e1
**Refreshed:** 2026-10-18

| Tool | Section | URL |
|------|---------|-----|
| Terraform | Command: plan (-detailed-exitcode, -out) | https://developer.hashicorp.com/terraform/cli/commands/plan |
| Terraform | JSON output format (resource_changes actions) | https://developer.hashicorp.com/terraform/internals/json-format |
| Terraform | Command: apply (saved plan) | https://developer.hashicorp.com/terraform/cli/commands/apply |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |

## Task 141 — Orchestrator Skills Default by Invocation; Rename new-greenfield-project to new-project; Prune Retired Skills

**Context fingerprint:** sha256:37f1f9a9ee7934b035ec365c3f113767395093394498dfce0a7e41ef3f71b3fc
**Refreshed:** 2026-10-18

| Tool | Section | URL |
|------|---------|-----|
| Claude Code | Agent Skills (invocation, user-invocable) | https://code.claude.com/docs/en/skills |
| Claude Code | Slash commands | https://code.claude.com/docs/en/slash-commands |
| Claude Code | Subagents | https://code.claude.com/docs/en/sub-agents |
| Claude Code | Memory (CLAUDE.md imports) | https://code.claude.com/docs/en/memory |
| GitHub Copilot | Agent skills | https://docs.github.com/en/copilot/concepts/agents/about-agent-skills |
| GitHub Copilot | VS Code agent skills | https://code.visualstudio.com/docs/agent-customization/agent-skills |
| GitHub Copilot | VS Code custom agents | https://code.visualstudio.com/docs/agent-customization/custom-agents |
| OpenAI Codex | Skills | https://learn.chatgpt.com/docs/build-skills |
| OpenAI Codex | Subagents | https://learn.chatgpt.com/docs/agent-configuration/subagents |
| OpenAI Codex | AGENTS.md | https://learn.chatgpt.com/docs/agent-configuration/agents-md |
| Go | Package embed | https://pkg.go.dev/embed |
| Go | Package io/fs (WalkDir) | https://pkg.go.dev/io/fs |
| Go | Package os (UserConfigDir, RemoveAll) | https://pkg.go.dev/os |
| Go | Package testing | https://pkg.go.dev/testing |
| Bash | Reference manual | https://www.gnu.org/software/bash/manual/bash.html |
| Python | shutil (copytree, rmtree) | https://docs.python.org/3/library/shutil.html |

## Task 142 — CI-Driven Deploys Run No Local Target Checks

**Context fingerprint:** sha256:d0040a28e7f266810f30663a2d9606dc4235b76cb85aa1379c9b22af5a638728
**Refreshed:** 2026-10-20

| Tool | Section | URL |
|------|---------|-----|
| GitHub CLI | gh run view (--json jobs, headSha) | https://cli.github.com/manual/gh_run_view |
| GitHub CLI | gh api (placeholders {owner}/{repo}) | https://cli.github.com/manual/gh_api |
| GitHub CLI | gh environment variables (GH_REPO, GH_TOKEN) | https://cli.github.com/manual/gh_help_environment |
| GitHub CLI | gh workflow run | https://cli.github.com/manual/gh_workflow_run |
| GitHub REST API | Compare two commits (status identical/ahead/behind/diverged) | https://docs.github.com/en/rest/commits/commits |
| GitHub REST API | Workflow jobs (steps, conclusion) | https://docs.github.com/en/rest/actions/workflow-jobs |
| GitHub REST API | Workflow runs (head_sha, status, conclusion) | https://docs.github.com/en/rest/actions/workflow-runs |
| GitHub Actions | Workflow syntax (jobs.<id>.steps[*].if) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Events that trigger workflows (workflow_dispatch GITHUB_SHA) | https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows |
| GitHub Actions | Contexts (github.sha, steps.conclusion) | https://docs.github.com/en/actions/reference/workflows-and-actions/contexts |
| Kubernetes (k3s) | k3s Documentation | https://docs.k3s.io/ |
| Kubernetes (k3s) | Jobs | https://kubernetes.io/docs/concepts/workloads/controllers/job/ |
| jq | Manual | https://jqlang.org/manual/ |

## Task 144 — All Deploys Go Through CI/CD Workflows (existing-k3s Enforced, VM Modes Transitional)

**Context fingerprint:** sha256:74a9fa190ff79a8263fedbb58f32cfd5866c5a02e84d2671d3fcab7b18a4bae1
**Refreshed:** 2026-10-23

| Tool | Section | URL |
|------|---------|-----|
| GitHub Actions | Events that trigger workflows (workflow_dispatch, default-branch requirement) | https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows |
| GitHub Actions | Manually running a workflow (workflow_dispatch inputs) | https://docs.github.com/en/actions/how-tos/manage-workflow-runs/manually-run-a-workflow |
| GitHub Actions | Deployments and environments (environment: <name>) | https://docs.github.com/en/actions/reference/workflows-and-actions/deployments-and-environments |
| GitHub Actions | Workflow syntax (on.workflow_dispatch.inputs, jobs.<id>.environment) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Contexts (github.sha, inputs) | https://docs.github.com/en/actions/reference/workflows-and-actions/contexts |
| GitHub CLI | gh workflow run (-f environment=…) | https://cli.github.com/manual/gh_workflow_run |
| GitHub CLI | gh run list (--commit, --workflow, --event) | https://cli.github.com/manual/gh_run_list |
| GitHub CLI | gh run watch (--exit-status) | https://cli.github.com/manual/gh_run_watch |
| GitHub CLI | gh run view (--json headSha,jobs) | https://cli.github.com/manual/gh_run_view |
| GitHub CLI | gh pr create | https://cli.github.com/manual/gh_pr_create |
| k3s | Documentation | https://docs.k3s.io/ |
| k3s | Cluster access (kubeconfig) | https://docs.k3s.io/cluster-access |
| Kubernetes | Organizing cluster access using kubeconfig files | https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/ |
| Kubernetes | kubectl reference | https://kubernetes.io/docs/reference/kubectl/ |
| Go | testing package | https://pkg.go.dev/testing |
| Go | Documentation | https://go.dev/doc/ |

## Task 145 — existing-k3s: CI-identity deploys without carried credentials (OIDC token, shared mesh key, platform-owned pulls)

**Context fingerprint:** sha256:bdb9d76b8f5f7140394d28247cea9752d1ffb069f6fdd555501d0e255001b268
**Refreshed:** 2026-10-25

| Tool | Section | URL |
|------|---------|-----|
| GitHub Actions OIDC | OpenID Connect in GitHub Actions (claims, audience) | https://docs.github.com/en/actions/concepts/security/openid-connect |
| GitHub Actions OIDC | OIDC reference (token claims, id-token permission) | https://docs.github.com/en/actions/reference/security/oidc |
| GitHub Actions | Workflow syntax (jobs.<id>.permissions) | https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax |
| GitHub Actions | Contexts (secrets context) | https://docs.github.com/en/actions/reference/workflows-and-actions/contexts |
| GitHub organisation secrets | Using secrets in GitHub Actions (organization secrets, access policy) | https://docs.github.com/en/actions/how-tos/write-workflows/choose-what-workflows-do/use-secrets |
| Kubernetes structured authentication | Authenticating (structured authentication configuration, JWT authenticator) | https://kubernetes.io/docs/reference/access-authn-authz/authentication/ |
| kubectl | client-go credential plugins (exec ExecCredential) | https://kubernetes.io/docs/reference/access-authn-authz/authentication/#client-go-credential-plugins |
| kubectl | kubectl auth can-i | https://kubernetes.io/docs/reference/kubectl/generated/kubectl_auth/kubectl_auth_can-i/ |
| k3s | Documentation | https://docs.k3s.io/ |
| k3s | Cluster access | https://docs.k3s.io/cluster-access |
| NetBird | Documentation | https://docs.netbird.io/ |
| NetBird | Setup keys | https://docs.netbird.io/manage/peers/register-machines-using-setup-keys |
| GHCR | Working with the Container registry | https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry |
| Go | testing package | https://pkg.go.dev/testing |
| Go | Documentation | https://go.dev/doc/ |
