---
version: "1.0.0"
---

# Phase Differences from `smaqit.new-greenfield-project`

| `smaqit.new-greenfield-project` | `smaqit.feature-new` | Difference |
|---|---|---|
| Phase 0 — Task Creation | Phase 0 — Task Creation | Adds an explicit deploy-now/defer decision. Creates 5 phase tasks, not 7. |
| Phase 1 — Requirements Extraction | *(none)* | Feature requirements come from session context; no `assets/raw/` sweep. |
| Phase 2 — Specification (from-scratch 5-layer generation) | Phase 1 — Spec Revalidation | Uses the Incremental Development model (`smaqit plan --phase=develop`, no `--regen`) and the Incremental Spec Updates decision table — see [references/spec-lifecycle-reference.md](spec-lifecycle-reference.md). Specs not touched by the feature are left unchanged. |
| Phase 3 — Development | Phase 2 — Development | Same mechanics (`/smaqit.development`, canonical `amendment:` tag). |
| Phase 4 — Dev Environment Sweep | *(none)* | Not present. |
| Provisioning Mode (applies to Phases 4/5) | Provisioning Mode (applies to Phase 3) | `smaqit.input-deployment`'s default (`provision`) is unchanged. `smaqit.feature-new` Phase 3 Step 1 resolves `existing-owned` first when `specs/infrastructure/*.md` already shows `status: deployed`, before falling through to `smaqit.input-deployment`'s elicitation. |
| Phase 5 — Production Deployment via CI/CD | Phase 3 — Deployment | Same CI/CD reuse (no IaC regeneration). Adds: amendment gate runs here on every cycle, deploy-now or defer; explicit defer path leaves the phase task open with a recorded reason instead of `Completed`. |
| Phase 6 — Domain + TLS | *(none)* | Not present. |
| Phase 7 — Validation | Phase 4 — Validation | Unchanged. |
| Phase 8 — Release | Phase 5 — Close-out | Same amendment re-scan. Release tagging runs only if the deploy-now path completed; skipped under `defer`. |
