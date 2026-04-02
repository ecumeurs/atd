---
id: mechanic_webui_search_handler
status: STABLE
version: 1.0
dependents: []
type: MECHANIC
parents: [[api_webui_search]]
tags: webui,mechanic,search
human_name: WebUI Search Handler Mechanic
layer: IMPLEMENTATION
priority: 3
---

# New Atom

## INTENT
Implement search by simple iterative substring checking.

## THE RULE / LOGIC
Iterate memory `Atoms`, string match `id`, `human_name`, `content`, `tags`. Return subsets.

## TECHNICAL INTERFACE

## EXPECTATION
Safely skip if query is empty.
