# Why Non-Functional Requirements in Business Layer

**Status:** Active Design Decision  
**Date:** 2026-01-19  
**Supersedes:** System Actor pattern (Session 007, 2025-12-16)

---

## Decision

Non-functional requirements (NFRs) driven by stakeholder needs belong in the Business layer, captured as measurable constraints separate from use case flows.

---

## Context

### Original Pattern (System Actor)

In Session 007 (2025-12-16), the "System Actor" pattern was introduced to handle stakeholder requirements about system properties like availability, auditability, and accessibility.

**Intent:** Provide a pattern for non-functional requirements that don't belong to specific human actors.

**Implementation:**
- Business specs could include a "System" actor in the Actors table
- System actor represented "the application as a whole"
- Goals captured "system-level properties stakeholders require"

### Problem Discovered

During v0.6.0-beta validation testing (Task 067, 2026-01-17), the System Actor pattern enabled severe scope creep:

**Evidence from generated Business spec:**
```markdown
| System | Console application runtime environment | Render appropriate output... |

Main flow:
- "System displays colorful console output" ← Stack concern
- "System detects console capabilities (color support, character encoding)" ← Stack concern
- "System displays Mario ASCII art with color formatting" ← Functional behavior

Alternative flows:
- Technical rendering logic
- Console capability detection
- Encoding issues and display limitations
```

**Violations:**
- Technical verbs: "displays," "renders," "detects," "handles"
- Stack artifacts: "console," "terminal," "color support," "character encoding"
- Functional behaviors: "colorful console output," "ASCII art formatting"
- Implementation details: "monochrome fallback," "rendering failures"

### Root Cause

1. **Vague boundaries** — Guidance said "system-level properties" but didn't prohibit behavioral/technical concerns
2. **No enforcement** — No word blacklist or validation to catch spillage
3. **Conceptual mismatch** — Agents interpreted "System" as catch-all for anything non-human
4. **Actor table format** — Using actor table implied behaviors/flows rather than constraints

---

## Rationale for NFRs in Business Layer

### Why Business Layer?

Non-functional requirements are **stakeholder-driven concerns about system-wide properties:**

| NFR Type | Business Driver (WHY) | Example |
|----------|----------------------|---------|
| Availability | Mission-critical operations require continuity | "Must maintain 99.9% uptime for payment processing" |
| Compliance | Regulatory requirements mandate audit capability | "Must provide audit logs for SOX examination" |
| Accessibility | Inclusion policy requires universal access | "Must be usable by screen reader users" |
| Platform | Client environments constrain deployment options | "Must run on Windows where clients operate" |

These are **business justifications**, not technical solutions. They answer "Why does this property matter?" before Functional answers "What behaviors achieve it?"

### Layer Separation

**Business Layer (WHY):**
- States WHAT property is required
- Explains WHY stakeholders need this property
- Defines measurable target or threshold
- Does NOT specify HOW it's achieved

**Functional Layer (WHAT behaviors):**
- Translates NFRs into specific behaviors
- Defines contracts, flows, and error handling
- Specifies WHAT the system must do

**Stack Layer (HOW/WITH WHAT):**
- Selects technologies that enable behaviors
- Specifies tools, frameworks, libraries
- Defines HOW requirements are implemented

### Example: Platform Constraint

**Scenario:** Stakeholder says "This needs to run on Windows, because most our clients are on Windows."

**Business Layer:**
```markdown
## Non-Functional Requirements

| Requirement | Rationale | Target |
|-------------|-----------|--------|
| Windows Platform Support | 85% of client base operates on Windows environments | Windows 10+ compatibility |
```

**Functional Layer:**
```markdown
## Platform Requirements

- FUN-PLATFORM-001: Application must execute on Windows 10 and later versions
- FUN-PLATFORM-002: Application must not require Linux-specific system calls
- FUN-PLATFORM-003: File paths must use Windows-compatible path separators
```

**Stack Layer:**
```markdown
## Platform Technology

- STK-RUNTIME-001: Python 3.8+ (cross-platform, Windows native support)
- STK-DEPS-001: Use pathlib for OS-agnostic path handling
- STK-BUILD-001: Test on Windows 10, Windows 11 environments
```

### Gray Area: Coherence Without Derivation

**Question:** If Functional doesn't derive requirements from Business, how does it know about the Windows constraint?

**Answer:** Layer Independence means each layer receives requirements from its **own prompt file**:

- **Business prompt:** User writes "Must run on Windows (85% client base)"
- **Functional prompt:** User writes "Application must execute on Windows 10+"
- **Stack prompt:** User writes "Python 3.8+ with Windows compatibility"

**Coherence validation** happens at implementation phase:
- Development agent reads all three specs
- Detects if Business says "Windows" but Stack chooses "Linux-only tool"
- Flags incoherence before proceeding

**NFR traceability** is explicit through references:
- Functional spec: "Implements [BUS-PLATFORM-001]"
- Stack spec: "Enables [FUN-PLATFORM-001]"

---

## Implementation

### Pattern Rename

