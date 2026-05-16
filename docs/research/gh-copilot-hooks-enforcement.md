# Research: GitHub Copilot Hooks as Enforcement Layer

**Source:** [GitHub Copilot hooks reference](https://docs.github.com/en/copilot/reference/hooks-reference)  
**Reviewed:** 2026-05-16  
**Purpose:** Assess hooks as a deterministic enforcement layer for smaqit orchestration validation gates (Task 082).  
**Related:** `docs/research/ms-ai-agent-orchestration-patterns.md` — identifies the gaps this layer addresses.

---

## Background

Task 082 implements orchestration-first phase agents via agent instructions. Instructions are LLM-interpreted — compliance is expected but not enforced. Two concerns in particular are fragile at the instruction level:

- **Inter-layer spec validation:** the orchestrator should not proceed to the next spec layer if the previous agent produced incomplete output. An instruction saying "verify before proceeding" can be ignored or misjudged.
- **Iteration caps:** the orchestrator should stop retrying after 3 attempts per layer. An instruction saying "max 3 retries" can drift.

GitHub Copilot Hooks execute shell commands at specific lifecycle points. They operate outside the LLM — their decisions are not subject to model interpretation.

---

## Relevant Hook Events

### `subagentStop`

Fires after each subagent completes a turn. Supports decision control:

| Output field | Values | Effect |
|---|---|---|
| `decision` | `"block"` | Forces another orchestrator turn; `reason` is injected as the prompt |
| `decision` | `"allow"` or empty | Orchestrator continues normally |
| `reason` | string | Required when `decision: "block"`; passed to the orchestrator as context |

This is the primary gate event. A shell script fires after each spec agent, reads the written spec file, and either passes (`{}`) or blocks with structured feedback.

### `subagentStart`

Fires before each subagent is spawned. Supports `additionalContext` in the output, which is prepended to the subagent's initial prompt. A hook script reads upstream spec files from disk and injects them directly — implementing scoped context passing externally, not relying on the orchestrator's instructions.

Matcher: `agentName` regex, anchored as `^(?:pattern)$`.

### `preToolUse`

Fires before each tool call. Supports `decision: "allow" | "deny" | "ask"` and `modifiedArgs`. Potentially useful for guarding writes during spec generation phases. Lower priority than `subagentStop` / `subagentStart` for current smaqit needs.

---

## Proposed Gates

### Gate 1 — Spec Output Validation (`subagentStop`)

After each spec agent completes, a hook script validates the spec file before the orchestrator moves to the next layer.

Script responsibilities:
1. Resolve the expected spec file path for the completed agent (e.g., `specs/business/*.md`)
2. Check required frontmatter fields (`status:`) and required document sections
3. Return `{}` on pass; return `{ "decision": "block", "reason": "..." }` with specific gap description on failure

Effect: the orchestrator **cannot** advance past a failed spec regardless of its instructions. Enforcement is deterministic.

```json
// .github/hooks/smaqit-gates.json
{
  "version": 1,
  "hooks": {
    "subagentStop": [
      {
        "matcher": "smaqit\\.business",
        "type": "command",
        "bash": ".github/hooks/scripts/smaqit-validate-spec.sh business",
        "timeoutSec": 10
      },
      {
        "matcher": "smaqit\\.functional",
        "type": "command",
        "bash": ".github/hooks/scripts/smaqit-validate-spec.sh functional",
        "timeoutSec": 10
      },
      {
        "matcher": "smaqit\\.stack",
        "type": "command",
        "bash": ".github/hooks/scripts/smaqit-validate-spec.sh stack",
        "timeoutSec": 10
      },
      {
        "matcher": "smaqit\\.infrastructure",
        "type": "command",
        "bash": ".github/hooks/scripts/smaqit-validate-spec.sh infrastructure",
        "timeoutSec": 10
      },
      {
        "matcher": "smaqit\\.coverage",
        "type": "command",
        "bash": ".github/hooks/scripts/smaqit-validate-spec.sh coverage",
        "timeoutSec": 10
      }
    ]
  }
}
```

### Gate 2 — Iteration Cap Enforcement (`subagentStop`, same script)

The validation script maintains per-layer retry counters in `.smaqit/.hook-state/retries-{layer}`. On each block decision, the counter increments. At 3, the script returns `{}` (allow through) — the agent's iteration-cap fallback instruction then escalates to the user.

```bash
# excerpt from smaqit-validate-spec.sh
LAYER="$1"
STATE_DIR=".smaqit/.hook-state"
COUNT_FILE="$STATE_DIR/retries-$LAYER"
mkdir -p "$STATE_DIR"
count=$(cat "$COUNT_FILE" 2>/dev/null || echo 0)

if validate_spec "$LAYER"; then
  echo 0 > "$COUNT_FILE"   # reset on success
  echo '{}'
  exit 0
fi

count=$((count + 1))
echo "$count" > "$COUNT_FILE"

if [ "$count" -ge 3 ]; then
  echo '{}'   # allow through; orchestrator escalates
  exit 0
fi

REASON=$(build_feedback "$LAYER")
echo "{\"decision\": \"block\", \"reason\": \"$REASON\"}"
```

This makes the cap enforced externally. LLM compliance is not required to prevent runaway loops.

### Gate 3 — Scoped Context Injection (`subagentStart`)

Before spec agents that have upstream dependencies are spawned, a hook reads the required upstream spec files and injects them as `additionalContext`. This guarantees the subagent receives upstream spec content even if the orchestrator failed to pass it explicitly.

```json
// add to .github/hooks/smaqit-gates.json
"subagentStart": [
  {
    "matcher": "smaqit\\.functional",
    "type": "command",
    "bash": ".github/hooks/scripts/smaqit-inject-context.sh functional",
    "timeoutSec": 10
  },
  {
    "matcher": "smaqit\\.stack",
    "type": "command",
    "bash": ".github/hooks/scripts/smaqit-inject-context.sh stack",
    "timeoutSec": 10
  },
  {
    "matcher": "smaqit\\.coverage",
    "type": "command",
    "bash": ".github/hooks/scripts/smaqit-inject-context.sh coverage",
    "timeoutSec": 10
  }
]
```

The inject script reads only the sections relevant to the downstream agent (not full upstream spec dumps) and formats the output as:

```json
{ "additionalContext": "## Upstream Business Spec\n\n[extracted content]" }
```

---

## Concern Mapping: Task 082

| Task 082 concern | Instruction-level (task 082) | Hook enforcement-level |
|---|---|---|
| Spec output validation between layers | Agent: verify completeness before proceeding | `subagentStop`: shell validates spec file; blocks if invalid |
| Iteration cap (3 per layer) | Agent: max 3 retries, escalate on cap | `subagentStop` script: counts retries in `.smaqit/.hook-state/`; enforces cap |
| Scoped context to subagents | Agent: pass user requirements + upstream specs only | `subagentStart`: reads and injects upstream spec files directly |

Agent instructions remain necessary for orchestration logic (sequence, skip-if-complete, mode selection). Hooks reinforce only the concerns where LLM compliance is most fragile.

---

## Limitations and Open Questions

1. **VS Code surface not explicitly confirmed.** The docs name Copilot CLI and Copilot cloud agent. A "VS Code compatible" payload format is documented (PascalCase event names, snake_case fields), implying VS Code Copilot extension support — but the surface is not stated explicitly. **Must confirm before implementing.** This is the single prerequisite for all three gates.

2. **`subagentStop` block counts against session timeout.** Each forced turn extends the session. The 3-iteration cap matters for session budget, not only output quality.

3. **Hook scripts depend on spec files written before the hook fires.** Spec agents must have committed their output to `specs/` before `subagentStop` fires. This is the expected flow but is an implicit contract between the agent and the hook.

4. **Installer scope expansion.** If hooks are part of the smaqit scaffold, the installer must copy `.github/hooks/` and `.github/hooks/scripts/` alongside agents. This is a new installer concern not currently in scope for task 082.

5. **Context injection size.** `additionalContext` prepends to the subagent's prompt. The inject script must be selective — extract relevant sections rather than dumping entire upstream spec files.

6. **State file cleanup.** `.smaqit/.hook-state/retries-*` counters must be reset between phase agent runs. The validation script resets on success (see Gate 2 excerpt). A session-start hook or a pre-phase cleanup step should reset all counters at the start of a new phase execution.

---

## Verdict

Hooks are a strong fit for the enforcement layer of smaqit's orchestration gates — specifically inter-layer spec validation and iteration cap enforcement. The instruction-based approach in task 082 is the primary mechanism (it covers orchestration logic hooks cannot express). Hooks add deterministic reinforcement on top.

Implementation is conditional on confirming VS Code Copilot chat/agent surface support. If confirmed, gates 1 and 2 (via `subagentStop`) are the highest-value additions. Gate 3 (`subagentStart` context injection) is lower priority since scoped context passing is already addressed in agent instructions.
