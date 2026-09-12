---
version: "1.0.0"
---

# Project Research Map
**Project:** smaqit
**Refreshed:** 2026-09-12

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
| PlantUML MCP | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@plantuml/mcp-js |
| Resvg WASM | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@resvg/resvg-wasm |
| Noto Sans | Package reference | Unreachable (npmjs.com returns 403 to automated requests) — https://www.npmjs.com/package/@fontsource/noto-sans |
| Git | Documentation | https://git-scm.com/doc |
| GitHub Actions | Documentation | https://docs.github.com/en/actions |
| VS Code | Documentation | https://code.visualstudio.com/docs |
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
