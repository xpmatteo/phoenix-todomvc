#!/usr/bin/env python3
"""Log every subagent run: duration, model, prompt, cost, and outcome.

Wired to the SubagentStop and Stop hooks. Reads the hook payload on stdin,
finds every subagent transcript belonging to the session, and writes one
record per subagent run to ~/.claude.personal/subagent-log.jsonl.

The script is idempotent: it is safe to run repeatedly. A record whose
outcome is not yet known (the parent has not received the tool_result) is
written as "unknown" and backfilled on a later run.
"""

import fcntl
import json
import os
import subprocess
import sys
import tempfile
from datetime import datetime
from pathlib import Path

# Seconds to allow a single credential scan before giving up on it.
SCAN_TIMEOUT_SEC = 15

# Where Claude Code keeps its transcripts. Only used to locate the subagent
# transcripts we read; nothing is written here.
CONFIG_DIR = Path(os.environ.get("CLAUDE_CONFIG_DIR", Path.home() / ".claude"))

# The log lives beside this script, in the project's own .claude/ directory.
LOG_DIR = Path(__file__).resolve().parent.parent
LOG_PATH = LOG_DIR / "subagent-log.jsonl"
LOCK_PATH = LOG_DIR / ".subagent-log.lock"

# ---------------------------------------------------------------------------
# PRICING — edit this block when rates change.
# ---------------------------------------------------------------------------
# Base rates in USD per million tokens, as (input, output).
FABLE_5 = (10.00, 50.00)
MYTHOS_5 = (10.00, 50.00)
OPUS_4_8 = (5.00, 25.00)
OPUS_4_7 = (5.00, 25.00)
OPUS_4_6 = (5.00, 25.00)
OPUS_4_5 = (5.00, 25.00)
OPUS_4_1 = (15.00, 75.00)
OPUS_4_0 = (15.00, 75.00)
SONNET_5 = (3.00, 15.00)  # introductory $2/$10 runs through 2026-08-31
SONNET_4_6 = (3.00, 15.00)
SONNET_4_5 = (3.00, 15.00)
SONNET_4_0 = (3.00, 15.00)
HAIKU_4_5 = (1.00, 5.00)

# Cache rates are multiples of the model's base input rate.
CACHE_WRITE_5M_MULTIPLIER = 1.25
CACHE_WRITE_1H_MULTIPLIER = 2.00
CACHE_READ_MULTIPLIER = 0.10

# Placeholder model on harness-generated messages; never billed.
SYNTHETIC_MODEL = "<synthetic>"

# Model ID prefix -> base rates. Matched by longest prefix, so date suffixes
# and variants like "[1m]" resolve to the right entry.
PRICES = {
    "claude-fable-5": FABLE_5,
    "claude-mythos-5": MYTHOS_5,
    "claude-opus-4-8": OPUS_4_8,
    "claude-opus-4-7": OPUS_4_7,
    "claude-opus-4-6": OPUS_4_6,
    "claude-opus-4-5": OPUS_4_5,
    "claude-opus-4-1": OPUS_4_1,
    "claude-opus-4": OPUS_4_0,
    "claude-sonnet-5": SONNET_5,
    "claude-sonnet-4-6": SONNET_4_6,
    "claude-sonnet-4-5": SONNET_4_5,
    "claude-sonnet-4": SONNET_4_0,
    "claude-haiku-4-5": HAIKU_4_5,
}


def price_for(model):
    """Longest-prefix match so '[1m]' and date suffixes still resolve."""
    if not model:
        return None
    best = None
    for name, rates in PRICES.items():
        if model.startswith(name) and (best is None or len(name) > len(best[0])):
            best = (name, rates)
    return best[1] if best else None


