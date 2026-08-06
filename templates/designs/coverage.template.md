---
id: DSG-COV-[CONCEPT]-REQUIREMENT-TRACE
status: draft # Must equal the least-advanced active linked specification; use smaqit.spec-status-update for status-only changes.
created: [TIMESTAMP]
layer: coverage
diagram_type: requirement-trace
notation: plantuml
specifications:
  - ../../../specs/coverage/[FILENAME].md
requirements:
  - COV-[CONCEPT]-001
source_sha256: "[SOURCE_SHA256]"
image_sha256: "[IMAGE_SHA256]"
visual_validation:
  status: pending
  validated_at: null
  source_sha256: null
  image_sha256: null
---

```plantuml
@startuml
rectangle "[SOURCE_REQUIREMENT]" as Requirement
rectangle "[TEST_CASE]" as TestCase
rectangle "[EXPECTED_OUTCOME]" as Outcome
Requirement --> TestCase : verified by
TestCase --> Outcome : proves
@enduml
```
