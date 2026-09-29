---
id: rule_atd_atom_self_sufficiency
human_name: "Atom Self-Sufficiency"
description: "An atom links out only through its structural edges (parents:, dependents:, @spec-link/@test-link). Any other pointer to a document outside the atom is forbidden. Reported by `atd lint`."
type: RULE
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 5
tags: [atd, authoring, self-sufficiency, lint]
parents:
  - [[domain_atd_structure]]
dependents: []
---

# Atom Self-Sufficiency

## INTENT
To make every atom fully understandable from its own content, so that no atom depends on a document outside it that can move, be deleted, or never be read.

## THE RULE / LOGIC
A reader who has never opened any other file must be able to understand an atom completely from its own frontmatter and sections. When explaining something seems to require pointing elsewhere, the explanation is written into the atom instead.

The only links an atom may carry are its structural graph edges:

- `parents:` and `dependents:` in the frontmatter;
- `@spec-link [[id]]` / `@test-link [[id]]` tags;
- a prose `[[id]]` naming the atom itself or one of its own declared `parents:`/`dependents:` — it restates an existing edge rather than adding a new dependency.

Everything else that sends the reader to another document is forbidden:

1. A markdown link or image `[text](target)`, or a reference-style link definition, whatever its target.
2. A URL (`http(s)://`, `ftp://`, `file://`) in prose.
3. A prose `[[id]]` naming an atom that is neither the atom itself nor one of its declared `parents:`/`dependents:`.
4. A citation of a document: a document-file path with a directory component (such as `<dir>/<report>.md`), or a document cited by section (such as `<GUIDE>.md §<n>`). This applies inside code spans too, because wrapping a citation in backticks does not make it self-contained.

What is not a reference to an outside document:

- A URL or `[[id]]` inside a code span or fenced block is a literal value — an endpoint, a config entry, a syntax example — and is allowed.
- Fenced code blocks (schemas, commands, examples) are not scanned.
- A bare filename naming an artifact the tool reads or writes (`task_list.md`) and a placeholder or glob (`<atom.md>`, `*.atom.md`) describe a kind of file, not a document to go and read.
- Source-code paths in TECHNICAL INTERFACE (`atd/cmd/atd/cmd/lint.go`) are the code bridge, not documents.

The frontmatter `description:` value is prose and is held to the same rule; every other frontmatter key is structural.

### Enforcement
- **`atd lint` reports.** Each distinct outside reference in an atom is one lint error naming the reference and its shape. Lint never edits; the author inlines the missing reasoning.
- **Deterministic.** Detection is purely lexical — no LLM, no network, no filesystem lookup of the referenced target — so a dangling reference is flagged exactly like a live one.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_atd_atom_self_sufficiency]]`
- **Detector:** `atd/pkg/atom/selfsufficiency.go` — `FindOutsideReferences`, `SplitFrontmatter`
- **Reported by:** `atd/cmd/atd/cmd/lint.go` — `runLint`, `selfSufficiencyAllowed`

## EXPECTATION (For Testing)
- `atd lint` on an atom whose body contains `[text](https://example.com)` exits non-zero and names the link target.
- `atd lint` on an atom whose prose contains a bare `https://` URL exits non-zero; the same URL inside a code span or fenced block passes.
- `atd lint` on an atom whose prose names `[[other_atom]]`, where `other_atom` is not in its `parents:`/`dependents:`, exits non-zero; naming one of its own parents passes.
- `atd lint` on an atom whose prose contains `@spec-link [[other_atom]]` or the code span `` `[[id]]` `` passes.
- `atd lint` on an atom citing a document path with a directory component (in prose or in backticks) or a document followed by a `§` section marker exits non-zero; mentioning `task_list.md` or `<atom.md>` passes.
- `atd lint` on an atom whose frontmatter `description:` cites a document path with a directory component exits non-zero.
