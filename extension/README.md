# ATD Linker for VS Code

This extension provides seamless navigation for Atomic Traceable Documentation (ATD) within VS Code.

## Features

- **Link Resolution**: Automatically transforms `[[atomic_id]]` patterns into clickable links (Ctrl+Click).
- **Go to Definition**: Right-click or use `F12` on an `[[atomic_id]]` to jump directly to the target `.atom.md` file.
- **Hover Information**: Hover over `@spec-link [[ID]]` in source code to see the atom's intent and live health metrics.
- **Code Lenses**: View implementation and test coverage percentages directly at the top of `.atom.md` files.
- **ATD Graph Explorer**: A dedicated sidebar tree view showing the parents and dependents of the currently open atom.
- **Full System Graph**: Visualize the entire project architecture with an interactive graph (Command: `ATD: Show Full System Graph`).
- **Config Awareness**: Reads `.atd` config to resolve the correct `docs_path`.

## Installation

Since this is a development extension, you can install it manually by creating a symlink in your extensions directory:

```bash
mkdir -p ~/.antigravity/extensions/
ln -s "$PWD/extension" "$HOME/.antigravity/extensions/local-dev.atd-linker"
```

## Devcontainer Integration

To use this extension within a Devcontainer, add the following to your `.devcontainer/devcontainer.json`:

```json
{
    "extensions": [
        "local-dev.atd-linker"
    ],
    "mounts": [
        "source=${localEnv:HOME}/.antigravity/extensions,target=/home/vscode/.vscode-server/extensions,type=bind,consistency=cached"
    ]
}
```

> [!NOTE]
> The exact target path might vary depending on your devcontainer user and whether you are using VS Code Desktop or Web.

## Documentation

This extension is documented using ATD itself. See `docs/specification_vscode_atd_linker.atom.md` for more details.
