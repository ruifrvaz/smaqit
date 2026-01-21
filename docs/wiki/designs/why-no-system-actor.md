# Why No System Actor

**Status:** Implemented (2026-01-21)  
**Category:** Layer Boundaries  
**Related:** Task 068, Session 007 (System Actor introduction), Session 042 (System Actor removal)

## Overview

The Business layer does not use a "System Actor" pattern. This document explains why the pattern was introduced, why it was removed, and how to properly capture non-functional requirements and system properties.

## Historical Context

### Introduction (Session 007, 2025-12-16)

The System Actor pattern was introduced to handle stakeholder requirements about system properties like availability, auditability, and accessibility—concerns that didn't seem to belong to specific human actors.

**Original pattern:**
```markdown
| Actor | Description | Goals |
|-------|-------------|-------|
| System | The application as a whole | [System-level properties stakeholders require] |
```

**Original rationale:** Provide a pattern for non-functional requirements that don't fit with interactive user actors.

### The Problem (Discovered Task 067, 2026-01-17)

During v0.6.0-beta validation testing, the Business Agent generated a specification with severe scope creep:

**Example violations:**
```markdown
| System | Console application runtime environment | Render appropriate output based on console capabilities and exit cleanly |
```

**Main flow included:**
- "System detects console capabilities (color support, character encoding)" ← Stack layer
- "System displays Mario ASCII art with color formatting" ← Functional layer  
- "System outputs colorful console text to enhance visual experience" ← Stack layer

**Alternative flows included:**
- Technical rendering logic
- Console capability detection
- Encoding issues and display limitations

### Why It Failed

1. **No enforcement mechanism** — Guidance said "system-level properties stakeholders require" but didn't prevent behavioral/technical concerns from leaking in

2. **Vague boundaries** — Agents interpreted "System" as a catch-all for anything non-human, allowing Functional and Stack concerns to pollute Business specs

3. **Conceptual mismatch** — Stakeholders don't think about "the system" as an actor. They think about outcomes they need and properties they require.

4. **Enabled scope creep by design** — The pattern provided a convenient dumping ground for anything that didn't fit cleanly into user actors

## The Solution: Named Stakeholder Actors

**Principle:** Non-functional requirements are expressed as actor goals in the Business layer, using named stakeholders rather than a generic "System" actor.

### Actor Concept

An actor is anyone who cares about some aspect of what is being built. Actors have goals—what they want to achieve or what properties they require.

**Actor goals may express:**
- **Interactive outcomes** — What an actor wants to accomplish through using the system
- **System properties** — What constraints or qualities an actor requires the system to have
- **Success criteria** — What measurable outcomes matter to an actor

### Actor Diversity Examples

**Instead of "System Actor", use specific stakeholders:**

| NFR Type | Wrong Pattern | Right Pattern |
|----------|--------------|---------------|
| Reliability | System maintains 99.9% uptime | **Operations Team** — Infrastructure operators — System maintains 99.9% uptime for business continuity |
| Compliance | System provides audit logs | **Compliance Officer** — Regulatory stakeholder — All data access logged for audit trail per GDPR requirements |
| Accessibility | System supports screen readers | **Accessibility Advocate** — Inclusion policy stakeholder — Application accessible to screen reader users |
| Performance | System responds in < 2 seconds | **Product Owner** — Business stakeholder — Users experience responsive interactions (< 2s response time) |
| Security | System encrypts data at rest | **Security Team** — Risk management stakeholder — Sensitive data protected through encryption |

### Key Differences

**System Actor (removed):**
- Generic, catch-all pattern
- Encourages dumping technical concerns
- No clear stakeholder ownership
- Vague boundaries enable scope creep

**Named Stakeholder Actors (current):**
- Specific, accountable stakeholders
- Each actor has clear motivations
- Properties expressed as stakeholder needs
- Natural boundary: if you can't name who cares, it may not be a business requirement

## Where Do Technical Details Belong?

The Business layer captures **WHAT** stakeholders need and **WHY** it matters. Technical details about **HOW** to satisfy those needs belong downstream.

### Functional Layer

Translates stakeholder goals into specific behaviors with measurable criteria:

**Business Layer:**
```markdown
**Operations Team** — System maintains 99.9% uptime for business continuity
```

**Functional Layer:**
```markdown
**Availability Behavior:**
- Application responds to health checks within 1 second
- Failed requests automatically retry with exponential backoff
- Graceful degradation when dependencies unavailable
- Circuit breaker pattern for external service failures

**Acceptance Criteria:**
- FUN-AVAILABILITY-001: Health endpoint returns 200 within 1s
- FUN-AVAILABILITY-002: Failed requests retry up to 3 times
- FUN-AVAILABILITY-003: Circuit opens after 5 consecutive failures
```

