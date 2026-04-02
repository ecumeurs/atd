---
id: mechanic_atd_assemble
human_name: "ATD Assemble"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, assemble, snapshot]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Assemble

## INTENT
To recursively stitch atoms together following their dependent chains, producing a unified document for human review or an LLM-ready narrative/summarization.

## THE RULE / LOGIC
Starting from one or more root atom IDs, recursively gathers content through the exploration graph (preventing cycles via visited set). Traversal can be restricted to only ascend (`--only-parents`) or only descend (`--only-dependents`) via flags. The tool automatically prompts the LLM to process the concatenated fragments according to the provided `--intent` (defaulting to Executive Summary) and `--length`.

If `--structured` is provided, the tool organizes gathered atoms by `Layer` and performs a multi-pass LLM procedure (summarizing Customer, Architecture, and Implementation layers individually before drawing a final intent conclusion). If no LLM is available, this mode falls back gracefully to sequentially grouped concatenation.

With `--json`, the system outputs the raw or LLM-summarized strings along with an array of involved atoms `metadata` for programmatic consumption by the WebUI.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd assemble --starts <ids> [--intent <text>] [--length <short|default|extended|long>] [--structured] [--json] [--only-parents] [--only-dependents]`
- **LLM Task:** `assemble`, `assemble_layer_*`, `assemble_final`
- **Code Tag:** `@spec-link [[mechanic_atd_assemble]]`
