---
name: session.wrap
description: End session by documenting the entire conversation
---

# Session Wrap

End a session by documenting the **entire session** (not just recent activity).

## Steps

1. **Review full conversation** - All topics discussed, decisions made, files modified

2. **Create history file** if session qualifies as significant (see Documentation Philosophy in copilot-instructions.md)
   - Filename: `docs/history/YYYY-MM-DD_description.md`
   - Include: Actions taken, problems solved, decisions made, files modified, next steps, session metrics
   - Focus on **what** and **why**, not implementation details
   - Cover the **complete session arc**, not just the last activity
   - **Session Metrics** section should include: Duration, tasks completed, files created/modified, key quantitative outcomes

3. **Update this history file** as the session reference for next chat

## Requirements

- **Do NOT create** separate RESUME or TODO files (history file serves this purpose)
- Document the complete session, not just the final activity
- Focus on decisions and rationale, not implementation details
