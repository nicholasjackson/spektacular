"""Verify that the spektacular plan workflow completed successfully.

Per-step pass/fail reporting for the full plan workflow, driven by four
layers of assertions:

  1. Workflow integrity — state reached finished, completed_steps matches
     the hand-maintained EXPECTED_STEP_ORDER, no placeholder markers in
     any artefact.
  2. Plan CLI invocation — for every step, the agent invoked the matching
     `plan new` / `plan goto` Bash call during that step's window.
  3. Skill retrieval and sub-agent spawning — for every (step, skill) pair
     in EXPECTED_SKILLS_PER_STEP, the agent retrieved that skill during
     that step's window; for every step in EXPECTED_SPAWN_STEPS, the agent
     spawned at least one sub-agent during that step's window.
  4. Artefact content and instruction `next_step` validity — plan.md,
     context.md, research.md exist with substantive on-topic content, and
     every rendered instruction's next_step directive references a valid
     state-machine step.

All expectation maps (step order, skills per step, sub-agent spawn steps)
are hand-maintained in this module — they are the independent behavioural
oracle. Do NOT derive them from `spektacular plan steps` or from the
`templates/steps/plan/*.md` files at runtime: that would make the test
tautological (asserting "the agent did what the state machine / templates
told it to do" is a closed loop, not a behavioural check).

When a legitimate change lands in `internal/steps/plan/steps.go` Steps() or
in `templates/steps/plan/*.md`, update the corresponding map below in the
same commit.
"""

import json
import re
from dataclasses import dataclass
from pathlib import Path

import pytest

# ---------------------------------------------------------------------------
# Paths and constants
# ---------------------------------------------------------------------------

PROJECT_DIR = Path("/app")
SPEK_DIR = PROJECT_DIR / ".spektacular"
TRANSCRIPT = Path("/logs/agent/claude-code.txt")

# Hand-maintained canonical step order — the behavioural oracle for
# state-machine correctness. When a legitimate state-machine change lands
# (add/rename/reorder a step), update this list AND
# internal/steps/plan/steps.go Steps() in the same commit.
EXPECTED_STEP_ORDER = [
    "new",
    "overview",
    "discovery",
    "architecture",
    "components",
    "data_structures",
    "implementation_detail",
    "dependencies",
    "testing_approach",
    "milestones",
    "phases",
    "open_questions",
    "out_of_scope",
    "verification",
    "write_plan",
    "write_context",
    "write_research",
    "finished",
]

# Hand-maintained skill retrievals the agent is expected to make during
# each step. Independent oracle — when a template under
# templates/steps/plan/ changes its skill references, this map must be
# updated in the same commit.
EXPECTED_SKILLS_PER_STEP = {
    "discovery": frozenset(
        {
            "discover-project-commands",
            "discover-test-patterns",
            "spawn-planning-agents",
        }
    ),
    "phases": frozenset({"spawn-implementation-agents"}),
    "verification": frozenset(
        {
            "gather-project-metadata",
            "determine-feature-slug",
        }
    ),
}

# Hand-maintained set of steps whose templates direct the agent to spawn
# sub-agents. Independent oracle — update when template language changes.
EXPECTED_SPAWN_STEPS = frozenset({"discovery"})

MIN_SECTION_LENGTH = 100

# Hand-maintained list of scaffold slot strings that must be replaced when
# the agent fills the scaffolds. An occurrence of any of these substrings in
# a written artefact means the agent left a slot unfilled. These are literal
# strings copied from templates/scaffold/{plan,context,research}.md — when
# the scaffolds change, this list must be updated in the same commit.
#
# Matching is case-sensitive and substring-based, which is fine because
# these are distinctive angle-bracketed slot labels that don't appear in
# legitimate prose.
SCAFFOLD_LEFTOVERS = (
    # plan.md scaffold
    "<user-facing title>",
    "<short title>",
    "<one-paragraph description",
    "<outcome statement in plain language>",
    # context.md scaffold
    "<title matching plan.md>",
    "<file:line>",
    "<description of change>",
    "<approach>",
    # research.md scaffold
    "<description>",
    "<reason with citation>",
    "<path:line>",
    "<one-line summary of what was learned>",
)