**Old:** "System Actor" (actor table format)  
**New:** "Non-Functional Requirements" (constraint table format)

### Template Structure

```markdown
## Non-Functional Requirements

<!-- System-wide properties driven by stakeholder needs -->
<!-- State WHAT property is required and WHY it matters (not HOW it's achieved) -->

| Requirement | Rationale | Target |
|-------------|-----------|--------|
| [NFR_NAME] | [Why stakeholders need this property] | [Measurable threshold or standard] |
```

### Scope Boundaries

**Business NFRs MUST:**
- State WHAT property is required
- Explain WHY it matters to stakeholders
- Define measurable target or threshold

**Business NFRs MUST NOT:**
- Use behavioral verbs (display, render, output, execute, process, detect, handle)
- Reference technical artifacts (console, terminal, screen, database, API, encoding)
- Describe HOW features work (behaviors belong in Functional layer)
- Include technical error handling or fallback mechanisms

### Warning Signs

If Business specs contain these patterns, they have layer spillage:
- Technical verbs: "displays," "renders," "detects," "handles," "executes"
- Implementation details: "console output," "fallback to monochrome," "ASCII art rendering"
- Error handling flows: "If X fails, do Y"
- Technology mentions: specific tools, frameworks, encodings

---

## Examples

### ✅ Good: Proper Layer Separation

**Business NFR:**
```markdown
| Accessibility | Company inclusion policy requires universal access | WCAG 2.1 AA compliance |
```

**Functional Behavior:**
```markdown
- FUN-ACCESS-001: All interactive elements must have keyboard navigation
- FUN-ACCESS-002: All images must have alternative text descriptions
- FUN-ACCESS-003: Color contrast must meet 4.5:1 ratio minimum
```

**Stack Technology:**
```markdown
- STK-UI-001: HTML semantic elements for screen reader compatibility
- STK-CSS-001: CSS color palette with WCAG-compliant contrast ratios
- STK-TEST-001: axe-core automated accessibility testing
```

### ❌ Bad: Layer Spillage

**Business spec with spillage:**
```markdown
| System | Console application | Display colorful ASCII art with fallback to monochrome |

Main Flow:
1. System detects console capabilities
2. System renders ASCII art with color codes
3. If color unsupported, System displays monochrome version
```

**Problems:**
- "Display," "detects," "renders" → Functional verbs
- "Console," "ASCII art," "color codes" → Stack artifacts
- "If color unsupported" → Functional error handling
- Describes HOW (rendering logic) not WHAT (accessibility property)

**Corrected Business NFR:**
```markdown
| Accessibility | Application must work on various console environments | Support standard terminal capabilities |
```

(Move rendering details to Functional; move color/encoding to Stack)

---

## Migration Guidance

### For Existing Projects

If your Business specs use System Actor pattern:

1. **Review each System Actor entry**
   - Is it a legitimate NFR (availability, compliance, accessibility)?
   - Or is it behavioral spillage (displays, renders, detects)?

2. **Legitimate NFRs → Convert to NFR table:**
   ```markdown
   Before: | System | Application | Maintain 99.9% uptime |
   After:  | Availability | Critical payment operations | 99.9% uptime |
   ```

3. **Behavioral spillage → Move to Functional:**
   - "System displays X" → Functional layer behavior
   - "System detects Y" → Functional layer behavior
   - "If Z fails, fallback" → Functional layer error handling

4. **Technical details → Move to Stack:**
   - "Console capabilities" → Stack layer
   - "Color support" → Stack layer
   - "Character encoding" → Stack layer

### Breaking Change Notice

This is a **breaking change** for projects using System Actor in Business specs.

**Impact:**
- Business agent will no longer recognize System Actor pattern
- Template no longer includes System row in Actors table
- Validation will flag behavioral verbs as scope violations

**Benefit:**
- Prevents layer spillage by design
- Clearer separation of concerns
- More maintainable specifications

---

## Related Decisions

- **Layer Independence** (Session 007, 2025-12-16) — Each layer receives requirements from own prompt file
- **Foundation vs Feature Specs** (Session 007, 2025-12-16) — Foundation specs serve multiple upstream specs
- **Traceability Across Layers** (SMAQIT.md) — References create explicit chains without derivation

---

## Success Criteria

Implementation is successful when:

- [ ] Business specs contain no behavioral verbs (display, render, output, execute)
- [ ] Business specs contain no technical artifacts (console, terminal, database)
- [ ] NFRs state WHAT property and WHY it matters (not HOW it's achieved)
- [ ] Functional layer captures behaviors that implement NFRs
- [ ] Stack layer captures technologies that enable behaviors
- [ ] Layer spillage is caught during validation
- [ ] Test cases demonstrate proper separation

---

## Historical Note

The System Actor pattern served its purpose during early framework development by providing a placeholder for NFRs. As the framework matured and real-world usage revealed scope creep patterns, we evolved to the more explicit Non-Functional Requirements structure with strengthened boundaries.

This evolution demonstrates the framework's commitment to learning from usage and refining principles based on evidence rather than theory.
