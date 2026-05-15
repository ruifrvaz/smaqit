---
id: INF-[CONCEPT]
status: draft
created: [TIMESTAMP]
---

# [CONCEPT_NAME]

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

## Constraints

| Constraint | Description | Impact |
|------------|-------------|--------|
| Target Environment | [dev/staging/prod] | [Environment-specific considerations] |
| Geographic | [Region/data residency requirements] | [How it affects resource placement] |
| Budget | [Cost limits or optimization goals] | [How it constrains resource choices] |

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

*Generated with smaqit v0.8.2-beta*
