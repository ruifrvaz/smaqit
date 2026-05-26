# Zero-to-Prod Workflow — Diagrams

Five complementary views of the `smaqit.new-greenfield-project` orchestrator workflow. Each diagram captures a different aspect of the journey.

---

## 1. Flowchart — The Journey

Best view of the **end-to-end path**: gates, conditional branches (Phase 6), the optional teardown step, and which steps are skills vs agents. Subgraphs group steps by phase.

```mermaid
flowchart TD
    Start([Operator triggers /zero.to.prod]) --> P0

    subgraph P0[Phase 0 — Task Creation Entry]
        P0a[/Operator decides mode/]
        P0b[smaqit.task-create x7]
        P0g{{Gate: tasks created, mode set}}
        P0a --> P0b --> P0g
    end

    P0 --> P1

    subgraph P1[Phase 1 — Requirements]
        P1s1[smaqit.task-start]
        P1s2[smaqit.requirements-extract]
        P1g{{Gate: inventory sufficient}}
        P1s3[smaqit.task-complete]
        P1s1 --> P1s2 --> P1g --> P1s3
    end

    P1 --> P2

    subgraph P2[Phase 2 — Specification]
        P2s1[smaqit.task-start]
        P2a1(["@smaqit.business"])
        P2a2(["@smaqit.functional"])
        P2a3(["@smaqit.stack"])
        P2a4(["@smaqit.infrastructure"])
        P2a5(["@smaqit.coverage"])
        P2g{{Gate: all specs draft + approved}}
        P2s2[smaqit.task-complete]
        P2s1 --> P2a1 --> P2a2 --> P2a3 --> P2a4 --> P2a5 --> P2g --> P2s2
    end

    P2 --> P3

    subgraph P3[Phase 3 — Development]
        P3s1[smaqit.task-start]
        P3a1(["@smaqit.development"])
        P3g{{Gate: build passes, specs implemented}}
        P3s2[smaqit.task-complete]
        P3s1 --> P3a1 --> P3g --> P3s2
    end

    P3 --> P4

    subgraph P4[Phase 4 — Dev Env Sweep]
        P4s1[smaqit.task-start]
        P4a1(["@smaqit.deployment IaC gen"])
        P4s2[smaqit.infrastructure-provision-cyso]
        P4s3[smaqit.infrastructure-vm-bootstrap]
        P4s4[smaqit.infrastructure-deploy-rsync]
        P4s5[smaqit.infrastructure-deploy-verify]
        P4c[commit IaC artifacts]
        P4g{{Gate: deploy-verify PASS}}
        P4s6[smaqit.task-complete]
        P4opt[/Optional: terraform destroy/]
        P4s1 --> P4a1 --> P4s2 --> P4s3 --> P4s4 --> P4s5 --> P4c --> P4g --> P4s6 --> P4opt
    end

    P4 --> P5

    subgraph P5[Phase 5 — Prod via CI/CD]
        P5s1[smaqit.task-start]
        P5s2[smaqit.infrastructure-repo-config]
        P5p[git push origin main]
        P5w[gh run watch]
        P5s3[smaqit.infrastructure-deploy-verify]
        P5g{{Gate: CI/CD success, specs deployed}}
        P5s4[smaqit.task-complete]
        P5s1 --> P5s2 --> P5p --> P5w --> P5s3 --> P5g --> P5s4
    end

    P5 --> P6decision{Domain + DNS ready?}
    P6decision -->|Yes| P6
    P6decision -->|No| P7

    subgraph P6[Phase 6 — Domain + TLS]
        P6s1[smaqit.task-start]
        P6s2[smaqit.infrastructure-domain-tls]
        P6g{{Gate: HTTPS live, renewal ok}}
        P6s3[smaqit.task-complete]
        P6s1 --> P6s2 --> P6g --> P6s3
    end

    P6 --> P7

    subgraph P7[Phase 7 — Validation]
        P7s1[smaqit.task-start]
        P7a1(["@smaqit.validation"])
        P7g{{Gate: validation passes}}
        P7s2[smaqit.task-complete]
        P7s1 --> P7a1 --> P7g --> P7s2
    end

    P7 --> P8

    subgraph P8[Phase 8 — Release]
        P8c[Confirm all tasks closed]
        P8r1[smaqit.release-analysis]
        P8r2[smaqit.release-approval]
        P8r3[smaqit.release-prepare-files]
        P8r4[smaqit.release-git-local]
        P8c --> P8r1 --> P8r2 --> P8r3 --> P8r4
    end

    P8 --> End([App running + release tagged])

    classDef skill fill:#e3f2fd,stroke:#1976d2,color:#0d47a1
    classDef agent fill:#fff3e0,stroke:#f57c00,color:#e65100
    classDef gate fill:#f3e5f5,stroke:#7b1fa2,color:#4a148c
    classDef human fill:#fce4ec,stroke:#c2185b,color:#880e4f

    class P0b,P1s1,P1s2,P1s3,P2s1,P2s2,P3s1,P3s2,P4s1,P4s2,P4s3,P4s4,P4s5,P4s6,P5s1,P5s2,P5s3,P5s4,P6s1,P6s2,P6s3,P7s1,P7s2,P8r1,P8r2,P8r3,P8r4 skill
    class P2a1,P2a2,P2a3,P2a4,P2a5,P3a1,P4a1,P7a1 agent
    class P0g,P1g,P2g,P3g,P4g,P5g,P6g,P7g gate
    class P0a,P4opt,P6decision human
```

