"""Claude Code process runner with stream-JSON parsing."""

import io
import json
import re
import subprocess
import threading
from dataclasses import dataclass, field
from datetime import datetime
from pathlib import Path
from typing import Generator

from .config import SpektacularConfig


@dataclass
class ClaudeEvent:
    """Represents a parsed event from Claude's stream-JSON output."""
    type: str
    data: dict = field(default_factory=dict)

    @property
    def session_id(self) -> str | None:
        return self.data.get("session_id")

    @property
    def is_result(self) -> bool:
        return self.type == "result"

    @property
    def is_error(self) -> bool:
        return self.is_result and self.data.get("is_error", False)

    @property
    def result_text(self) -> str | None:
        if self.is_result:
            return self.data.get("result")
        return None

    @property
    def text_content(self) -> str | None:
        """Extract text blocks from assistant messages."""
        if self.type != "assistant":
            return None
        message = self.data.get("message", {})
        content = message.get("content", [])
        texts = [block["text"] for block in content if block.get("type") == "text"]
        return "\n".join(texts) if texts else None

    @property
    def tool_uses(self) -> list[dict]:
        """Extract tool_use blocks from assistant messages."""
        if self.type != "assistant":
            return []
        message = self.data.get("message", {})
        content = message.get("content", [])
        return [block for block in content if block.get("type") == "tool_use"]


@dataclass
class Question:
    """A structured question detected in Claude output."""
    question: str
    header: str
    options: list[dict]


QUESTION_PATTERN = re.compile(r"<!--QUESTION:(.*?)-->", re.DOTALL)


def detect_questions(text: str) -> list[Question]:
    """Detect structured question markers in text."""
    questions = []
    for match in QUESTION_PATTERN.finditer(text):
        try:
            data = json.loads(match.group(1))
            for q in data.get("questions", []):
                questions.append(Question(
                    question=q["question"],
                    header=q.get("header", ""),
                    options=q.get("options", []),
                ))
        except (json.JSONDecodeError, KeyError):
            continue
    return questions


def build_prompt(
    spec_content: str,
    agent_prompt: str,
    knowledge_contents: dict[str, str],
) -> str:
    """Build the combined prompt for the claude process.

    Concatenates: agent instructions + knowledge base + spec.
    """
    parts = [agent_prompt, "\n\n---\n\n# Knowledge Base\n"]
    for filename, content in knowledge_contents.items():
        parts.append(f"\n## {filename}\n{content}\n")
    parts.append(f"\n---\n\n# Specification to Plan\n\n{spec_content}")
    return "\n".join(parts)


def _open_debug_log(config: SpektacularConfig, command: str, cwd: Path) -> io.TextIOBase | None:
    """Open a debug log file if debug is enabled. Returns file handle or None."""
    if not config.debug.enabled:
        return None
    log_dir = cwd / config.debug.log_dir
    log_dir.mkdir(parents=True, exist_ok=True)
    timestamp = datetime.now().strftime("%Y-%m-%d_%H%M%S")
    tool_name = config.agent.command
    filename = f"{timestamp}_{tool_name}_{command}.log"
    return open(log_dir / filename, "w", encoding="utf-8")


def run_claude(
    prompt: str,
    config: SpektacularConfig,
    session_id: str | None = None,
    cwd: Path | None = None,
    command: str = "unknown",
) -> Generator[ClaudeEvent, None, None]:
    """Spawn claude process and yield parsed events."""
    cmd = [config.agent.command, "-p", prompt]
    cmd.extend(config.agent.args)

    if config.agent.allowed_tools:
        cmd.extend(["--allowedTools", ",".join(config.agent.allowed_tools)])

    if config.agent.dangerously_skip_permissions:
        cmd.append("--dangerously-skip-permissions")

    if session_id:
        cmd.extend(["--resume", session_id])

    effective_cwd = cwd or Path.cwd()

    process = subprocess.Popen(
        cmd,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        cwd=effective_cwd,
    )

    stderr_buf = io.StringIO()

    def drain_stderr():
        if process.stderr:
            stderr_buf.write(process.stderr.read())

    stderr_thread = threading.Thread(target=drain_stderr, daemon=True)
    stderr_thread.start()

    debug_log = _open_debug_log(config, command, effective_cwd)
    try:
        for line in (process.stdout or []):
            line = line.strip()
            if not line:
                continue
            if debug_log:
                debug_log.write(line + "\n")
                debug_log.flush()
            try:
                data = json.loads(line)
                yield ClaudeEvent(type=data.get("type", "unknown"), data=data)
            except json.JSONDecodeError:
                continue

        stderr_thread.join()
        process.wait()
        if process.returncode != 0:
            raise RuntimeError(
                f"Claude process exited with code {process.returncode}: {stderr_buf.getvalue()}"
            )
    finally:
        if debug_log:
            debug_log.close()
