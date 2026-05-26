# Skill Definition: smaqit.infrastructure-domain-tls

## Identity
- **Name:** smaqit.infrastructure-domain-tls
- **Version:** 1.0.0
- **Description:** Configure a custom domain and provision a Let's Encrypt TLS certificate for a running nginx server. Covers DNS A record verification, nginx server_name update, Certbot certificate issuance, HTTPS smoke tests, and auto-renewal verification.

## Pre-conditions (operator must complete before invoking)
- Domain purchased and DNS A record pointing at the VM's public IP
- nginx installed and serving the app on HTTP
- `python3-certbot-nginx` installed on the VM (provisioned by cloud-init in this project)
- SSH access to the VM

## Steps
All commands that run on the VM are executed via SSH.

1. **Fix any stale spec reference:** Check `specs/infrastructure/deployment-topology.md` for the `DNS prerequisite` constraint. If it references "Floating IP", update it to the fixed IP (`81.24.10.203` for this project) and commit before proceeding.
2. **Verify DNS propagation:** Run `dig <domain> A +short` locally. Must return the VM's public IP. If not resolved yet, wait and retry — do NOT proceed until propagation is confirmed.
3. **Update nginx `server_name`:**
   ```
   sudo sed -i 's/server_name _;/server_name <domain>;/' /etc/nginx/sites-available/him
   sudo nginx -t
   sudo systemctl reload nginx
   ```
4. **Run Certbot:**
   ```
   sudo certbot --nginx -d <domain>
   ```
   If `www.<domain>` is also desired: `sudo certbot --nginx -d <domain> -d www.<domain>`.
   Certbot will obtain the certificate, rewrite the nginx config for HTTPS, and add the HTTP→HTTPS 301 redirect automatically.
5. **Smoke tests:**
   a. `curl -sI https://<domain>/` — assert HTTP 200, `Content-Type: text/html`
   b. `curl -sI http://<domain>/` — assert HTTP 301 redirect to `https://`
   c. `curl -sf https://<domain>/api/health` — assert HTTP 200
6. **Certificate verification:** `curl -vI https://<domain>/` — confirm `issuer: Let's Encrypt`, not expired.
7. **Auto-renewal test:** `sudo certbot renew --dry-run` — must succeed without errors.
8. **Update specs:** Invoke `smaqit.spec-status-update` for `specs/infrastructure/deployment-topology.md` — mark INF-TOPOLOGY-004 through INF-TOPOLOGY-007 as `[x]`.

## Output
- nginx configured with real `server_name` and TLS
- Let's Encrypt certificate issued and auto-renewal configured
- INF-TOPOLOGY-004 through 007 checked off

## Scope
- Does NOT purchase the domain. That is a human action (joker.com or equivalent registrar).
- Does NOT handle multi-tenant or wildcard certificates.
- Does NOT configure HSTS or other security headers (those can be added to nginx separately).

## Gotchas
- `dig` propagation can take minutes to hours depending on DNS TTL. Do NOT run Certbot before DNS is confirmed — Certbot will fail the ACME HTTP-01 challenge if the domain doesn't resolve to the VM.
- Certbot rewrites the nginx config file. After issuance, the `server_name _;` replacement and the `listen 443 ssl` blocks are managed by Certbot. Do not manually edit the nginx file for TLS settings after Certbot runs.
- `81.24.10.90` (Cyso floating IP) does not route on the flat network. Always use the fixed IP `81.24.10.203` as the DNS A record target for this project.
- `python3-certbot-nginx` is pre-installed by cloud-init in this project. Verify it is present (`which certbot`) before running.
- `sudo certbot renew --dry-run` — run this even if auto-renewal seems configured. The systemd timer may not be active on first boot.

## Completion
- [ ] Stale spec reference fixed (if applicable)
- [ ] DNS propagation confirmed: `dig <domain>` returns VM IP
- [ ] nginx `server_name` updated and config tested
- [ ] Certbot certificate issued without errors
- [ ] HTTPS smoke test passes (200 + text/html)
- [ ] HTTP → HTTPS redirect confirmed (301)
- [ ] `/api/health` accessible via HTTPS
- [ ] Certificate issuer: Let's Encrypt, not expired
- [ ] Auto-renewal dry-run passes
- [ ] INF-TOPOLOGY-004 through 007 checked off in spec

## Failure Handling
| Situation | Action |
|-----------|--------|
| DNS not propagated | Wait 5 minutes and retry `dig`; do not proceed to Certbot until confirmed |
| Certbot ACME challenge fails | Verify nginx is serving on port 80 with the correct `server_name`; check firewall (port 80 must be open) |
| `certbot` command not found | Install: `sudo apt install python3-certbot-nginx` |
| nginx config test fails after `server_name` update | Show the exact nginx error; restore with `sudo sed -i 's/server_name <domain>;/server_name _;/'` |

## Examples
**Input:** Domain `himcorp.com` purchased, A record set to `81.24.10.203`. User invokes `/domain.tls himcorp.com`.
**Output:** TLS certificate issued, HTTPS live at `https://himcorp.com/`, HTTP redirects to HTTPS, health check passing, INF-TOPOLOGY-004–007 checked off.
