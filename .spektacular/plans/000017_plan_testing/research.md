# Research: 17_plan_testing

## Alternatives considered and rejected

### Option A: Fully static mirror of `test_spec_workflow.py`

Hard-code `EXPECTED_STEP_ORDER`, `EXPECTED_SKILLS_PER_STEP`, `EXPECTED_SPAWN_STEPS`, and `EXPECTED_PLAN_SECTIONS` directly in the pytest verifier as literal constants, matching the existing spec test pattern one-to-one.

**Chosen — this is the final design.** Initially I read the spec's line 102 ("Parse skill and sub-agent references directly from `templates/plan-steps/*.md`") as a hard requirement and designed a runtime template parser. During implementation review the user pushed back: "if you parse everything at runtime you are not testing the behaviour". Correct — deriving the expected values from the same templates the agent reads makes the test tautological ("the agent did what the template said"), not a behavioural check on the template itself. The spec's line 102 was a convenience ask, not a soundness requirement; Option A is the correct design. Maintenance cost: when a step template adds or renames a skill reference, the corresponding entry in `EXPECTED_SKILLS_PER_STEP` must be updated in the same commit. That manual update is the behavioural confirmation — a human consciously acknowledges "yes, this change is intentional".

### Option C: Static skeleton + external YAML metadata file

Hand-maintain a `plan_steps_meta.yml` alongside the test listing `{step_name: {skills: [...], spawns_agents: bool, artefact: ...}}`. Verifier reads it at startup and runs parameterised checks over it.

**Rejected**: Has the same failure mode as Option A — the metadata file drifts silently from the templates exactly like a hard-coded map would. The only benefit is a cleaner diff when a human updates it, which is insufficient given that the whole point of the test is to catch drift a human did *not* notice.

### Option D: Unit-test the plan step templates in Go

Write Go unit tests that parse the template files at compile time and assert structural properties, bypassing the harbor/agent layer entirely.

**Rejected**: Would not exercise the plan CLI, state machine transitions, agent-driven retrieval of skills, or sub-agent spawning at runtime. The spec is explicit that the test "drives the full plan workflow from start to terminal step" (`.spektacular/specs/17_plan_testing.md:28`) and "asserts the agent invoked the corresponding `plan new` / `plan goto` command" (`.spektacular/specs/17_plan_testing.md:34`). Runtime agent behaviour cannot be asserted from a pure Go unit test; the harbor layer is load-bearing.

### Option E: Derive the canonical step list from `spektacular plan steps` at runtime

Shell out to `spektacular plan steps` inside the verifier container and use its JSON output directly as the expected order. This is literally what the spec's technical approach suggests: "Derive the canonical step list from the plan state machine source of truth" (`.spektacular/specs/17_plan_testing.md:101`).

**Rejected**: the spec's guidance is wrong for this particular check, and the user identified why during plan review. If the verifier consults the state machine for the expected order and then asserts `completed_steps` matches that order, a reordering bug in `Steps()` passes silently — the test simply defers to the broken source. There is no independent behavioural check on whether `Steps()` itself is correct. The test needs an oracle that lives *outside* the state machine. The chosen approach is a hand-maintained `EXPECTED_STEP_ORDER` list in the verifier. Because the agent follows the state machine's order at runtime, a reorder in `Steps()` surfaces as a diff between `completed_steps` and `EXPECTED_STEP_ORDER` in the existing `test_steps_executed_in_order` assertion — no separate introspection test is needed. Derived-from-templates via file-number prefix has the same failure mode — if someone renames `02-discovery.md` to `02-diskovery.md` while also updating `Steps()`, the templates-and-state-machine-agree check still passes, but both are now wrong relative to the original spec. Only a hand-maintained list in the verifier is immune.

### Option F: Assert sub-agent spawn by counting tool_use blocks whole-transcript

Drop the per-step window requirement and just assert that `Task`/`Agent` blocks appear *somewhere* in the transcript.

**Rejected**: The spec is explicit that the assertion is per-step: "Sub-agents referenced in a step are spawned **during that step**" (`.spektacular/specs/17_plan_testing.md:80`). Whole-transcript counting would pass even if the agent spawned sub-agents in an unrelated step, defeating the point.

