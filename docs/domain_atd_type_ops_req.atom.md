---
id: domain_atd_type_ops_req
human_name: "Operations and Requirement Atoms"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, types, operations, requirements]
parents:
  - [[domain_atd_structure]]
dependents: []
layer: BUSINESS
---

# Operations and Requirement Atoms

## INTENT
To define the usage and specificities of DATA, USAGE, BUILD, REQUIREMENT, and SPECIFICATION atoms.

## THE RULE / LOGIC
These types handle the ecosystem and external constraints:

- **`DATA`**: Static configuration, database schemas, or initialization data.
- **`USAGE`**: Tutorials, examples, and human-facing "How-to-use" guides.
- **`BUILD`**: CI/CD pipelines, environment setup, and deployment logic.
- **`REQUIREMENT`**: External or high-level constraints (e.g., "The system must support Linux"). Often less granular than `RULE`.
- **`SPECIFICATION`**: Detailed technical specs that might group multiple rules into a cohesive document.

### Specificity
- **Linking**: `DATA` links go to config files or DB migrations. `BUILD` links go to CI/CD yaml files or scripts. `USAGE` links often appear in READMEs.
- **Verification**: `REQUIREMENT` and `SPECIFICATION` are often verified via manual audits or top-level integration tests.

## TECHNICAL INTERFACE
- **Type Tag:** `type: DATA | USAGE | BUILD | REQUIREMENT | SPECIFICATION`.

## EXPECTATION
- `DATA` atoms match configuration files.
- `BUILD` atoms reflect the actual CI/CD workflow.
- `USAGE` atoms are clear and helpful for humans.
