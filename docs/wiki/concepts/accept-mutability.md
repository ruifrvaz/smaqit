# Accept Mutability, Validate Behavior

## Definition

Accept Mutability, Validate Behavior is the principle that LLM-generated artifacts (code, configs, specs) may vary between runs, but the behavior they produce (passing acceptance criteria) should be consistent. smaqit embraces non-determinism in artifacts while enforcing determinism in outcomes.

## The Challenge

LLMs are non-deterministic:
- Same prompt rarely generates identical output twice
- Code style varies (spacing, naming, comments)
- Documentation phrasing differs
- Implementation approaches change

Attempting to fight this variance is futile and counterproductive.

## The Solution

**Accept that artifacts are mutable:**
- Code structure may vary
- Variable names may differ
- Comments may be reworded
- File organization may change

**Validate that behavior is immutable:**
- Acceptance criteria pass or fail consistently
- Functionality works as specified
- Tests verify requirements
- Performance meets thresholds

## How It Works in smaqit

### Specifications Define Behavior

Specs include testable acceptance criteria:

```markdown
## Acceptance Criteria

- [ ] BUS-LOGIN-001: User can authenticate with valid credentials
- [ ] BUS-LOGIN-002: System rejects invalid credentials
- [ ] BUS-LOGIN-003: Account locks after 5 failed attempts
```

These criteria define **what must be true**, not **how to make it true**.

### Implementation Varies

Development agent might generate:

**Run 1:**
```python
def authenticate(username: str, password: str) -> Token:
    user = db.get_user(username)
    if user and verify_password(password, user.password_hash):
        return generate_token(user)
    raise AuthenticationError()
```

**Run 2:**
```python
def authenticate(credentials: dict) -> Token:
    username, password = credentials['user'], credentials['pass']
    if user := db.find_by_username(username):
        if bcrypt.verify(password, user.hashed_password):
            return create_jwt_token(user.id)
    raise InvalidCredentialsError()
```

Different code, same behavior: both satisfy BUS-LOGIN-001.

### Validation Focuses on Outcomes

Validation agent tests against specs:

```gherkin
# COV-LOGIN-001: Maps to BUS-LOGIN-001
Scenario: Successful authentication
  Given a user with valid credentials
  When the user submits login request
  Then the system returns an authentication token
```

Test passes with both implementations. Code variance doesn't matter.

## Why This Matters

**Reduces false negatives:**
- Don't fail because variable names changed
- Don't fail because comments differ
- Don't fail because of stylistic variations

**Focuses on what matters:**
- Does it meet requirements? (functional correctness)
- Does it pass tests? (behavioral validation)
- Does it satisfy acceptance criteria? (stakeholder value)

**Enables iteration:**
- Can regenerate code without breaking workflows
- Can refactor without spec changes
- Can optimize without re-validation

## Trade-offs

**Benefits:**
- Works with LLM non-determinism instead of against it
- Reduces brittleness in validation
- Focuses effort on meaningful validation

**Costs:**
- Can't rely on exact output matching
- Code reviews focus on behavior, not style
- Version control shows more churn

The cost is acceptable—behavioral correctness is more valuable than artifact consistency.

## Related

- [Template Constraints](../designs/template-constraints.md) — How templates reduce variance in specs
- [Self-Validating Agents](self-validating-agents.md) — How agents verify behavior, not artifacts