def redact(text):
    """Replace recognisable credentials in text before it is written.

    Returns (text, status). Status is "clean" if nothing was found,
    "redacted" if something was replaced, or "unavailable" if the scan could
    not run — recorded explicitly so unscanned text is never mistaken for
    scanned text.

    Detection is delegated to gitleaks so it stays in step with the
    pre-commit gate. It matches credentials by shape; a password written in
    prose has no shape to match and will not be caught.
    """
    if not text:
        return text, "clean"

    handle, report = tempfile.mkstemp(prefix="subagent-scan-", suffix=".json")
    os.close(handle)
    try:
        subprocess.run(
            ["gitleaks", "stdin", "--report-format", "json",
             "--report-path", report, "--no-banner"],
            input=text.encode(), stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL, timeout=SCAN_TIMEOUT_SEC, check=False,
        )
        with open(report) as f:
            findings = json.load(f)
    except FileNotFoundError:
        return text, "unavailable"  # gitleaks not installed
    except Exception:
        return text, "unavailable"
    finally:
        try:
            os.unlink(report)
        except OSError:
            pass

    if not findings:
        return text, "clean"

    # Longest first, so a secret contained within another is not half-replaced.
    for finding in sorted(findings, key=lambda f: len(f.get("Secret") or ""), reverse=True):
        secret = finding.get("Secret")
        if secret:
            text = text.replace(secret, f"[REDACTED:{finding.get('RuleID', 'secret')}]")
    return text, "redacted"


def read_jsonl(path):
    rows = []
    try:
        with open(path) as f:
            for line in f:
                line = line.strip()
                if line:
                    try:
                        rows.append(json.loads(line))
                    except json.JSONDecodeError:
                        pass
    except OSError:
        pass
    return rows


def parse_ts(value):
    if not value:
        return None
    try:
        return datetime.fromisoformat(value.replace("Z", "+00:00"))
    except ValueError:
        return None


def text_of(content):
    """Flatten a message content field to plain text."""
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        parts = []
        for block in content:
            if isinstance(block, dict) and block.get("type") == "text":
                parts.append(block.get("text", ""))
        return "\n".join(parts)
    return ""


def index_parent(parent_rows):
    """Map tool_use_id -> Task invocation details and result outcome."""
    tasks, results = {}, {}
    for row in parent_rows:
        content = (row.get("message") or {}).get("content")
        if not isinstance(content, list):
            continue
        for block in content:
            if not isinstance(block, dict):
                continue
            if block.get("type") == "tool_use" and block.get("name") in ("Task", "Agent"):
                tasks[block["id"]] = block.get("input") or {}
            elif block.get("type") == "tool_result":
                results[block.get("tool_use_id")] = block
    return tasks, results


def read_journal(path):
    """Map agentId -> {'status': ..., 'result': ...} from a workflow journal.

    The journal records a "started" line when an agent is spawned and a
    "result" line when it returns. An agent with a start but no result was
    still running when we looked.
    """
    journal = {}
    for row in read_jsonl(path):
        agent_id = row.get("agentId")
        if not agent_id:
            continue
        entry = journal.setdefault(agent_id, {"status": "unknown", "result": None})
        kind = row.get("type")
        if kind == "started":
            entry["status"] = "running"
        elif kind == "result":
            entry["status"] = "ok"
            entry["result"] = row.get("result")
        elif kind == "error":
            entry["status"] = "error"
            entry["result"] = row.get("error")
    return journal


