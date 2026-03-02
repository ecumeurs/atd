# ATD: Atomic Traceable Documentation Skillset

**Version:** 1.0  
**Status:** Implementation Ready

## Core Philosophy
Documentation is code. Logic must be traceable, secable, and verifiable without constant LLM overhead. By separating documentation into highly granular "Atoms", updating documentation logic becomes a rigid contract constraint that ripples cleanly across the codebase.

---

## The System Components
This skillset transforms standard markdown rendering into a **Development Governance System**. It defines explicit parameters that bridge Architect-level intention with Developer-level code generation.

### 1. The Atom Template
Every functional document fragment is treated as an "Atom". It consists of rigid YAML frontmatter targeting programmatic deterministic analysis and human-readable bodies.
*See section 1 of ATD.md or refer to `template.atom.md`.*

### 2. The Personas (`role.md`, `draft_agent.md`)
The system mandates distinct operational modes for the AI assisting on this repository.
* **Architect Mode:** Focuses on the "Contract." Merges, dissects, and outlines "Universe Rules" while guarding against logic drifts or redundant mechanics.
* **Developer Mode:** Focuses on "Implementation." Never changes a rule without explicit Architect approval. Refuses to implement contradictory code and runs the Archaeology engine to adopt legacy implementations.
* **Analyst Mode:** Read-only mode used for data generation, presentation synthesis, and "Ripple Reporting" to human stakeholders.

### 3. The Toolset (`tool.md`, `draft_skill.md`)
ATD avoids brute-force LLM inference, employing parsing heuristics explicitly to keep API tokens down to near zero. 
* **Deterministic Tools (0% LLM):** `atd-crawl`, `atd_query`, `atd_report_gaps`
* **Surgical LLM (High Value):** `atd-audit`, `atd_dissect`, `atd_reconcile`
* **Flow LLM (Rich Token):** `atd_generate_snapshot`

---

## Getting Started

To install this framework onto a project, map this folder inside your AI Assistant's `.agent/skills/` directory.

### Initiating on a New or Legacy Project

When you apply this skill to an existing repository, you must transition the project into the ATD framework. The LLM Agent should follow these steps:

1. **The Automated Pipeline (Cold Start):** The agent must execute `scripts/atd-cold-start.sh <project_dir>`. This orchestrates the generation of a structural roadmap, semantic index databases, and isolates high-density code targets without spending LLM tokens.
2. **The Deconstruction Phase (Architect Mode):** If you possess existing monolithic documentation (GDDs, wikis, massive readmes), the Agent reads the extracted text (found in `pipeline_output/domain_*.md.txt`) and generates the isolated `.atom.md` files in the `docs/` folder.
3. **The Mechanical Phase (Developer Mode):** The Agent reads the generated `pipeline_output/dissect_*.json` files to generate Mechanic and API atoms matching the structural reality of the code, saving them to the `docs/` folder.
4. **The Weaving & Recon Phase:** The pipeline will automatically connect Atom parents/dependents in the background, and then deploy the local *Search-Then-Recon* protocol to hunt down matching code files and apply `@spec-link` tags.

By strictly isolating these modes, your agent will securely transition standard projects into governed systems without blindly overwriting your code or hallucinating mechanics!
