# Cyso Cloud Deployment Setup

Operational runbook for deploying HIM Corporate to Cyso Cloud (OpenStack, region `ams2`).  
Covers all manual pre-conditions required before `terraform init` can run.

---

## Prerequisites

- Cyso Cloud account active and project created
- Terraform 1.14.7+ installed locally (`terraform -version`)
- SSH client available

---

## Step 1 — Create an Application Credential

Application Credentials are scoped, revokable tokens used for automation and Terraform. They avoid storing your personal password in shell scripts.

1. Log in to [Cyso Cloud Portal](https://my.cyso.cloud)
2. In the top-right menu, open your **Project** dropdown → select your project
3. Navigate to **Access** → **Credentials**
4. Click **Create Application Credential**
5. Fill in:
   - **Name:** `him-corporate-terraform`
   - **Description:** `Terraform deployment for HIM Corporate`
   - **Expiration:** set a date appropriate for your deploy window (or leave blank for no expiry)
   - **Roles:** leave as default (inherits your project roles)
6. Click **Create**
7. **Copy both the ID and Secret immediately** — the secret is shown only once and cannot be retrieved again

> **Security:** Store the secret in a password manager. If lost, delete the credential and create a new one.

---

## Step 2 — Download OpenRC credentials

OpenRC is a shell script that exports the OpenStack environment variables Terraform reads automatically (`OS_AUTH_URL`, `OS_APPLICATION_CREDENTIAL_ID`, `OS_APPLICATION_CREDENTIAL_SECRET`, etc.).

**How to download:**

1. Still on the **Application Credentials** page, find the credential you just created
2. Click **Download OpenRC File** next to it
3. Save as `openrc.sh` outside the git repo (e.g. `~/him-openrc.sh`)

**Load credentials in your terminal:**

```bash
source ~/him-openrc.sh
# No password prompt — the application credential secret is embedded
```

Verify it worked:

```bash
openstack token issue
```

> **Security:** `openrc.sh` contains your application credential secret. Never commit this file. Add `openrc.sh` and `*openrc*` to `.gitignore` if placed near the project directory.

---

## Step 3 — Create Terraform remote state bucket

Terraform state is stored in a Cyso Object Storage container (S3-compatible).  
This bucket must exist **before** `terraform init` runs — it cannot be created by Terraform itself.

1. In the Cyso Cloud Portal, navigate to **Object Storage**
2. Create a new container named: `him-corporate-tfstate`
3. Set visibility to **Private**
4. Under **S3 Credentials**, generate an access key + secret key pair
5. Note down both values — you will need them for `backend.tf`

---

## Step 4 — Generate and upload SSH keypair

```bash
ssh-keygen -t ed25519 -f ~/.ssh/him_key -C "him-corporate-deploy"
```

Upload the **public key** (`~/.ssh/him_key.pub`) to Cyso:

1. In the Cyso Cloud Portal, navigate to **Compute** → **Key Pairs**
2. Click **Import Key Pair**
3. Name: `him-key`
4. Paste the contents of `~/.ssh/him_key.pub`
5. Save

---

## Step 5 — Confirm Ubuntu image ID

The Terraform configuration uses a specific Ubuntu 24.04 LTS image ID for region `ams2`.

Verify the image ID is still valid:

```bash
openstack image list --name "Ubuntu 24.04" --format table
```

Expected ID: `fd91e198-f162-4b6b-a23e-123304fb408a`  
If the ID differs, update `variables.tf` before applying.

---

## Step 6 — Confirm VM flavor name

```bash
openstack flavor list --format table | grep -i small
```

Expected flavor: `s5.small` (2 vCPU / 8 GB RAM / 50 GB disk, € 17.50/month).

---

## Step 7 — Terraform deployment

With all pre-conditions met:

```bash
# Export Terraform credentials (never hardcode in files)
export TF_VAR_app_credential_id=<application-credential-id>
export TF_VAR_app_credential_secret=<application-credential-secret>
export TF_VAR_ssh_public_key="$(cat ~/.ssh/him_key.pub)"

# Export Object Storage credentials for remote state backend
export AWS_ACCESS_KEY_ID=<object-storage-access-key>
export AWS_SECRET_ACCESS_KEY=<object-storage-secret-key>

cd deployment/terraform/

# Initialise with remote state backend
terraform init

# Review planned changes
terraform plan

# Apply (creates VM, boot volume, data volume, floating IP, security group)
terraform apply
```

Note the `floating_ip` value from the output — you'll need it for all subsequent steps.

---

## Step 8 — Post-apply: format data volume and deploy app

SSH into the VM (cloud-init may still be running — wait ~60 seconds after apply):

```bash
ssh -i ~/.ssh/him_key ubuntu@<FLOATING_IP>
```

**Format and mount the Cinder data volume** (run once):

```bash
sudo mkfs.ext4 /dev/vdb
sudo mkdir -p /data
sudo mount /dev/vdb /data
echo '/dev/vdb /data ext4 defaults,nofail 0 2' | sudo tee -a /etc/fstab
```

**Create the `.env` file** (never in Terraform state):

```bash
sudo tee /opt/him/.env > /dev/null <<EOF
NODE_ENV=production
PORT=3001
DB_PATH=/data/him.db
ANTHROPIC_API_KEY=<your-key-here>
EOF
sudo chmod 600 /opt/him/.env
```

**Deploy app files** from your local machine:

```bash
# Build backend locally first
cd backend && npm run build && cd ..

# Copy compiled backend and Docker Compose file to VM
scp -r -i ~/.ssh/him_key backend/dist ubuntu@<FLOATING_IP>:/opt/him/backend/
scp -i ~/.ssh/him_key backend/package.json ubuntu@<FLOATING_IP>:/opt/him/backend/
scp -i ~/.ssh/him_key deployment/docker-compose.yml ubuntu@<FLOATING_IP>:/opt/him/

# Install production dependencies on VM
ssh -i ~/.ssh/him_key ubuntu@<FLOATING_IP> \
  "cd /opt/him/backend && npm install --omit=dev"

# Copy built frontend assets for Nginx to serve
cd frontend && npm run build && cd ..
scp -r -i ~/.ssh/him_key frontend/dist ubuntu@<FLOATING_IP>:/opt/him/frontend/
```

**Start the backend container:**

```bash
ssh -i ~/.ssh/him_key ubuntu@<FLOATING_IP> \
  "cd /opt/him && docker compose up -d"
```

---

## Step 9 — TLS and DNS

1. Point your domain's A record to the Floating IP from `terraform output`
2. SSH into the VM and configure Nginx + Certbot:

```bash
sudo certbot --nginx -d <your-domain>
```

Certbot will add the HTTPS server block and configure auto-renewal.

---

## Smoke test

```bash
# Via HTTPS (after DNS propagation + Certbot)
curl https://<your-domain>/api/health
# Expected: {"status":"ok"}

# OpenStack CLI health checks
source ~/him-openrc.sh
openstack server show him-corporate          # VM in ACTIVE state
openstack floating ip list                   # Floating IP attached
openstack volume show him-data              # Data volume in-use
```

---

## About the Cyso Dockerfile download

The Cyso Cloud portal also offers a `Dockerfile` download alongside the OpenRC and Terraform files. This is a **Cyso CLI tooling image** (`fugacloud/openstackcli`) — it is not a Node.js container. It is useful only if you need to run `openstack` CLI commands inside a Docker container (e.g. in a CI/CD pipeline without the CLI installed). It is not used in this deployment.

The backend container uses `node:22-alpine` (defined in `deployment/docker-compose.yml`).

---

## References

- [Cyso Cloud — Application Credentials & OpenRC](https://cyso.cloud/docs/cloud/api-automation/how-to-download-your-openrc-file)
- [Cyso Cloud — Deploy instance with Terraform](https://cyso.cloud/docs/cloud/compute/how-to-deploy-an-instance-using-terraform/)
- [Cyso Cloud — Object Storage](https://cyso.cloud/docs/cloud/storage/object-storage/)
- [Terraform OpenStack Provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs)
