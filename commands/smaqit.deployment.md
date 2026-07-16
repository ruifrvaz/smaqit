---
description: Orchestrate the Deployment phase — Infrastructure specs, then deploy to the target environment.
---

Invoke the `smaqit-deployment` subagent via the Task tool, passing the user's current request and full session context as the subagent's input. Relay the subagent's complete output back to the user, including any clarifying questions it raises.
