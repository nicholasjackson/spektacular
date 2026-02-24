Plan updated with all three decisions incorporated. Here's the summary of changes:

## Updated Plan Summary

**4 phases** (down from 5 — removed the `config set/get` phase):

| Phase | What | Files |
|-------|------|-------|
| **1** | Protocol message models (Pydantic) | `src/spektacular/protocol.py` (new) |
| **2** | Replace `run` placeholder with CLI mode | `cli.py:38-44` (replaced), `config.py` (+1 line), `adhoc.py` (new) |
| **3** | JSON mode with JSONL stdin/stdout event loop | `adhoc.py` (extended) |
| **4** | Tests (26 test cases) | `test_protocol.py` + `test_adhoc.py` (new), `test_config.py` (extended) |

**Key changes from original plan:**
- `spektacular run` replaces the TODO at `cli.py:38-44` instead of adding a new `adhoc` command
- No `spektacular config set/get` commands — output mode persisted only via manual config.yaml editing
- Full integration with `run_claude()` — no stubs or dry-run mode

All 3 plan files are at `.spektacular/plans/5_adhoc_json_protocol/`:
- `plan.md` — Full implementation plan with code examples
- `research.md` — Research notes, patterns, design decisions
- `context.md` — Quick reference for execution