Legend: blue = skill, orange = agent, purple = gate, pink = human/conditional.

---

## 2. Sequence Diagram — Who Calls What, In Order

Best view of the **invocation order** between operator, orchestrator, agents, skills, and external systems (GitHub Actions, Cyso VM). Makes the "agent invokes its own input skill internally" pattern explicit.

```mermaid
sequenceDiagram
    autonumber
    actor Op as Operator
    participant Orch as zero-to-prod Orchestrator
    participant TC as smaqit.task-* skills
    participant RE as smaqit.requirements-extract
    participant SA as Spec Agents (5)
    participant Dev as @smaqit.development
    participant Dep as @smaqit.deployment
    participant Prov as smaqit.infrastructure-provision-cyso
    participant Boot as smaqit.infrastructure-vm-bootstrap
    participant Rsync as smaqit.infrastructure-deploy-rsync
    participant Ver as smaqit.infrastructure-deploy-verify
    participant GH as smaqit.infrastructure-repo-config
    participant TLS as smaqit.infrastructure-domain-tls
    participant Val as @smaqit.validation
    participant Rel as smaqit.release-*
    participant CI as GitHub Actions
    participant VM as Cyso VM

    Op->>Orch: trigger /zero.to.prod (mode: assisted/autonomous)

    rect rgb(20,40,70)
    note over Orch,TC: Phase 0 — Task Creation
    Orch->>TC: task-create x7 (one per phase)
    end

    rect rgb(70,40,20)
    note over Orch,RE: Phase 1 — Requirements
    Orch->>TC: task-start (P1)
    Orch->>RE: extract requirements
    RE-->>Orch: inventory + ambiguities
    Orch->>Op: review ambiguities
    Op-->>Orch: resolutions
    Orch->>TC: task-complete (P1)
    end

    rect rgb(20,60,30)
    note over Orch,SA: Phase 2 — Specification
    Orch->>TC: task-start (P2)
    loop business → functional → stack → infrastructure → coverage
        Orch->>SA: invoke agent
        SA->>SA: invoke own input skill internally
        SA-->>Orch: spec written (status: draft)
    end
    Orch->>Op: approve spec set
    Orch->>TC: task-complete (P2)
    end

    rect rgb(60,20,60)
    note over Orch,Dev: Phase 3 — Development
    Orch->>TC: task-start (P3)
    Orch->>Dev: implement specs
    Dev-->>Orch: code + specs status: implemented
    Orch->>TC: task-complete (P3)
    end

    rect rgb(20,35,65)
    note over Orch,VM: Phase 4 — Dev Env Sweep
    Orch->>TC: task-start (P4)
    Orch->>Dep: generate IaC (state: dev/)
    Dep-->>Orch: terraform + workflow files
    Orch->>Prov: provision dev VM
    Prov->>VM: terraform apply
    Prov-->>Orch: fixed_ip
    Orch->>Boot: bootstrap VM
    Boot->>VM: install Docker + .env + /data mount
    Orch->>Rsync: deploy artifacts
    Rsync->>VM: rsync + docker compose
    Orch->>Ver: verify dev
    Ver->>VM: health + SHA + SPA checks
    Ver-->>Orch: PASS
    Orch->>Orch: commit IaC artifacts
    Orch->>TC: task-complete (P4)
    end

    rect rgb(65,45,20)
    note over Orch,CI: Phase 5 — Prod via CI/CD
    Orch->>TC: task-start (P5)
    Orch->>GH: set secrets + variables
    Orch->>CI: git push origin main
    CI->>VM: provision + deploy (prod state)
    CI-->>Orch: workflow run complete
    Orch->>Ver: verify prod
    Ver-->>Orch: PASS
    Orch->>TC: task-complete (P5)
    end

    rect rgb(45,20,65)
    note over Orch,VM: Phase 6 — Domain + TLS (conditional)
    alt domain + DNS ready
        Orch->>TC: task-start (P6)
        Orch->>TLS: configure domain + cert
        TLS->>VM: certbot + nginx reload
        TLS-->>Orch: HTTPS live
        Orch->>TC: task-complete (P6)
    else skipped
        Orch->>Orch: document open item
    end
    end

    rect rgb(20,55,55)
    note over Orch,Val: Phase 7 — Validation
    Orch->>TC: task-start (P7)
    Orch->>Val: run validation
    Val-->>Orch: checks pass
    Orch->>Op: sign-off
    Orch->>TC: task-complete (P7)
    end

    rect rgb(55,55,20)
    note over Orch,Rel: Phase 8 — Release
    Orch->>Orch: confirm all tasks closed
    Orch->>Rel: release-analysis → approval → prepare → git
    Rel-->>Orch: tagged release
    end

    Orch-->>Op: application running + release tagged
```

