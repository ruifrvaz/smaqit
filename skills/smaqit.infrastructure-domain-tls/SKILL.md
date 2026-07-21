---
name: smaqit.infrastructure-domain-tls
description: Use when configuring a custom domain with a Let's Encrypt TLS certificate for a running nginx server. Covers DNS propagation verification, nginx server_name update, Certbot certificate issuance, HTTPS smoke tests, and auto-renewal verification. Also use when activating a purchased domain for a deployed application, setting up HTTPS post-deployment, or when the infrastructure spec references pending TLS/domain acceptance criteria.
metadata:
  version: "1.1.0"
---

# Configure Domain TLS with Let's Encrypt

## Pre-conditions

Operator must complete before invoking:
- Domain purchased and DNS A record pointing at the VM's fixed IP (`<vm-fixed-ip>`)
- nginx installed and serving the app on HTTP
- SSH access to the VM

Verify Certbot is available before Step 4: `which certbot`. If cloud-init did not pre-install it
for this project, install with `sudo apt install python3-certbot-nginx` first.

## Steps

All commands that run on the VM are executed via SSH.

1. **Fix any stale spec reference:**
   Check the infrastructure spec (e.g. `specs/infrastructure/deployment.md`) for a `DNS
   prerequisite` constraint. If it references "Floating IP", update it to the fixed IP
   (`<vm-fixed-ip>`) and commit before proceeding.

2. **Verify DNS propagation:**
   ```bash
   dig <domain> A +short
   ```
   Must return `<vm-fixed-ip>`. If not resolved, wait and retry — do NOT proceed until confirmed.

3. **Update nginx `server_name`:**
   ```bash
   sudo sed -i 's/server_name _;/server_name <domain>;/' /etc/nginx/sites-available/<project-slug>
   sudo nginx -t
   sudo systemctl reload nginx
   ```

4. **Run Certbot:**
   ```bash
   sudo certbot --nginx -d <domain>
   ```
   If `www.<domain>` is also desired:
   ```bash
   sudo certbot --nginx -d <domain> -d www.<domain>
   ```
   Certbot obtains the certificate, rewrites the nginx config for HTTPS, and adds the HTTP→HTTPS 301 redirect automatically.

5. **Smoke tests:**
   ```bash
   curl -sI https://<domain>/            # assert HTTP 200, Content-Type: text/html
   curl -sI http://<domain>/             # assert HTTP 301 redirect to https://
   curl -sf https://<domain>/<health-check-path>  # assert HTTP 200
   ```

6. **Certificate verification:**
   ```bash
   curl -vI https://<domain>/
   ```
   Confirm `issuer: Let's Encrypt` and certificate not expired.

7. **Auto-renewal test:**
   ```bash
   sudo certbot renew --dry-run
   ```
   Must succeed without errors. Run this even if auto-renewal appears configured — the systemd timer may not be active on first boot.

8. **Update specs:** Invoke `smaqit.spec-status-update` for the infrastructure spec (e.g.
   `specs/infrastructure/deployment.md`) — mark the domain/TLS acceptance criteria as `[x]`.

## Output

- nginx configured with real `server_name` and TLS
- Let's Encrypt certificate issued and auto-renewal configured
- Domain/TLS acceptance criteria checked off in the infrastructure spec

## Scope

- Does NOT purchase the domain — that is a human action (joker.com or equivalent registrar).
- Does NOT handle multi-tenant or wildcard certificates.
- Does NOT configure HSTS or other security headers (those can be added to nginx separately post-issuance).

## Examples

**Input:** Domain `<domain>` purchased, A record set to `<vm-fixed-ip>`. User invokes `/domain.tls <domain>`.
**Output:** TLS certificate issued, HTTPS live at `https://<domain>/`, HTTP redirects to HTTPS, health check passing, domain/TLS acceptance criteria checked off.

## Gotchas

- `dig` propagation can take minutes to hours depending on DNS TTL. Do NOT run Certbot before DNS is confirmed — Certbot will fail the ACME HTTP-01 challenge if the domain doesn't resolve to the VM.
- Certbot rewrites the nginx config file. After issuance, the `listen 443 ssl` blocks are managed by Certbot — do not manually edit the nginx file for TLS settings post-issuance.
- A cloud provider's floating IP may not route on a flat network (this is the case on Cyso, for
  example). Always use the VM's actual fixed/routable IP as the DNS A record target — confirm
  which one applies for the target provider before configuring DNS.
- Port 80 must be open in the VM firewall during the ACME HTTP-01 challenge. If Certbot fails, check `ufw status` and ensure port 80 is allowed.

## Completion

- [ ] Stale spec reference fixed (if applicable)
- [ ] DNS propagation confirmed: `dig <domain>` returns `<vm-fixed-ip>`
- [ ] nginx `server_name` updated and config tested (`nginx -t` passes)
- [ ] Certbot certificate issued without errors
- [ ] HTTPS smoke test passes (HTTP 200 + `text/html`)
- [ ] HTTP → HTTPS redirect confirmed (301)
- [ ] `<health-check-path>` accessible via HTTPS (HTTP 200)
- [ ] Certificate issuer: Let's Encrypt, not expired
- [ ] Auto-renewal dry-run passes
- [ ] Domain/TLS acceptance criteria checked off in the infrastructure spec

## Failure Handling

| Situation | Action |
|-----------|--------|
| Required input not provided | Request the missing information before proceeding |
| Gathered input is ambiguous | Flag the ambiguity and ask for clarification |
| Subagent invocation fails | Report the failure with context; do not silently retry |
| Output artifact already exists | Confirm with user before overwriting |
| DNS not propagated | Wait 5 minutes and retry `dig`; do not proceed to Certbot until confirmed |
| Certbot ACME challenge fails | Verify nginx is serving on port 80 with the correct `server_name`; check firewall (port 80 must be open) |
| `certbot` command not found | Install: `sudo apt install python3-certbot-nginx` |
| nginx config test fails after `server_name` update | Show the exact nginx error; restore with `sudo sed -i 's/server_name <domain>;/server_name _;/'` |
