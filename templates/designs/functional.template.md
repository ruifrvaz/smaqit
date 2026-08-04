---
id: DSG-FUN-[CONCEPT]-SYSTEM-SEQUENCE
status: draft
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
participant "[SYSTEM]" as System
Actor -> System: [REQUEST]
activate System
System --> Actor: [RESPONSE]
deactivate System
@enduml
```