# Section headings (lowercase) expected to appear in plan.md after the
# verification step writes it.
EXPECTED_PLAN_SECTIONS = (
    "overview",
    "architecture & design decisions",
    "component breakdown",
    "data structures & interfaces",
    "implementation detail",
    "dependencies",
    "testing approach",
    "milestones & phases",
    "open questions",
    "out of scope",
)

PLAN_GOTO_STEP_RE = re.compile(r'"step"\s*:\s*"([a-z_]+)"')
INSTRUCTION_NEXT_STEP_RE = re.compile(
    r"plan\s+goto\s+--data\s+'\{\"step\":\"([a-z_]+)\"\}'"
)


# ---------------------------------------------------------------------------
# Data classes
# ---------------------------------------------------------------------------


@dataclass(frozen=True)
class ToolCall:
    index: int
    type: str  # "Bash" | "Skill" | "Task" | "Agent"
    name: str
    input: dict
    tool_use_id: str


@dataclass(frozen=True)
class StepWindow:
    step: str
    start: int  # inclusive
    end: int  # exclusive


# ---------------------------------------------------------------------------
# State / artefact helpers
# ---------------------------------------------------------------------------


def find_plan_state_file() -> Path:
    """Locate the unified .spektacular/state.json."""
    state_file = SPEK_DIR / "state.json"
    assert state_file.exists(), f"No state file found at {state_file}"
    return state_file


def find_plan_name() -> str:
    state = load_state()
    name = (state.get("data") or {}).get("name")
    assert name, "state.json has no data.name — cannot resolve plan name"
    return name


def load_state() -> dict:
    return json.loads(find_plan_state_file().read_text())


def plan_artefact_paths() -> tuple:
    name = find_plan_name()
    base = SPEK_DIR / "plans" / name
    return base / "plan.md", base / "context.md", base / "research.md"


def parse_sections(text: str) -> dict:
    """Split markdown into sections keyed by lowercase heading.

    HTML comments are stripped so they do not count toward content length.
    `##` headings become section keys.
    """
    cleaned = re.sub(r"<!--.*?-->", "", text, flags=re.DOTALL)
    sections: dict = {}
    current = None
    lines: list = []
    for line in cleaned.splitlines():
        m = re.match(r"^#{2}\s+(.+)", line)
        if m:
            if current is not None:
                sections[current] = "\n".join(lines).strip()
            current = m.group(1).strip().lower()
            lines = []
        elif current is not None:
            lines.append(line)
    if current is not None:
        sections[current] = "\n".join(lines).strip()
    return sections


# ---------------------------------------------------------------------------
# Transcript extraction
# ---------------------------------------------------------------------------


def _iter_transcript_objects():
    assert TRANSCRIPT.exists(), f"Agent transcript not found at {TRANSCRIPT}"
    for line in TRANSCRIPT.read_text().splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            yield json.loads(line)
        except json.JSONDecodeError:
            continue


def extract_tool_calls() -> list:
    """Return an ordered list of ToolCall entries from the transcript.

    Captures Bash, Skill, Task, and Agent tool_use blocks. Preserves
    transcript order so step-window attribution works.
    """
    calls: list = []
    for obj in _iter_transcript_objects():
        if obj.get("type") != "assistant":
            continue
        msg = obj.get("message", {})
        for block in msg.get("content", []):
            if block.get("type") != "tool_use":
                continue
            name = block.get("name", "")
            if name in ("Bash", "Skill", "Task", "Agent"):
                calls.append(
                    ToolCall(
                        index=len(calls),
                        type=name,
                        name=name,
                        input=block.get("input", {}) or {},
                        tool_use_id=block.get("id", "") or "",
                    )
                )
    return calls


def extract_tool_results() -> dict:
    """Return a map of tool_use_id → result text extracted from user turns."""
    results: dict = {}
    for obj in _iter_transcript_objects():
        if obj.get("type") != "user":
            continue
        msg = obj.get("message", {})
        for block in msg.get("content", []):
            if block.get("type") != "tool_result":
                continue
            tid = block.get("tool_use_id", "")
            if not tid:
                continue
            content = block.get("content", "")
            if isinstance(content, list):
                text = "".join(
                    (part.get("text", "") if isinstance(part, dict) else str(part))
                    for part in content
                )
            else:
                text = str(content)
            results[tid] = text
    return results


