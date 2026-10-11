# Cyso Cloud Reference

Authoritative reference for Cyso Cloud (OpenStack, region `ams2`) infrastructure facts, applicable
to any project deployed there.
All facts verified against live documentation as of **2026-04-05**. Volume attachment, volume
types, retype, image ID, and GitHub token permissions were corrected from live provisioning on
**2026-08-24**.

---

## Platform overview

| Property | Value |
|---|---|
| Platform | OpenStack (Cyso-managed) |
| Primary region | `ams2` (Amsterdam) |
| Secondary region | `fra` (Frankfurt) |
| Dashboard | https://my.cyso.cloud |
| Auth endpoint (Keystone) | `https://core.fuga.cloud:5000/v3` |
| Object Storage S3 endpoint | `https://core.fuga.cloud:8080` |
| Documentation | https://cyso.cloud/docs/cloud/ |

---

## Authentication and credentials

Cyso Cloud uses **Application Credentials** for all automation. These are OpenStack `v3applicationcredential` tokens, scoped to a project, with no personal password embedded.

### Creating application credentials

1. Log in at https://my.cyso.cloud
2. Navigate to **Access → Credentials**
3. Click **Create** — choose a name (e.g. `<project-slug>-terraform`), optionally set an expiry
4. **Copy the ID and Secret immediately** — the secret is shown only once
5. Download the credential in one of four formats (see below)

> The Object Storage (S3) credentials are **separate**. Go to **Access → Credentials → EC2/S3 credentials tab** and create an S3-specific key pair there.

### Credential download formats

| Format | Use |
|---|---|
| **OpenRC** (`openrc.sh`) | OpenStack CLI, shell scripts (`source openrc.sh`) |
| **clouds.yaml** | OpenStack SDK, multi-environment CLI |
| **Terraform** | `provider "openstack"` block, pre-filled |
| **Dockerfile** | CI/CD containers with `fugacloud/openstackcli` base image |

### OpenRC environment variables

After `source openrc.sh`, the following vars are available in the shell:

```bash
OS_AUTH_URL="https://core.fuga.cloud:5000/v3"
OS_APPLICATION_CREDENTIAL_ID="<id>"
OS_APPLICATION_CREDENTIAL_SECRET="<secret>"
OS_REGION_NAME="ams2"
OS_AUTH_TYPE="v3applicationcredential"
OS_IDENTITY_API_VERSION=3
```

### Terraform provider block (from dashboard download)

```hcl
provider "openstack" {
  auth_url                      = "https://core.fuga.cloud:5000/v3"
  region                        = "ams2"
  application_credential_id     = "<id>"
  application_credential_secret = "<secret>"
}
```

Credentials are passed via `TF_VAR_*` environment variables — never hardcoded. The `auth_url` is hardcoded in `main.tf` as it is a public platform constant.

### clouds.yaml format

```yaml
clouds:
  fuga:
    auth:
      auth_url: "https://core.fuga.cloud:5000/v3"
      application_credential_id: "<id>"
      application_credential_secret: "<secret>"
    region_name: "ams2"
    auth_type: v3applicationcredential
    identity_api_version: 3
```

Place at `~/.config/openstack/clouds.yaml`. Use with: `openstack --os-cloud fuga server list`

---

## Service catalog

All services available in regions `ams2` (AMS) and `fra` (FRA). Project-specific endpoint URLs are shown at https://my.cyso.cloud/account/api-endpoints.

| Service | OpenStack component | Typical use |
|---|---|---|
| `compute` | Nova | ✓ VM provisioning |
| `volumev3` | Cinder | ✓ Boot volume + data volume |
| `network` | Neutron | ✓ Floating IP, security groups |
| `identity` | Keystone | ✓ Auth (`core.fuga.cloud:5000/v3`) |
| `image` | Glance | ✓ Ubuntu 24.04 image |
| `object-store` | Ceph (Swift) | ✓ Terraform remote state |
| `s3` | Ceph (S3 API) | ✓ Terraform S3 backend (`core.fuga.cloud:8080`) |
| `load-balancer` | Octavia | — deferred post-MVP |
| `dns` | Designate | — operator managed externally |
| `orchestration` | Heat | — not used |

---

## Compute

### Confirmed images

| Image | ID | Source |
|---|---|---|
| **Ubuntu 24.04 LTS** | `fd91e198-f162-4b6b-a23e-123304fb408a` | Verified via `openstack image list` (2026-04-05) |

Verify before use:
```bash
openstack image list --name "Ubuntu 24.04" --format table
```

### Flavor families

Three families. Naming scheme: `{family}.{size}`.

#### Standard (s5) — validated production choice