## Chosen approach — evidence

**Step-order oracle**: a hand-maintained `EXPECTED_STEP_ORDER` list in the verifier module, mirroring the convention in `test_spec_workflow.py:21-31`. The state machine is *not* the source of truth for expected order — it is the subject under test. Because the agent follows the state machine's order at runtime, any reorder or rename in `Steps()` surfaces as a diff between `completed_steps` and `EXPECTED_STEP_ORDER` in `test_steps_executed_in_order`. No separate state-machine introspection test is needed — that would be a second assertion catching the same bug class.

**State file layout**: `cmd/plan.go:69-75` (`planDataDir`) and `cmd/plan.go:78-80` (`planStateFilePath`) confirm the state file lives at `.spektacular/plan-<name>/state.json`. The artefact files live at `.spektacular/plans/<name>/{plan.md,context.md,research.md}` because `store.NewFileStore(filepath.Join(dataDir, ".."))` at `cmd/plan.go:136` sets the store root to `.spektacular/` and `internal/plan/steps.go:15-27` defines the `plans/<name>/<file>` subpaths.

**State struct shape**: `internal/workflow/state.go:14-20` defines `State{CurrentStep, CompletedSteps, CreatedAt, UpdatedAt, Data}` with canonical JSON tags. This is the struct `.spektacular/plan-<name>/state.json` deserialises to and the verifier parses as plain JSON.

**Plan Result JSON shape**: `internal/plan/result.go` defines `Result{Step, PlanPath, PlanName, Instruction}`. `cmd/plan.go:16-24` documents the same shape as `planResultOutputSchema`. The verifier can rely on these four field names in the JSON emitted by every `plan new` / `plan goto` call.

**Plan step templates and their skill references**:
- `templates/plan-steps/02-discovery.md:9-10,22` — references `discover-project-commands`, `discover-test-patterns`, `spawn-planning-agents`, and contains the phrase "agent orchestration capability to parallelize".
- `templates/plan-steps/10-phases.md:30` — references `spawn-implementation-agents`.
- `templates/plan-steps/13-verification.md:7-13` — references `gather-project-metadata` and `determine-feature-slug`.
- `templates/plan-steps/14-finished.md:13` — mentions `share-docs` as an optional follow-up, not a required retrieval in the happy path.
- All other plan-step templates (`01`, `03`-`09`, `11`-`12`) contain no skill references.

**Skill CLI entrypoint**: `cmd/skill.go` implements `spektacular skill <name>` which reads from the embedded skills filesystem and returns `{name, title, instructions}` JSON. Agents can retrieve a skill either via the Claude Code `Skill` tool or via this CLI, and the verifier must accept either as evidence of retrieval.

**Reference harbor task to model after**: `tests/harbor/spec-workflow/` with its full five-file layout (`task.toml`, `instruction.md`, `environment/Dockerfile`, `solution/solve.sh`, `tests/{test.sh,test_spec_workflow.py}`). `test_spec_workflow.py:80-106` and `:109-123` are direct templates for the transcript extractor and CLI-call finder. `test_spec_workflow.py:51-77` is the template for the `parse_sections` helper.

**The drift bug that motivates the `next_step` validity invariant**: `internal/plan/steps.go:65` passes `"approach"` as the discovery step's `nextStep` template variable, but the state machine at `internal/plan/steps.go:35` expects `"architecture"` as the source state for the next transition. Running `go run . plan goto --data '{"step":"approach"}'` returns `{"error":"event approach does not exist"}`, confirmed during this planning session. This is exactly the class of regression the Phase 4.2 `test_every_rendered_next_step_is_valid` invariant is meant to catch. The bug is fixed in Phase 1.1 of this plan as a precondition for the harbor test to complete its happy-path run.

**Harbor Makefile pattern**: `Makefile:28-34` shows the existing `harbor-test` target shape. The new `plan-harbor-test` target is a near-copy with the path replaced and a different output glob.

## Files examined