# ---------------------------------------------------------------------------
# Plan-CLI call detection and step windowing
# ---------------------------------------------------------------------------


def _bash_command(call: ToolCall) -> str:
    return call.input.get("command", "") if call.type == "Bash" else ""


def _is_plan_new_call(cmd: str) -> bool:
    return ("spektacular plan new" in cmd) or (
        "plan new --data" in cmd and "spektacular" in cmd
    )


def _is_plan_goto_call(cmd: str) -> bool:
    return "plan goto" in cmd and "spektacular" in cmd


def find_plan_cli_calls(calls: list) -> list:
    """Return an ordered list of step names touched by plan new / plan goto calls.

    `plan new` emits the implicit "new" step. `plan goto --data '{"step":"X"}'`
    emits "X".
    """
    results: list = []
    for call in calls:
        cmd = _bash_command(call)
        if not cmd:
            continue
        if _is_plan_new_call(cmd):
            results.append("new")
        elif _is_plan_goto_call(cmd):
            m = PLAN_GOTO_STEP_RE.search(cmd)
            if m:
                results.append(m.group(1))
    return results


def resolve_step_windows(calls: list) -> dict:
    """Build step-attribution windows over the tool-call list.

    Each transition call (`plan new` or `plan goto`) opens a window keyed by
    the step it transitions INTO. The window ends at the next transition
    call, or at len(calls) for the last one. The implicit `overview`
    transition inside `plan new` is given a shared window with `new` so
    assertions against `overview` still resolve.
    """
    transitions: list = []  # (call_index, step_name)
    for call in calls:
        cmd = _bash_command(call)
        if not cmd:
            continue
        if _is_plan_new_call(cmd):
            transitions.append((call.index, "new"))
        elif _is_plan_goto_call(cmd):
            m = PLAN_GOTO_STEP_RE.search(cmd)
            if m:
                transitions.append((call.index, m.group(1)))

    windows: dict = {}
    for i, (start_idx, step) in enumerate(transitions):
        end = transitions[i + 1][0] if i + 1 < len(transitions) else len(calls)
        windows[step] = StepWindow(step=step, start=start_idx, end=end)

    # `overview` is entered automatically by `plan new`. If it has no
    # explicit goto, share `new`'s window so expectations still attach.
    if "new" in windows and "overview" not in windows:
        nw = windows["new"]
        windows["overview"] = StepWindow(
            step="overview",
            start=nw.start,
            end=nw.end,
        )
    return windows


# ---------------------------------------------------------------------------
# Parameter expansion and caches
# ---------------------------------------------------------------------------

# Lazy-populated caches so transcript parsing happens once per test session.
_CACHE: dict = {}


def _skill_params() -> list:
    return sorted(
        (step, skill)
        for step, skills in EXPECTED_SKILLS_PER_STEP.items()
        for skill in skills
    )


def _spawn_params() -> list:
    return sorted(EXPECTED_SPAWN_STEPS)


def _calls_cache() -> list:
    if "calls" not in _CACHE:
        _CACHE["calls"] = extract_tool_calls()
    return _CACHE["calls"]


def _windows_cache() -> dict:
    if "windows" not in _CACHE:
        _CACHE["windows"] = resolve_step_windows(_calls_cache())
    return _CACHE["windows"]


def _results_cache() -> dict:
    if "results" not in _CACHE:
        _CACHE["results"] = extract_tool_results()
    return _CACHE["results"]


# ---------------------------------------------------------------------------
# Workflow-level tests
# ---------------------------------------------------------------------------


