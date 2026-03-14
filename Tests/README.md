# ATD Test Suite

This folder contains the full test suite for the ATD toolchain. Tests are designed to be run by an IDE Agent (Gemini Flash 3) using the `atd` unified CLI binary and — in a second pass — via the MCP server interface.

## Structure

Tests are organized as chunked task files, numbered sequentially. Each file is self-contained and includes prerequisites, exact steps, expected outputs, and acceptance criteria.

## Phases

| Phase | Tasks | Scope | Model |
|---|---|---|---|
| 0 — Setup | 01 | `atd init` + binary smoke test | None |
| 1 — upsilonbattle CLI | 02–06 | Dissect, create atoms, tag, index, search | IDE Agent + local Ollama |
| 2 — upsilon Coldstart | 07–08 | Dissect coldstart docs, create atoms | IDE Agent |
| 3 — MCP Round | 09–13 | Same operations via `atd serve` MCP tools | IDE Agent (MCP) |
| 4 — Cold Start Pipeline | 14 | Full `atd-cold-start.sh` on upsilonbattle | Token-heavy — run last |
| 5 — Audit | 15 | Full `atd audit` + `atd congruence` | Token-heavy — run last |

## Prerequisites

1. **Binary compiled**: `cd scripts/cmd/atd && go build -o /home/bastien/.local/bin/atd .`
2. **Ollama available** (optional, for LLM tasks): remote at `192.168.1.10:11434` or local at `localhost:11434`
3. **Models pulled**: `llama3.2`, `nomic-embed-text`, optionally `qwen2.5-coder:14b`
4. **Start with task 01** — it initializes `.atd` in `upsilonbattle/` which every subsequent task depends on
5. **Skill directory**: `atd_management_skill/.agent/skill/SKILL.md`

## How to Read Task Files

Each task file follows this structure:
- **Objective** — what success looks like
- **Prerequisites** — tasks and tools needed before starting
- **Steps** — exact CLI commands or MCP JSON calls (copy-paste ready)
- **Expected Output** — what to observe
- **Acceptance Criteria** — checklist to mark the task done

## Running MCP Tests (Phase 3)

Before starting task 09:
1. Build and place `atd` binary somewhere accessible
2. Configure `.mcp.json` in the project root (see `atd_serve.atom.md` for examples)
3. Enable MCP in the IDE Agent — the agent will then use `atd_*` tools directly
