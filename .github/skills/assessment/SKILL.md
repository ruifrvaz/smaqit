---
name: assessment
description: Critical assessment skill for handling ambiguous requirements, conflicting inputs, and insufficient detail in complex planning scenarios. Provides approval gate with iterative refinement to prevent wasted execution on poor-quality inputs. Invoke when user explicitly requests assessment using words "assess" or "assessment" or when requirements are ambiguous, conflicting, or underspecified.
---

# Assessment

Assessment prevents wasted execution on poor-quality inputs by providing:

- Critical evaluation of input quality and completeness
- Identification of assumptions, gaps, and conflicts
- Generation of execution plan with explicit steps
- User review and approval mechanism
- Iterative refinement based on feedback

## Workflow

1. **Question the premise**
   - Evaluate necessity
   - Check for duplication
   - Identify hidden assumptions
   - Question maintenance burden

2. **Check existing state**
   - Read relevant files first
   - Search for similar patterns
   - Verify problem exists as described
   - Empirically verify without guessing

3. **Identify trade-offs**
   - List downsides of proposed approach
   - Compare alternatives
   - Analyze cost-benefit
   - Consider token/complexity/maintenance costs

4. **Flag problems upfront**
   - State flaws clearly before proceeding
   - Surface redundancy or conflicts
   - Identify better approaches
   - Question if simpler solutions suffice

5. **Present assessment and request direction**
   - Summarize findings with rationale
   - Present alternatives with analysis
   - Request confirmation before proceeding
   - Challenge incomplete framing

## Stop and Explain Risks

Before implementing anything that:
- Modifies user configuration files (dotfiles, shell configs)
- Violates security best practices
- Breaks established conventions without clear justification
- Could affect system stability or user experience negatively
- Duplicates existing functionality in another location

Stop immediately and explain the risks explicitly before proceeding.

## Output

Return structured results including:

- **Premise Evaluation**: Summary of necessity analysis, duplication checks, and hidden assumptions
- **Current State Findings**: Verified existing patterns, confirmed problem scope (empirically validated, not assumed)
- **Trade-off Analysis**: Comparison of alternatives with cost-benefit analysis and recommendations
- **Flagged Problems**: Clear statement of flaws, conflicts, redundancy requiring attention
- **Execution Plan**: Explicit steps with coordination requirements for complex work
- **Approval Status**: User confirmation received or autonomous selection made with rationale