| Flavor | vCPU | RAM | Disk | Monthly |
|---|---|---|---|---|
| **s5.small** | 2 | 8 GB | 50 GB | € 17.50 |
| s5.medium | 4 | 16 GB | 100 GB | € 37.50 |
| s5.large | 8 | 32 GB | 200 GB | € 75.00 |

**`s5.small` is a proven baseline** — sufficient for a Node.js backend + SQLite at MVP scale.

#### CPU Optimised (c5)

| Flavor | vCPU | RAM | Disk | Monthly |
|---|---|---|---|---|
| c5.small | 2 | 4 GB | 25 GB | € 22.00 |
| c5.medium | 4 | 8 GB | 50 GB | € 43.75 |

#### Memory Optimised (m5)

| Flavor | vCPU | RAM | Disk | Monthly |
|---|---|---|---|---|
| m5.small | 2 | 16 GB | 100 GB | € 26.75 |
| m5.medium | 4 | 32 GB | 200 GB | € 53.50 |

Verify flavor names:
```bash
openstack flavor list --format table | grep -i s5
```

### Networks

| Network name | Protocol | Use |
|---|---|---|
| `public` | IPv4 | Use for `network { name = "public" }` in Terraform |
| `public6` | IPv6 | Alternative for IPv6 instances |

The instance network block in Terraform:
```hcl
network {
  name = "public"
}
```

### SSH access

Ubuntu 24.04 instances use `ubuntu` as the default user:
```bash
ssh -i ~/.ssh/<project-slug>_key ubuntu@<FLOATING_IP>
```

### Dashboard sidebar structure

```
Compute
EMK - Kubernetes
Storage
  Object Store
  Volume Store
  Images
  Snapshots
Network
  Network topology
  Networks
  Ports
  Routers
  IP addresses
  Security Groups
  Load Balancers
Service
  DNS
Access
  Key Pairs
  Credentials
  API Endpoints
```

---

## Storage — Cinder volumes

### Volume types

| Type (`volume_type` value) | IOPS/GB | Max IOPS | Multi-attach | Price |
|---|---|---|---|---|
| **`tier-1`** | 5 | 25k | No | € 0.095 /GB/month |
| `tier-1m` | 5 | 25k | Yes | € 0.145 /GB/month |
| `tier-2` | 25 | 25k | No | € 0.250 /GB/month |
| `tier-2m` | 25 | 25k | Yes | € 0.300 /GB/month |

The selectable names are lowercase exactly as above (`openstack volume type list`).

**Always set `volume_type` explicitly.** A volume created without one resolves to **`tier-2`**,
about 2.6× the cost of `tier-1`, not to the cheaper type. `tier-1` is a proven baseline
(10 GB ≈ € 0.95/month) for a SQLite data volume, but only when requested.

### Changing the type of an existing volume (retype)

Retyping an **attached** volume is refused (HTTP 400, even though the error text suggests an
in-use volume is acceptable). Retype it detached:

1. On the host: stop services using the mount, then `sudo umount /data`.
2. Detach: `openstack server remove volume <server> <volume>`.
3. Retype: `openstack volume set --type tier-1 <volume>`. If the two types live on different
   storage backends and the retype is refused, add `--retype-policy on-demand` to allow migration.
4. Reattach: `openstack server add volume <server> <volume>`.
5. On the host: `sudo mount -a` and confirm the mount.

When Terraform manages the volume, set `volume_type` in its configuration to the new value
afterwards, so the next plan shows no diff.

### Volume attachment

Cinder volumes attach over **virtio-scsi**, not virtio-blk: the first data volume appears as
`/dev/sdb`, **not** `/dev/vdb`. Kernel names can change between boots, so address the volume by
its stable udev symlink, which embeds the full volume ID:

- `/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_<full-volume-id>`
- a `scsi-S…` variant of the same link

The virtio-blk convention (`/dev/disk/by-id/virtio-<first-20-chars-of-id>`) does **not** exist on
Cyso hosts; a first-boot script written against it fails.

Format and mount on first attach, by ID:

```bash
DEV=/dev/disk/by-id/scsi-0QEMU_QEMU_HARDDISK_<full-volume-id>
sudo blkid "$DEV" || sudo mkfs.ext4 "$DEV"   # format only if empty
sudo mkdir -p /data
echo "$DEV /data ext4 defaults,nofail 0 2" | sudo tee -a /etc/fstab
sudo mount /data
```

The `nofail` option prevents boot failure if the volume is temporarily unavailable. A first-boot
script should fail loudly when the by-id link is absent rather than fall back to a kernel name.

### Terraform Cinder resources

