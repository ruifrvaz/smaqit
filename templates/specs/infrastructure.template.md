---
id: INF-[CONCEPT]
status: draft
created: [TIMESTAMP]
---

# [CONCEPT_NAME]

## Design References

- [DSG-INF-[CONCEPT]-DEPLOYMENT](../../docs/designs/infrastructure/dsg-inf-[concept]-deployment.md) · [Image](../../docs/designs/infrastructure/dsg-inf-[concept]-deployment.png)

## References

<!-- References establish traceability and coherence, not requirement derivation -->
<!-- Infrastructure specs enable the entire application to run in a target environment -->
<!-- Use Implements for feature specs (1:1 mapping) -->
<!-- Use Enables for foundation specs (1:many mapping) -->
<!-- Use same-layer references when feature specs extend foundation specs -->

### Foundation Reference

<!-- Same-layer reference: use when this feature spec extends a foundation spec -->
<!-- Omit this section if this spec doesn't depend on a foundation spec in the same layer -->

- [INF-[FOUNDATION-CONCEPT]](./[FOUNDATION-FILENAME].md) — [Shared requirements referenced here]

### Implements

#### Business
- [BUS-CONCEPT](../business/[FILENAME].md) — [Business requirement this infrastructure implements]

#### Functional
- [FUN-CONCEPT](../functional/[FILENAME].md) — [Functional behavior this infrastructure implements]

#### Stack
- [STK-CONCEPT](../stack/[FILENAME].md) — [Technology constraint this infrastructure implements]

### Enables

#### Business
- [BUS-CONCEPT](../business/[FILENAME].md) — [Business requirement this infrastructure enables]

#### Functional
- [FUN-CONCEPT](../functional/[FILENAME].md) — [Functional behavior this infrastructure supports]

#### Stack
- [STK-CONCEPT](../stack/[FILENAME].md) — [Technology constraint this infrastructure accommodates]

## Scope

### Included

[What this specification covers]

### Excluded

[What this specification explicitly does not cover]

## Compute Resources

### Service Topology

[How the application is structured for deployment — monolith, microservices, serverless functions, etc.]

### Resources

| Service | Type | CPU | Memory | Storage | Purpose |
|---------|------|-----|--------|---------|---------|
| [SERVICE_NAME] | [Container/VM/Serverless] | [Limit] | [Limit] | [Limit] | [What this service does] |

## Networking

### Topology

[Network structure — VPC, subnets, load balancers, CDN, etc.]

### Security Boundaries

| Boundary | Ingress | Egress | Purpose |
|----------|---------|--------|---------|
| [BOUNDARY_NAME] | [Allowed sources] | [Allowed destinations] | [What this boundary protects] |

### Integration Points

| System | Protocol | Direction | Purpose |
|--------|----------|-----------|---------|
| [EXTERNAL_SYSTEM] | [HTTP/gRPC/etc.] | [Inbound/Outbound] | [What this integration provides] |

## Scaling

| Service | Metric | Threshold | Min | Max | Cooldown |
|---------|--------|-----------|-----|-----|----------|
| [SERVICE_NAME] | [CPU/Memory/Requests] | [Trigger value] | [Min instances] | [Max instances] | [Seconds] |

## Observability

### Logging

| Log Type | Destination | Retention | Purpose |
|----------|-------------|-----------|---------|
| [Application/Access/Error] | [Where logs are sent] | [How long kept] | [What it captures] |

### Metrics

| Metric | Source | Alert Threshold | Purpose |
|--------|--------|-----------------|---------|
| [METRIC_NAME] | [Service/Infrastructure] | [When to alert] | [What it measures] |

### Tracing

[Distributed tracing approach — sampling rate, trace propagation, storage]

## Secrets Management

| Secret | Type | Source | Rotation |
|--------|------|--------|----------|
| [SECRET_NAME] | [API Key/Password/Certificate] | [Vault/KMS/etc.] | [Rotation policy] |

## Environments

