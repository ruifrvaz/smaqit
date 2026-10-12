#!/usr/bin/env python3
"""Checks over a `claude -p --output-format stream-json --verbose` session log.

Used by scripts/e2e-claude.sh. Every check reads the structured events, never the
command or skill text Claude Code expands into the prompt, so a phrase that only
appears in smaqit's own instructions (for example "NEEDS-INPUT") never counts as
evidence.

  log_check.py session-id LOG          print the run's session id
  log_check.py summary LOG             print a one-line JSON run summary
  log_check.py signal LOG NAME         exit 0 if the main thread received or wrote a NAME block
  log_check.py delegated LOG AGENT     exit 0 if the main thread delegated to AGENT
  log_check.py errored LOG             exit 0 (and print why) if the run ended with an error
"""
import json
import re
import sys


def events(path):
    with open(path, encoding="utf-8") as handle:
        for line in handle:
            line = line.strip()
            if not line:
                continue
            try:
                yield json.loads(line)
            except json.JSONDecodeError:
                continue


def blocks(event):
    message = event.get("message")
    if not isinstance(message, dict):  # e.g. system/permission_denied carries a plain string
        return []
    content = message.get("content")
    return content if isinstance(content, list) else []


def text_of(value):
    if isinstance(value, str):
        return value
    if isinstance(value, list):
        return "\n".join(text_of(item) for item in value)
    if isinstance(value, dict):
        return text_of(value.get("text", "")) if "text" in value else text_of(value.get("content", ""))
    return ""


def main_thread_texts(path):
    """Text the main conversation wrote, and results that came back to it.

    A foreground subagent's result arrives as a tool_result block. A background
    subagent's result arrives as a system/task_notification event (its summary or
    handback report), and the main conversation relays it in a result event.
    """
    for event in events(path):
        if event.get("parent_tool_use_id"):
            continue
        if event.get("type") == "system" and event.get("subtype") == "task_notification":
            yield text_of(event.get("summary", ""))
            yield text_of((event.get("handback_report") or {}).get("text", ""))
            continue
        if event.get("type") == "result":
            yield text_of(event.get("result", ""))
            continue
        for block in blocks(event):
            if event.get("type") == "assistant" and block.get("type") == "text":
                yield block.get("text", "")
            elif event.get("type") == "user" and block.get("type") == "tool_result":
                yield text_of(block.get("content"))


def session_id(path):
    found = ""
    for event in events(path):
        found = event.get("session_id") or found
    return found


def summary(path):
    """One run can emit several result events: a headless run that delegates to a
    background subagent reports once when the main turn ends, again when it relays the
    subagent's result, and finally with an empty closing event. Combine them all."""
    init, results = {}, []
    for event in events(path):
        if event.get("type") == "system" and event.get("subtype") == "init" and not init:
            init = event
        elif event.get("type") == "result":
            results.append(event)
    result = next((r for r in reversed(results) if r.get("result")), results[-1] if results else {})
    denials = [d for r in results for d in (r.get("permission_denials") or [])]
    return {
        "claude_code_version": init.get("claude_code_version"),
        "session_id": result.get("session_id") or init.get("session_id"),
        "result_events": len(results),
        "subtype": result.get("subtype"),
        "is_error": any(r.get("is_error") for r in results),
        "num_turns": sum(r.get("num_turns") or 0 for r in results),
        "duration_ms": sum(r.get("duration_ms") or 0 for r in results),
        "total_cost_usd": max((r.get("total_cost_usd") or 0 for r in results), default=None),
        "mcp_servers": [
            {"name": server.get("name"), "status": server.get("status")}
            for server in init.get("mcp_servers") or []
        ],
        "permission_denials": sorted({denial.get("tool_name", "?") for denial in denials}),
        "api_error_status": next((r.get("api_error_status") for r in results if r.get("api_error_status")), None),
        "result": str(result.get("result") or "")[:300],
    }


def has_signal(path, name):
    pattern = re.compile(r"^\s*(?:`{3}\w*\s*)?" + re.escape(name) + r"\b", re.MULTILINE)
    return any(pattern.search(text) for text in main_thread_texts(path))


def delegated(path, agent):
    """True when the main thread invoked AGENT as a subagent, skill, or command."""
    command = agent.replace("smaqit-", "smaqit.", 1)
    for event in events(path):
        if event.get("type") != "assistant" or event.get("parent_tool_use_id"):
            continue
        for block in blocks(event):
            if block.get("type") != "tool_use":
                continue
            payload = block.get("input") or {}
            if block.get("name") in ("Task", "Agent") and payload.get("subagent_type") == agent:
                return f"{block['name']}(subagent_type={agent})"
            if block.get("name") == "Skill" and command in {str(v).lstrip("/") for v in payload.values()}:
                return f"Skill({command})"
    return ""


def main(argv):
    if len(argv) < 3:
        print(__doc__, file=sys.stderr)
        return 2
    command, path = argv[1], argv[2]
    if command == "session-id":
        print(session_id(path))
        return 0 if session_id(path) else 1
    if command == "summary":
        print(json.dumps(summary(path)))
        return 0
    if command == "errored":
        run = summary(path)
        if run["is_error"] or not run["subtype"]:
            print(run["result"] or "no result event (the run did not finish)")
            return 0
        return 1
    if command == "signal" and len(argv) == 4:
        return 0 if has_signal(path, argv[3]) else 1
    if command == "delegated" and len(argv) == 4:
        match = delegated(path, argv[3])
        if match:
            print(match)
        return 0 if match else 1
    print(__doc__, file=sys.stderr)
    return 2


if __name__ == "__main__":
    sys.exit(main(sys.argv))