```hcl
# Boot volume (created inline in block_device)
block_device {
  uuid                  = "<ubuntu-24.04-image-id>"
  source_type           = "image"
  volume_size           = 20
  boot_index            = 0
  destination_type      = "volume"
  delete_on_termination = true
}

# Separate data volume — always set volume_type (the default is tier-2)
resource "openstack_blockstorage_volume_v3" "data" {
  name        = "<project-slug>-data"
  size        = 10
  volume_type = "tier-1"
}

resource "openstack_compute_volume_attach_v2" "data" {
  instance_id = openstack_compute_instance_v2.app.id
  volume_id   = openstack_blockstorage_volume_v3.data.id
}
```

---

## Storage — Object Storage (S3-compatible)

### Endpoint

```
https://core.fuga.cloud:8080
```

S3-compatible (Ceph). Used as Terraform remote state backend and general blob storage.

- **Ingress**: free
- **Egress**: € 0.055 /GB
- **Price**: € 0.055 /GB/month

### Object Storage credentials

Object Storage uses **EC2/S3 credentials**, which are separate from application credentials:

1. Dashboard → **Access → Credentials → EC2/S3 credentials** tab
2. Click **EC2/S3 credential** button → select region → name it
3. Note the access key and secret key — used as `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`

### Creating a container (bucket)

1. Dashboard → **Storage → Object Store**
2. Click **Create a container** → give it a name
3. Set visibility to **Private** for Terraform state

### Terraform S3 backend configuration

```hcl
terraform {
  backend "s3" {
    bucket   = "<project-slug>-tfstate"
    key      = "<project-slug>/terraform.tfstate"
    region   = "ams2"
    endpoint = "https://core.fuga.cloud:8080"

    skip_credentials_validation = true
    skip_metadata_api_check     = true
    skip_region_validation      = true
    force_path_style            = true
  }
}
```

Credentials passed at runtime:
```bash
export AWS_ACCESS_KEY_ID=<object-storage-access-key>
export AWS_SECRET_ACCESS_KEY=<object-storage-secret-key>
terraform init
```

The bucket must be created manually in the dashboard before `terraform init` — Terraform cannot create the bucket it uses for state.

---

## Networking

### Floating IPs

- Cost: **€ 1.99/month** per IPv4 Floating IP
- IPv6 `/48`: free
- Allocated from the `public` pool

Terraform:
```hcl
resource "openstack_networking_floatingip_v2" "app" {
  pool = "public"
}

resource "openstack_compute_floatingip_associate_v2" "app" {
  floating_ip = openstack_networking_floatingip_v2.app.address
  instance_id = openstack_compute_instance_v2.app.id
}
```

### Security groups

Security groups are virtual firewalls. Instances have no external access unless rules explicitly allow it. The `default` security group exists per project — Cyso recommends creating project-specific groups alongside `default`, not modifying `default` directly.

To allow SSH and ICMP (ping):

| Rule | Protocol | Port | CIDR |
|---|---|---|---|
| SSH | TCP | 22 | `0.0.0.0/0` |
| ICMP | ALL ICMP | — | `0.0.0.0/0` |
| HTTP | TCP | 80 | `0.0.0.0/0` |
| HTTPS | TCP | 443 | `0.0.0.0/0` |

Terraform security group resources:
```hcl
resource "openstack_networking_secgroup_v2" "app" {
  name        = "<project-slug>-sg"
  description = "Allow SSH, HTTP, HTTPS"
}

resource "openstack_networking_secgroup_rule_v2" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.app.id
}
```

---

## Terraform patterns

### Provider block

```hcl
terraform {
  required_providers {
    openstack = {
      source = "terraform-provider-openstack/openstack"
    }
  }
}

provider "openstack" {
  auth_url                      = "https://core.fuga.cloud:5000/v3"
  region                        = "ams2"
  application_credential_id     = var.app_credential_id
  application_credential_secret = var.app_credential_secret
}
```

The provider reads credentials from environment variables if not set in the block. OpenRC variables (`OS_APPLICATION_CREDENTIAL_ID`, etc.) are honoured automatically.

### Key pair resource

```hcl
resource "openstack_compute_keypair_v2" "app" {
  name       = "<project-slug>-key"
  public_key = var.ssh_public_key
}
```

Key pairs can also be uploaded via the dashboard: **Access → Key Pairs → Import Key Pair**.

### Instance resource

```hcl
resource "openstack_compute_instance_v2" "app" {
  name            = "<project-slug>"
  flavor_name     = "s5.small"
  key_pair        = openstack_compute_keypair_v2.app.name
  security_groups = [openstack_networking_secgroup_v2.app.name]

  block_device {
    uuid                  = "fd91e198-f162-4b6b-a23e-123304fb408a"  # Ubuntu 24.04 LTS (see Confirmed images)
    source_type           = "image"
    volume_size           = 20
    boot_index            = 0
    destination_type      = "volume"
    delete_on_termination = true
  }

  network {
    name = "public"
  }
}
```

### Workflow