<!-- `provisioning_mode: existing-k3s` only — omit this section otherwise -->
<!-- One row per environment; every value comes from the platform team's onboarding documentation, never invented -->
<!-- `smaqit slots plan --mode=existing-k3s` refuses to plan an environment whose row is incomplete; `smaqit spec environments` validates and prints this table -->
<!-- Leave a cell as — when the rules below say it must be empty -->
<!-- Reachability: public (CI reaches the API and the app over the internet) | private-mesh (reachable only from mesh peers; CI joins the mesh first) -->
<!-- Ingress Mode: own-tls (the app's Ingress requests its own certificate from ClusterIssuer) | platform-default-tls (Ingress without tls; the platform's default certificate serves it) | platform-owned (the platform owns the Ingress; the app ships none) -->
<!-- Host: required unless platform-owned. Ingress Class: required unless platform-owned, empty for it. ClusterIssuer: own-tls only -->
<!-- Health Path: absolute path the verification requests on the app (e.g. /health.json) -->
<!-- Mesh Setup-Key Secret (GitHub secret name in this environment) and Mesh Wait-For Host (the k3s API's mesh name): private-mesh only -->

| Environment | Machine Slug | Host | Reachability | Ingress Mode | Ingress Class | ClusterIssuer | Health Path | Mesh Setup-Key Secret | Mesh Wait-For Host |
|-------------|--------------|------|--------------|--------------|---------------|---------------|-------------|-----------------------|--------------------|
| [ENVIRONMENT] | [MACHINE_SLUG] | [HOST] | [public/private-mesh] | [own-tls/platform-default-tls/platform-owned] | [INGRESS_CLASS] | [CLUSTER_ISSUER] | [HEALTH_PATH] | [MESH_SETUP_KEY_SECRET] | [MESH_WAIT_FOR_HOST] |

## Constraints

| Constraint | Description | Impact |
|------------|-------------|--------|
| Target Environment | [dev/staging/prod] | [Environment-specific considerations] |
| Geographic | [Region/data residency requirements] | [How it affects resource placement] |
| Budget | [Cost limits or optimization goals] | [How it constrains resource choices] |
| Platform Repo | [owner/repo hosting the k3s onboarding registry — `provisioning_mode: existing-k3s` only, omit otherwise] | [`smaqit.infrastructure-request-k3s-onboarding`'s PR must target this repo, never this project's own] |
| Registry File Path | [path within Platform Repo to the onboarding registry file — `existing-k3s` only, omit otherwise; sourced from the platform team's own onboarding documentation, never invented] | [`smaqit.infrastructure-request-k3s-onboarding` appends this project's entry to this exact path] |
| Mesh Join Action | [pinned `uses:` reference of the platform's mesh-join GitHub Action, e.g. "owner/repo@v1.2.3" — required when any environment is `private-mesh`, omit otherwise; sourced from the platform team, never invented] | [the generated deploy workflow joins the mesh with this action before any `kubectl` call] |
| Mesh Setup-Key Input | [that action's input name for the setup key — `private-mesh` only, omit otherwise; from the action's own documentation] | [the workflow passes the environment's setup-key secret through this input] |
| Mesh Wait-For Input | [that action's input name for the host to wait for — `private-mesh` only, omit otherwise; from the action's own documentation] | [the workflow passes the environment's Mesh Wait-For Host through this input] |
| Probe Client | [`curl` or `wget`, whichever the app's own image contains — required when any environment is `private-mesh` or `platform-owned`, omit otherwise] | [the in-cluster probe Job runs the app's image with this client against the Service's health path] |
| Namespace Quota | [the onboarded Namespace's `ResourceQuota`/`LimitRange` numbers — `existing-k3s` only, omit otherwise; sourced from the platform team's own onboarding documentation, never invented] | [application resource requests/limits declared elsewhere in this spec must fit within this ceiling] |
| Container Registry | [registry host and visibility, e.g. "ghcr.io, private" — `existing-k3s` only, omit otherwise; a per-project decision, not sourced from the platform team] | [`smaqit.infrastructure-cicd-generate`'s `build` job pushes here; if private, `smaqit.infrastructure-repo-config` syncs a pull credential and the manifest needs `imagePullSecrets`] |

## Acceptance Criteria

Requirements use format: `INF-[CONCEPT]-[NNN]`

- [ ] INF-[CONCEPT]-001: [Criterion — must be measurable, observable, unambiguous]
- [ ] INF-[CONCEPT]-002: [Criterion]

### Untestable Criteria

If any criterion cannot be automatically validated, flag it:

- [ ] INF-[CONCEPT]-NNN: [Criterion] *(untestable)*
  - **Reason:** [Why it cannot be tested automatically]
  - **Proposal:** [Measurable proxy or alternative approach]
  - **Resolution:** [How it will be verified]

---

*Generated with smaqit [SMAQIT_VERSION]*
