"""Verify that the spektacular spec workflow completed successfully."""

import json
from pathlib import Path

PROJECT_DIR = Path("/app")
SPEK_DIR = PROJECT_DIR / ".spektacular"
STATE_FILE = SPEK_DIR / "state.json"
SPEC_FILE = SPEK_DIR / "specs" / "user-auth.md"

EXPECTED_STEPS = [
    "new",
    "overview",
    "requirements",
    "acceptance_criteria",
    "constraints",
    "technical_approach",
    "success_metrics",
    "non_goals",
    "verification",
]


def load_state() -> dict:
    assert STATE_FILE.exists(), f"State file not found at {STATE_FILE}"
    return json.loads(STATE_FILE.read_text())


def test_project_initialized():
    """spektacular init was run successfully."""
    assert (SPEK_DIR / "config.yaml").exists(), "config.yaml missing - init not run"


def test_spec_file_exists():
    """The spec markdown file was created."""
    assert SPEC_FILE.exists(), f"Spec file not found at {SPEC_FILE}"


def test_workflow_reached_finished():
    """Workflow current_step is finished or done."""
    state = load_state()
    assert state["current_step"] in ("finished", "done"), (
        f"Workflow did not finish, stuck at: {state['current_step']}"
    )


def test_all_steps_completed():
    """Every expected step appears in completed_steps."""
    state = load_state()
    completed = state.get("completed_steps", [])
    for step in EXPECTED_STEPS:
        assert step in completed, f"Step '{step}' not in completed_steps: {completed}"


def test_spec_has_meaningful_content():
    """Spec file has substantial content, not just a scaffold."""
    content = SPEC_FILE.read_text()
    assert len(content) > 500, (
        f"Spec content too short ({len(content)} chars), likely just scaffold"
    )


def test_spec_has_sections():
    """Spec file contains expected markdown sections."""
    content = SPEC_FILE.read_text()
    expected_headings = [
        "overview",
        "requirements",
        "acceptance criteria",
        "constraints",
        "technical approach",
        "success metrics",
        "non-goals",
    ]
    content_lower = content.lower()
    for heading in expected_headings:
        assert heading in content_lower, (
            f"Section '{heading}' not found in spec content"
        )


def test_spec_not_placeholder_text():
    """Spec sections contain real content, not TODO markers."""
    content = SPEC_FILE.read_text()
    placeholder_markers = ["TODO", "TBD", "FIXME", "placeholder", "[insert"]
    for marker in placeholder_markers:
        assert marker not in content, (
            f"Found placeholder marker '{marker}' in spec content"
        )
