# ATD System Architect & Tooling

## Intent of the Project
The ATD (Atomic Traceable Document) project provides a framework and an associated AI agent skill to strictly manage software architecture, rules, and requirements. It ensures that code acts as a perfect reflection of atomic specifications. By structuring documentation into highly focused, single-responsibility files (Atoms), it bridges the gap between high-level game design/architecture and low-level code implementation. The system is designed to provide AI agents with a deterministic, searchable source of truth that enforces rules before code is ever written or modified.

## What are ATDs?
ATDs are **Atomic Traceable Documents**. Each ATD is a single Markdown file representing a unique, isolated piece of logic, domain knowledge, or specification. To minimize context windows and allow for automated parsing, ATDs employ a strict YAML header that defines metadata such as ID, type (e.g., `DOMAIN`, `MECHANIC`, `API`, `RULE`), status, and dependency relationships.

Because an ATD is atomic, it describes only one primary rule or concept. This strict granularity prevents ambiguous requirements and ensures that every rule can be individually tested, verified, and linked directly to the application's source code.

## How it Works
1. **The ATD Framework**: Developers and architects write ATDs referencing systems, mechanics, and entities.
2. **The Tools**: A suite of Go-based CLI tools (e.g., `atd-audit`, `atd-crawl`, `atd-dissect`, `atd-link-weaver`) processes these ATDs. The tools can validate formatting, extract legacy logic into new Atoms, weave dependency graphs, and verify congruence between the documentation and the codebase.
3. **The Link (@spec-link)**: Code objects (functions, classes) are annotated with `@spec-link [[ATOM_ID]]`. The ATD agents and tools trace these links to verify that code implementations align with current architecture definitions. If a developer or an AI agent attempts to violate an ATD rule, the discrepancy is flagged.
4. **Agent Integration**: The AI assistant (having the ATD skill) operates under specific modes (Architect, Developer, Analyst) to either create/manage Atoms, write compliant code, or audit the system respectively.

## Reference Project
**`upsilonbattle`** serves as the primary reference project used to test and validate this skill. It demonstrates how ATD mechanics, API routes, and domain elements interact in a real-world scenario, acting as the testbed for the ATD toolchain's extraction, auditing, and generation capabilities.
