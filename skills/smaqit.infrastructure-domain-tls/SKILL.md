---
name: smaqit.infrastructure-domain-tls
description: Use when configuring a custom domain with a Let's Encrypt TLS certificate for a running nginx server. Covers DNS propagation verification, nginx server_name update, Certbot certificate issuance, HTTPS smoke tests, and auto-renewal verification. Also use when activating a purchased domain for a deployed application, setting up HTTPS post-deployment, or when a spec references INF-TOPOLOGY-004 through INF-TOPOLOGY-007 as pending TLS configuration.
metadata:
  version: "1.0.0"
---

# Configure Domain TLS with Let's Encrypt

## Pre-conditions

Operator must complete before invoking:
- Domain purchased and DNS A record pointing at the VM's public IP (`81.24.10.203` for this project)
- nginx installed and serving the app on HTTP
- SSH access to the VM

Verify Certbot is available before Step 4: `which certbot`. It is pre-installed by cloud-init in this project.

## Steps

All commands that run on the VM are executed via SSH.

1. **Fix any stale spec reference:**
   Check `specs/infrastructure/deployment-topology.md` for the `DNS prerequisite` constraint. If it references "Floating IP", update it to the fixed IP (`81.24.10.203`) and commit before proceeding.

2. **Verify DNS propagation:**
   ```bash
   dig <domain> A +short
   ```
   Must return `81.24.10.203`. If not resolved, wait and retry — do NOT proceed until confirmed.

3. **Update nginx `server_name`:**
   ```bash
   sudo sed -i 's/server_name _;/server_name <domain>;/' /etc/nginx/sites-available/him
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
   curl -sf https://<domain>/api/health  # assert HTTP 200
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

8. **Update specs:** Invoke `smaqit.spec-status-update` for `specs/infrastructure/deployment-topology.md` — mark INF-TOPOLOGY-004 through INF-TOPOLOGY-007 as `[x]`.

## Output

- nginx configured with real `server_name` and TLS
- Let's Encrypt certificate issued and auto-renewal configured
- INF-TOPOLOGY-004 through INF-TOPOLOGY-007 checked off in `specs/infrastructure/deployment-topology.md`

## Scope

- Does NOT purchase the domain — that is a human action (joker.com or equivalent registrar).
- Does NOT handle multi-tenant or wildcard certificates.
- Does NOT configure HSTS or other security headers (those can be added to nginx separately post-issuance).

## Examples

**Input:** Domain `himcorp.com` purchased, A record set to `81.24.10.203`. User invokes `/domain.tls himcorp.com`.
**Output:** TLS certificate issued, HTTPS live at `https://himcorp.com/`, HTTP redirects to HTTPS, health check passing, INF-TOPOLOGY-004–007 checked off.

## Gotchas

- `dig` propagation can take minutes to hours depending on DNS TTL. Do NOT run Certbot before DNS is confirmed — Certbot will fail the ACME HTTP-01 challenge if the domain doesn't resolve to the VM.
- Certbot rewrites the nginx config file. After issuance, the `listen 443 ssl` blocks are managed by Certbot — do not manually edit the nginx file for TLS settings post-issuance.
- `81.24.10.90` (Cyso floating IP) does not route on the flat network. Always use the fixed IP `81.24.10.203` as the DNS A record target.
- Port 80 must be open in the VM firewall during the ACME HTTP-01 challenge. If Certbot fails, check `ufw status` and ensure port 80 is allowed.

## Completion

- [ ] Stale spec reference fixed (if applicable)
- [ ] DNS propagation confirmed: `dig <domain>` returns `81.24.10.203`
- [ ] nginx `server_name` updated and config tested (`nginx -t` passes)
- [ ] Certbot certificate issued without errors
- [ ] HTTPS smoke test passes (HTTP 200 + `text/html`)
- [ ] HTTP → HTTPS redirect confirmed (301)
- [ ] `/api/health` accessible via HTTPS (HTTP 200)
- [ ] Certificate issuer: Let's Encrypt, not expired
- [ ] Auto-renewal dry-run passes
- [ ] INF-TOPOLOGY-004 through INF-TOPOLOGY-007 checked off in spec

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
