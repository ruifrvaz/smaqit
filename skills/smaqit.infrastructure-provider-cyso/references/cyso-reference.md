# Cyso Cloud Reference

Authoritative reference for HIM Corporate infrastructure on Cyso Cloud (OpenStack, region `ams2`).  
All facts verified against live documentation as of **2026-04-05**.

---

## Platform overview

| Property | Value |
|---|---|
| Platform | OpenStack (Cyso-managed) |
| Primary region | `ams2` (Amsterdam) |
| Secondary region | `fra` (Frankfurt) — not used by HIM Corporate |
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
3. Click **Create** — choose a name (e.g. `him-corporate-terraform`), optionally set an expiry
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

In HIM Corporate's Terraform, credentials are passed via `TF_VAR_*` environment variables — never hardcoded. The `auth_url` is hardcoded in `main.tf` as it is a public platform constant.

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

| Service | OpenStack component | Used by HIM Corporate |
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

#### Standard (s5) — chosen for HIM Corporate

| Flavor | vCPU | RAM | Disk | Monthly |
|---|---|---|---|---|
| **s5.small** | 2 | 8 GB | 50 GB | € 17.50 |
| s5.medium | 4 | 16 GB | 100 GB | € 37.50 |
| s5.large | 8 | 32 GB | 200 GB | € 75.00 |

**HIM Corporate uses `s5.small`** — sufficient for Node.js backend + SQLite at MVP scale.

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
ssh -i ~/.ssh/him_key ubuntu@<FLOATING_IP>
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

| Type | IOPS/GB | Max IOPS | Multi-attach | Price |
|---|---|---|---|---|
| **Tier-1** | 5 | 25k | No | € 0.095 /GB/month |
| Tier-1m | 5 | 25k | Yes | € 0.145 /GB/month |
| Tier-2 | 25 | 25k | No | € 0.250 /GB/month |
| Tier-2m | 25 | 25k | Yes | € 0.300 /GB/month |

**HIM Corporate uses Tier-1** (10 GB ≈ € 0.95/month) for the SQLite data volume.

### Volume attachment

Data volumes attach at `/dev/vdb` (first non-boot Cinder volume). Format and mount on first attach:

```bash
sudo mkfs.ext4 /dev/vdb
sudo mkdir -p /data
sudo mount /dev/vdb /data
echo '/dev/vdb /data ext4 defaults,nofail 0 2' | sudo tee -a /etc/fstab
```

The `nofail` option prevents boot failure if the volume is temporarily unavailable.

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

# Separate data volume
resource "openstack_blockstorage_volume_v3" "data" {
  name = "him-data"
  size = 10
}

resource "openstack_compute_volume_attach_v2" "data" {
  instance_id = openstack_compute_instance_v2.him.id
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
    bucket   = "him-corporate-tfstate"
    key      = "him/terraform.tfstate"
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
resource "openstack_networking_floatingip_v2" "him" {
  pool = "public"
}

resource "openstack_compute_floatingip_associate_v2" "him" {
  floating_ip = openstack_networking_floatingip_v2.him.address
  instance_id = openstack_compute_instance_v2.him.id
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
resource "openstack_networking_secgroup_v2" "him" {
  name        = "him-corporate-sg"
  description = "Allow SSH, HTTP, HTTPS"
}

resource "openstack_networking_secgroup_rule_v2" "ssh" {
  direction         = "ingress"
  ethertype         = "IPv4"
  protocol          = "tcp"
  port_range_min    = 22
  port_range_max    = 22
  remote_ip_prefix  = "0.0.0.0/0"
  security_group_id = openstack_networking_secgroup_v2.him.id
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
resource "openstack_compute_keypair_v2" "him" {
  name       = "him-key"
  public_key = var.ssh_public_key
}
```

Key pairs can also be uploaded via the dashboard: **Access → Key Pairs → Import Key Pair**.

### Instance resource

```hcl
resource "openstack_compute_instance_v2" "him" {
  name            = "him-corporate"
  flavor_name     = "s5.small"
  key_pair        = openstack_compute_keypair_v2.him.name
  security_groups = [openstack_networking_secgroup_v2.him.name]

  block_device {
    uuid                  = "bcdcd204-65b6-440c-afd0-b91ac812d43c"  # Ubuntu 24.04 LTS
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
source ~/him-openrc.sh

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

Useful commands for HIM Corporate pre-deploy verification:
```bash
source ~/him-openrc.sh

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

## HIM Corporate cost summary

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