def summarize(agent_id, meta, agent_rows, tasks, results, session_id, cwd,
              workflow_run_id=None, journal_entry=None):
    stamps = [t for t in (parse_ts(r.get("timestamp")) for r in agent_rows) if t]
    assistants = [r for r in agent_rows if r.get("type") == "assistant"]

    tokens = {
        "input": 0,
        "output": 0,
        "cache_write_5m": 0,
        "cache_write_1h": 0,
        "cache_read": 0,
    }
    models = []
    for row in assistants:
        message = row.get("message") or {}
        model = message.get("model")
        # "<synthetic>" marks harness-generated stubs (API errors, interrupts).
        # They carry no real usage and would defeat the price lookup.
        if model == SYNTHETIC_MODEL:
            continue
        if model and model not in models:
            models.append(model)
        usage = message.get("usage") or {}
        tokens["input"] += usage.get("input_tokens") or 0
        tokens["output"] += usage.get("output_tokens") or 0
        tokens["cache_read"] += usage.get("cache_read_input_tokens") or 0
        breakdown = usage.get("cache_creation") or {}
        five = breakdown.get("ephemeral_5m_input_tokens")
        hour = breakdown.get("ephemeral_1h_input_tokens")
        if five is None and hour is None:
            # Older transcripts only carry the aggregate; assume the 5m TTL.
            tokens["cache_write_5m"] += usage.get("cache_creation_input_tokens") or 0
        else:
            tokens["cache_write_5m"] += five or 0
            tokens["cache_write_1h"] += hour or 0

    # Cost is billed against whichever model produced each message; the fleet
    # is near-always homogeneous, so price the total against the first model.
    rates = price_for(models[0] if models else None)
    if rates:
        inp, out = rates
        cost = (
            tokens["input"] * inp
            + tokens["output"] * out
            + tokens["cache_write_5m"] * inp * CACHE_WRITE_5M_MULTIPLIER
            + tokens["cache_write_1h"] * inp * CACHE_WRITE_1H_MULTIPLIER
            + tokens["cache_read"] * inp * CACHE_READ_MULTIPLIER
        ) / 1_000_000
        cost = round(cost, 6)
    else:
        cost = None

    tool_use_id = meta.get("toolUseId")
    task = tasks.get(tool_use_id, {})
    result = results.get(tool_use_id)

    # Outcome comes from the parent's tool_result for Task spawns, and from
    # the run's journal for workflow agents. Forks produce neither, so there
    # is nothing to wait for.
    if meta.get("stoppedByUser"):
        outcome = "stopped_by_user"
    elif journal_entry is not None:
        outcome = journal_entry["status"]
    elif result is not None:
        outcome = "error" if result.get("is_error") else "ok"
    elif tool_use_id is None:
        outcome = "no_result"
    else:
        outcome = "unknown"

    prompt = task.get("prompt")
    if not prompt:
        first_user = next((r for r in agent_rows if r.get("type") == "user"), None)
        if first_user:
            prompt = text_of((first_user.get("message") or {}).get("content"))

    error_text = text_of(result.get("content"))[:2000] if outcome == "error" else None

    # Scrub before anything reaches disk, so a pasted credential never lands
    # in the log at all.
    prompt, prompt_status = redact(prompt)
    error_text, error_status = redact(error_text)
    redaction = "unavailable" if "unavailable" in (prompt_status, error_status) else (
        "redacted" if "redacted" in (prompt_status, error_status) else "clean"
    )

    record = {
        "agent_id": agent_id,
        "session_id": session_id,
        "cwd": cwd,
        "agent_type": meta.get("agentType"),
        "is_fork": bool(meta.get("isFork")),
        "workflow_run_id": workflow_run_id,
        "spawn_depth": meta.get("spawnDepth"),
        "tool_use_id": tool_use_id,
        # Task spawns carry the description on the tool call; forks carry it
        # on their own meta, since there is no tool call.
        "description": task.get("description") or meta.get("description"),
        "background": bool(task.get("run_in_background")),
        "requested_model": task.get("model"),
        "models": models,
        "started_at": stamps[0].isoformat() if stamps else None,
        "ended_at": stamps[-1].isoformat() if stamps else None,
        "duration_sec": round((stamps[-1] - stamps[0]).total_seconds(), 1) if len(stamps) > 1 else None,
        "turns": len(assistants),
        "tokens": tokens,
        "cost_usd": cost,
        "outcome": outcome,
        "redaction": redaction,
        "prompt": prompt,
    }
    if error_text is not None:
        record["error"] = error_text
    return record


