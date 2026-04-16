# Context: 17_plan_testing

## Current State Analysis

The repo has one existing harbor task, `tests/harbor/spec-workflow/`, that exercises the `spec` workflow end-to-end. It is built as a linux/amd64 binary into `tests/harbor/spec-workflow/environment/spektacular` by the Makefile target `harbor-test` (`Makefile:28-34`). The verifier is a pytest module at `tests/harbor/spec-workflow/tests/test_spec_workflow.py` that:

- Hard-codes `EXPECTED_STEP_ORDER` (`test_spec_workflow.py:21-31`) as a list of spec step names, mirroring but not importing from `internal/spec/steps.go`.
- Loads `.spektacular/state.json` via `load_state()` (`test_spec_workflow.py:41-43`).
- Reads the produced spec markdown via `load_spec()` and splits it into lowercase-keyed sections via `parse_sections()` (`test_spec_workflow.py:46-77`). HTML comments are stripped before section splitting.
- Parses the JSONL transcript at `/logs/agent/claude-code.txt` via `extract_tool_calls()` (`test_spec_workflow.py:80-106`) — currently only captures `Bash` `tool_use` blocks and returns a list of `{"command": "..."}` dicts.
- Greps bash commands for `spektacular spec new` / `spektacular spec goto --data '{"step":"..."}'` via `find_spektacular_calls()` (`test_spec_workflow.py:109-123`).
- Has one pytest class per spec step plus a `TestWorkflow` class for cross-cutting checks.

The plan workflow implementation lives at:

- `internal/plan/steps.go` — the state machine. `Steps()` at `internal/plan/steps.go:30-48` declares the canonical 15-step order: `new → overview → discovery → architecture → components → data_structures → implementation_detail → dependencies → testing_approach → milestones → phases → open_questions → out_of_scope → verification → finished`. Terminal state is `"finished"`.
- `internal/plan/steps.go:65` — **contains the drift bug Phase 1.1 fixes**: `writeStepResult("discovery", "approach", "plan-steps/02-discovery.md", ...)`. The second argument is the `next_step` template variable; it should be `"architecture"` (the actual source state for the next step at `internal/plan/steps.go:35`).
- `internal/plan/result.go` — defines `Result{Step, PlanPath, PlanName, Instruction}` (the JSON response type for every plan CLI command).
- `cmd/plan.go` — CLI surface. `planDataDir()` at `cmd/plan.go:69-75` returns `.spektacular/plan-<name>/`, `planStateFilePath()` at `cmd/plan.go:78-80` appends `state.json`. `runPlanNew()` at `cmd/plan.go:82-148` creates the data dir, seeds state, and runs the first step. `runPlanGoto()` at `cmd/plan.go:150-216` reads the most-recently-modified `plan-*` dir via `findActivePlan()` and transitions. `planResultOutputSchema` at `cmd/plan.go:16-24` documents the output JSON. `runPlanSteps()` at `cmd/plan.go:253-270` returns the canonical step list via `plan.StepsResult`.
- `internal/workflow/state.go:14-20` — the `State{CurrentStep, CompletedSteps, CreatedAt, UpdatedAt, Data}` struct persisted as `state.json`.
- `templates/plan-steps/` — 14 step templates (`01-overview.md` through `14-finished.md`). All rendered via mustache at `internal/plan/steps.go:213-219`, with template variables including `{{config.command}}`, `{{step}}`, `{{title}}`, `{{next_step}}`, `{{plan_path}}`, `{{context_path}}`, `{{research_path}}`, `{{spec_path}}`, etc.
- `cmd/skill.go` — the skill retrieval command. `go run . skill <name>` reads from an embedded filesystem and returns `{name, title, instructions}` JSON.

