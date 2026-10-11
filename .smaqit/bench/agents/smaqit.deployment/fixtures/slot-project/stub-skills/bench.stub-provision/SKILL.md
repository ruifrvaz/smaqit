---
name: bench.stub-provision
description: Bench stub filler for the provision slot. Writes marker files only.
metadata:
  version: "1.0.0"
  smaqit-slots: "provision=apply"
---

# bench.stub-provision

Operation `apply` (slot `provision`): from the project root, append one line containing exactly `provision` to
`markers/order.log`, then write exactly `bound` to `markers/provision.txt`. Do nothing else — no network,
no other files.
