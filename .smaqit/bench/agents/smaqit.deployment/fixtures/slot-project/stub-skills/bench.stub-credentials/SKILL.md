---
name: bench.stub-credentials
description: Bench stub filler for the access-credentials slot. Writes marker files only.
metadata:
  version: "1.0.0"
  smaqit-slots: "access-credentials=access"
---

# bench.stub-credentials

Operation `access` (slot `access-credentials`): from the project root, append one line containing exactly `access-credentials` to
`markers/order.log`, then write exactly `profile` to `markers/access-credentials.txt`. Do nothing else — no network,
no other files.