---

## 3. State Diagram — Phase Transitions + Re-entry

Best view of **re-entry semantics**: each phase is a state with a gate-controlled transition. Highlights the loop-back on `deploy-verify` failure and the Phase 6 skip branch.

```mermaid
stateDiagram-v2
    [*] --> Phase0

    Phase0: Phase 0 — Task Creation\n(operator picks mode)
    Phase1: Phase 1 — Requirements
    Phase2: Phase 2 — Specification
    Phase3: Phase 3 — Development
    Phase4: Phase 4 — Dev Env Sweep
    Phase5: Phase 5 — Prod CI/CD
    Phase6: Phase 6 — Domain + TLS
    Phase7: Phase 7 — Validation
    Phase8: Phase 8 — Release

    Phase0 --> Phase1: tasks created
    Phase1 --> Phase2: inventory ok
    Phase2 --> Phase3: specs approved
    Phase3 --> Phase4: build passes
    Phase4 --> Phase5: deploy-verify PASS\n(IaC committed)
    Phase4 --> Phase4: deploy-verify FAIL\n(fix + retry)
    Phase5 --> Phase6: prod verified\n(domain ready)
    Phase5 --> Phase7: prod verified\n(no domain)
    Phase6 --> Phase7: HTTPS live
    Phase7 --> Phase8: sign-off
    Phase8 --> [*]: release tagged

    note right of Phase0
        Re-entry: any phase can be resumed
        from its first incomplete step.
        Each phase: task-start → work → gate → task-complete
    end note
```

---

## 4. Component / Block Diagram — Architecture View

