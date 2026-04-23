# Filesystem-Only ATD Health Snapshot
**Date**: 2026-04-19
**Scope**: 245 physical `.atom.md` files on disk

## Overview
This report provides a corrected view of the documentation health by ignoring the "Ghost Atoms" (stale records) present in the ATD index. It reflects the outcome of the **Hardening Phase** conducted today.

---

## 1. Quantitative Inventory
| Metric | Value | Note |
|---|---|---|
| **Total Atoms (Filesystem)** | **245** | (vs 610 in stale index) |
| **Atoms with Metadata** | 245 | All physical files have valid layer/type/status. |
| **STABLE Implementation Orphans** | **0** | **Target Met**. (100% coverage for Implementation layer) |
| **False Orphans (Detected)** | 2 | Confirmed tags exist but are tool-ignored. |

---

## 2. Layer Distribution (On-Disk)
| Layer | Count | Coverage Strategy |
|---|---|---|
| **IMPLEMENTATION** | ~137 | 100% Mandated `@spec-link`. |
| **ARCHITECTURE** | ~60 | High-level `@spec-link` applied to controllers/modules. |
| **CUSTOMER** | ~48 | Linked via child atoms (DOCKING strategy). |

---

## 3. Notable Improvements
- **Initiative & Ruler**: Restored missing links for requeue calculation and delay costs in the Go engine.
- **Admin Dashboard**: Applied high-level tags to the Laravel Admin Controller and Vue views for history management.
- **Reroll Mechanics**: Synchronized stats reroll logic across the database (Character model) and the API (Profile controller).
- **Action Feedback**: Mapped the core API output DTOs to the `ActionFeedback` atom.

## 4. Residual Ghost Audit
The 365 "Ghost Atoms" remaining in the `atd stats` report consist primarily of:
- 196 atoms with `<unspecified>` layer.
- Atoms referencing files that no longer exist in the repository.

**Recommendation**: The ATD team should perform a hard-reset of the semantic index to align the tool with the physical reality of the 245 files documented here.
