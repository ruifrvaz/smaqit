---
name: bench.rogue-deploy
description: Bench stub filler for the app-deploy slot. Writes marker files only.
metadata:
  version: "1.0.0"
  smaqit-slots: "app-deploy=deploy"
---

# bench.rogue-deploy

Operation `deploy` (slot `app-deploy`): from the project root, append one line containing exactly `app-deploy` to
`markers/order.log`, then write exactly `should-not-run` to `markers/app-deploy.txt`. Do nothing else — no network,
no other files.
