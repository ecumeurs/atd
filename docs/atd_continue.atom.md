---
id: atd_continue
human_name: "ATD Continue Protocol"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, ide, fallback, pipeline]
parents:
  - [[atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Continue Protocol

## INTENT
To enable seamless resumption of multi-step ATD workflows after the IDE Agent completes delegated LLM tasks, by parsing a task_list.md file, checking for completed .result files, and advancing the pipeline.

## THE RULE / LOGIC
When an LLM-requiring tool detects IDE Agent fallback, it:
1. Writes the prompt to `pipeline_output/<tool>_<target>.prompt`
2. Writes a structured `task_list.md` with checkable items, output format specs, and context
3. Exits with a message directing the IDE Agent to the task list

The IDE Agent reads the task list, processes each pending item, writes results to `.result` files, then calls `atd continue <task_list.md>`. The continue command:
1. Parses `task_list.md` for `[ ]` entries
2. Checks if referenced `.result` files exist
3. Marks completed entries as `[x]`
4. Reports next pending task or "all complete"

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd continue <task_list.md>`
- **Pipeline dir:** `pipeline_output/` relative to project root
- **Code Tag:** `@spec-link [[atd_continue]]`
