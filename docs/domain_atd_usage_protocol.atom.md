---
id: domain_atd_usage_protocol
human_name: "ATD Usage Protocol"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, process, lifecycle, workflow]
parents:
  - [[domain_atd_philosophy]]
  - [[domain_atd_structure]]
dependents:
  - [[usage_atd_use_case_auditing]]
  - [[usage_atd_use_case_code_sync]]
  - [[usage_atd_use_case_cold_start]]
  - [[usage_atd_use_case_impact_analysis]]
layer: BUSINESS
---

# ATD Usage Protocol

## INTENT
To define the standard workflow for creating, linking, and verifying ATD atoms during the software development lifecycle.

## THE RULE / LOGIC
The ATD lifecycle operates in two distinct contexts:

### Cold Start (Bootstrapping)
Used once, at project inception or when onboarding ATD to a legacy codebase:
1. `roadmap` → Identify high-density files to prioritize.
2. `index` → Build the vector DB for search and matching.
3. Author atoms for the prioritized source files in `docs/`.
4. `weave` → Establish bidirectional parent/dependent links.
5. `discover` → Recommend `@spec-link` placements in code.
6. `recon` → Confirm that discovered links are correct.
7. `audit` → Check for bloat, collisions, and orphans.

### Day-to-Day Development Lifecycle
Once the ATD base is established, atoms and code co-evolve through five recurring stages:

1. **Plan**: Explore existing atoms, brainstorm new features. Create `DRAFT` atoms for ideas still being considered.
2. **Specify**: Create or update atoms (`REQUIREMENT`, `USECASE`, `SPECIFICATION`, `MODULE`, `SERVICE`). Link to parents. Run `weave`.
3. **Implement**: Write code, annotate with `@spec-link` and `@test-link`. Update IMPLEMENTATION-domain atoms as code solidifies. Promote to `REVIEW`.
4. **Verify**: Run `audit`, `verify`, `test-links`, `crawl --gaps` to validate alignment and coverage.
5. **Evolve**: Run `crawl` for impact analysis before modifying atoms. Apply surgical edits via `update`. Promote to `STABLE`. Loop back to Plan.

## TECHNICAL INTERFACE
- **Core Workflow Tool:** `atd` CLI.
- **MCP Integration:** IDE agents use `mcp_atd_*` tools to automate these steps.

## EXPECTATION
- All core logic has a corresponding `@spec-link`.
- No orphan atoms exist in the `STABLE` state.
- Documentation updates trigger a verification loop.