def collect_dir(directory, session_id, cwd, tasks, results,
                workflow_run_id=None, journal=None):
    """Summarize every agent transcript in one directory."""
    records = []
    for meta_path in sorted(directory.glob("agent-*.meta.json")):
        agent_id = meta_path.name[: -len(".meta.json")]
        transcript = directory / f"{agent_id}.jsonl"
        if not transcript.exists():
            continue
        try:
            meta = json.loads(meta_path.read_text())
        except (OSError, json.JSONDecodeError):
            meta = {}
        rows = read_jsonl(transcript)
        if not rows:
            continue
        # Journal keys drop the "agent-" filename prefix.
        entry = journal.get(agent_id[len("agent-"):]) if journal is not None else None
        records.append(summarize(
            agent_id, meta, rows, tasks, results, session_id, cwd,
            workflow_run_id=workflow_run_id, journal_entry=entry,
        ))
    return records


def collect(session_dir, session_id, cwd, parent_rows):
    """Agents live in two places: Task spawns and forks sit directly under
    subagents/; workflow agents sit under subagents/workflows/<run-id>/,
    with a journal.jsonl recording each agent's outcome.
    """
    tasks, results = index_parent(parent_rows)
    subagents = session_dir / "subagents"
    if not subagents.is_dir():
        return []

    records = collect_dir(subagents, session_id, cwd, tasks, results)

    workflows = subagents / "workflows"
    if workflows.is_dir():
        for run_dir in sorted(p for p in workflows.iterdir() if p.is_dir()):
            records.extend(collect_dir(
                run_dir, session_id, cwd, tasks, results,
                workflow_run_id=run_dir.name,
                journal=read_journal(run_dir / "journal.jsonl"),
            ))
    return records


def merge(records):
    """Rewrite the log with these records merged in, keyed by agent_id."""
    if not records:
        return
    LOG_DIR.mkdir(parents=True, exist_ok=True)
    with open(LOCK_PATH, "w") as lock:
        fcntl.flock(lock, fcntl.LOCK_EX)
        existing = {r["agent_id"]: r for r in read_jsonl(LOG_PATH) if r.get("agent_id")}
        changed = False
        for record in records:
            previous = existing.get(record["agent_id"])
            # Never regress a known outcome back to "unknown".
            if previous and record["outcome"] == "unknown" and previous.get("outcome") != "unknown":
                record["outcome"] = previous["outcome"]
                record["error"] = previous.get("error")
            if previous != record:
                existing[record["agent_id"]] = record
                changed = True
        if not changed:
            return
        tmp = LOG_PATH.with_suffix(".jsonl.tmp")
        with open(tmp, "w") as out:
            for record in sorted(existing.values(), key=lambda r: r.get("started_at") or ""):
                out.write(json.dumps(record) + "\n")
        os.replace(tmp, LOG_PATH)


def main():
    try:
        payload = json.load(sys.stdin)
    except (json.JSONDecodeError, ValueError):
        return 0

    session_id = payload.get("session_id")
    cwd = payload.get("cwd")
    transcript_path = payload.get("transcript_path")
    if not session_id:
        return 0

    # The parent transcript sits next to the session's subagents/ directory.
    if transcript_path:
        parent = Path(transcript_path)
    else:
        encoded = str(cwd or "").replace("/", "-")
        parent = CONFIG_DIR / "projects" / encoded / f"{session_id}.jsonl"

    # Scan every session in this project, not just the current one. The log is
    # derived data: rebuilding it from all transcripts means a deleted or
    # truncated log heals itself instead of silently losing history.
    project_dir = parent.parent
    records = []
    for session_dir in sorted(p for p in project_dir.iterdir() if p.is_dir()):
        if not (session_dir / "subagents").is_dir():
            continue  # most sessions spawned no agents; skip the transcript read
        sid = session_dir.name
        records.extend(collect(
            session_dir, sid, cwd, read_jsonl(project_dir / f"{sid}.jsonl"),
        ))
    merge(records)
    return 0


if __name__ == "__main__":
    try:
        sys.exit(main())
    except Exception:
        # A hook must never break the session it is observing.
        sys.exit(0)
