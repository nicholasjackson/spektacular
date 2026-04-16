# Feature: 19_codex_integration

## Overview

Spektacular will separate its agent-facing interface into two distinct layers — commands (for agents like Claude Code and Bob that support custom slash commands) and skills (instruction files that work across all agents). Each supported agent will have a defined integration interface capturing install details and invocation method. This enables Spektacular to ship Codex support out of the box, while also making it straightforward to add future agent integrations. Development teams benefit regardless of which AI coding agent they use — they get the same Spec Driven Development workflow everywhere.

## Requirements

- **Agent init command** — Users can run `spektacular init <agent>` to install the Spektacular integration for a specific agent (e.g. `codex`, `claude`, `bob`)
- **Agent-appropriate artefacts** — Each agent integration installs the appropriate artefacts for that agent (commands, skills, or both) based on what the agent supports
- **Commands and skills separation** — The system must separate agent-facing artefacts into commands and skills, where skills work across all agents
- **Agent integration interface** — Each agent must have a defined integration interface that captures its install behaviour and invocation method
- **Extensibility** — Adding support for a new agent requires only implementing that agent's integration — no changes to core workflow logic
- **Codex integration** — A new Codex integration must be implemented using the agent integration interface

## Constraints

No constraints. Breaking changes are acceptable.

## Acceptance Criteria

- **Init succeeds** — Running `spektacular init codex` (or `claude`, `bob`) completes without error and the expected artefacts are present on disk afterward
- **Artefact correctness** — After `init`, only artefacts valid for that agent exist (e.g. Codex gets skills but not Claude Code commands)
- **Artefact separation** — The installed artefacts for an agent that supports both commands and skills are split into distinct files for each type
- **Unknown agent error** — Every supported agent has a working integration; attempting to init an unsupported agent name returns a clear error
- **Isolated extensibility** — A new agent can be integrated by adding a single new implementation; existing agent integrations and workflow logic are unchanged
- **Codex workflow invocation** — After `spektacular init codex`, a Codex user can invoke spec, plan, and implement workflows using skill syntax (e.g. `$spek-plan`) without error

## Technical Approach

- Define an `Agent` interface in Go that each integration (Claude, Bob, Codex) implements, capturing install behaviour and the artefact types it supports
- Skills are markdown instruction files installable by any agent integration
- Commands are the existing Claude Code / Bob slash command format, only installed by agents that support them
- The `init` command dispatches to the appropriate `Agent` implementation based on the argument provided
- Codex skills use `$skill-name` invocation syntax and are installed into the path Codex expects (`.codex/skills/`)

## Success Metrics

Manual testing by the developer after delivery.

## Non-Goals

None defined.
