---
id: atd_type_interface
human_name: "Interface Atoms: API, UI"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, types, api, ui]
parents:
  - [[atd_structure]]
dependents: []
layer: CUSTOMER
---

# Interface Atoms: API, UI

## INTENT
To define the usage and specificities of interface-level atoms.

## THE RULE / LOGIC
These types define the boundary of the system:

- **`API`**: External or internal technical contracts (e.g., HTTP endpoints, RPC signatures). It defines the payload structure, headers, and expected responses.
- **`UI`**: Visual requirements, user interaction flows, and layout constraints (e.g., "Login button must be centered", "Hovering shows a tooltip").

### Specificity
- **Linking**: `@spec-link` for an `API` is placed above the route handler. `UI` links go near the component definition or in UI-specific rule files.
- **Contracts**: `API` atoms should include sample payloads in the `THE RULE / LOGIC` or `TECHNICAL INTERFACE` sections.

## TECHNICAL INTERFACE
- **Type Tag:** `type: API | UI`.
- **Code Tag Placement:** Route handlers, component declarations.

## EXPECTATION
- `API` atoms must match the actual request/response schema.
- `UI` atoms should be verified through visual auditing or automated UI tests.
