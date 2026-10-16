# Platform Repositories

A **platform repository** owns machines rather than an application. It holds the cloud resources (instances, volumes, security groups, keypairs), each machine's Terraform state, the host baseline every machine runs, and the contract through which applications become tenants. Application repositories then deploy onto those machines without owning any of them.

This page describes how such a repository fits smaqit, the shape the pattern has taken in real use, and the failures that only showed up when it ran against live machines. It is guidance, not a mechanism: smaqit does not ship the platform's provisioning, baseline, or tenancy skills, for the reason below.

## Why smaqit Doesn't Ship the Mechanisms

The mechanisms a platform repository runs — how a machine is composed, what its baseline installs, how a tenant is admitted — are that repository's own decisions, and they change as the platform evolves. In the first platform repository built this way, the tenancy contract was replaced wholesale within two weeks (from per-tenant Unix users and nginx sites to Kubernetes Namespaces), and the host baseline was rewritten around a different runtime and a private mesh network. A canonical smaqit skill copied from either version would have been wrong for the next one.

So the platform repository keeps its mechanisms as its own project skills and connects them to smaqit through [deployment slots](deployment-slots.md): smaqit defines the hand-off points and the invariants, the repository fills them. This is the same split as the k3s onboarding skills, where smaqit ships the application side (`smaqit.infrastructure-request-k3s-onboarding`, `smaqit.infrastructure-deploy-k3s-app`) and the platform owns the receiving side. What the platform tells its tenants about each machine — whether it is public or reachable only over its mesh, and whether the platform serves TLS or owns the Ingress itself — goes into each app's own Infrastructure spec as an `## Environments` row (see [existing-k3s targets](deployment-slots.md#existing-k3s-targets)), so the platform's documentation should state those facts per machine.

## Binding the Slots

A platform repository runs the deployment phase like any project, with its own `.smaqit/slots.yml`:

| Slot | Filler |
|---|---|
| `target-credentials` | shipped: `smaqit.infrastructure-vault-loader`, operation `register` — first-time population of `secret/machines/<machine-slug>/*` |
| `environment-config` | shipped: `smaqit.infrastructure-repo-config`, operation `sync-machine` — one GitHub Environment per machine |
| `provision` | the repository's own skill (composition, gated apply) |
| `host-bootstrap` | the repository's own baseline skill |
| `verify` | the repository's own machine checks (see the checklist below) |
| `replace` (day two) | the repository's own replacement procedure, run as a single-slot run |
| `app-deploy`, `tls` | usually `none` — a machine has no application of its own |

Its skills declare their slots in frontmatter (`metadata.smaqit-slots: "provision=apply, replace=replace"`) and live in the repository's own skill directories, which are searched before the global ones. Every filler runs under the Deployment agent's invariants: guarded IaC plan/apply, credential references only, no out-of-band changes to IaC-managed resources.

## The Shape of the Pattern

### Layout

- **Shared modules, per-machine roots.** Machine-agnostic modules (compute instance, data volume, security group, often a wrapper that composes them) and one Terraform root per machine that does little more than call the wrapper with that machine's values. Adding a machine is copying a root, not editing shared code.
- **Volume attachment at the composition level.** First-boot content usually needs the data volume's ID to mount it by its stable link, so the volume is created first and attached alongside the instance rather than inside the instance module.
- **State per machine.** Each root has its own state key; production machines may get a dedicated bucket. See `smaqit.infrastructure-provider-cyso` for the backend block proven against Cyso Object Storage.
- **One GitHub Environment per machine**, named after the machine, holding that machine's credentials. A machine-class file (for example, public versus private-network-only) is cross-checked against the Environment and the Terraform outputs before anything runs, so a mismatched configuration fails loudly instead of targeting the wrong place.

### Governance

- **Pull requests plan; only a gated dispatch applies.** A production Environment with a main-only branch policy rejects jobs from pull-request branches, so the PR check there is credential-free (`fmt -check`, `init -backend=false`, `validate`); a test machine without that policy can run a real plan.
- **Apply requires an exact confirmation input** — the dispatch fails unless a `confirm_machine` field equals the machine's name — plus a human approval on the Environment.
- **Every plan goes through the guard.** `plan-guard.sh` (from `smaqit.infrastructure-provision-cyso`) blocks any planned delete. A deliberate replacement names its address: `plan-guard.sh <dir> --allow-replace <address> --out tfplan`, then `terraform apply tfplan`. The guard passes only that replacement and keeps the exact plan it inspected.
- **Addresses and secrets flow from outputs and Vault, not hand-maintained variables.** Workflows read a machine's IP and volume ID from its Terraform outputs; credentials come from Vault through the operator's own scripts, synced into the machine's Environment.

