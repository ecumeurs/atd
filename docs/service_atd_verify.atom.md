---
id: service_atd_verify
human_name: "ATD Verify"
type: SERVICE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, verify, git, compliance]
parents:
  - [[module_atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Verify

## INTENT
To audit modified files against their linked ATD specifications by combining git diff analysis, native test execution, and LLM-powered compliance verification.

## THE RULE / LOGIC
Iterates through every @spec-link instance in modified files (or full project via --full). For each instance, it assembles a surgical context containing the atom logic, its full parent ancestry, a code snippet around the tag, and any associated @test-link verification proof. It executes native tests for stability and generates a bundled JSON-structured audit prompt for individual compliance assessment.

Flags:
- --full: Audit the entire project.
- --file/--line: Target specific tags.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd verify`
- **LLM Task:** None (stdout passthrough — IDE Agent processes the prompt)
- **Code Tag:** `@spec-link [[service_atd_verify]]`