Plan artefact files live at `.spektacular/plans/<name>/plan.md`, `.spektacular/plans/<name>/context.md`, `.spektacular/plans/<name>/research.md` — note the path is `plans/` (plural) whereas the state file is under `plan-<name>/` (singular prefix). This asymmetry is because `store.NewFileStore(filepath.Join(dataDir, ".."))` at `cmd/plan.go:136` sets the store root to `.spektacular/`, and `PlanFilePath(name)` at `internal/plan/steps.go:15-17` returns `plans/<name>/plan.md`. The `finished` step asserts all three files exist at `internal/plan/steps.go:153-163`.

Template-referenced skills (confirmed by grep at research time):

- `templates/plan-steps/02-discovery.md:9` — `discover-project-commands`
- `templates/plan-steps/02-discovery.md:10` — `discover-test-patterns`
- `templates/plan-steps/02-discovery.md:22` — `spawn-planning-agents` (+ "agent orchestration capability to parallelize")
- `templates/plan-steps/10-phases.md:30` — `spawn-implementation-agents`
- `templates/plan-steps/13-verification.md:8` — `gather-project-metadata`
- `templates/plan-steps/13-verification.md:13` — `determine-feature-slug`
- `templates/plan-steps/14-finished.md:13` — `share-docs` (aspirational, not required in happy path)

Discovery is the only step whose template unambiguously invites sub-agent spawning (`templates/plan-steps/02-discovery.md:22`).

## Per-Phase Technical Notes

### Phase 1.1: Fix discovery next_step

Change `internal/plan/steps.go:65` from:

```go
return "", writeStepResult("discovery", "approach", "plan-steps/02-discovery.md", data, out, st, cfg)
```

to:

```go
return "", writeStepResult("discovery", "architecture", "plan-steps/02-discovery.md", data, out, st, cfg)
```

- `internal/plan/steps.go:65` — the one-line change.
- Verify by rendering: `go run . plan new --data '{"name":"smoke"}'` then advancing to discovery and inspecting the `instruction` field of the goto response. The rendered instruction's tail should print `plan goto --data '{"step":"architecture"}'`.
- No callers or templates reference the string `"approach"` — grep returns only `internal/plan/steps.go:65` and `templates/plan-steps/02-discovery.md:46` (the `{{next_step}}` placeholder).
- `internal/plan/steps.go:35` (the `architecture` step declaration `Src: []string{"discovery"}`) already accepts `discovery` as the source state, so no state-machine edit is needed.

**Complexity**: Low
**Token estimate**: ~3k
**Agent strategy**: Single agent, sequential (trivial edit + manual smoke)

### Phase 2.1: Harbor task scaffold

Create `tests/harbor/plan-workflow/` with files modelled after `tests/harbor/spec-workflow/`:

