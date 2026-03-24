---
id: atd_stats
human_name: "ATD Stats"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, stats, metrics, health]
parents:
  - [[atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Stats

## INTENT
To provide a quantitative overview of documentation health and coverage metrics for an ATD-enabled project.

## THE RULE / LOGIC
Analyzes all ATD atoms in the project to compute:
- Total atom count.
- Breakdown by `type` (e.g., MECHANIC, SERVICE, DOMAIN).
- Breakdown by `status` (DRAFT, REVIEW, STABLE).
- Breakdown by `domain`.
- Coverage ratio: Percentage of `STABLE` atoms that have at least one `@spec-link` implementation in the source code.
- Orphan count: Total number of `STABLE` atoms without any implementation.

The tool is deterministic and does not use LLMs.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd stats [--src <path>] [--docs <path>]`
- **Output:** JSON-formatted report.
- **Code Tag:** `@spec-link [[atd_stats]]`

## EXPECTATION
Output must be valid JSON and accurately reflect the counts and coverage based on the provided source and docs directories.
