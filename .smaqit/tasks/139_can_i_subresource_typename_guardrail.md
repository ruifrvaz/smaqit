---
status: Abandoned
created: "2026-10-18"
---

# Guardrail: Reject `kubectl auth can-i <verb> <resource>/<subresource>` in Project Workflows and Scripts

## Description

`kubectl auth can-i create pods/exec` does not ask whether the deployer may exec into pods. kubectl
parses its argument as `TYPE`, `TYPE/NAME`, or a non-resource URL, so `pods/exec` means "a pod
named `exec`". Any deployer allowed to create pods gets `yes`, even when exec is denied. The correct
form is `kubectl auth can-i create pods --subresource=exec`.

Task 136 fixed smaqit's own checks: `namespace-guard.sh` uses `--subresource`, and a test pins it.
It also documented the trap in the `cicd-generate` and `deploy-k3s-app` skills. A check a project
writes itself is still not inspected: a downstream project that adds its own `can-i` step can hit the
trap again, exactly as the first live private-mesh deploy did. This task turns the documentation into
a deterministic guardrail that refuses the `TYPE/NAME` form wherever a project writes it.

This is a follow-up of task 136. It was meant as a child of 136, but 136 was already `PR Open`
(PR #105), and a child must join an `In Progress` parent and ship inside the parent's PR.

## Issue Triage Context

**Mode:** Auto
**Technologies:** Kubernetes (k3s), kubectl, GitHub Actions, Bash, Go
**Platforms/Environments:** None
**Features/Integrations:** kubectl auth can-i subresource check, smaqit validate, generated k3s deploy workflow, project workflow and script scanning
**Versions/Constraints:** deterministic and fail-closed, no false positives on legitimate TYPE/NAME checks, app- and platform-agnostic

## Design Decisions

TBD — to be confirmed during assessment. Recorded starting points:

- **Match only known subresources, not every `TYPE/NAME`.** `kubectl auth can-i get pods/my-pod` is a
  legitimate check about a named pod, so the guardrail must not reject every slash. It should flag a
  `can-i` argument whose part after the slash is a known subresource name (`exec`, `log`, `attach`,
  `portforward`, `proxy`, `status`, `scale`, `eviction`, `ephemeralcontainers`, `binding`, `token`,
  and similar). The exact list is to be settled against the Kubernetes API reference.
- **Fail, not warn, with the fix in the message.** For example: "use `kubectl auth can-i create pods
  --subresource=exec`".
- **Where it runs — open.** Options:
  - (a) `smaqit validate` (Go) scans `.github/workflows/*.yml` and the project's shell scripts
    locally;
  - (b) a vendored lint script that the generated k3s `deploy.yml` runs before any `kubectl` call,
    so the check also runs in CI, where the smaqit binary is absent;
  - (c) both, sharing one pattern definition.
- **Scope of scanned files — open.** At least `.github/workflows/` and `deployment/**/*.sh`. Whether
  to scan more of the repository is to be decided.
- **Out of scope:** the `bash -e` trap. Whether a command is expected to fail is a matter of intent,
  so it can't be detected statically; a separate helper (`expect-refused.sh`) may be considered on
  its own.

## Implementation Steps

1. Settle the open Design Decisions with `task.plan 139`: where the check runs, which files it scans,
   and the subresource list.
2. Implement the pattern check: one definition, shared by every place it runs.
3. Wire it into the chosen place(s): `smaqit validate`, and/or a vendored lint step in
   `deploy.yml.k3s.template` with its renderer, golden files and vendoring list in
   `smaqit.infrastructure-cicd-generate`.
4. Tests:
   - the `TYPE/NAME` subresource form is rejected in a workflow `run:` block and in a shell script;
   - the `--subresource=` form passes;
   - a legitimate named-resource check (`can-i get pods/my-pod`) passes;
   - comments that quote the bad form (such as `namespace-guard.sh`'s own warning) are handled
     deliberately.
5. Update the skill gotchas from "document" to "enforced", plus the wiki, compendium and CHANGELOG.

## Known Issues Triage

[Populated by smaqit.task-start via smaqit.utils.triage-issues. Do not edit manually.]

## Acceptance Criteria

- [ ] A `kubectl auth can-i <verb> <resource>/<subresource>` call (known subresource) in a project
      workflow or script fails the guardrail, naming the file, the line, and the `--subresource=` fix
- [ ] The `--subresource=` form and legitimate named-resource checks (`can-i get pods/my-pod`) pass
- [ ] The guardrail runs where the plan decides (locally and/or in the generated CI workflow), and is
      tested there
- [ ] smaqit's own shipped scripts and templates pass the guardrail
- [ ] Skill text, wiki, compendium and CHANGELOG describe the enforced check; released

## Findings

[Populated by smaqit.task-complete. Do not fill in manually before task is complete.]

**Implementation approach:**
- TBD

**Decisions made:**
- TBD

**Blockers encountered:**
- TBD

**Follow-up identified:**
- TBD

## Files to Create / Modify

| File | Action |
|------|--------|
| installer/main.go and/or a new installer source file (smaqit validate check) | Modify / Create |
| skills/smaqit.infrastructure-deploy-k3s-app/scripts/ (vendored lint script, if chosen) | Create |
| skills/smaqit.infrastructure-cicd-generate/assets/deploy.yml.k3s.template, scripts/render-k3s-deploy.sh, SKILL.md | Modify (if the CI step is chosen) |
| skills/smaqit.infrastructure-deploy-k3s-app/SKILL.md | Modify |
| installer tests and installer/testdata/k3s-deploy/* | Modify |
| docs/wiki/workflows/deployment-slots.md, .smaqit/compendium.md, CHANGELOG.md | Modify |

## Notes

- **Abandoned 2026-10-18 before starting:** the user judged a static pattern check too brittle. The
  traps are covered instead by gotchas (shipped in v3.16.0 with task 136) and failure-handling rows
  in the `deploy-k3s-app` and `cicd-generate` skills. The rows (commit fc568e0) were pushed to the
  task 136 branch after PR #105 had merged, so they still need their own PR.

- Follow-up of task 136 (PR #105, v3.16.0). Start it after #105 merges, so it builds on the shipped
  `namespace-guard.sh` and the documented gotchas.
- Source: the first live private-mesh deploy, where a downstream project's own `can-i create pods/exec`
  check answered `yes` for a deployer that could not exec.
- Triage note from task 136: kubectl#1217 means `can-i --subresource` doesn't validate that the
  subresource exists, so a misspelled `--subresource=` value also answers silently. Consider
  checking the `--subresource=` value against the same known list.
