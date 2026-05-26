# Cyso Cloud Pricing Reference

Source: https://cyso.cloud/pricing — as of 2026-04-05. All prices in EUR, excluding VAT.  
Billing model: hourly usage, invoiced monthly.

---

## Compute

Three flavor families:
- **s5** — Standard. Balanced vCPU/RAM ratio. General-purpose workloads.
- **c5** — CPU Optimised. Higher vCPU relative to RAM. Compute-heavy workloads.
- **m5** — Memory Optimised. Higher RAM relative to vCPU. Memory-intensive workloads.

Ephemeral storage (root disk) is included in the instance price.

### Standard (s5)

| Flavor | vCPU | RAM | Disk | Hourly | Monthly |
|---|---|---|---|---|---|
| s5.small | 2 | 8 GB | 50 GB | € 0.02604 | € 17.50 |
| s5.medium | 4 | 16 GB | 100 GB | € 0.05580 | € 37.50 |
| s5.large | 8 | 32 GB | 200 GB | € 0.11161 | € 75.00 |
| s5.xlarge | 16 | 64 GB | 200 GB | € 0.22321 | € 150.00 |
| s5.2xlarge | 32 | 128 GB | 400 GB | € 0.44643 | € 300.00 |
| s5.4xlarge | 64 | 256 GB | 400 GB | € 0.89286 | € 600.00 |

### CPU Optimised (c5)

| Flavor | vCPU | RAM | Disk | Hourly | Monthly |
|---|---|---|---|---|---|
| c5.small | 2 | 4 GB | 25 GB | € 0.03274 | € 22.00 |
| c5.medium | 4 | 8 GB | 50 GB | € 0.06510 | € 43.75 |
| c5.large | 8 | 16 GB | 100 GB | € 0.13021 | € 87.50 |
| c5.xlarge | 16 | 32 GB | 200 GB | € 0.26040 | € 175.00 |
| c5.2xlarge | 32 | 64 GB | 400 GB | € 0.52083 | € 350.00 |
| c5.4xlarge | 64 | 128 GB | 400 GB | € 1.04167 | € 700.00 |

### Memory Optimised (m5)

| Flavor | vCPU | RAM | Disk | Hourly | Monthly |
|---|---|---|---|---|---|
| m5.small | 2 | 16 GB | 100 GB | € 0.03981 | € 26.75 |
| m5.medium | 4 | 32 GB | 200 GB | € 0.07961 | € 53.50 |
| m5.large | 8 | 64 GB | 200 GB | € 0.18787 | € 126.25 |
| m5.xlarge | 16 | 128 GB | 400 GB | € 0.31622 | € 212.50 |
| m5.2xlarge | 32 | 256 GB | 400 GB | € 0.63244 | € 425.00 |
| m5.4xlarge | 64 | 512 GB | 800 GB | € 1.26488 | € 850.00 |

> **HIM Corporate uses `s5.small`** — € 17.50/month. Single-tenant Node.js backend + SQLite. Sufficient for initial production load.

---

## Storage

| Type | Description | Price |
|---|---|---|
| Tier-1 | Standard persistent volume. Triple replication, 5 IOPS/GB, up to 25k IOPS. | € 0.095 /GB/month |
| Tier-1m | Multi-attach (multiple hosts). Same performance as Tier-1. | € 0.145 /GB/month |
| Tier-2 | High-performance persistent volume. Triple replication, 25 IOPS/GB, up to 25k IOPS. | € 0.250 /GB/month |
| Tier-2m | Multi-attach, high-performance. | € 0.300 /GB/month |
| Image Storage | VM disk image storage. | € 0.080 /GB/month |
| Snapshot Storage | Point-in-time volume snapshots. | € 0.080 /GB/month |
| Volume Backup | Automated volume backup copies. | € 0.080 /GB/month |
| Object Storage | S3-compatible, NVMe-backed. GDPR / ISO 27001. 99.99% SLA. | € 0.055 /GB/month |

### Object Storage operations

| Operation | Price |
|---|---|
| Ingress (upload) | FREE |
| Egress (download) | € 0.055 /GB |
| Category A (PUT, LIST) | € 0.055 per 10,000 ops |
| Category B (other) | € 0.004 per 10,000 ops |

> **HIM Corporate uses:** 10 GB Tier-1 Cinder data volume (≈ € 0.95/month) + Object Storage bucket for Terraform state (negligible cost).

---

## Network

| Resource | Monthly |
|---|---|
| IPv4 / Floating IP | € 1.99 |
| IPv6 /48 | FREE |
| LBaaS Small | € 10.95 |
| LBaaS Medium | € 27.45 |
| LBaaS Large | € 54.95 |

> **HIM Corporate uses:** 1 Floating IP — € 1.99/month.

---

## Services

| Service | Price |
|---|---|
| EMK Control Plane (managed Kubernetes) | € 29.99/month |
| EMK Control Plane HA | € 59.99/month |
| DNS as a Service — European AnyCast | € 1.09/month |
| DNS as a Service — Global AnyCast | € 5.49/month |
| Transactional Email (first 1,000 in first project) | FREE |
| Transactional Email (up to 10,000/project) | € 5.00/month |
| Transactional Email (above 10,000) | € 0.50 per 1,000 |

---

## HIM Corporate estimated monthly cost

| Resource | Spec | Cost |
|---|---|---|
| VM instance | `s5.small` | € 17.50 |
| Data volume | 10 GB Tier-1 | € 0.95 |
| Floating IP | 1× IPv4 | € 1.99 |
| Object Storage | < 1 GB Terraform state | < € 0.10 |
| **Total** | | **≈ € 20.54 / month** |
