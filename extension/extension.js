// @ts-check
// @spec-link [[specification_vscode_atd_linker]]
/** @typedef {import('vscode')} vscode */
const vscode = require('vscode');
const fs = require('fs');
const path = require('path');

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {
    let docsPath = '';

    // Get the first open workspace folder
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) return;
    const workspaceRoot = workspaceFolder.uri.fsPath;

    // 1. Function to read the .atd configuration file
    const loadConfig = () => {
        const atdFile = path.join(workspaceRoot, '.atd');
        try {
            if (fs.existsSync(atdFile)) {
                const data = fs.readFileSync(atdFile, 'utf8');
                const config = JSON.parse(data);
                docsPath = config.docs_path || '';
                console.log(`[ATD Linker] Loaded docs_path: ${docsPath}`);
            }
        } catch (err) {
            console.error('[ATD Linker] Failed to parse .atd config:', err);
        }
    };

    // Initial load
    loadConfig();

    // 2. Watch the .atd file so we reload if the config changes
    const watcher = vscode.workspace.createFileSystemWatcher(
        new vscode.RelativePattern(workspaceRoot, '.atd')
    );
    watcher.onDidChange(loadConfig);
    watcher.onDidCreate(loadConfig);
    context.subscriptions.push(watcher);

    // Common helper to get the target file URI
    const getTargetUri = (ATDId) => {
        const targetPath = path.join(workspaceRoot, docsPath, `${ATDId}.atom.md`);
        return vscode.Uri.file(targetPath);
    };

    // 3. The Link Provider (Ctrl + Click)
    const linkProvider = vscode.languages.registerDocumentLinkProvider('*', {
        provideDocumentLinks(document) {
            const text = document.getText();
            const links = [];
            const regex = /\[\[([^\]]+)\]\]/g;
            let match;

            while ((match = regex.exec(text))) {
                const ATDId = match[1];

                const startPos = document.positionAt(match.index);
                const endPos = document.positionAt(match.index + match[0].length);
                const range = new vscode.Range(startPos, endPos);

                links.push(new vscode.DocumentLink(range, getTargetUri(ATDId)));
            }

            return links;
        }
    });

    // 4. The Definition Provider (Peek / Go to Definition)
    const definitionProvider = vscode.languages.registerDefinitionProvider('*', {
        provideDefinition(document, position) {
            // Find if the cursor is currently resting on a [[link]] pattern
            const wordRange = document.getWordRangeAtPosition(position, /\[\[[^\]]+\]\]/);
            if (!wordRange) return null;

            // Extract the text and strip the brackets to get the raw ATD_id
            const text = document.getText(wordRange);
            const ATDId = text.replace('[[', '').replace(']]', '');

            // Return the location pointing to the very top (line 0, char 0) of the target file
            return new vscode.Location(
                getTargetUri(ATDId),
                new vscode.Position(0, 0)
            );
        }
    });

    // Register both providers
    context.subscriptions.push(linkProvider, definitionProvider);
}

function deactivate() { }

module.exports = {
    activate,
    deactivate
};