- `tests/harbor/plan-workflow/task.toml` — copy of `tests/harbor/spec-workflow/task.toml` with metadata strings updated. Keep `[agent] timeout_sec = 600.0` and `[verifier] timeout_sec = 120.0` as baseline. Update `difficulty_explanation`, `tags`, and the `[[artifacts]] source` (still `/app/.spektacular`).
- `tests/harbor/plan-workflow/instruction.md` — modelled on `tests/harbor/spec-workflow/instruction.md`. Must: (a) instruct the agent to run `spektacular init claude` first, (b) create a seed spec via the existing `/spek:new` skill or by writing the spec file directly so the plan workflow has something to plan against, (c) drive the plan workflow with `/spek:plan <spec-name>`, (d) copy `/app/.spektacular` to `/logs/artifacts/` at the end.
- `tests/harbor/plan-workflow/environment/Dockerfile` — copy of `tests/harbor/spec-workflow/environment/Dockerfile:1-15`. The spektacular binary goes at `/usr/local/bin/spektacular`. Templates are NOT baked into the image — they ride in via the harbor-mounted `/tests/` dir (see Phase 2.3).
- `tests/harbor/plan-workflow/solution/solve.sh` — reference solution. Drives `plan new` then walks every canonical step with `plan goto --data '{"step":"<name>"}'`, piping draft content via `--stdin` where the step template expects it. At the verification step, writes the three artefact files directly (the workflow's verification template uses `Write` tool semantics; in a bash solution we write them with heredocs). At `finished`, copies `.spektacular` to `/logs/artifacts/`.
- `tests/harbor/plan-workflow/tests/test.sh` — identical shape to `tests/harbor/spec-workflow/tests/test.sh:1-10`: `pytest /tests/test_plan_workflow.py -v` then write `1`/`0` to `/logs/verifier/reward.txt`.
- `tests/harbor/plan-workflow/tests/test_plan_workflow.py` — initial stub: a single `test_not_implemented()` that asserts `False` with a clear message. Will be filled in during Phase 2.2+.

Paths the verifier will need to know:
- Plan state: `/app/.spektacular/plan-<name>/state.json`
- Plan artefacts: `/app/.spektacular/plans/<name>/{plan.md,context.md,research.md}`
- Templates: NOT read at runtime. Expectation maps are hand-maintained in `test_plan_workflow.py` as literal constants.
- Transcript: `/logs/agent/claude-code.txt`

**Complexity**: Low
**Token estimate**: ~8k
**Agent strategy**: Single agent, sequential — touches only new files, no cross-cutting edits.

### Phase 2.2: Baseline pytest verifier

Fill in `tests/harbor/plan-workflow/tests/test_plan_workflow.py`. Module structure:

```python
PROJECT_DIR = Path("/app")
SPEK_DIR = PROJECT_DIR / ".spektacular"
TRANSCRIPT = Path("/logs/agent/claude-code.txt")

# Hand-maintained expectation maps. Do NOT derive any of these at runtime
# from the subject under test (state machine, templates) — that would make
# the verifier tautological.
EXPECTED_STEP_ORDER = [
    "new", "overview", "discovery", "architecture", "components",
    "data_structures", "implementation_detail", "dependencies",
    "testing_approach", "milestones", "phases", "open_questions",
    "out_of_scope", "verification", "finished",
]

EXPECTED_SKILLS_PER_STEP = {
    "discovery": frozenset({
        "discover-project-commands",
        "discover-test-patterns",
        "spawn-planning-agents",
    }),
    "phases": frozenset({"spawn-implementation-agents"}),
    "verification": frozenset({
        "gather-project-metadata",
        "determine-feature-slug",
    }),
}

EXPECTED_SPAWN_STEPS = frozenset({"discovery"})

def find_plan_state_file() -> Path:
    # Glob SPEK_DIR / "plan-*" / "state.json"; return the only match.
    ...

def load_state() -> dict: ...
def extract_tool_calls() -> list[dict]: ...   # Bash only at this phase
def find_plan_cli_calls(calls) -> list[str]: ...  # "plan new", "plan goto <step>"
```

Pytest classes:

- `TestWorkflow`:
  - `test_workflow_reached_finished` — `state.current_step == "finished"`.
  - `test_all_steps_completed` — every name in `EXPECTED_STEP_ORDER` appears in `state.completed_steps`.
  - `test_steps_executed_in_order` — `state.completed_steps == EXPECTED_STEP_ORDER`. **This is the load-bearing catch for reorder bugs in `Steps()`**: because the agent follows the state machine's order at runtime, any reorder surfaces here as a diff between the expected and observed lists. No separate state-machine introspection test is needed.
  - `test_no_placeholder_text` — scan all three artefact files for TODO/TBD/FIXME/placeholder/[insert.
- `TestStepX` (parameterised) — for each step in `EXPECTED_STEP_ORDER`, a test that the step is in `completed_steps` and that the expected CLI call is in the transcript. Use `@pytest.mark.parametrize("step", EXPECTED_STEP_ORDER)` on a single test function to generate per-step cases.

Key file:line references this phase depends on:

- `test_spec_workflow.py:21-31` — the existing hard-coded `EXPECTED_STEP_ORDER` pattern this verifier mirrors.
- `test_spec_workflow.py:80-106` — template for `extract_tool_calls`.
- `test_spec_workflow.py:109-123` — template for `find_*_calls`.
- `test_spec_workflow.py:148-163` — template for order-checking assertions.
- `cmd/plan.go:16-24` — the `planResultOutputSchema` we rely on for `step`/`plan_path`/`plan_name`/`instruction` field names.

**Maintenance note**: when a legitimate state-machine change lands (a new step added, a step renamed), `EXPECTED_STEP_ORDER` and `internal/plan/steps.go Steps()` must be updated in the same commit. `test_steps_executed_in_order` will fail loudly on the next harbor run if they diverge.

**Complexity**: Medium
**Token estimate**: ~15k
**Agent strategy**: Single agent, sequential. The file is small (~250 lines) and cohesive; parallelism offers no benefit.

### Phase 2.3: Makefile wiring

Extract the OAuth-token extraction pattern into a shared `HARBOR_AUTH` variable at the top of the Makefile, then add a `plan-harbor-test` target that builds the linux/amd64 binary and runs `harbor run`. Refactor the existing `harbor-test` target to use the same `HARBOR_AUTH` variable so both targets stay in sync:

```make
HARBOR_AUTH := ANTHROPIC_AUTH_TOKEN=$$(python3 -c "...")
HARBOR_MODEL := claude-sonnet-4-6

plan-harbor-test:
	GOOS=linux GOARCH=amd64 go build -o tests/harbor/plan-workflow/environment/spektacular .
	$(HARBOR_AUTH) harbor run -p tests/harbor/plan-workflow -a claude-code -m $(HARBOR_MODEL) -o tests/harbor/jobs
	@echo ""
	@echo "=== Test Results ==="
	@cat $$(ls -td tests/harbor/jobs/*/plan-workflow__*/verifier/test-stdout.txt 2>/dev/null | head -1)
```

- Add `plan-harbor-test` to the `.PHONY` line.
- No template copy step and no `.gitignore` change — the verifier's expectation maps are hand-maintained in `test_plan_workflow.py`, not read from the templates at runtime.
- Optional convenience target `all-harbor-tests: harbor-test plan-harbor-test` can be added later; not required by any acceptance criterion.

**Complexity**: Low
**Token estimate**: ~3k
**Agent strategy**: Single agent, sequential — one file.

### Phase 3.1: Step-window resolver

Add the step-window resolver helper:

```python
def resolve_step_windows(calls: list[ToolCall]) -> dict[str, StepWindow]:
    # Walk calls, find `plan new` and `plan goto --data '{"step":"<name>"}'` Bash
    # commands, identify the step name each transitions INTO, build [start, end)
    # windows. The implicit `overview` transition inside `plan new` shares `new`'s
    # window so assertions against `overview` still resolve.
    ...
```

No template parser. The `EXPECTED_SKILLS_PER_STEP` and `EXPECTED_SPAWN_STEPS` constants declared in Phase 2.2 are the hand-maintained oracles; Phases 3.2 and 3.3 iterate over them directly.

**Complexity**: Low
**Token estimate**: ~6k
**Agent strategy**: Single agent, sequential.

### Phase 3.2: Skill retrieval assertions

Extend `extract_tool_calls` to return `ToolCall` dataclasses and to capture `Skill` tool_use blocks and bash commands matching `spektacular skill <name>` (not just `plan` calls). Add a module-level parametrisation that iterates over the hand-maintained `EXPECTED_SKILLS_PER_STEP`:

```python
def _skill_params():
    return sorted(
        (step, skill)
        for step, skills in EXPECTED_SKILLS_PER_STEP.items()
        for skill in skills
    )

@pytest.mark.parametrize("step,skill", _skill_params())
def test_skill_retrieved(step, skill):
    window = _windows_cache().get(step)
    assert window is not None, f"Step '{step}' was never entered"
    window_calls = _calls_cache()[window.start:window.end]
    retrieved = any(
        (c.type == "Skill" and c.input.get("skill") == skill)
        or (c.type == "Bash" and f"skill {skill}" in c.input.get("command", ""))
        for c in window_calls
    )
    assert retrieved, f"Step '{step}' did not retrieve skill '{skill}'"
```

- The exact `Skill` tool_use schema comes from the Claude Code tool catalog. The input field is `{"skill": "<name>", "args": "..."}`. Depending on model version the block `name` may be `Skill` literally.
- The `(c.type == "Bash" and f"skill {skill}" in ...)` fallback catches the case where the agent uses the `spektacular skill <name>` CLI rather than the Skill tool — both count as "retrieved".

**Complexity**: Medium
**Token estimate**: ~10k
**Agent strategy**: Single agent, sequential.

### Phase 3.3: Sub-agent spawn assertions

Extend `extract_tool_calls` to capture `Task` and `Agent` tool_use blocks. Add a parametrised test:

```python
def _spawn_params():
    return sorted(EXPECTED_SPAWN_STEPS)

@pytest.mark.parametrize("step", _spawn_params())
def test_subagent_spawned(step, calls, windows):
    window = windows.get(step)
    if window is None:
        pytest.fail(f"Step '{step}' was never entered")
    window_calls = calls[window.start:window.end]
    spawned = any(c.type in ("Task", "Agent") for c in window_calls)
    assert spawned, f"Step '{step}' expected sub-agent spawn but none found"
```

**Complexity**: Low
**Token estimate**: ~6k
**Agent strategy**: Single agent, sequential.

### Phase 4.1: Artefact content assertions

Add artefact readers and per-section tests:

```python
PLAN_PATH = SPEK_DIR / "plans" / "<name>" / "plan.md"  # name from find_plan_state_file()
CONTEXT_PATH = SPEK_DIR / "plans" / "<name>" / "context.md"
RESEARCH_PATH = SPEK_DIR / "plans" / "<name>" / "research.md"
MIN_SECTION_LEN = 100
PLACEHOLDERS = ("TODO", "TBD", "FIXME", "placeholder", "[insert")

EXPECTED_PLAN_SECTIONS = (
    "overview", "architecture & design decisions", "component breakdown",
    "data structures & interfaces", "implementation detail", "dependencies",
    "testing approach", "milestones & phases", "open questions", "out of scope",
)

def parse_sections(text: str) -> dict[str, str]: ...  # port from test_spec_workflow.py:51-77
```

One parameterised test per expected section asserts presence and minimum length. One cross-cutting test scans all three files for placeholder markers. Topic-keyword checks for sections where ambiguity is low (Architecture → "harbor"/"test"/"template"; Component Breakdown → "parser"/"verifier").

The plan.md section list is hand-maintained in `EXPECTED_PLAN_SECTIONS` alongside the other expectation maps. Deriving it from the scaffold templates at runtime would violate the oracle-independence rule (see plan.md Architecture decision #1).

**Complexity**: Medium
**Token estimate**: ~15k
**Agent strategy**: Single agent, sequential.

### Phase 4.2: Instruction next_step validity invariant

Capture every JSON response from `plan new` / `plan goto` during the run. The reference solution's `solve.sh` can write each response to a `.jsonl` log file, or — more directly — the verifier can walk the transcript, find every assistant Bash call to `plan new` or `plan goto`, locate the corresponding `tool_result` block in the transcript (`type: "user"` with `tool_use_id` matching), parse its JSON, and extract the `next_step` from the `instruction` field using a regex like `plan goto --data '\{"step":"([^"]+)"\}'`. Then assert every extracted `next_step` is in `canonical_steps()` (or is the terminal `finished` with an empty next_step).

```python
NEXT_STEP_RE = re.compile(r"plan goto --data '\{\"step\":\"([a-z_]+)\"\}'")

def test_every_rendered_next_step_is_valid():
    canonical = set(canonical_steps())
    results = extract_plan_cli_tool_results()  # list of parsed JSON dicts
    offenders = []
    for r in results:
        m = NEXT_STEP_RE.search(r.get("instruction", ""))
        if m and m.group(1) not in canonical and m.group(1) != "":
            offenders.append((r.get("step"), m.group(1)))
    assert not offenders, f"Rendered instructions reference invalid next steps: {offenders}"
```

- The `tool_result` blocks for `Bash` calls in the Claude Code transcript contain the bash command's stdout. For our commands the stdout is the JSON response, which the verifier can `json.loads`.
- The regex is tolerant of whitespace — production instructions use the literal form shown above, but a small amount of normalisation (e.g. `re.sub(r"\s+", " ", instruction)` before matching) keeps it robust.

**Complexity**: Medium
**Token estimate**: ~10k
**Agent strategy**: Single agent, sequential.

## Testing Strategy

Every phase's implementation is verified by running `make plan-harbor-test` and reading the pytest output. Phases 2.2 through 4.2 are cumulative: each adds tests, each adds ways the harbor run can fail. The reference `solve.sh` is the ground truth for "correct agent behaviour" — if the solution passes and a regression in `internal/plan/` or a template causes it to fail, the test is doing its job.

A small regression matrix to run at the end of each milestone:

- Milestone 2 done → pass.
- Reorder `Steps()` → order test fails.
- Delete a `plan goto` from `solve.sh` → that step's CLI test fails.
- Milestone 3 done → still pass.
- Rename `discover-project-commands` in `02-discovery.md` → one skill test fails naming old vs new.
- Remove "agent orchestration" from `02-discovery.md` (and strip spawn-planning-agents) → discovery sub-agent test is dropped.
- Milestone 4 done → still pass.
- Delete the Component Breakdown section in `solve.sh` → section content test fails.
- Re-introduce Phase 1.1 bug → `next_step` validity invariant fails.

## Project References

- `thoughts/notes/commands.md` — Project commands (including `make harbor-test` and the new `make plan-harbor-test`).
- `thoughts/notes/testing.md` — Testing patterns including harbor verifier conventions.
- `tests/harbor/spec-workflow/` — Reference implementation for the harbor task layout.
- `.spektacular/specs/17_plan_testing.md` — The source spec this plan implements.

## Token Management Strategy

| Tier | Token Budget | Agent Strategy |
|------|-------------|----------------|
| Low | ~10k | Single agent, sequential |
| Medium | ~25k | 2-3 parallel agents |
| High | ~50k+ | Parallel analysis, sequential integration |

All phases in this plan are Low or Medium. None require parallel analysis because the surface area is narrow (one new directory of files plus one Go one-liner). Implementation agents should prefer single-session execution per phase rather than sharding.

## Migration Notes

None. The Phase 1.1 Go edit is backwards-compatible — the current string `"approach"` is never consumed by any state-machine transition, so replacing it with `"architecture"` removes an error path rather than creating a breaking change. Existing state files from prior plan runs are unaffected (the discovery step's persisted `CurrentStep` is `"discovery"`, not `"approach"`).

## Performance Considerations

- The harbor test adds one new container run to the `make plan-harbor-test` target, with timeouts inherited from the spec-workflow task (`600s` agent, `120s` verifier). Expected runtime is comparable to or slightly longer than `make harbor-test` because the plan workflow has 15 steps versus spec's 10.
- The verifier's template parsing happens once at module collection and caches results as module-level constants; pytest cost is negligible.
- The transcript extractor reads the JSONL file line by line; for typical run sizes (a few hundred to a few thousand lines) this is well under a second.
- No database, network, or disk-intensive work in the verifier beyond reading five files inside the container.
