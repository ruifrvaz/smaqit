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

## Sources

| Topic | URL |
|---|---|
| `openstack_compute_instance_v2` resource schema | https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs/resources/compute_instance_v2 |
| Terraform `ForceNew` concept | https://developer.hashicorp.com/terraform/plugin/sdkv2/schemas/schema-behaviors#forcenew |
