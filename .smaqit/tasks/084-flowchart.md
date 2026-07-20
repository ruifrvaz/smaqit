# Deploy Target Resolution — Decision Flow (diagram)

Companion diagram for [`084_deploy_target_resolution_provisioning_branch.md`](084_deploy_target_resolution_provisioning_branch.md).
An interactive rendered version also exists at
https://claude.ai/code/artifact/a035b9bd-b1e4-4750-a2c7-4a9326b4f235 — this file is the durable,
in-repo source of truth; the link is a nicer-to-read copy, not a dependency.

```mermaid
flowchart TD
    Start(["New deploy request"]) --> Q1{"VM already exists?"}

    Q1 -->|"no"| D1flow["<b>Full provisioning</b><br/>vault-loader: all 4 paths<br/>provision-cyso: terraform apply<br/>vm-bootstrap<br/>cicd-generate: provision+deploy<br/>repo-config: all 4 paths<br/>deploy-rsync &middot; first site &rarr; default_server"]

    Q1 -->|"yes"| Q2{"Owned by THIS<br/>project's Terraform state?"}

    Q2 -->|"yes, redeploying own app"| D2flow["<b>Idempotent re-apply</b><br/>vault-loader: paths already populated<br/>provision-cyso: plan-guard.sh &rarr; no-op<br/>cicd-generate: unchanged<br/>deploy-rsync &middot; same VM"]

    Q2 -->|"no, different project"| Q3{"Does this project need<br/>its own Terraform state?"}

    Q3 -->|"yes"| D3stop["<b>Out of scope</b><br/>Two Terraform states targeting<br/>one VM &mdash; not solved here"]

    Q3 -->|"no &mdash; recommended"| Q4{"SSH access to<br/>the shared VM?"}

    Q4 -->|"reuse owner's key"| B1["Copy private key into<br/>this project's Vault namespace"]
    Q4 -->|"generate own key"| B2["Generate keypair,<br/>append pubkey to<br/>authorized_keys manually"]

    B1 --> D4flow["vault-loader: ssh + github ONLY<br/>SKIP provision-cyso entirely<br/>gh variable set VM_HOST manually<br/>cicd-generate: deploy job ONLY<br/>repo-config: ssh + github ONLY"]
    B2 --> D4flow

    D4flow --> Q5{"First site<br/>on this VM?"}

    Q5 -->|"yes"| B3["nginx vhost:<br/>default_server"]
    Q5 -->|"no, co-hosted"| B4["nginx vhost:<br/>name-based only,<br/>never default_server"]

    B3 --> Verify["deploy-rsync &rarr; deploy-verify"]
    B4 --> Verify

    D1flow --> VerifyA["deploy-verify"]
    D2flow --> VerifyB["deploy-verify"]

    classDef question fill:#f2e3d5,stroke:#b5651d,stroke-width:1.5px,color:#1c232d;
    classDef flow fill:#ffffff,stroke:#d8dde3,stroke-width:1px,color:#1c232d,text-align:left;
    classDef stop fill:#f5e4e2,stroke:#9c3b3b,stroke-width:1.5px,color:#1c232d;
    classDef branch fill:#eef1f4,stroke:#9aa7b6,stroke-width:1px,color:#1c232d;
    classDef term fill:#e3efe8,stroke:#2f6f4f,stroke-width:1.5px,color:#1c232d;

    class Q1,Q2,Q3,Q4,Q5 question;
    class D1flow,D2flow,D4flow flow;
    class D3stop stop;
    class B1,B2,B3,B4 branch;
    class Verify,VerifyA,VerifyB term;
```

## Legend

- **Diamond** — decision point
- **White box** — sequence of skill invocations for that path
- **Grey box** — sub-branch / mechanism choice
- **Green box** — terminal, verified working
- **Red box** — explicitly out of scope for this task

## The five decisions, in order

1. **Does the target VM already exist?** The only question the current flow asks correctly — but
   implicitly, by whether `smaqit.infrastructure-provision-cyso` happens to be invoked, not as an
   explicit gate anything else can read.
2. **Owned by this project's own Terraform state?** Redeploying to your own VM and deploying onto a
   *different* project's VM both look like "VM exists" from the outside but need entirely different
   handling. The "yes" path (idempotent re-apply via `plan-guard.sh`'s no-op) already works today;
   the "no" path is the entire unhandled case.
3. **Does the deploying project need its own Terraform state?** Two Terraform states with opinions
   about the same VM is a coordination problem this task doesn't try to solve — recommended answer
   is always no for a co-hosted app; "yes" is flagged out of scope.
4. **How does this project get SSH access to a VM it doesn't own?** `vault-loader`'s SSH step
   auto-generates a fresh keypair whenever one is missing — correct default for a new VM, silently
   wrong here, since a brand-new key was never added to the shared VM's `authorized_keys`. Two valid
   mechanisms, not one automated default: reuse the owning project's key material, or generate a new
   one and manually authorize it once.
5. **Is this the first site on the VM, or a co-host?** The one decision already designed for
   correctly elsewhere (nginx `default_server` vs. name-based vhost) — included here because it's
   the payoff of getting decisions 1–4 right.
