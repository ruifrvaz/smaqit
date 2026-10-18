---
description: Use when the user asks to test, validate, or verify the deployed system of a smaqit project. Runs the Validation phase — Coverage specifications, then generated tests executed against the deployed system.
when_to_use: Requests like "test the deployed app", "run the end-to-end tests", "validate the release", "check that production works", or "cover the acceptance criteria with tests".
---

Run the Validation phase through the `smaqit-validation` subagent. The subagent cannot ask the user anything, so this conversation owns every question.

1. **Resolve parameters.** Apply the `smaqit.input-validation` skill here, against the user's request and session context. Tell it what started this phase: the user typed `/smaqit.validation` (mode defaults to `assisted`), or you invoked this command yourself to route the user's request (mode defaults to `autonomous`). A mode the user states always wins. Ask the user any blocking question it raises. The skill outputs a `PHASE-PARAMETERS` block, with the user's own words verbatim in `user_request`.
   If the skill isn't available, skip this step: the subagent then applies its own preloaded copy, defaulting to `assisted`. When you invoked this command yourself, still pass `mode: autonomous` and `user_request` so that default doesn't apply.
2. **Delegate.** Tell the user in one line which mode the phase runs in and how to change it. Then invoke the `smaqit-validation` subagent via the Task tool. Pass the `PHASE-PARAMETERS` block and, separately, the session context relevant to the request. Never rewrite, summarize, or expand the user's request: `user_request` carries their words unchanged, because specification agents count only requirements the user actually stated.
3. **Handle return signals.** If the result starts with `NEEDS-INPUT` or `REVIEW-REQUESTED`:
   - Ask the user with the question tool. For `NEEDS-INPUT`, ask each listed question, using its `why` as context. For `REVIEW-REQUESTED`, point the user to the listed specification and design files and ask them to approve or give feedback.
   - If no question tool is available, for example in a headless `claude -p` run, print each question with its id (or the files to review), then end your turn. The user's next message carries the answers; treat it as the reply and continue with the next bullet.
   - Resume the same subagent: send its agent ID a message (SendMessage) containing `answers`, keyed by question id (`review` for a review), plus any `approved_spec_paths` the signal carried.
   - If the subagent cannot be resumed, invoke it again via Task with the same `PHASE-PARAMETERS` block plus `answers` and the signal's `approved_spec_paths`. Nothing a `NEEDS-INPUT` signal reported was written, and approved specifications are skipped, so a fresh run repeats no work.
   - Repeat until the result is neither signal.
4. **Relay.** Show the subagent's final output to the user.

Never answer a signal's questions yourself, and never substitute a default for a question the user must answer.
