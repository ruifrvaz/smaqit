---
id: DSG-STK-[CONCEPT]-COMPONENT
status: draft # Must equal the least-advanced active linked specification; use smaqit.spec-status-update for status-only changes.
created: [TIMESTAMP]
layer: stack
diagram_type: component
notation: plantuml
specifications:
  - ../../../specs/stack/[FILENAME].md
requirements:
  - STK-[CONCEPT]-001
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
title DSG-STK-[CONCEPT]-COMPONENT
component "[COMPONENT]" as Component
component "[DEPENDENCY]" as Dependency
Component --> Dependency : [INTERFACE]
@enduml
```
