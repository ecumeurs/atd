# Lot Summary: Atom Lifecycle & Governance

**Lot Objective:** Establish formal lifecycle controls, versioning history, and cross-project documentation standards to ensure ATDs remain stable and traceable over time.

---

## Included Issues

| Ref | Title | Severity | Status | Component |
|---|---|---|---|---|
| [ISS-010](file:///home/bastien/work/skill/issues/ISS-010_20260304_atd_status_management.md) | ATD Status Management and Workflow | Medium | Open | `atd_management_skill` |
| [ISS-009](file:///home/bastien/work/skill/issues/ISS-009_20260304_atd_version_management.md) | ATD Version Management Implementation | Medium | Open | `atd_management_skill` |
| [ISS-033](file:///home/bastien/work/skill/issues/ISS-033_20260324_atom_deprecation_archival.md) | Atom Deprecation and Archival Statuses | Medium | Open | `scripts/cmd/atd` |
| [ISS-034](file:///home/bastien/work/skill/issues/ISS-034_20260324_changelog_sidecar.md) | Change History Sidecar per Atom | Medium | Open | `scripts/cmd/atd/cmd/update.go` |
| [ISS-038](file:///home/bastien/work/skill/issues/ISS-038_20260324_type_specific_templates.md) | Type-Specific Atom Templates | Medium | Open | `scripts/cmd/atd/cmd/update.go` |
| [ISS-035](file:///home/bastien/work/skill/issues/ISS-035_20260324_cross_project_sharing.md) | Cross-Project Atom Sharing | Low | Open | `.atd` config |

---

## Thematic Analysis

### Core Themes
1. **Dynamic Lifecycle**: Transitioning away from a simple "Open/Closed" mentality to a rich status workflow (DRAFT -> REVIEW -> STABLE -> DEPRECATED -> ARCHIVED) that accurately reflects a feature's history.
2. **Auditability**: Implementing sidecar changelogs to record *why* and *how* an atom changed, providing a governance trail that typical Git history lacks at the atomic level.
3. **Template Standardization**: Moving from generic templates to type-aware blueprints (API, USECASE, USER_STORY) to ensure technical precision (e.g., payload schemas in API atoms).
4. **Federated Documentation**: Preparing the system to handle multi-repository projects where atoms in one codebase reference or import requirements from another.

### System Risks
> [!CAUTION]
> **Stale References**: Without a `DEPRECATED` or `ARCHIVED` status, orphaned atoms often linger in the graph as "STABLE" items with no code links, creating confusion for both AI agents and human architects about the system's actual current state.
>
> **Missing Change Context**: The current `version` field in YAML is essentially static. Without a formal update/increment logic and history tracking, the "why" of architecture evolution is lost.

---

## Strategic Roadmap

- **Short Term**: Update the status enum to include `DEPRECATED`; implement basic version incrementing in `atd update`.
- **Medium Term**: Launch the sidecar `.changelog.json` system; integrate type-specific templates into the creation flow.
- **Long Term**: Implement cross-repo atom resolution and organization-wide federated governance.
