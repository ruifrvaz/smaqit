You are the smaqit Deployment agent. Read the Case brief's treatment table: the `deployment-agent`
treatment is your agent definition (`agents/deployment.md`) and the `input-deployment` treatment is
the `smaqit.input-deployment` skill. Follow them for this run.

Run a deployment for the project in `project/` with these execution parameters:

- `specification_mode: prevalidated`, handoff paths `specs/infrastructure/inf-target.md` and
  `specs/stack/platform-stack.md` (both already `status: deployed`)
- `provisioning_mode: existing-owned`, environment `prod`, `deployment_path: standard`
- No `slot_plan` is passed in: resolve it yourself as the agent definition and the input skill say

Constraints of this bench environment:

- The smaqit CLI for this project is `project/bin/smaqit`. Use it for every `smaqit` command; never
  use another `smaqit` binary on the machine.
- Every filler skill is a local stub; there is no real infrastructure, Vault, network, or CI. Do not
  invoke any skill, script, or tool that the slot plan does not name.
- Spec status updates and design gates are out of scope; the specs are already deployed.
- When you finish or stop, write one line to `project/deployment-outcome.txt`: `COMPLETED` if every
  slot ran as planned, or `STOPPED: ` followed by the exact diagnostic that stopped you.

No clarifying question is needed. You are running non-interactively via `codex exec`; no further
input is coming.
