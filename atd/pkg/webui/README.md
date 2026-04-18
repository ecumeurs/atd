# ATD WebUI

The WebUI is an advanced, specialized graphical interface for exploring, managing, and reasoning about Atomic Traceable Documents (ATDs) and their links to source code. Its primary goal is to surface system health and architectural alignment instantly, translating flat markdown files into a navigable "Waterfall of Intent".

## Architecture & Technologies

The WebUI operates as a lightweight, modular system integrated directly into the ATD CLI:

*   **Backend (Go + Gin):**
    *   `server.go`: Initializes the Gin engine, sets up routes, and handles server lifecycle.
    *   `handlers.go`: Consolidated API layer for ATD operations (tree, details, search, updates), document generation, and Gemini integration.
    *   `embed.go`: Manages production asset delivery via Go's `embed` filesystem.
*   **Frontend (Vanilla JS + ES Modules):** 
    *   `static/index.html`: The main single-page application entry point.
    *   `static/js/app.js`: Orchestrates the modular frontend, managing application state and tab navigation.
    *   **Modular Components (`static/js/`):** Separated concerns for API (`api.js`), State (`state.js`), Waterfall Explorer (`explorer.js`), Details Panel (`details.js`), Search Overlay (`search.js`), and Document Management (`documents.js`).
    *   `static/spec-builder.js`: Dedicated logic for the interactive Gemini Spec Builder.
*   **Rendering:**
    *   Markdown is rendered on the client via `marked.js`.
    *   Traceability highlights and dependency pathing are managed dynamically through the DOM and CSS.

## Key Features

- **Waterfall of Intent:** A powerful three-column visualization sorting ATDs by their layer (`CUSTOMER`, `ARCHITECTURE`, `IMPLEMENTATION`). Provides an immediate, readable birds-eye view of how business requirements translate to implementation logic.
- **Health-based Categorization:** Visual signaling of atom health (Done, Almost Done, WIP, Draft) based on implementation links and test coverage.
- **Traceability Explorer:** Interactive Detail Panel mapping an ATD's complete ancestry (Parents upstream, Dependents downstream), implementation links, and real-time coverage.
- **Gemini Spec Builder:** An interactive chat interface embedded within the WebUI, specifically tuned to act as an architectural discussion partner for scoping and creating compliant ATDs before code is written.
- **AI Context Summarization:** Built-in integration with LLMs to generate holistic summaries of an ATD by automatically extracting and synthesizing the contexts of its entire dependency chain.
- **Global Search:** Fast, server-side search overlay (`Ctrl+K` command palette) covering ATD IDs, names, types, tags, and content.
- **Interactive Multi-select:** Perform bulk status updates and select multiple atoms for synthetic document generation.

## Usage

The WebUI is part of the ATD toolkit. To launch it, run the following command from your project root:

```bash
atd webui
```

For local development of the WebUI itself, use the `--dev` flag to serve static files directly from the filesystem:

```bash
atd webui --dev
```

The server will bind to the host and port specified in your `.atd` configuration (defaulting to [http://localhost:8080](http://localhost:8080)).

## Configuration

The WebUI's behavior is governed by the `.atd` configuration file located in the root of your project:

```json
{
  "webui": {
    "host": "localhost",
    "port": 8080,
    "toolkit_path": "scripts/pkg/webui"
  }
}
```

- **`host`**: The address the server binds to.
- **`port`**: The port selection for the local server.
- **`toolkit_path`**: (Dev Mode) The relative path to the directory containing `static/` files when running with `--dev`.
