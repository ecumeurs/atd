// @ts-check
// @spec-link [[specification_vscode_atd_linker]]
/** @typedef {import('vscode')} vscode */
const vscode = require('vscode');
const fs = require('fs');
const path = require('path');
const cp = require('child_process');
const util = require('util');

const execAsync = util.promisify(cp.exec);

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {
    const outputChannel = vscode.window.createOutputChannel("ATD Linker");
    context.subscriptions.push(outputChannel);
    outputChannel.appendLine("[ATD Linker] Extension activation started.");

    let docsPath = '';

    // Get the first open workspace folder
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        outputChannel.appendLine("[ATD Linker] No workspace folder found.");
        return;
    }
    const workspaceRoot = workspaceFolder.uri.fsPath;

    // 1. Function to read the .atd configuration file
    const loadConfig = () => {
        const atdFile = path.join(workspaceRoot, '.atd');
        try {
            if (fs.existsSync(atdFile)) {
                const data = fs.readFileSync(atdFile, 'utf8');
                const config = JSON.parse(data);
                docsPath = config.docs_path || '';
                outputChannel.appendLine(`[ATD Linker] Loaded docs_path: ${docsPath}`);
            }
        } catch (err) {
            outputChannel.appendLine(`[ATD Linker ERROR] Failed to parse .atd config: ${err}`);
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
            const wordRange = document.getWordRangeAtPosition(position, /\[\[[^\]]+\]\]/);
            if (!wordRange) return null;

            const text = document.getText(wordRange);
            const ATDId = text.replace('[[', '').replace(']]', '');

            return new vscode.Location(
                getTargetUri(ATDId),
                new vscode.Position(0, 0)
            );
        }
    });

    // 5. Shared CodeLens Emitter to force UI refresh
    const lensEmitter = new vscode.EventEmitter();
    context.subscriptions.push(lensEmitter);


    // 6. The ATD Health CodeLens Provider
    const codeLensProvider = vscode.languages.registerCodeLensProvider({ scheme: 'file', language: 'markdown' }, {
        onDidChangeCodeLenses: lensEmitter.event,
        async provideCodeLenses(document, token) {
            outputChannel.appendLine(`[ATD Linker] Markdown Lens Provider checking file: ${document.fileName}`);
            // Only run on .atom.md files
            if (!document.fileName.endsWith('.atom.md')) {
                return [];
            }

            outputChannel.appendLine(`[ATD Linker] Generating lenses for atom: ${document.fileName}`);
            // Read the first 20 lines to find the ID in the frontmatter
            const text = document.getText(new vscode.Range(0, 0, 20, 0));
            const idMatch = text.match(/^id:\s*([a-zA-Z0-9_-]+)/m);

            if (!idMatch) return [];

            const atomId = idMatch[1];
            const topOfFile = new vscode.Range(0, 0, 0, 0);

            try {
                // Execute the CLI command in the workspace directory
                const { stdout } = await execAsync(`atd trace ${atomId}`, { cwd: workspaceRoot });
                const traceData = JSON.parse(stdout);

                // Format the metrics
                const ancestryIcon = traceData.health_summary.ancestry_complete ? '✅' : '⚠️';
                const originText = traceData.health_summary.has_customer_origin ? 'Cust' : 'No Cust';

                const implPercent = Math.round(traceData.health_summary.implementation_rate * 100);
                const testPercent = Math.round(traceData.health_summary.test_coverage_rate * 100);
                const warningCount = traceData.warnings ? traceData.warnings.length : 0;
                const warningText = warningCount > 0 ? ` | ⚠️ ${warningCount} Warn` : '';

                const title = `${ancestryIcon} Ancestry: ${originText} | ⚙️ Impl: ${implPercent}% (${traceData.metrics.implemented_dependents}/${traceData.metrics.total_dependents}) | 🧪 Tests: ${testPercent}%${warningText}`;

                // Return the CodeLens attached to the top of the file
                const lens = new vscode.CodeLens(topOfFile, {
                    title: title,
                    command: 'atd.showDetails',
                    arguments: [traceData]
                });

                return [lens];
            } catch (error) {
                outputChannel.appendLine(`[ATD Linker ERROR] Trace failed for ${atomId}: ${error}`);
                return [new vscode.CodeLens(topOfFile, {
                    title: `⚠️ ATD: Trace Failed for ${atomId}`,
                    command: "atd.showDetails",
                    arguments: [{ target_id: atomId, warnings: ["CLI execution failed.", "Is the 'atd' binary compiled and in your system PATH?", String(error)] }]
                })];
            }
        }
    });

    // 6. Command to handle CodeLens clicks
    const showDetailsCommand = vscode.commands.registerCommand('atd.showDetails', (traceData) => {
        if (traceData.warnings && traceData.warnings.length > 0) {
            const warningMsg = traceData.warnings.join(' \n• ');
            vscode.window.showWarningMessage(`ATD Warnings for ${traceData.target_id}:\n• ${warningMsg}`, { modal: true });
        } else {
            vscode.window.showInformationMessage(`ATD Branch Health for ${traceData.target_id} is looking great!`);
        }
    });

    // 8. The Implementation CodeLens Provider (Above @spec-link in source code)
    const implCodeLensProvider = vscode.languages.registerCodeLensProvider({ scheme: 'file' }, {
        onDidChangeCodeLenses: lensEmitter.event,
        async provideCodeLenses(document, token) {
            // Skip markdown files to avoid overlapping with our Header Lens
            if (document.fileName.endsWith('.atom.md') || document.fileName.endsWith('.md')) {
                return [];
            }
            outputChannel.appendLine(`[ATD Linker] Source Code Lens Provider running on: ${document.fileName}`);

            const text = document.getText();
            const regex = /@spec-link\s+\[\[([a-zA-Z0-9_-]+)\]\]/g;
            let match;

            // 1. Collect all matches in the file
            const matches = [];
            while ((match = regex.exec(text))) {
                matches.push({
                    id: match[1],
                    // Get the line number where the tag was found
                    line: document.positionAt(match.index).line
                });
            }

            if (matches.length === 0) return [];

            // 2. Fetch trace data for all found IDs concurrently
            const tracePromises = matches.map(async (m) => {
                const range = new vscode.Range(m.line, 0, m.line, 0);

                try {
                    const { stdout } = await execAsync(`atd trace ${m.id}`, { cwd: workspaceRoot });
                    const traceData = JSON.parse(stdout);

                    // Format the inline UI based on the trace.go output
                    const layer = traceData.layer || 'UNKNOWN';
                    const implPercent = Math.round(traceData.health_summary.implementation_rate * 100);
                    const testPercent = Math.round(traceData.health_summary.test_coverage_rate * 100);
                    const warnCount = traceData.warnings ? traceData.warnings.length : 0;

                    let title = `ATD [${layer}] | Impl: ${implPercent}% | Tests: ${testPercent}%`;
                    if (warnCount > 0) title += ` | ⚠️ ${warnCount} Warn`;

                    return new vscode.CodeLens(range, {
                        title: title,
                        command: 'atd.showDetails', // Reuse the same click command!
                        arguments: [traceData]
                    });
                } catch (err) {
                    outputChannel.appendLine(`[ATD Linker ERROR] Source trace failed for ${m.id}: ${err}`);
                    return new vscode.CodeLens(range, {
                        title: `⚠️ ATD: Trace Failed for ${m.id}`,
                        command: "atd.showDetails",
                        arguments: [{ target_id: m.id, warnings: ["CLI execution failed.", "Is the 'atd' binary compiled and in your system PATH?", String(err)] }]
                    });
                }
            });

            // 3. Wait for all CLI calls to finish and return the lenses
            const resolvedLenses = await Promise.all(tracePromises);
            return resolvedLenses;
        }
    });

    // Register all providers and commands
    context.subscriptions.push(
        watcher,
        codeLensProvider,
        implCodeLensProvider,
        linkProvider,
        definitionProvider,
        showDetailsCommand
    );

}

function deactivate() { }

module.exports = {
    activate,
    deactivate
};