---
id: DSG-INF-[CONCEPT]-DEPLOYMENT
status: draft # Must equal the least-advanced active linked specification; use smaqit.spec-status-update for status-only changes.
created: [TIMESTAMP]
layer: infrastructure
diagram_type: deployment
notation: plantuml
specifications:
  - ../../../specs/infrastructure/[FILENAME].md
requirements:
  - INF-[CONCEPT]-001
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
node "[ENVIRONMENT]" {
  node "[COMPUTE]" {
    artifact "[DEPLOYABLE]" as Deployable
  }
}
cloud "[EXTERNAL_SYSTEM]" as External
Deployable --> External : [PROTOCOL]
@enduml
```
