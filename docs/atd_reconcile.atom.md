---
id: atd_reconcile
human_name: "ATD Reconcile"
type: MECHANIC
version: 1.0
status: DRAFT
priority: 5
tags: [atd, cli, reconcile, diff]
parents:
  - [[atd_cli]]
dependents: []
layer: IMPLEMENTATION
---

# ATD Reconcile

## INTENT
To compute a semantic diff between new inbound documentation and the existing ATD atom store, identifying which atoms need creation, update, or parent linking.

## THE RULE / LOGIC
Reads the new inbound edits file and the existing atom store, constructs a reconcile prompt asking the LLM (task `reconcile`) to identify relationships between new content and existing atoms. Returns JSON proposals: `[{proposed_id, relationship (CREATE|UPDATE|LINK), change_context}]`.

## TECHNICAL INTERFACE (The Bridge)
- **Command:** `atd reconcile -new <inbound_edits> -store <existing_atoms>`
- **LLM Task:** `reconcile`
- **Code Tag:** `@spec-link [[atd_reconcile]]`
