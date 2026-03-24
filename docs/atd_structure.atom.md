---
id: atd_structure
human_name: "ATD File Structure"
type: DOMAIN
version: 1.0
status: STABLE
priority: 5
tags: [atd, structure, format, yaml]
parents:
  - [[atd_philosophy]]
dependents: [[[atd_type_architectural]], [[atd_type_interface]], [[atd_type_logic]], [[atd_type_ops_req]], [[atd_usage_protocol]]]], [[atd_type_interface]], [[atd_type_logic]], [[atd_type_ops_req]], [[atd_usage_protocol]]]], [[atd_type_interface]], [[atd_type_logic]], [[atd_type_ops_req]], [[atd_usage_protocol]]]
layer: CUSTOMER
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
- **Code Tag:** `@spec-link [[atd_structure]]`
- **Parser Logic:** `atd-dissect` and `Go` YAML unmarshalers.

## EXPECTATION
- Files must pass `atd audit` without formatting errors.
- YAML header must be valid and contain all mandatory fields.
