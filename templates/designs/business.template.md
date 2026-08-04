---
id: DSG-BUS-[CONCEPT]-USE-CASE
status: draft
created: [TIMESTAMP]
layer: business
diagram_type: use-case
notation: plantuml
specifications:
  - ../../../specs/business/[FILENAME].md
requirements:
  - BUS-[CONCEPT]-001
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
left to right direction
actor "[ACTOR]" as Actor
rectangle "[SYSTEM_BOUNDARY]" {
  usecase "[USE_CASE]" as UseCase
}
Actor --> UseCase
@enduml
```
