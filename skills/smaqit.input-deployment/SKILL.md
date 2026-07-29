---
name: smaqit.input-deployment
description: Validate and elicit deployment phase parameters from session context before deployment begins. Invoke automatically before starting the Deployment phase to confirm or default execution preferences.
metadata:
  version: "1.2.0"
---

# Deployment Input

Validate session context for deployment execution preferences before the Deployment phase begins. All parameters have defaults — the phase can always proceed. Only elicit when a parameter is explicitly referenced but unspecified, or when ambiguity would cause incorrect execution.

## Steps

1. **Extract from session context** — Scan for any explicit deployment target overrides, verification preferences, or output requirements
2. **Apply defaults** — Use standard defaults for all unspecified parameters (listed below)
3. **Elicit only when blocking** — Ask only if a parameter is referenced but missing, or if the deployment target cannot be resolved from Infrastructure specs
4. **Proceed** — Begin deployment with confirmed or defaulted parameters

## Parameters

### Deployment Target
**Default:** Target environment as defined in Infrastructure specs  
**Elicit when:** Multiple deployment targets exist and the active target is ambiguous  
**Question:** "Any target-specific preferences? (e.g., deploy to staging first, use blue-green strategy)"

### Verification
**Default:** Run standard post-deployment health checks as defined in Coverage specs  
**Elicit when:** User has indicated they want to skip or extend verification  
**Question:** "How should deployment be verified? (run smoke tests, skip health checks, full suite)"

### Output Preferences
**Default:** Standard output; deployment steps and results shown  
**Elicit when:** Sensitive environment where log scrubbing is needed  
**Question:** "How should output be displayed? (verbose, scrub sensitive data from logs)"

### Provisioning Mode
**Default:** `provision`  
**Elicit when:** Session context mentions co-hosting, an existing/shared VM, a dedicated VM that will never be Terraform-managed, or names infrastructure another project already owns, and the mode cannot be resolved unambiguously from that context  
**Question:** "Is this a new VM, a redeploy of your own existing VM, a VM another project already manages, or a dedicated VM that's provisioned out-of-band and will never be Terraform-managed?"

**Values:**
- `provision` — this project provisions its own new VM via Terraform. Unchanged, current behavior. Default whenever nothing in session context suggests otherwise.
- `existing-owned` — redeploying to a VM this project's own Terraform state already manages. Same skill sequence as `provision`; `terraform apply` is expected to no-op.
- `existing-shared` — targeting a VM a *different* project owns and manages via its own Terraform state (e.g., co-hosting a second app on a VM another project provisioned). Skips Terraform provisioning for this project entirely; downstream skills (`smaqit.infrastructure-vault-loader`, `smaqit.infrastructure-provision-cyso`, `smaqit.infrastructure-repo-config`, `smaqit.infrastructure-cicd-generate`) branch on this value — see `smaqit.new-greenfield-project` Phase 4/5.
- `existing-unmanaged` — a VM dedicated to this project (not co-hosted with anything else) but never managed by any Terraform state — provisioned out-of-band (manually, or via a provider's own tooling) and staying that way by design. Skips Terraform entirely, exactly like `existing-shared`, but is not a co-hosting scenario: the VM is this project's alone, just outside Terraform's view. The same downstream skills branch on this value — see `smaqit.new-greenfield-project` Phase 4/5 for the two points where it diverges from `existing-shared` (machine registration typically starts fresh rather than assuming an already-registered machine; the nginx vhost gets `default_server` since there's no co-hosting to guard against).

Surface the resolved `provisioning_mode` value alongside the Deployment Target so downstream phases can read it from session context without re-eliciting.

## Readiness Condition

No required sections — defaults are always valid. Proceed immediately unless the deployment target cannot be determined from Infrastructure specs.
