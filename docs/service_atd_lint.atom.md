---
id: service_atd_lint
human_name: "ATD Structural Linter"
type: MECHANIC
version: 0.1.0
status: STABLE
priority: 3
tags: [atd, cli, lint, validation]
parents:
  - [[module_atd_cli]]
dependents:
  - [[api_atd_serve_lint]]
layer: IMPLEMENTATION
---

# ATD Structural Linter

## INTENT
Perform a fast, deterministic structural validation across all ATD atoms in a documentation directory. This ensures that all atoms adhere to the mandatory schema and reference integrity before they are used by other tools or agents.

## THE RULE / LOGIC
- Scans for all `*.atom.md` files in the specified directory (default: configured docs dir).
- **Mandatory Fields**: Verifies presence of `id`, `human_name`, `type`, `layer`, `version`, `status`, and `priority`.
- **Enum Validation**:
    - `layer` must be one of: `BUSINESS`, `ARCHITECTURE`, `IMPLEMENTATION`.
    - `priority` must be one of: `1`, `2`, `3`, `4`, `5`, `CORE`.
- **Mandatory Sections**: Verifies non-empty content for `## INTENT`, `## THE RULE / LOGIC`, `## TECHNICAL INTERFACE`, and `## EXPECTATION`.
- **Reference Integrity**: Verifies that all IDs listed in `parents` and `dependents` exist within the scanned set of atoms.
- **Governance Isolation**: `CONTRACT`/`VISION` atoms sit outside the ancestry graph. Flags a governance atom that declares a non-empty `parents` or `dependents`, and flags any atom that names a governance atom in either of its own link fields.
- **Project Governance**: Flags more than one `CONTRACT` or `VISION` atom, and — once the corpus holds at least one `BUSINESS`-layer atom — a missing one.
- **Self-Sufficiency**: An atom may link out only through its structural edges — `parents`, `dependents`, and `@spec-link`/`@test-link` tags. Flags, per atom and once per distinct reference:
    - any markdown link or image `[text](target)` and any reference-style link definition;
    - any URL (`http(s)://`, `ftp://`, `file://`) in prose — a URL inside a code span or fenced block is a literal value (an endpoint, a config value), not a link;
    - any `[[id]]` wiki-link in prose naming an atom other than the atom itself or one of its declared `parents`/`dependents` — `@spec-link`/`@test-link` tags and code-span syntax examples such as `` `[[id]]` `` are exempt;
    - any document citation, in prose or in a code span: a document-file path (`.md`, `.markdown`, `.mdx`, `.rst`, `.adoc`, `.pdf`, `.doc`, `.docx`, `.odt`) that carries a directory component, or a document file immediately followed by a `§` section marker. A bare filename naming an artifact the tool reads or writes, and placeholders or globs (`<file.md>`, `*.atom.md`), are not citations.
  Fenced code blocks are skipped. The frontmatter `description` value is scanned as prose; other frontmatter keys are structural and ignored.
- **Deterministic**: No LLM or network calls.

## TECHNICAL INTERFACE (The Bridge)
- **Binary:** `atd/cmd/atd/cmd/lint.go`
- **Self-sufficiency detector:** `atd/pkg/atom/selfsufficiency.go` — `FindOutsideReferences`
- **Usage:** `atd lint [dir]`
- **Code Tag:** `@spec-link [[service_atd_lint]]`

## EXPECTATION (For Testing)
- Running `atd lint` on a directory with a missing `id` field in an atom returns a non-zero exit code and a descriptive error message.
- Running `atd lint` on a directory with an unresolved parent link returns an error identifying the broken reference.
- Running `atd lint` on a directory where a `CONTRACT` or `VISION` atom declares any `parents` or `dependents` returns an error naming the offending references.
- Running `atd lint` on a directory where an atom's body contains a markdown link, a prose URL, a prose `[[id]]` naming an atom outside its `parents`/`dependents`, or a document citation (a document path with a directory component, or a document followed by a `§` marker) returns an error naming each reference.
- Running `atd lint` on an atom whose prose names only its own parents/dependents, or whose `[[id]]` and URLs appear only inside code spans or fenced blocks, reports no self-sufficiency error.
- Running `atd lint` on a fully compliant documentation directory returns "All atoms passed structural validation." and exit code 0.
