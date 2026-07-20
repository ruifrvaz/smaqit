# GitHub Provider (`integrations/github`) — Import ID Formats

`terraform import` ID formats for `integrations/github` provider resources, used alongside Cyso in
this workflow (e.g. publishing the provisioned VM's IP as a `github_actions_variable`).

## `github_actions_variable`

**Import ID format:** `<repository>:<variable_name>` (colon-separated, exactly two parts — not
`<owner>/<repo>`).

```bash
terraform import github_actions_variable.vm_host <repository>:VM_HOST
```

`<repository>` is the repo name only — the provider is already scoped to a single owner via
`provider "github" { owner = "..." }`.

## `github_actions_secret`

Same pattern: `<repository>:<secret_name>`.

## General pattern

Most `integrations/github` resources scoped to a single repository use
`<repository>:<sub-resource-name>` as the import ID (owner excluded). Verify per-resource before
assuming — check the resource's import documentation page (linked below).

## Resource ownership — pick one, never both

A resource declared in Terraform whose real-world object is also mutated out-of-band (e.g.
`gh variable set` run manually while `github_actions_variable.vm_host` is still declared in
`main.tf`) produces permanent state drift: every `apply` that doesn't know about the manual value
tries to create it and gets a 409. Resolve by choosing one:

1. **Fully Terraform-managed** — no manual `gh variable set`/`gh secret set` for that name. If drift
   occurs, `terraform import` it once with the format above, then let Terraform own all future
   updates.
2. **Fully operator-managed** — remove the `resource` block from `main.tf`; document in the relevant
   workflow that the value is set manually and Terraform must never declare it.

## Sources

| Topic | URL |
|---|---|
| `github_actions_variable` resource + import docs | https://registry.terraform.io/providers/integrations/github/latest/docs/resources/actions_variable |
| `github_actions_secret` resource + import docs | https://registry.terraform.io/providers/integrations/github/latest/docs/resources/actions_secret |
