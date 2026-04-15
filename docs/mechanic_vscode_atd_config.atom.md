---
id: mechanic_vscode_atd_config
status: DRAFT
type: MECHANIC
version: 1.0.0
priority: 3
parents:
  - [[service_vscode_linker_features]]
dependents: []
human_name: "ATD Config Loader"
layer: IMPLEMENTATION
---

# New Atom

## INTENT
Manage loading and filesystem watching of the .atd configuration file.

## THE RULE / LOGIC
1. Read `.atd` file from workspace root on activation.
2. Parse JSON and extract `docs_path`.
3. Create a FileSystemWatcher on `.atd` and reload config on change.

## TECHNICAL INTERFACE

## EXPECTATION
Extension correctly identifies docs_path and reloads when .atd changes.
