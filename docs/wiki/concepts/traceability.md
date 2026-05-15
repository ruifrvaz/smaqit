# Traceability

## Definition

Traceability is the ability to track every requirement from its origin (session context) through all layers to its verification (test case). In smaqit, every acceptance criterion, implementation decision, and test case can be traced back to its source.

## How It Works

### Requirement Identifiers

Every acceptance criterion has a unique ID following the pattern:
```
[LAYER_PREFIX]-[CONCEPT]-[NNN]
```

Examples:
- `BUS-LOGIN-001`: Business requirement for login
- `FUN-AUTH-TOKEN-001`: Functional requirement for token behavior
- `STK-FRAMEWORK-001`: Stack choice for framework
- `INF-SCALING-001`: Infrastructure scaling rule
- `COV-LOGIN-001`: Test case verifying BUS-LOGIN-001

### Reference Chains

Specs reference upstream specs explicitly:

```markdown
## References

### Implements
- [BUS-LOGIN](../business/uc1-login.md) — Implements login use case

### Enables
- [BUS-CHECKOUT](../business/uc2-checkout.md) — Requires authenticated session
- [BUS-PROFILE](../business/uc3-profile.md) — Requires authenticated session
```

This creates an explicit chain: Business → Functional → Stack → Infrastructure → Coverage

### Coverage Maps

Coverage specs create explicit mappings:

| Requirement ID | Source Spec | Test Case ID | Expected Outcome |
|----------------|-------------|--------------|------------------|
| BUS-LOGIN-001 | business/uc1-login.md | COV-LOGIN-001 | User can authenticate |
| FUN-AUTH-TOKEN-001 | functional/auth.md | COV-TOKEN-001 | Token expires after 24h |

## Why Traceability Matters

**Impact analysis:**
- When BUS-LOGIN-001 changes, find all referencing specs
- See which functional behaviors, stack choices, and tests depend on it
- Update downstream artifacts explicitly

**Coverage verification:**
- Count requirements vs test cases
- Identify untested requirements (gaps)
- Report spec coverage percentage: `(tested / total) × 100`

**Audit trail:**
- Every decision traces to a requirement
- Every requirement traces to user session input
- Complete chain from user input to validation

**Change management:**
- Know what breaks when requirements change
- Update all affected specs systematically
- No silent omissions or forgotten dependencies

## Example

**Traceability chain for authentication:**

```
User Input (session context)
  ↓
BUS-LOGIN-001: User can authenticate with valid credentials
  ↓ (Implements)
FUN-AUTH-001: System issues JWT token upon successful login
  ↓ (Enables)
STK-JWT-001: Use jsonwebtoken library v9.0+
  ↓ (Referenced by)
INF-SECRETS-001: Store JWT secret in environment variable
  ↓ (Verified by)
COV-LOGIN-001: Test successful authentication returns token
```

Each step is explicit. Each ID is unique. Each reference is documented.

## Traceability Matrix

For complex projects, maintain a cross-reference:

| Business | Functional | Stack | Infrastructure | Coverage |
|----------|------------|-------|----------------|----------|
| BUS-LOGIN-001 | FUN-AUTH-001 | STK-JWT-001 | INF-SECRETS-001 | COV-LOGIN-001 |
| BUS-CHECKOUT-002 | FUN-CART-002, FUN-AUTH-001 | STK-PAYMENT-002 | INF-PCI-002 | COV-CHECKOUT-002 |

Shows which requirements span multiple layers and how they connect.

## Related

- [Explicit Over Implicit](../designs/explicit-over-implicit.md) — Why references are mandatory
- [Progressive Refinement](../designs/progressive-refinement.md) — How layers build on each other