- `.spektacular/specs/17_plan_testing.md` — source spec; drove requirements, acceptance criteria, and technical approach direction.
- `tests/harbor/spec-workflow/task.toml:1-25` — harbor task metadata template including `[agent]`/`[verifier]` timeouts and `[[artifacts]]` capture.
- `tests/harbor/spec-workflow/instruction.md:1-49` — agent instruction template; reusable for the plan task with the spec-writing step replaced by a plan-driving step.
- `tests/harbor/spec-workflow/environment/Dockerfile:1-15` — Dockerfile template; plan task reuses it verbatim plus a baked-in seed spec copy.
- `tests/harbor/spec-workflow/tests/test.sh:1-10` — test runner entrypoint template (one-line change: script name).
- `tests/harbor/spec-workflow/tests/test_spec_workflow.py:1-382` — entire spec verifier as a pattern source. Key pieces: helpers at `:41-123`, `TestWorkflow` class at `:130-170`, per-step classes at `:177-382`.
- `tests/harbor/spec-workflow/solution/solve.sh:1-99` — reference solution pattern. Plan task's solution drives 15 steps instead of 10 and writes three files at the verification step instead of one.
- `Makefile:1-34` — build and harbor-test targets; pattern for `plan-harbor-test`.
- `internal/plan/steps.go:1-229` — plan state machine and template rendering. `Steps()` at `:30-48` is the canonical step list. `:65` holds the drift bug. `:153-163` (finished callback) asserts the three artefact files exist. `:166-202` (writeStepResult) shows the template variable set passed to mustache.
- `cmd/plan.go:1-319` — CLI surface. `:16-24` output schema; `:69-75` data dir path; `:78-80` state file path; `:82-148` new handler; `:150-216` goto handler; `:253-270` steps handler; `:273-307` findActivePlan.
- `cmd/skill.go` — skill retrieval CLI; returns `{name,title,instructions}` JSON for a given skill name.
- `internal/workflow/state.go:14-20` — State struct and JSON field names.
- `internal/plan/result.go` — Result struct with the four-field JSON shape every plan command returns.
- `templates/plan-steps/01-overview.md` through `14-finished.md` — all 14 step templates; inspected at research time to extract the skill and sub-agent expectations hand-copied into `EXPECTED_SKILLS_PER_STEP` and `EXPECTED_SPAWN_STEPS`. Referenced skill counts per file: `02` has 3, `10` has 1, `13` has 2, `14` has 1 (aspirational, excluded from the map). All others have 0.
- `thoughts/notes/commands.md` (created during discovery) — project command reference.
- `thoughts/notes/testing.md` (created during discovery) — testing conventions reference.

## External references

