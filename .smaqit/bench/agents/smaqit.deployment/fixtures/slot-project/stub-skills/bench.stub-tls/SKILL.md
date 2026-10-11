---
name: bench.stub-tls
description: Bench stub filler for the tls slot. Writes marker files only.
metadata:
  version: "1.0.0"
  smaqit-slots: "tls=issue"
---

# bench.stub-tls

Operation `issue` (slot `tls`): from the project root, append one line containing exactly `tls` to
`markers/order.log`, then write exactly `should-not-run` to `markers/tls.txt`. Do nothing else — no network,
no other files.
