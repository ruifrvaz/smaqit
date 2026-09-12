---
id: DSG-DSD-[CONCEPT]-DESIGN-SEQUENCE
status: implemented # Not coupled to the linked specification's lifecycle rank — this diagram is Development-phase output, generated once the spec is already implemented.
created: [TIMESTAMP]
layer: design-sequence
diagram_type: design-sequence
notation: plantuml
realizes: DSG-FUN-[CONCEPT]-SYSTEM-SEQUENCE
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
title DSG-DSD-[CONCEPT]-DESIGN-SEQUENCE
participant "[HANDLER]" as Handler
participant "[SERVICE]" as Service
hide footbox
Handler -> Service: [OPERATION]
' impl: [RELATIVE/PATH/TO/FILE.EXT]:[LINE]
activate Service
Service --> Handler: [RESULT]
deactivate Service
@enduml
```
