# Troubleshooting

Common issues when working with smaqit and their solutions.

## Context Pollution in Multi-Agent Workflows

**Issue:** When invoking multiple layer agents in the same GitHub Copilot session (e.g., `/smaqit.business` → `/smaqit.functional` → `/smaqit.stack`), agents may retain context from previous invocations. This can cause mode confusion where agents reference inappropriate context from other layers.

**When it occurs:** Multi-layer specification workflows where you invoke different agents sequentially in the same Copilot chat session.

**Symptoms:**
- Agent mentions awareness of a previous layer's execution
- Agent references context from a different layer inappropriately
- Uncertainty about which agent mode is currently active
- Agent produces specifications for the wrong layer

**Root Cause:** GitHub Copilot maintains session state across prompt invocations, causing context carryover between agent invocations.

### Workaround

**Recommended:** Start a fresh Copilot chat session between layer invocations
1. Click "New Chat" in the Copilot panel before invoking the next layer agent
2. Context clearing ensures each agent operates in clean mode
3. Each agent will explicitly state its layer at the start of execution

**When fresh session recommended:**
- Always recommended between specification layers (Business → Functional → Stack)
- Especially important when switching between phases (Phase 1 → Phase 2 → Phase 3)
- When working on multiple unrelated features

**When fresh session required:**
- If agent exhibits confusion about current mode
- If agent references inappropriate context from previous invocation
- If agent produces specs for the wrong layer

### Agent Awareness Feature

Starting in v0.5.0, all specification agents explicitly state their layer identity at the start of every response:

- Business Agent: "I am the Business Agent, operating in Business layer mode."
- Functional Agent: "I am the Functional Agent, operating in Functional layer mode."
- Stack Agent: "I am the Stack Agent, operating in Stack layer mode."
- Infrastructure Agent: "I am the Infrastructure Agent, operating in Infrastructure layer mode."
- Coverage Agent: "I am the Coverage Agent, operating in Coverage layer mode."

This helps reduce confusion and makes layer boundaries explicit, even when context pollution occurs.

### Future Solution

**v0.6.0 (planned):** Orchestrator agent pattern (`/smaqit.orchestrator`) will invoke sub-agents with isolated contexts, eliminating this issue entirely. The orchestrator will coordinate all layer agents in a single invocation while maintaining proper context isolation.

---

## Other Common Issues

*More troubleshooting guidance will be added as common issues are identified.*

### Getting Help

If you encounter issues not covered here:
1. Check [Issues](https://github.com/ruifrvaz/smaqit/issues) for similar problems
2. Review the [documentation](../../README.md) for additional guidance
3. Open a new issue with details about your problem
