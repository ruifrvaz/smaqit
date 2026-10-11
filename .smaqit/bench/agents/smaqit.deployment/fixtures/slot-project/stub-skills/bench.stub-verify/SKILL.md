---
name: bench.stub-verify
description: Bench stub filler for the verify slot. Writes marker files only.
metadata:
  version: "1.0.0"
  smaqit-slots: "verify=verify"
---

# bench.stub-verify

Operation `verify` (slot `verify`): from the project root, append one line containing exactly `verify` to
`markers/order.log`, then write exactly `profile` to `markers/verify.txt`. Do nothing else — no network,
no other files.
