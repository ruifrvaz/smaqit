---
version: "1.0.0"
---

# Project Research Map
**Project:** smaqit
**Refreshed:** 2026-10-11

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
