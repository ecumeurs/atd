# ATD WebUI

The WebUI provides a graphical interface for exploring and managing Atomic Traceable Documents (ATDs) and their links to source code.

## Quick Start

To start the WebUI, navigate to the `webui` directory and run the Go application:

```bash
cd webui
go run main.go
```

The server will start (by default on [http://localhost:8081](http://localhost:8081)).

## Configuration (`config.json`)

The WebUI behavior is controlled by `config.json`. Ensure the following paths are correctly set relative to the `webui` directory:

```json
{
    "host": "localhost",
    "port": 8081,
    "project_path": "../upsilon",
    "atd_path": "docs",
    "toolkit_path": "../atd_management_skill/.agent/skills/atd/tool"
}
```

- **`project_path`**: Path to the root of the project you want to audit (e.g., `../upsilon` or `../upsilonbattle`).
- **`atd_path`**: The sub-directory within `project_path` that contains the `.atom.md` files.
- **`toolkit_path`**: Path to the ATD binary toolchain (used for enriched analysis).

## Features

- **ATD Tree**: Browse the hierarchy of your documentation.
- **Traceability**: See which lines of code are linked to specific Atoms via `@spec-link`.
- **Status Tracking**: Identify "Green" atoms that are fully implemented and tested.