### Day Two Runs Through Workflows

First-boot content (cloud-init) runs once, so fixing a script in Terraform never heals a machine that already exists. The working rule is **no ad-hoc remediation**: fix the provisioning source, then run the change through a workflow.

- A host-baseline workflow takes the machine as a choice input, reads its targets from Terraform outputs, and streams each baseline script to the host in order (`ssh … sudo bash -s -- <args> < script`).
- First boot and day two run **the same scripts** — cloud-init embeds them, the workflow streams them — so there is one source for what a baselined host looks like.
- Every script is idempotent (see the gotchas), so re-running the baseline is the remediation.

## Verifying a Machine

`smaqit.infrastructure-deploy-verify` checks an application (health endpoint, revision, page). A machine needs its own checks, and which ones apply depends on its class, so this is a checklist for the repository's `verify` filler rather than a shipped skill:

- **Reachable the intended way.** The management path works (SSH over the mesh, or a break-glass address), and the paths that must be closed are closed — test them against a known-open control so a broken network can't pass as "closed".
- **First boot finished.** `cloud-init status` reports done, with no errors in its log.
- **Data volume.** Mounted at the expected path (`findmnt`), its `fstab` entry pinned by UUID or stable by-id link, and a write test succeeds.
- **Services.** The expected services are active and answer locally.
- **Port surface from the cloud API**, not from Terraform state: the security group lists exactly the declared ingress rules and nothing else. Checking state would only prove Terraform agrees with itself.
- **Reboot survival.** Reboot, wait for it to return, and re-check the mount and services.
- **No drift.** A plan of the machine's root reports no changes.

## Gotchas Found Only on Live Machines

None of these were caught by `terraform validate`, YAML parsing, or code review; each cost a failed run.

- **ssh joins its arguments into one string.** `ssh host cmd arg1 arg2` sends `cmd arg1 arg2` to the remote shell, which splits it again — a multi-word argument silently becomes several, and backslashes are stripped. Pass the whole remote command as **one** argument, quoting every value inside it, or stream a script on stdin.
- **ssh in a loop reads the loop's input.** Inside `while read …; do ssh …; done`, ssh consumes the rest of the list from stdin. Give every such ssh `</dev/null`.
- **Secrets travel on stdin, never in arguments.** Arguments are visible in the process list on both ends. Prefix the streamed script with `export NAME=<quoted value>` lines and pipe both into `sudo bash -s`.
- **Idempotency is a filesystem property.** Guard every mutation (check, then act); an unconditional `touch` or `chmod` changes mtimes on every run, so "nothing changed" stops being true even when the content is identical. Verify by re-running and confirming mtimes didn't move.
- **Long remote steps need keepalives.** A quiet wait (pulling images, waiting for a rollout) can drop the connection with "Broken pipe"; use `-o ServerAliveInterval=30 -o ServerAliveCountMax=10`.
- **`user_data` has a size limit.** Nova rejects more than 65,535 bytes base64-encoded, and the error shows as an empty box. Gzip embedded scripts and guard the size at plan time (see `smaqit.infrastructure-provider-cyso`).
- **`ignore_changes = [user_data]` hides content changes.** It protects the machine from incidental replacement, but a new first-boot script then applies only to a new instance. Rebuild deliberately with a guarded `--allow-replace`.
- **Host keys change on replacement.** A new instance on a reused address presents a new host key. Ephemeral CI runners can use `StrictHostKeyChecking=accept-new`; on a workstation, remove the stale entry with `ssh-keygen -R <host>` rather than treating the warning as an attack — after confirming the replacement was yours.
- **Optional variables may go unused.** `smaqit.infrastructure-repo-config`'s `sync-machine` also sets a `FIXED_IP` variable; a repository whose workflows read the address from Terraform outputs can ignore it.