class TestWorkflow:
    """Overall workflow integrity checks."""

    def test_project_initialized(self):
        assert (SPEK_DIR / "config.yaml").exists(), (
            "config.yaml missing — `spektacular init` was not run"
        )

    def test_state_file_exists(self):
        find_plan_state_file()

    def test_workflow_reached_finished(self):
        state = load_state()
        assert state.get("current_step") == "finished", (
            f"Workflow did not finish; current_step={state.get('current_step')}"
        )

    def test_all_steps_completed(self):
        state = load_state()
        completed = set(state.get("completed_steps", []))
        missing = [s for s in EXPECTED_STEP_ORDER if s not in completed]
        assert not missing, f"Steps not completed: {missing}"

    def test_steps_executed_in_order(self):
        """completed_steps == EXPECTED_STEP_ORDER catches state-machine
        reordering because the agent follows the state machine's order, so
        any reorder in Steps() surfaces as a diff here.
        """
        state = load_state()
        completed = state.get("completed_steps", [])
        assert completed == EXPECTED_STEP_ORDER, (
            f"Steps out of order.\n"
            f"  Expected: {EXPECTED_STEP_ORDER}\n"
            f"  Got:      {completed}"
        )

    def test_no_unfilled_scaffold_slots(self):
        """Every scaffold slot must be replaced with real content.

        This catches partial fills where the agent left a `<placeholder>`
        span from the scaffold unchanged. It does NOT scan for words like
        "TODO" or "TBD" in prose — those are legitimate when discussing
        genuine implementation uncertainty.
        """
        offenders = []
        for path in plan_artefact_paths():
            if not path.exists():
                continue
            text = path.read_text()
            for slot in SCAFFOLD_LEFTOVERS:
                if slot in text:
                    offenders.append((path.name, slot))
        assert not offenders, (
            f"Unfilled scaffold slots in artefacts: {offenders}"
        )

    def test_artefact_files_exist(self):
        for path in plan_artefact_paths():
            assert path.exists(), f"Artefact file missing: {path}"


# ---------------------------------------------------------------------------
# Per-step CLI-call tests
# ---------------------------------------------------------------------------


class TestPlanCLICalls:
    """For every expected step, the agent invoked the matching plan CLI call."""

    @pytest.mark.parametrize("step", EXPECTED_STEP_ORDER)
    def test_step_completed(self, step):
        state = load_state()
        assert step in state.get("completed_steps", []), (
            f"Step '{step}' not in completed_steps"
        )

    @pytest.mark.parametrize("step", EXPECTED_STEP_ORDER)
    def test_cli_call_for_step(self, step):
        calls = _calls_cache()
        observed = find_plan_cli_calls(calls)
        if step == "overview":
            # Overview is reached implicitly by `plan new`; accept either
            # an explicit goto or the implicit `plan new` transition.
            assert "new" in observed or "overview" in observed, (
                f"Neither 'plan new' nor 'plan goto overview' observed; got: {observed}"
            )
        else:
            assert step in observed, (
                f"No plan CLI call observed for step '{step}'; got: {observed}"
            )


# ---------------------------------------------------------------------------
# Template-driven skill-retrieval tests
# ---------------------------------------------------------------------------


class TestSkillRetrieval:
    """Every skill a step template references is retrieved during that step."""

    @pytest.mark.parametrize("step,skill", _skill_params())
    def test_skill_retrieved_in_step(self, step, skill):
        windows = _windows_cache()
        calls = _calls_cache()
        window = windows.get(step)
        assert window is not None, (
            f"Step '{step}' was never entered — cannot verify skill '{skill}' retrieval"
        )
        window_calls = calls[window.start : window.end]
        retrieved = False
        for c in window_calls:
            if c.type == "Skill":
                skill_name = c.input.get("skill") or c.input.get("name") or ""
                if skill_name == skill:
                    retrieved = True
                    break
            elif c.type == "Bash":
                cmd = c.input.get("command", "")
                if re.search(rf"\bskill\s+{re.escape(skill)}\b", cmd):
                    retrieved = True
                    break
        assert retrieved, (
            f"Step '{step}' did not retrieve skill '{skill}' — checked "
            f"{len(window_calls)} tool calls in the step's window"
        )


# ---------------------------------------------------------------------------
# Template-driven sub-agent spawn tests
# ---------------------------------------------------------------------------


class TestSubAgentSpawning:
    """Steps whose templates expect sub-agent spawning actually spawn one."""

    @pytest.mark.parametrize("step", _spawn_params())
    def test_subagent_spawned_in_step(self, step):
        windows = _windows_cache()
        calls = _calls_cache()
        window = windows.get(step)
        assert window is not None, (
            f"Step '{step}' was never entered — cannot verify sub-agent spawning"
        )
        window_calls = calls[window.start : window.end]
        spawned = any(c.type in ("Task", "Agent") for c in window_calls)
        assert spawned, (
            f"Step '{step}' expected sub-agent spawning but found no "
            f"Task/Agent tool_use in its window "
            f"({len(window_calls)} tool calls)"
        )


