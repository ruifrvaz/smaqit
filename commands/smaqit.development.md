---
description: Orchestrate the Development phase — Business/Functional/Stack specs, then a working application.
---

Run the Development phase through the `smaqit-development` subagent. The subagent cannot ask the user anything, so this conversation owns every question.

1. **Resolve parameters.** Apply the `smaqit.input-development` skill here, against the user's request and session context. Ask the user any blocking question it raises. The skill outputs a `PHASE-PARAMETERS` block.
   If the skill isn't available, skip this step: the subagent then applies its own preloaded copy and returns `NEEDS-INPUT` when a parameter is blocking.
2. **Delegate.** Invoke the `smaqit-development` subagent via the Task tool. Pass the user's request, the session context relevant to it, and the `PHASE-PARAMETERS` block.
3. **Handle return signals.** If the result starts with `NEEDS-INPUT` or `REVIEW-REQUESTED`:
   - Ask the user with the question tool. For `NEEDS-INPUT`, ask each listed question, using its `why` as context. For `REVIEW-REQUESTED`, point the user to the listed specification and design files and ask them to approve or give feedback.
   - Resume the same subagent: send its agent ID a message (SendMessage) containing `answers`, keyed by question id (`review` for a review), plus any `approved_spec_paths` the signal carried.
   - If the subagent cannot be resumed, invoke it again via Task with the same `PHASE-PARAMETERS` block plus `answers` and the signal's `approved_spec_paths`. Nothing a `NEEDS-INPUT` signal reported was written, and approved specifications are skipped, so a fresh run repeats no work.
   - Repeat until the result is neither signal.
4. **Relay.** Show the subagent's final output to the user.

Never answer a signal's questions yourself, and never substitute a default for a question the user must answer.
