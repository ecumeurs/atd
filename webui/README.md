# ATD WebUI

The WebUI is an advanced, specialized graphical interface for exploring, managing, and reasoning about Atomic Traceable Documents (ATDs) and their links to source code. Its primary goal is to surface system health and architectural alignment instantly, translating flat markdown files into a navigable "Waterfall of Intent".

## Architecture & Technologies

The WebUI operates as a lightweight, decomposed monolith without any heavy build systems (zero-build architecture):

*   **Backend (Go + Gin):** 
    *   Decomposed routing layer handling static file serving and dedicated API scopes.
    *   `atoms.go` and `handlers_atd.go`: ATD state management, graph traversal, and native API exposure (tree, detail, update, search, local LLM summarization).
    *   `handlers_gemini.go`: Dedicated routes for interfacing with cloud-based Gemini features (spec building, proposals).
*   **Frontend (Vanilla JS + ES Modules):** 
    *   `app.js` runs as the module entry point, orchestrating state management and event-driven view rendering.
    *   Separated modules for API (`api.js`), State (`state.js`), Explorer visualization (`explorer.js`), Details (`details.js`), and Search (`search.js`).
*   **Rendering & Parsing:**
    *   Markdown is rendered securely on the client via `marked.js`.
    *   Complex path rendering (SVG) handles traceability lines dynamically.

## Key Features

- **Waterfall of Intent:** A powerful three-column visualization sorting ATDs by their layer (`CUSTOMER`, `ARCHITECTURE`, `IMPLEMENTATION`). Provides an immediate, readable birds-eye view of how business requirements translate to implementation logic.
- **Traceability Explorer:** Features an interactive Detail Panel that maps an ATD's complete ancestry (Parents upstream, Dependents downstream), implementation links, and real-time coverage signaling. 
- **AI Context Summarization:** Built-in integration with local LLMs (via Ollama) to generate holistic summaries of an ATD by automatically extracting and synthesizing the contexts of its entire dependency chain.
- **Global Search:** Fast, server-side search overlay (`Ctrl+K` command palette) covering ATD IDs, names, types, tags, and content.
- **Ctrl+K Document Generation:** Dynamically assemble and generate comprehensive markdown documentation by semantically selecting ATD nodes and providing an intent narrative, powered by the ATD assembly pipeline. Generates fully synthetic docs directly via an interactive overlay.
- **Gemini Spec Builder:** An interactive chat interface embedded within the WebUI, specifically tuned to act as an architectural discussion partner for scoping and creating compliant ATDs before code is written.

## Quick Start

To run the WebUI Analyzer, start the Go application from the `webui` directory:

```bash
cd webui
go run .
```

The server will bind to port 8081 by default. Access the UI via [http://localhost:8081](http://localhost:8081).

## Configuration (`config.json` / `.env`)

The WebUI's behavior is dictated by `config.json` (and `GEMINI_API_KEY` environmental variable for cloud AI tools). Make sure the relative paths correctly map to your workspace setup.

```json
{
    "host": "localhost",
    "port": 8081,
    "project_path": "../",
    "atd_path": "docs",
    "toolkit_path": "../scripts/cmd/atd"
}
```

- **`project_path`**: The root of the audited project.
- **`atd_path`**: Directory containing `.atom.md` specs.
- **`toolkit_path`**: Location of the compiled ATD binary (to execute updates and heavy ATD operations).