# ---------------------------------------------------------------------------
# Artefact content tests
# ---------------------------------------------------------------------------


def _plan_sections():
    plan_path, _, _ = plan_artefact_paths()
    if not plan_path.exists():
        return {}
    return parse_sections(plan_path.read_text())


class TestPlanMdContent:
    """plan.md contains every expected section with substantive content."""

    @pytest.mark.parametrize("section", EXPECTED_PLAN_SECTIONS)
    def test_section_present(self, section):
        sections = _plan_sections()
        assert section in sections, (
            f"Section '{section}' missing from plan.md; "
            f"found sections: {sorted(sections.keys())}"
        )

    @pytest.mark.parametrize("section", EXPECTED_PLAN_SECTIONS)
    def test_section_has_content(self, section):
        sections = _plan_sections()
        content = sections.get(section, "")
        assert len(content) >= MIN_SECTION_LENGTH, (
            f"Section '{section}' too short ({len(content)} chars, "
            f"need ≥ {MIN_SECTION_LENGTH})"
        )


class TestContextAndResearch:
    """context.md and research.md exist with substantive content."""

    def test_context_md_has_content(self):
        _, context_path, _ = plan_artefact_paths()
        assert context_path.exists(), f"context.md missing at {context_path}"
        content = context_path.read_text()
        # Arbitrary but nontrivial floor: at least 500 chars of real content.
        cleaned = re.sub(r"<!--.*?-->", "", content, flags=re.DOTALL)
        assert len(cleaned.strip()) >= 500, (
            f"context.md too short ({len(cleaned.strip())} chars)"
        )

    def test_research_md_has_content(self):
        _, _, research_path = plan_artefact_paths()
        assert research_path.exists(), f"research.md missing at {research_path}"
        content = research_path.read_text()
        cleaned = re.sub(r"<!--.*?-->", "", content, flags=re.DOTALL)
        assert len(cleaned.strip()) >= 500, (
            f"research.md too short ({len(cleaned.strip())} chars)"
        )

    def test_research_names_alternatives(self):
        """The research log should enumerate at least one rejected option."""
        _, _, research_path = plan_artefact_paths()
        if not research_path.exists():
            pytest.skip("research.md missing")
        text = research_path.read_text().lower()
        assert "reject" in text or "alternative" in text, (
            "research.md does not mention 'rejected' or 'alternative' "
            "— the decision log is likely incomplete"
        )


# ---------------------------------------------------------------------------
# Instruction next_step validity invariant
# ---------------------------------------------------------------------------


class TestInstructionNextStepValidity:
    """Every rendered instruction's next_step directive must reference a
    step name the state machine actually accepts.

    This is the formal guard for the discovery → approach drift class of
    bug. Each plan-CLI Bash call's tool_result is parsed as JSON and the
    `instruction` string is scanned for `plan goto --data '{"step":"X"}'`;
    X must appear in EXPECTED_STEP_ORDER.
    """

    def test_every_rendered_next_step_is_valid(self):
        calls = _calls_cache()
        results = _results_cache()
        valid = set(EXPECTED_STEP_ORDER)
        offenders = []
        for call in calls:
            if call.type != "Bash":
                continue
            cmd = _bash_command(call)
            if not (_is_plan_new_call(cmd) or _is_plan_goto_call(cmd)):
                continue
            result_text = results.get(call.tool_use_id, "")
            if not result_text:
                continue
            try:
                payload = json.loads(result_text)
            except json.JSONDecodeError:
                # Not all stdout lines are pure JSON (error paths, wrapped
                # harness output). Skip.
                continue
            instruction = (
                payload.get("instruction", "") if isinstance(payload, dict) else ""
            )
            if not instruction:
                continue
            m = INSTRUCTION_NEXT_STEP_RE.search(instruction)
            if not m:
                continue
            next_step = m.group(1)
            if next_step not in valid:
                offenders.append(
                    {
                        "from_step": payload.get("step"),
                        "rendered_next_step": next_step,
                    }
                )
        assert not offenders, (
            f"Rendered instructions reference invalid next steps: {offenders}"
        )
