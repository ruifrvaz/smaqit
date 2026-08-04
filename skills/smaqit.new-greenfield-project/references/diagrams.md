# Zero-to-Prod Workflow — Diagrams

Five complementary PlantUML views of the `smaqit.new-greenfield-project` orchestrator workflow.

## 1. Activity — End-to-End Journey

```plantuml
@startuml
start
:Phase 0 — Create phase tasks and select mode;
:Phase 1 — Extract and approve requirements;
:Phase 2 — Generate five spec layers;
:Render and visually approve linked design PNGs;
if (Specs and designs approved?) then (yes)
  :Phase 3 — Build and test application;
else (no)
  :Return to owning specification agent;
  stop
endif
:Phase 4 — Provision and verify dev environment;
while (Dev deploy verified?) is (no)
  :Correct deployment and retry;
endwhile (yes)
:Phase 5 — Deploy production through CI/CD;
if (Domain and DNS ready?) then (yes)
  :Phase 6 — Configure domain and TLS;
else (no)
  :Record domain/TLS open item;
endif
:Phase 7 — Execute validation phase;
:Phase 8 — Analyze, approve, prepare, and release;
stop
@enduml
```

## 2. Sequence — Invocation Order

```plantuml
@startuml
autonumber
actor Operator
participant "zero-to-prod\nOrchestrator" as Orch
participant "task-* skills" as Tasks
participant "requirements-extract" as Req
participant "Specification Agents" as Specs
participant "design-validate" as Design
participant "Development" as Dev
participant "Deployment" as Deploy
participant "GitHub Actions" as CI
participant "Validation" as Validate
participant "release-* skills" as Release
participant "Target VM" as VM

Operator -> Orch: start workflow (assisted/autonomous)
Orch -> Tasks: create phase tasks
Orch -> Req: extract requirements
Req --> Orch: inventory and ambiguities
Orch -> Operator: resolve blocking ambiguities
Operator --> Orch: approved requirements
loop business -> functional -> stack -> infrastructure -> coverage
  Orch -> Specs: invoke owning agent
  Specs -> Design: render, open PNG, attest, validate
  Design --> Specs: current validated pair
  Specs --> Orch: draft spec and design paths
end
Orch -> Operator: approve spec/design set
Orch -> Dev: implement prevalidated specs/designs
Dev --> Orch: build passes; lifecycle synchronized
Orch -> Deploy: generate infrastructure and deploy dev
Deploy -> VM: provision, bootstrap, deploy, verify
VM --> Deploy: verification result
Deploy --> Orch: dev PASS
Orch -> CI: push production deployment
CI -> VM: deploy production
CI --> Orch: workflow complete
Orch -> Validate: execute coverage specs/designs
Validate --> Orch: validation PASS
Orch -> Release: analysis -> approval -> prepare -> git
Release --> Operator: tagged release
@enduml
```

## 3. State — Gates and Re-entry

```plantuml
@startuml
[*] --> TaskCreation
TaskCreation : Phase 0 — task creation and mode
Requirements : Phase 1 — requirements
Specification : Phase 2 — specs and visual designs
Development : Phase 3 — development
DevEnvironment : Phase 4 — dev environment sweep
Production : Phase 5 — production CI/CD
DomainTLS : Phase 6 — domain and TLS
Validation : Phase 7 — validation
Release : Phase 8 — release

TaskCreation --> Requirements : tasks created
Requirements --> Specification : inventory sufficient
Specification --> Development : specs/designs approved
Development --> DevEnvironment : build passes
DevEnvironment --> DevEnvironment : deploy verification fails / correct
DevEnvironment --> Production : deploy verification passes
Production --> DomainTLS : production verified and domain ready
Production --> Validation : production verified; domain deferred
DomainTLS --> Validation : HTTPS live
Validation --> Release : sign-off
Release --> [*] : release tagged

note right of Specification
  Resume any phase at its first incomplete gate.
  Missing, stale, or unreadable design PNGs block progress.
end note
@enduml
```

## 4. Component — Responsibility Boundaries

```plantuml
@startuml
left to right direction
actor Operator
component "Zero-to-Prod\nOrchestrator" as Orch
component "Specification Agents\nBusiness / Functional / Stack\nInfrastructure / Coverage" as Specs
component "PlantUML MCP +\nVisual Design Gate" as Design
component "Development Agent" as Dev
component "Deployment Agent" as Deploy
component "Validation Agent" as Validate
component "Task + Release Skills" as Skills
cloud "GitHub\nRepository + Actions" as GitHub
cloud "Cyso OpenStack" as Cyso
cloud "DNS Registrar" as DNS
cloud "Let's Encrypt" as LE
node "Target VM" as VM

Operator --> Orch
Orch --> Specs
Specs --> Design
Orch --> Dev
Orch --> Deploy
Orch --> Validate
Orch --> Skills
Skills --> GitHub
Deploy --> GitHub
Deploy --> Cyso
Cyso --> VM
Deploy --> VM
Deploy --> DNS
Deploy --> LE
Validate --> VM
@enduml
```

## 5. Gantt — Relative Phase Sequencing

Durations are illustrative.

```plantuml
@startgantt
Project starts 2026-01-01
[Phase 0 — Task Creation] lasts 1 day
[Phase 1 — Requirements] lasts 2 days
[Phase 1 — Requirements] starts at [Phase 0 — Task Creation]'s end
[Phase 2 — Specs + Visual Designs] lasts 5 days
[Phase 2 — Specs + Visual Designs] starts at [Phase 1 — Requirements]'s end
[Phase 3 — Development] lasts 8 days
[Phase 3 — Development] starts at [Phase 2 — Specs + Visual Designs]'s end
[Phase 4 — Dev Environment] lasts 4 days
[Phase 4 — Dev Environment] starts at [Phase 3 — Development]'s end
[Phase 5 — Production CI/CD] lasts 2 days
[Phase 5 — Production CI/CD] starts at [Phase 4 — Dev Environment]'s end
[Phase 6 — Domain + TLS] lasts 1 day
[Phase 6 — Domain + TLS] starts at [Phase 5 — Production CI/CD]'s end
[Phase 7 — Validation] lasts 2 days
[Phase 7 — Validation] starts at [Phase 6 — Domain + TLS]'s end
[Phase 8 — Release] happens at [Phase 7 — Validation]'s end
@endgantt
```

## Selection Guide

| View | Best for |
|---|---|
| Activity | Onboarding and gate/branch explanation |
| Sequence | Invocation ownership and ordering |
| State | Re-entry and failure-loop behavior |
| Component | External-system reach and security review |
| Gantt | Relative duration and dependency planning |