Best view of **what connects to what across boundaries**: which skills touch which external systems (GitHub, Cyso, DNS, Let's Encrypt). No order — just structural composition.

```mermaid
block-beta
  columns 4

  block:actors:4
    columns 4
    Operator(("Operator"))
    Orchestrator["zero-to-prod orchestrator"]
    CodingAgent["GitHub Coding Agent"]
    space
  end

  space space space space

  block:agents:4
    columns 5
    Business["@smaqit.business"]
    Functional["@smaqit.functional"]
    Stack["@smaqit.stack"]
    Infra["@smaqit.infrastructure"]
    Coverage["@smaqit.coverage"]
    Dev["@smaqit.development"]
    Deploy["@smaqit.deployment"]
    Validation["@smaqit.validation"]
    space
    space
  end

  space space space space

  block:taskSkills:1
    columns 1
    TC["task-create / task-start / task-complete"]
  end
  block:reqSkills:1
    columns 1
    RE["requirements-extract"]
  end
  block:devSkills:1
    columns 1
    DevSkills["(no skills — agent owned)"]
  end
  block:depSkills:1
    columns 1
    Prov["provision-target-cyso"]
  end

  block:bootSkills:1
    columns 1
    Boot["vm-bootstrap"]
  end
  block:rsyncSkills:1
    columns 1
    Rsync["app-deploy-rsync"]
  end
  block:verSkills:1
    columns 1
    Verify["deploy-verify"]
  end
  block:ghSkills:1
    columns 1
    GH["github-repo-config"]
  end

  block:cicdSkills:1
    columns 1
    CICD["cicd-generate (reference)"]
  end
  block:tlsSkills:1
    columns 1
    TLS["domain-tls"]
  end
  block:scaffoldSkills:1
    columns 1
    Scaffold["project-scaffold"]
  end
  block:relSkills:1
    columns 1
    Release["release-analysis / approval / prepare / git"]
  end

  space space space space

  block:external:4
    columns 4
    GitHub["GitHub (Actions + repo)"]
    Cyso["Cyso OpenStack (VM + storage)"]
    DNS["DNS Registrar"]
    LE["Let's Encrypt"]
  end

  Operator --> Orchestrator
  Orchestrator --> Business
  Orchestrator --> Dev
  Orchestrator --> Deploy
  Orchestrator --> Validation
  Orchestrator --> TC
  Orchestrator --> RE
  Deploy --> Prov
  Prov --> Cyso
  Boot --> Cyso
  Rsync --> Cyso
  Verify --> Cyso
  GH --> GitHub
  CICD --> GitHub
  TLS --> DNS
  TLS --> LE
  Release --> GitHub
  CodingAgent --> GitHub
```

---

## 5. Gantt Chart — Phase Sequencing

Best view of **relative phase length and dependencies**. Phase 6 is marked critical/optional. Phase 8 is a milestone. Durations are illustrative.

```mermaid
gantt
    title Zero-to-Prod Workflow — Phase Sequence
    dateFormat X
    axisFormat %s

    section Setup
    Phase 0 — Task Creation         :p0, 0, 1
    Phase 1 — Requirements          :p1, after p0, 2

    section Specs
    Phase 2 — Specification         :p2, after p1, 5

    section Build
    Phase 3 — Development           :p3, after p2, 8

    section Deploy
    Phase 4 — Dev Env Sweep         :p4, after p3, 4
    Phase 5 — Prod via CI/CD        :p5, after p4, 2
    Phase 6 — Domain + TLS (opt)    :crit, p6, after p5, 1

    section Verify + Ship
    Phase 7 — Validation            :p7, after p6, 2
    Phase 8 — Release               :milestone, p8, after p7, 0
```

---

## When to use which diagram

| Diagram | Best for |
|---|---|
| Flowchart | Onboarding new operators; explaining gates + branches |
| Sequence | Debugging "who invokes what"; agent vs skill distinction |
| State | Understanding re-entry and failure loop-back |
| Component | Auditing external system reach (security/compliance) |
| Gantt | Estimating duration; identifying long phases |
