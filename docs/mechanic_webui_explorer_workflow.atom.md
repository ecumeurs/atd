---
id: mechanic_webui_explorer_workflow
human_name: "Explorer Interaction Workflow"
type: MECHANIC
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [logic, explorer, interaction]
parents:
  - [[ui_webui_traceability_explorer]]
dependents: []
---

# Explorer Interaction Workflow

## INTENT
Define the step-by-step logic and state changes when users interact with the Health-Based Explorer view.

## THE RULE / LOGIC
1. **Overview:** All atoms are bucketed into four health blobs (Done, Almost Done, WIP, Doc Jungle).
2. **Blob Selection:** Upon clicking a blob, the container transitions to a 'zoomed' state.
3. **Zoom Hierarchy:** The selected blob expands to fill the view, revealing the architectural hierarchy (nested treemap) of atoms within that category.
4. **Atom Selection:** Clicking an atom within a zoomed hierarchy selects it and displays its metadata in the details panel.
5. **Deselection:** Clicking the breadcrumb back to 'All' reverts the view to the initial four blobs.

## TECHNICAL INTERFACE (The Bridge)
- **JS State:** `currentViewScope` (all, done, wip, etc.)
- **D3 Transition:** `.transition().duration(500)`
- **Code Tag:** `@spec-link [[mechanic_webui_explorer_workflow]]`

## EXPECTATION (For Testing)
When a user clicks on the 'WIP' blob, the `currentViewScope` must change to 'wip' and the treemap must re-render to display only atoms within the WIP category.
