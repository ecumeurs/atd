---
id: atd_type_architectural
human_name: "Architectural Atoms: MODULE, SERVICE, ENTITY"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, types, architecture]
parents:
  - [[atd_structure]]
dependents: []
layer: CUSTOMER
---

# Architectural Atoms: MODULE, SERVICE, ENTITY

## INTENT
To define the usage and specificities of architectural-level atoms.

## THE RULE / LOGIC
These types represent the structural backbone of the system:

- **`MODULE`**: High-level grouping of related components (e.g., "The Auth System", "Database Layer"). It often serves as a parent to multiple services or rules. Typically `layer: ARCHITECTURE`.
- **`SERVICE`**: A logical orchestrator or manager that performs specific operations (e.g., `DocumentManager`, `AuthValidator`). Typically `layer: ARCHITECTURE` or `IMPLEMENTATION`.
- **`ENTITY`**: Definitions of data structures, state models, or domain objects (e.g., `UserRecord`, `SpecFile`). Entities belong to the `ARCHITECTURE` layer because they define the system's data contracts and are sensitive to change — modifying an entity can have cascading impact across services.

### Specificity
- **Granularity**: Modules can be broader, but Services and Entities should focus on a single responsibility.
- **Linking**: `@spec-link` for a `MODULE` is usually placed at the package/directory level or in a main entry point. `SERVICE` links go above class/struct definitions. `ENTITY` links go above type/struct definitions.

## TECHNICAL INTERFACE
- **Type Tag:** `type: MODULE | SERVICE | ENTITY`.
- **Code Tag Placement:** Structural headers or class definitions.

## EXPECTATION
- `MODULE` atoms should have multiple `SERVICE` or `RULE` dependents.
- `ENTITY` atoms must match the actual data structures in code.