### Stack Layer

Selects technologies that enable the functional behaviors:

**Stack Layer:**
```markdown
**Availability Stack:**
- Container orchestration: Kubernetes for automatic restarts
- Load balancer: nginx for health check routing
- Circuit breaker library: resilience4j
- Monitoring: Prometheus + Grafana for uptime metrics
```

### Infrastructure Layer

Defines where and how the application runs:

**Infrastructure Layer:**
```markdown
**Availability Infrastructure:**
- Multi-zone deployment (3 availability zones)
- Auto-scaling: 2-10 replicas based on CPU/memory
- Load balancer with health checks every 10s
- 99.9% uptime SLA monitoring alerts
```

## Common Misconceptions

### "NFRs are technical, so they belong in Functional layer"

**Wrong.** NFRs express stakeholder needs and success criteria, making them Business layer concerns. The **behaviors** that satisfy NFRs belong in Functional layer.

**Example:**
- **Business:** Operations Team requires 99.9% uptime for business continuity (stakeholder need)
- **Functional:** Health checks, retries, circuit breakers (behaviors satisfying need)
- **Stack:** Kubernetes, nginx, resilience4j (technologies enabling behaviors)

### "System properties have no human stakeholder"

**Wrong.** Every system property has a stakeholder who cares about it:
- Availability → Operations Team, End Users
- Security → Security Team, Compliance Officer, End Users
- Performance → Product Owner, End Users
- Accessibility → Accessibility Advocate, Legal Team, End Users

If you can't identify who cares about a property, question whether it's actually a requirement.

### "Business layer shouldn't have technical terms like 'uptime' or 'latency'"

**Context matters.** Measurable properties are acceptable when they're stakeholder-facing success metrics:
- ✅ "99.9% uptime" — Measurable outcome Operations Team requires
- ✅ "Response time < 2 seconds" — Observable user experience criterion
- ❌ "Circuit breaker pattern" — Technical implementation mechanism
- ❌ "ANSI color codes" — Technical artifact detail

## Validation: Does My Business Spec Have Scope Creep?

**Red flags indicating Functional/Stack layer pollution:**

**Technical verbs (describing HOW):**
- display, render, output, execute, process, detect, handle, parse, format
- **If your spec says "System displays..." → Probably scope creep**

**Technical artifacts:**
- console, terminal, screen, database, API, server, client, encoding, color support
- **If your spec mentions implementation artifacts → Probably scope creep**

**Behavioral mechanisms:**
- How features work internally, technical error handling, fallback behaviors
- **If your spec describes mechanisms → Belongs in Functional layer**

**Ask:** Does this describe what a stakeholder needs (outcome) or how to build it (mechanism)?

## Migration Guidance

If you have existing Business specs with System Actor:

1. **Identify the real stakeholder** — Who actually cares about this property?
2. **Express as stakeholder goal** — What outcome do they need?
3. **Make it measurable** — What observable criteria define success?
4. **Move technical details downstream** — Behaviors → Functional, Technologies → Stack

**Example transformation:**

**Before (with System Actor):**
```markdown
| Actor | Description | Goals |
|-------|-------------|-------|
| System | Console application | Display colorful output with fallback to monochrome |

**Main Flow:**
1. System detects console color support
2. System renders Mario ASCII art with ANSI colors
3. System falls back to monochrome if colors unsupported
```

**After (named stakeholders):**
```markdown
| Actor | Description | Goals |
|-------|-------------|-------|
| Mario Fan | Nintendo franchise enthusiast | Experience authentic Mario character greeting |
| Accessibility Advocate | Inclusion policy stakeholder | Application works regardless of console capabilities |

**Main Flow:**
1. Mario Fan runs application
2. Mario Fan experiences authentic Mario greeting
3. Application completes successfully

**Success Metrics:**
- Users recognize Mario character
- Application works on all standard terminals
```

*Technical details about color rendering, ASCII art, and fallback logic belong in Functional/Stack specs.*

## Summary

- **No System Actor** — Use named stakeholders with specific goals
- **NFRs are Business concerns** — Express as stakeholder needs with measurable criteria
- **Technical details go downstream** — Behaviors in Functional, technologies in Stack
- **Every requirement has a stakeholder** — If you can't name who cares, question if it's needed
- **Focus on WHAT and WHY** — Business describes outcomes, not mechanisms

The System Actor pattern was well-intentioned but enabled scope creep by design. Named stakeholder actors provide clearer boundaries and better traceability while still capturing all necessary requirements.
