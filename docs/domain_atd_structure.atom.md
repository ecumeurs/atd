---
id: domain_atd_structure
human_name: "ATD File Structure"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, structure, format, yaml]
parents:
  - [[domain_atd_philosophy]]
dependents:
  - [[domain_atd_type_architectural]]
  - [[domain_atd_type_interface]]
  - [[domain_atd_type_logic]]
  - [[domain_atd_type_ops_req]]
  - [[domain_atd_usage_protocol]]
  - [[rule_atd_naming_convention]]
  - [[rule_webui_health_classification]]
layer: BUSINESS
---

# ATD File Structure

## INTENT
To define the mandatory fields and sections of an Atomic Traceable Document (ATD) file, ensuring consistency for automated parsing and human readability.

## THE RULE / LOGIC
Every ATD must be a Markdown file with a strict YAML frontmatter and specific Markdown headers.

### 1. YAML Frontmatter
- `id`: Unique slug-style identifier (e.g., `my_feature_rule`).
- `human_name`: Readable title.
- `type`: Category (e.g., `MECHANIC`, `API`, `DOMAIN`, `RULE`).
- `domain`: Optional grouping for large multi-domain projects (e.g., `core`, `extension`, `ui`).
- `layer`: Documentation hierarchy: `CUSTOMER`, `ARCHITECTURE`, or `IMPLEMENTATION`. Named `layer` (not `domain`) to avoid collision with the `DOMAIN` type.
- `version`: Document version (e.g., `1.0`).
- `status`: `DRAFT`, `REVIEW`, or `STABLE`.
- `priority`: Numeric importance level, integer 1 (low) to 5 (highest).
- `tags`: List of relevant keywords.
- `parents`: List of parent atom IDs in `[[id]]` format.
- `dependents`: List of dependent atom IDs in `[[id]]` format.

### 2. Mandatory Sections (Markdown H2)
- `## INTENT`: A single-sentence "Why?".
- `## THE RULE / LOGIC`: The core technical or functional specification.
- `## TECHNICAL INTERFACE`: Linking information (API endpoints, `@spec-link`, tests).
- `## EXPECTATION`: Traceable criteria for verification.

### 3. Granularity & Bloat Control
ATD enforces the "Minimum Atomic Scale" to prevent overly broad atoms:
- **The "One Rule" Rule:** If a section contains more than one state-changing rule, it must be split.
- **Intent Clarity:** If an intent requires "and" or "also", the granularity is too low.
- **Bloat Factor:** Configured per type in `.atd` config (1.0 = strictest, 0.0 = no limit). High factors (≥0.7) enforce single-rule atoms. Low factors (≤0.3) allow broader narrative.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[domain_atd_structure]]`
- **Parser Logic:** `atd`'s atom parser and `Go` YAML unmarshalers.

## EXPECTATION
- Files must pass `atd audit` without formatting errors.
- YAML header must be valid and contain all mandatory fields.