```bash
# 1. Source application credentials
source ~/<project-slug>-openrc.sh

# 2. Export Object Storage credentials for S3 backend
export AWS_ACCESS_KEY_ID=<s3-access-key>
export AWS_SECRET_ACCESS_KEY=<s3-secret-key>

# 3. Terraform workflow
terraform init     # downloads provider plugin, connects to S3 backend
terraform plan     # dry-run, no changes
terraform apply    # provisions all resources
```

---

## Nginx + Let's Encrypt (TLS)

Install packages (Ubuntu 24.04):
```bash
sudo apt update && sudo apt install -y nginx certbot python3-certbot-nginx
```

Or via cloud-init:
```yaml
#cloud-config
packages:
  - nginx
  - certbot
  - python3-certbot-nginx
```

Obtain certificate (certbot nginx plugin):
```bash
sudo certbot --nginx -d <your-domain>
```

Certbot writes certificates to `/etc/letsencrypt/live/<domain>/` and patches the Nginx config automatically. Auto-renewal is configured by the certbot package via a systemd timer or cron.

Manual webroot method (alternative):
```bash
sudo certbot certonly -a webroot \
  --webroot-path=/var/www/<domain>/html/ \
  -d <domain>
```

Renewal script (if using manual method, run `@weekly` via cron):
```bash
#!/bin/sh
if ! letsencrypt renew > /var/log/letsencrypt/renew.log 2>&1 ; then
    echo "Automated renewal failed:"; cat /var/log/letsencrypt/renew.log; exit 1
fi
nginx -t && nginx -s reload
```

---

## OpenStack CLI

Install on Ubuntu/Debian:
```bash
sudo apt install -y python3-pip python3-dev
sudo pip3 install python-openstackclient
```

Or via pipx (recommended on Ubuntu 24.04 where pip is externally managed):
```bash
pipx install python-openstackclient
```

Useful commands for pre-deploy verification:
```bash
source ~/<project-slug>-openrc.sh

# Verify auth works
openstack token issue

# Confirm Ubuntu 24.04 image ID
openstack image list --name "Ubuntu 24.04" --format table

# Confirm s5.small flavor exists
openstack flavor list --format table | grep s5

# Check running instances
openstack server list

# Check volumes
openstack volume list

# Check floating IPs
openstack floating ip list
```

---

## GitHub automation credentials (fine-grained PAT)

Automation that writes a **GitHub Environment's** secrets or variables (for example
machine-scoped Environments holding Cyso credentials) needs more than repository-level grants on a
fine-grained personal access token. Repository Environments + Secrets + Variables read/write is
**not sufficient**: writes return 403 until **organization-level Secrets and Variables read/write**
is also enabled on the token.

| Signature | Meaning |
|---|---|
| Environment create → `403 Resource not accessible by personal access token` | Token lacks the Environments/org-level grants |
| `gh variable set --env <name>` → `404` | The Environment does not exist yet; create it first |
| `gh variable set --env <name>` → `403` with repository-only grants | Org-level Secrets/Variables read/write missing |

`smaqit.infrastructure-repo-config`'s `check-pat-permissions.sh` maps these signatures before any
write.

---

## Example monthly cost (single s5.small deployment)

| Resource | Spec | Monthly cost |
|---|---|---|
| VM instance | `s5.small` (2 vCPU / 8 GB / 50 GB) | € 17.50 |
| Data volume | 10 GB Tier-1 Cinder | € 0.95 |
| Floating IP | 1× IPv4 | € 1.99 |
| Object Storage | < 1 GB Terraform state | < € 0.10 |
| **Total** | | **≈ € 20.54 / month** |

---

## Sources

All facts sourced from live Cyso Cloud documentation on 2026-04-05:

| Topic | URL |
|---|---|
| Getting started | https://cyso.cloud/docs/cloud/getting-started/ |
| Terraform deployment | https://cyso.cloud/docs/cloud/compute/how-to-deploy-an-instance-using-terraform/ |
| Credential formats (OpenRC, Terraform) | https://cyso.cloud/docs/cloud/api-automation/how-to-download-your-openrc-file/ |
| Service endpoints | https://cyso.cloud/docs/cloud/api-automation/service-endpoints/ |
| Object Storage getting started | https://cyso.cloud/docs/object-storage/getting-started/ |
| Volume attachment | https://cyso.cloud/docs/cloud/volume/how-to-attach-a-volume-to-your-instance/ |
| Let's Encrypt + Nginx | https://cyso.cloud/docs/cloud/extra/how-to-use-lets-encrypt-certificates-to-secure-nginxs-ssl-configuration/ |
| OpenStack CLI on Linux | https://cyso.cloud/docs/cloud/extra/how-to-use-the-openstack-cli-tools-on-linux/ |
| Pricing | https://cyso.cloud/pricing (captured in `docs/wiki/cyso-pricing.md`) |
