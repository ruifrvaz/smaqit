# How to Set Up the OpenStack CLI and Run Pre-flight Checks

Before running `terraform init`, you need the OpenStack CLI installed locally and your Cyso credentials loaded. This guide covers local tooling setup and how to verify that the values in `variables.tf` match what Cyso actually has in your project.

Runbook reference: [cyso-deployment-setup.md](./cyso-deployment-setup.md)

---

## Step 1 — Install the OpenStack CLI

The OpenStack CLI (`openstack`) is needed to verify cloud resources before running Terraform. It must not be installed into the system Python — use `pipx`, which isolates CLI tools in their own environments while keeping them globally accessible.

```bash
sudo apt install pipx
pipx ensurepath
exec $SHELL
pipx install python-openstackclient
openstack --version
```

> **Why pipx and not pip:** Ubuntu protects its system Python (PEP 668). `pip install` without `--break-system-packages` will fail. `pipx` creates a dedicated isolated environment per tool and symlinks the binary to `~/.local/bin` — no system conflicts, no activating venvs, no interference with other projects.

---

## Step 2 — Load Cyso credentials

Source your OpenRC file to export the OpenStack environment variables:

```bash
source deployment/configuration/openrc.sh
```

The script prompts for your Application Credential secret. After entering it, verify auth works:

```bash
openstack token issue
```

A successful response returns a table with a token ID, expiry, and project details. If you see an authentication error, the secret is wrong or the credential has been deleted.

> Your OpenRC file lives in `deployment/configuration/` which is gitignored. Never commit it.

---

## Step 3 — Verify Terraform variable defaults

`deployment/terraform/variables.tf` contains default values for the Ubuntu image ID, VM flavor, and external network name. These need to match what actually exists in your Cyso project. Run the following with your credentials loaded:

**Ubuntu 24.04 image ID:**

```bash
openstack image list --name "Ubuntu 24.04" --format table
```

Compare the `ID` column against the `image_id` default in `variables.tf`. Verified value: `fd91e198-f162-4b6b-a23e-123304fb408a`.

**VM flavor:**

```bash
openstack flavor list --format table
```

Confirm `s5.small` appears (2 vCPU / 8 GB RAM / 50 GB disk). Update `flavor_name` in `variables.tf` if needed.

**External network name:**

```bash
openstack network list --external --format table
```

The `Name` column value must match the `external_network` default in `variables.tf`. This is used to allocate the Floating IP. Update if the name is not `public`.

---

## Step 4 — Update variables.tf if needed

If any value from Step 3 differs from the defaults, update `deployment/terraform/variables.tf`:

```hcl
variable "image_id" {
  default = "<actual-id-from-openstack-image-list>"
}

variable "flavor_name" {
  default = "<actual-name-from-openstack-flavor-list>"
}

variable "external_network" {
  default = "<actual-name-from-openstack-network-list>"
}
```

Only update what differs. Do not change values that already match.