None. This plan is entirely internal — no RFCs, papers, library docs, or blog posts were consulted. The only external contracts depended on are (a) the Claude Code agent transcript JSONL format at `/logs/agent/claude-code.txt` (documented by `test_spec_workflow.py`'s existing usage) and (b) the harbor CLI's container orchestration contract (documented by the existing `make harbor-test` target). Both are already in production use.

## Prior plans / specs consulted

- `.spektacular/specs/17_plan_testing.md` — the source spec; everything traces back to this.
- `.spektacular/plans/15_implementation/` and `.spektacular/plans/16_plan_format/` — visible in `git status` as deleted state files on the current `skills` branch. Not consulted for technical content; only noted as evidence that `findActivePlan()`'s "most-recently-modified" selection logic (`cmd/plan.go:273-307`) can encounter multiple plan directories in a single `.spektacular/`. The harbor container starts from a clean `/app`, so this is not a concern for the test itself.

## Open assumptions

The following assumptions were made during planning but not independently verified. If any turn out wrong, the implement workflow must STOP and ask before proceeding.

1. **Claude Code transcript schema stability.** We assume the JSONL at `/logs/agent/claude-code.txt` contains `{"type":"assistant","message":{"content":[...]}}` records with `tool_use` blocks whose `name` field is literally `"Bash"`, `"Skill"`, `"Task"`, or `"Agent"` depending on the tool invoked, and that `tool_result` blocks from `"type":"user"` records carry the command's stdout verbatim in their `content`. This is how `test_spec_workflow.py:80-106` currently parses transcripts and we assume the same shape extends to the additional tool types. **If the Skill/Task tool_use block names differ (e.g. model-version-specific aliases), the Phase 3.2 and 3.3 assertions will need adjusting.**
2. **`phases` step sub-agent expectation.** Our heuristic flags `phases` as expecting sub-agent spawn because `templates/plan-steps/10-phases.md:30` references `spawn-implementation-agents`. The template actually uses that skill to *describe* per-phase agent strategy in context.md rather than to direct the planner agent to spawn sub-agents itself. **If this produces false-positive failures in Phase 3.3, the heuristic should be tightened to require both a spawn skill reference and an explicit "parallelize/orchestrate now" instruction — or the `phases` step should be manually excluded.**
3. **`spektacular plan steps` JSON stability.** The verifier does not consume this command at runtime; it is mentioned here only as a tool maintainers can use to sanity-check `EXPECTED_STEP_ORDER` by hand when editing `Steps()`. If `cmd/plan.go:253-270` changes its output shape, the verifier is unaffected.
4. **Happy-path reference solution can drive every step.** Some step templates instruct the agent to present drafts to the user for review; a bash `solve.sh` cannot meaningfully "present to a user" — it just writes content and advances. We assume the state machine does not enforce any human-in-the-loop gating and will accept successive `plan goto` calls without intermediate confirmation. This matches the existing spec workflow's reference solution pattern.
5. **Skill and spawn expectations are hand-synced with templates.** `EXPECTED_SKILLS_PER_STEP` and `EXPECTED_SPAWN_STEPS` mirror the current content of `templates/plan-steps/*.md`. If a template changes its skill references or adds sub-agent orchestration language, the maps must be updated in the same commit. There is no runtime check enforcing this — it is the maintainer's responsibility, and the resulting test failure when drift occurs is the alert mechanism.
6. **`share-docs` skill in `14-finished.md` is not a required retrieval.** The text "When ready to share the plan with the team, use the `share-docs` skill" (`templates/plan-steps/14-finished.md:13`) is aspirational guidance. The verifier should either exclude this skill from the expectation set for `finished` or treat its absence as non-blocking. **If the team wants `share-docs` to be a hard requirement in CI, the heuristic needs an explicit allowlist/blocklist mechanism.**

## Rehydration cues

If context is lost mid-implementation and a cold session needs to pick up this plan:

1. **Start here**: read `.spektacular/specs/17_plan_testing.md` (the source spec) and this research file. Together they give the full scope.
2. **See what already exists**: `ls tests/harbor/` — if `plan-workflow/` exists, inspect `tests/harbor/plan-workflow/tests/test_plan_workflow.py` to see which phases have landed.
3. **Canonical step list at runtime**: `go run . plan steps` returns the authoritative ordered list. Do not hand-maintain it.
4. **Drift bug verification**: `go run . plan new --data '{"name":"smoke"}'` then advance through each step manually by copying the `next_step` command from each rendered instruction. If the discovery step's instruction says `approach` instead of `architecture`, Phase 1.1 has not landed.
5. **Project command reference**: `thoughts/notes/commands.md`.
6. **Testing pattern reference**: `thoughts/notes/testing.md` — covers harbor verifier conventions including the per-step class structure and the `/logs/verifier/reward.txt` contract.
7. **Reference implementation to mirror**: `tests/harbor/spec-workflow/` — every file has a sibling in the new plan-workflow task with small deltas.
8. **Skill details for a template-referenced skill**: `go run . skill <name>`. Useful when verifying the heuristic picked up the right skill.
9. **State file location for running a manual test**: `.spektacular/plan-<name>/state.json` (singular prefix). The artefact files live under `.spektacular/plans/<name>/` (plural).
10. **Relevant Go source for any CLI contract surprises**: `cmd/plan.go` (CLI handlers), `internal/plan/steps.go` (state machine and template rendering), `internal/workflow/state.go` (state struct), `internal/plan/result.go` (CLI JSON shape).
