---
id: rule_webui_blob_sizing
human_name: "Blob Sizing Logic"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [logic, sizing, blob]
parents:
  - [[ui_webui_explorer_layout]]
dependents: []
---

# Blob Sizing Logic

## INTENT
Define the logic for calculating the relative size (area) of each health blob in the Explorer view.

## THE RULE / LOGIC
For the first level of categorizations:
- **Relative Size:** Each health blob's area in the treemap will be directly proportional to the **total count of ATD atoms** it contains.
- **Minimum Size:** To ensure readability, a minimum base size must be assigned even if a bucket is empty (displaying "0 items").
- **Future Changeability:** This sizing logic is designed for easy future swapping (e.g., weighting by complexity or priority).

## TECHNICAL INTERFACE (The Bridge)
- **JS Property:** `d3.hierarchy().sum(d => d.value)`
- **D3 Property:** `value` (assigned based on item count)
- **Code Tag:** `@spec-link [[rule_webui_blob_sizing]]`

## EXPECTATION (For Testing)
If the 'Done' bucket contains 10 atoms and 'WIP' contains 5, the area for 'Done' must be twice as large as 'WIP' in the initial treemap view.
