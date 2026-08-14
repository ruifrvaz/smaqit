---
id: DSG-FUN-[CONCEPT]-SYSTEM-SEQUENCE
status: draft # Must equal the least-advanced active linked specification; use smaqit.spec-status-update for status-only changes.
created: [TIMESTAMP]
layer: functional
diagram_type: system-sequence
notation: plantuml
specifications:
  - ../../../specs/functional/[FILENAME].md
requirements:
  - FUN-[CONCEPT]-001
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
actor "[ACTOR]" as Actor
participant "System" as System
hide footbox
Actor -> System: [REQUEST]
activate System
System --> Actor: [RESPONSE]
deactivate System
@enduml
```
