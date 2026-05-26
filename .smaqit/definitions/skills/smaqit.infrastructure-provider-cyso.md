---
name: smaqit.infrastructure-provider-cyso
description: Knowledge router for Cyso Cloud. Tells an agent which reference file to read for any Cyso-related decision — pricing, authentication, deployment setup, or CLI pre-flight. Does not execute steps; surfaces the right documentation at the right moment to prevent guessing.
---

# smaqit.infrastructure-provider-cyso

## Steps

This skill has no procedural steps. It is a knowledge router: a set of load conditions that map agent decisions to the correct reference file. An agent invoking this skill must read the files indicated by the matching conditions before acting.

### Load conditions

| Condition | File to read |
|-----------|--------------|
| Selecting or comparing compute flavors, estimating cost, or choosing storage tier | `references/cyso-pricing.md` |
| Creating an Application Credential, downloading OpenRC, generating S3 keys, or any manual cloud portal action | `references/cyso-deployment-setup.md` |
| Looking up platform endpoints (auth URL, S3 endpoint), image IDs, network names, OpenStack resource IDs, or the Terraform provider block | `references/cyso-reference.md` |
| Installing the OpenStack CLI, sourcing credentials locally, or running pre-flight checks against `variables.tf` | `references/openstack-cli-setup-and-preflight.md` |
| Any uncertainty about Cyso Cloud — if the load condition is ambiguous | Read all four reference files before proceeding |
| Question not answered by the reference files | Fetch `https://cyso.cloud/docs/cloud/` and navigate to the relevant section |

## Output

No artifact produced. Output is the agent's informed reasoning derived from the referenced files.

## Scope

- **In scope:** Routing to the correct Cyso reference for any cloud decision.
- **Out of scope:** Executing Terraform commands, writing infrastructure code, provisioning resources, or resolving non-Cyso infrastructure questions.

## Completion criteria

- [ ] The correct reference file(s) have been read for the active decision.
- [ ] No Cyso-specific fact (endpoint, flavor name, credential format) has been assumed without reading the relevant reference.

## Failure handling

| Situation | Action |
|-----------|--------|
| A decision spans multiple load conditions | Read all implicated reference files before acting |
| A reference file is missing from `references/` | Fetch `https://cyso.cloud/docs/cloud/` as a fallback; report the missing file to the user |
| A fact in the reference files is stale (e.g., a price changed) | Fetch `https://cyso.cloud/docs/cloud/` to verify the current value; note that local references were last verified 2026-04-05 |

## Gotchas

- Cyso Cloud is OpenStack-based but branded as Cyso. The auth endpoint is `https://core.fuga.cloud:5000/v3` (Fuga Cloud, which powers Cyso). Do not substitute generic OpenStack defaults.
- Application Credentials and S3/Object Storage credentials are **separate** — they are created through different portal flows. Read `references/cyso-reference.md` to see both.
- The `openrc.sh` file embeds the application credential secret. It must never be committed to git. This constraint is documented in `references/cyso-deployment-setup.md`.
- The HIM Corporate production VM uses `s5.small` at €17.50/month (verified 2026-04-05). Pricing is hourly-billed, invoiced monthly.
- Reference files are copies of `docs/wiki/` content kept inside the skill's `references/` subdirectory for portability. If the wiki files are updated, the skill references must be updated too.

## Examples

**Example 1 — Flavor selection**
Input: "Which VM flavor should I use for a Node.js + SQLite backend?"
Action: Read `references/cyso-pricing.md`. Identify s5.small as the current production choice.
Output: Recommendation of `s5.small` with justification from the pricing table.

**Example 2 — Terraform provider credentials**
Input: "What auth_url and credential format does the Terraform OpenStack provider need?"
Action: Read `references/cyso-reference.md`. Load the provider block and environment variable list.
Output: The correct `provider "openstack"` HCL block with Cyso-specific values.

**Example 3 — First-time setup**
Input: "How do I configure my machine to run terraform init against Cyso?"
Action: Read `references/cyso-deployment-setup.md` and `references/openstack-cli-setup-and-preflight.md`.
Output: Step-by-step instructions covering app credential creation, OpenRC download, S3 bucket setup, and CLI pre-flight checks.

## Compatibility

- Scoped to the SPECtacular project. Reference files are located at `.github/skills/smaqit.infrastructure-provider-cyso/references/` relative to the workspace root.
- Reference content was originally sourced from `docs/wiki/` in the guida (HIM Corporate) project (verified 2026-04-05). Keep references in sync when Cyso facts change.
