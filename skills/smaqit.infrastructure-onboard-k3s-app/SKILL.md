---
name: smaqit.infrastructure-onboard-k3s-app
description: Use when a project's Infrastructure spec targets a k3s cluster owned by another (infrastructure-owning) repository. Onboarding a new tenant app onto such a cluster — Namespace, RBAC, kubeconfig issuance, and whatever else that specific cluster's operator requires — is that infra repo's own responsibility and contract; it varies per infra repo and must never be assumed or replicated here. This skill's only job is to recognize that hand-off point and defer entirely to the infra repo's own onboarding skill, workflow, or instructions.
metadata:
  version: "1.0.0"
---

# Defer k3s App Onboarding to the Infra Repo

When this project's target deployment environment is a k3s cluster owned and administered by a
different repository, this project does not self-onboard. Locate that infra repo and follow *its
own* onboarding process end to end — whatever shape it takes — rather than assuming, inventing, or
reusing any particular mechanism here.

## Pre-conditions

- The target k3s cluster is owned and administered by a repository other than this project's own.
- This project's Infrastructure spec identifies that infra repo (or the operator can identify it).

## Steps

1. **Identify the infra repo** that owns the target k3s cluster, from this project's Infrastructure
   spec.
2. **Locate that repo's own onboarding mechanism** — its own skill, workflow, or documented process
   for registering a new tenant app. Do not assume it matches any other infra repo's shape (a
   registry file plus a converge workflow is common, but not guaranteed).
3. **Follow that process exactly as that repo defines it** to register this app and receive
   whatever credential or access it issues — typically a scoped kubeconfig, but follow the infra
   repo's own handoff convention rather than assuming one.
4. **Store the received credential** via this project's own `smaqit.infrastructure-vault-loader`
   conventions, ready for this project's own deployment mechanism to consume.

## Output

This app registered as a tenant of the target k3s cluster, with whatever credential the infra
repo's own onboarding process issues, stored in this project's Vault.

## Scope

- Does NOT prescribe, assume, or replicate any infra repo's onboarding mechanics (registry format,
  RBAC scheme, workflow shape) — those are that repo's own contract to define.
- Does NOT provision the cluster itself.
- Does NOT deploy this app's own workloads onto the cluster — a separate, app-side deploy mechanism
  (not yet a smaqit skill) consumes the credential this process yields.

## Failure Handling

| Situation | Action |
|-----------|--------|
| The infra repo's onboarding mechanism cannot be located | Stop and ask the operator to point to it directly — do not guess or invent one |
| The infra repo's process issues no credential, or an unusable one | Stop and report to the operator; do not proceed to deployment without a valid credential |

## Allowed Tools

Bash(git:*), Bash(gh:*)
