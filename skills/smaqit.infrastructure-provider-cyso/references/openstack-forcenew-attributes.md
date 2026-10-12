# OpenStack Provider — `ForceNew` Attributes

`terraform-provider-openstack/openstack` resource attributes that destroy and recreate the resource
on change, rather than updating it in place.

## `openstack_compute_instance_v2`

| Attribute | ForceNew? |
|---|---|
| `user_data` | Yes — cloud-init is only injected at first boot; the provider has no in-place update path. Any diff (including comment-only or whitespace-only) is a destroy+recreate. |
| `image_id` / `image_name` | Yes |
| `key_pair` | Yes — injected via the metadata service at boot, same as `user_data` |
| `flavor_name` / `flavor_id` | No — resizable in place |
| `security_groups` | No — updatable via the Neutron/Nova API |
| `network` blocks | Depends — adding a block is safe; removing/reordering can force replacement depending on provider version |

**Rule:** treat `user_data`, `image_id`, `image_name`, and `key_pair` as destroy-triggers. Never edit
them as part of a routine/day-2 config change. If day-2 configuration needs to change (e.g. adding a
package, editing a `write_files` block), that change belongs in a separate, re-runnable provisioning
mechanism (a script executed over SSH by `smaqit.infrastructure-vm-bootstrap` or
`smaqit.infrastructure-deploy-rsync`), not in `user_data` itself.

If `user_data` is loaded via `file("${path.module}/cloud-init.yaml")` rather than an inline heredoc,
the ENTIRE file's content becomes the hashed `user_data` string — including any comments in that
file. Keep explanatory comments about the ForceNew risk in the `.tf` file itself (as an HCL comment
above the resource/attribute), not inside the referenced cloud-init file, or an edit meant only to
add documentation will itself force a replacement.

To verify ForceNew status for an attribute not listed above, don't guess from documentation prose —
check the pinned provider's actual schema:

```bash
terraform providers schema -json | jq '.provider_schemas | to_entries[] | select(.key | test("openstack")) | .value.resource_schemas.openstack_compute_instance_v2.block.attributes'
```

Attributes with `"force_new": true` in that output are ForceNew.

Regardless of what this document says, `smaqit.infrastructure-provision-cyso/scripts/plan-guard.sh`
is the actual enforcement point — it inspects the real `terraform plan` output and blocks the apply
if any resource shows a `delete` action.

## `ignore_changes = [user_data]` and deliberate replacement

Repositories that embed first-boot scripts in `user_data` commonly add
`lifecycle { ignore_changes = [user_data] }` to the instance, so an edit to those scripts can never
destroy the machine through an incidental diff. The trade-off: a content change is now invisible to
`plan` and `apply` — it takes effect only on an instance created after the change. Day-2 changes go
through a re-runnable mechanism (scripts over SSH, ideally the same scripts first boot uses). When
the instance itself must be rebuilt so new first-boot content applies, replace it explicitly and
through the guard, never by removing the `ignore_changes` entry:

```bash
plan-guard.sh <terraform-dir> --allow-replace <instance-address> --out tfplan
terraform apply tfplan
```

## Nova's `user_data` size limit

Nova rejects `user_data` larger than **65,535 bytes after base64 encoding**. The failure is hard
to read: the provider's error quotes the `user_data` value, which Terraform treats as sensitive,
so the plan or apply shows an empty error box. Embedding several host scripts reaches the limit
quickly.

- **Fail at plan time** with a precondition on the instance:

  ```hcl
  lifecycle {
    precondition {
      condition     = length(base64encode(var.user_data)) <= 65535
      error_message = "Rendered user_data exceeds Nova's 65535-byte limit (base64-encoded). Shrink the embedded scripts."
    }
  }
  ```

- **Shrink embedded files** by gzipping them: render each with `base64gzip(file(...))` and mark the
  cloud-init `write_files` entry `encoding: gz+b64`, so cloud-init decompresses it on the host.
- Any change to the rendered content is still a `user_data` change — the ForceNew rule above
  applies unless `ignore_changes = [user_data]` is set.

## Sources

| Topic | URL |
|---|---|
| `openstack_compute_instance_v2` resource schema | https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/compute_instance_v2 |
| Terraform `ForceNew` concept | https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew |
