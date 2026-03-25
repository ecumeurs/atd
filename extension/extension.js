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

    // @spec-link [[mechanic_vscode_atd_config]]
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

    loadConfig();

    const watcher = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(workspaceRoot, '.atd'));
    watcher.onDidChange(loadConfig);
    watcher.onDidCreate(loadConfig);

    const getTargetUri = (ATDId) => {
        const targetPath = path.join(workspaceRoot, docsPath, `${ATDId}.atom.md`);
        return vscode.Uri.file(targetPath);
    };

    // @spec-link [[mechanic_vscode_atom_parser]]
    // Helper to read and parse an atom file lightly
    const parseAtomMetadata = (atomId) => {
        try {
            const uri = getTargetUri(atomId);
            if (!fs.existsSync(uri.fsPath)) return null;
            const content = fs.readFileSync(uri.fsPath, 'utf8');

            const layerMatch = content.match(/^layer:\s*(.+)$/m);
            const statusMatch = content.match(/^status:\s*(.+)$/m);
            const priorityMatch = content.match(/^priority:\s*(.+)$/m);
            const intentMatch = content.match(/## INTENT\n([^#]+)/);

            const getLinks = (section) => {
                const sectionRegex = new RegExp(`${section}:\\s*\\n(?:\\s*-\\s*\\[\\[([a-zA-Z0-9_-]+)\\]\\]\\n?)*`, 'm');
                const match = content.match(sectionRegex);
                if (!match) return [];
                return [...match[0].matchAll(/\[\[([a-zA-Z0-9_-]+)\]\]/g)].map(m => m[1]);
            };

            return {
                layer: layerMatch ? layerMatch[1].trim() : 'UNKNOWN',
                status: statusMatch ? statusMatch[1].trim() : 'UNKNOWN',
                priority: priorityMatch ? priorityMatch[1].trim() : '?',
                intent: intentMatch ? intentMatch[1].trim() : 'No intent documented.',
                parents: getLinks('parents'),
                dependents: getLinks('dependents')
            };
        } catch (e) {
            return null;
        }
    };

    // @spec-link [[service_vscode_linker_features]]
    // 2. Link & Definition Providers
    // @spec-link [[mechanic_vscode_link_provider]]
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
                links.push(new vscode.DocumentLink(new vscode.Range(startPos, endPos), getTargetUri(ATDId)));
            }
            return links;
        }
    });

    const definitionProvider = vscode.languages.registerDefinitionProvider('*', {
        provideDefinition(document, position) {
            const wordRange = document.getWordRangeAtPosition(position, /\[\[[^\]]+\]\]/);
            if (!wordRange) return null;
            const ATDId = document.getText(wordRange).replace('[[', '').replace(']]', '');
            return new vscode.Location(getTargetUri(ATDId), new vscode.Position(0, 0));
        }
    });

    // 3. Header CodeLens Provider (Markdown files)
    // @spec-link [[mechanic_vscode_codelens_provider]]
    const lensEmitter = new vscode.EventEmitter();
    const codeLensProvider = vscode.languages.registerCodeLensProvider({ scheme: 'file', language: 'markdown' }, {
        onDidChangeCodeLenses: lensEmitter.event,
        async provideCodeLenses(document) {
            if (!document.fileName.endsWith('.atom.md')) return [];
            const text = document.getText(new vscode.Range(0, 0, 20, 0));
            const idMatch = text.match(/^id:\s*([a-zA-Z0-9_-]+)/m);
            if (!idMatch) return [];

            const atomId = idMatch[1];
            try {
                const { stdout } = await execAsync(`atd trace ${atomId}`, { cwd: workspaceRoot });
                const traceData = JSON.parse(stdout);

                const implPercent = Math.round(traceData.health_summary.implementation_rate * 100);
                const testPercent = Math.round(traceData.health_summary.test_coverage_rate * 100);
                const warningCount = traceData.warnings ? traceData.warnings.length : 0;

                return [new vscode.CodeLens(new vscode.Range(0, 0, 0, 0), {
                    title: `✅ Ancestry | ⚙️ Impl: ${implPercent}% | 🧪 Tests: ${testPercent}% ${warningCount > 0 ? `| ⚠️ ${warningCount}` : ''}`,
                    command: 'atd.showDetails',
                    arguments: [traceData]
                })];
            } catch (e) {
                return [];
            }
        }
    });

    const showDetailsCommand = vscode.commands.registerCommand('atd.showDetails', (traceData) => {
        if (traceData.warnings && traceData.warnings.length > 0) {
            vscode.window.showWarningMessage(`Warnings for ${traceData.target_id}:\n• ${traceData.warnings.join('\n• ')}`, { modal: true });
        } else {
            vscode.window.showInformationMessage(`Health for ${traceData.target_id} looks great!`);
        }
    });

    // 4. THE NEW HOVER PROVIDER FOR SOURCE CODE
    // @spec-link [[mechanic_vscode_hover_provider]]
    const hoverProvider = vscode.languages.registerHoverProvider('*', {
        async provideHover(document, position) {
            const range = document.getWordRangeAtPosition(position, /@spec-link\s+\[\[([a-zA-Z0-9_-]+)\]\]/);
            if (!range) return null;

            const text = document.getText(range);
            const match = text.match(/\[\[([a-zA-Z0-9_-]+)\]\]/);
            if (!match) return null;

            const atomId = match[1];
            const meta = parseAtomMetadata(atomId);

            let traceData = null;
            try {
                const { stdout } = await execAsync(`atd trace ${atomId}`, { cwd: workspaceRoot });
                traceData = JSON.parse(stdout);
            } catch (e) { }

            const md = new vscode.MarkdownString();
            md.isTrusted = true;

            if (meta) {
                md.appendMarkdown(`### 📄 ${atomId} \n\n`);
                md.appendMarkdown(`**Layer:** ${meta.layer} | **Status:** ${meta.status} | **Priority:** ${meta.priority}\n\n`);

                if (traceData) {
                    const implPercent = Math.round(traceData.health_summary.implementation_rate * 100);
                    const testPercent = Math.round(traceData.health_summary.test_coverage_rate * 100);
                    const warnIcon = (traceData.warnings && traceData.warnings.length > 0) ? '🔴' : '🟢';
                    md.appendMarkdown(`**Health:** ⚙️ ${implPercent}% Impl | 🧪 ${testPercent}% Tests ${warnIcon}\n\n`);
                }

                md.appendMarkdown(`---\n**INTENT:**\n*${meta.intent}*\n\n---\n`);

                // Add a command link to open the file directly from the hover
                const args = encodeURIComponent(JSON.stringify([getTargetUri(atomId)]));
                md.appendMarkdown(`[📂 Open Document](command:vscode.open?${args})`);
            } else {
                md.appendMarkdown(`⚠️ **Atom not found:** \`${atomId}\``);
            }

            return new vscode.Hover(md, range);
        }
    });

    // @spec-link [[service_vscode_atd_ui]]
    // 5. ATD GRAPH EXPLORER (SIDEBAR TREE)
    // @spec-link [[mechanic_vscode_sidebar_tree]]
    class ATDGraphProvider {
        constructor() {
            this._onDidChangeTreeData = new vscode.EventEmitter();
            this.onDidChangeTreeData = this._onDidChangeTreeData.event;
            this.currentAtomId = null;
        }

        refresh(atomId) {
            this.currentAtomId = atomId;
            this._onDidChangeTreeData.fire();
        }

        getTreeItem(element) { return element; }

        async getChildren(element) {
            if (!this.currentAtomId) return [new vscode.TreeItem("Open an .atom.md file to view its graph")];

            if (!element) {
                const meta = parseAtomMetadata(this.currentAtomId);
                const cur = new vscode.TreeItem(`📍 CURRENT: ${this.currentAtomId}`, vscode.TreeItemCollapsibleState.None);
                const parents = new vscode.TreeItem(`🔻 PARENTS (${meta?.parents.length || 0})`, vscode.TreeItemCollapsibleState.Expanded);
                parents.contextValue = 'parents';
                const deps = new vscode.TreeItem(`🔻 DEPENDENTS (${meta?.dependents.length || 0})`, vscode.TreeItemCollapsibleState.Expanded);
                deps.contextValue = 'deps';
                return [cur, parents, deps];
            }

            const meta = parseAtomMetadata(this.currentAtomId);
            if (!meta) return [];

            let listToProcess = [];
            if (element.contextValue === 'parents') listToProcess = meta.parents;
            if (element.contextValue === 'deps') listToProcess = meta.dependents;

            // Fetch health for each child asynchronously
            const children = await Promise.all(listToProcess.map(async (id) => {
                const childMeta = parseAtomMetadata(id);
                const item = new vscode.TreeItem(`${id}`, vscode.TreeItemCollapsibleState.None);
                item.description = childMeta ? childMeta.layer : 'Unknown';

                try {
                    const { stdout } = await execAsync(`atd trace ${id}`, { cwd: workspaceRoot });
                    const trace = JSON.parse(stdout);
                    if (trace.health_summary.implementation_rate === 1 && trace.health_summary.test_coverage_rate === 1) {
                        item.iconPath = new vscode.ThemeIcon('pass');
                    } else {
                        item.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('problemsWarningIcon.foreground'));
                    }
                } catch (e) {
                    item.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('problemsErrorIcon.foreground'));
                }

                item.command = {
                    command: 'vscode.open',
                    title: "Open File",
                    arguments: [getTargetUri(id)]
                };
                return item;
            }));

            return children;
        }
    }

    const atdGraphProvider = new ATDGraphProvider();
    vscode.window.registerTreeDataProvider('atdGraphExplorer', atdGraphProvider);

    // Update Sidebar when active editor changes
    vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor && editor.document.fileName.endsWith('.atom.md')) {
            const fileName = path.basename(editor.document.fileName);
            const atomId = fileName.replace('.atom.md', '');
            atdGraphProvider.refresh(atomId);
        }
    });
    // @spec-link [[mechanic_vscode_webview_graph]]
    // 6. ATOM NEIGHBORHOOD GRAPH (WEBVIEW)
    let graphPanel = undefined; // Track the open panel

    // Helper function to update or create the graph
    const updateGraphPanel = async (editor) => {
        if (!editor || !editor.document.fileName.endsWith('.atom.md')) {
            return; // Ignore non-ATD files
        }

        const atomId = path.basename(editor.document.fileName, '.atom.md');

        // If the panel isn't open, create it
        if (!graphPanel) {
            graphPanel = vscode.window.createWebviewPanel(
                'atdGraph',
                `ATD Graph`,
                vscode.ViewColumn.Beside, // Open beside the current editor
                {
                    enableScripts: true,
                    retainContextWhenHidden: true // Keeps the graph loaded when switching tabs
                }
            );

            // Clean up the reference if the user closes the tab
            graphPanel.onDidDispose(() => {
                graphPanel = undefined;
            });
        }

        // Update the panel's title and show a loading state
        graphPanel.title = `ATD: ${atomId}`;
        graphPanel.webview.html = `
            <div style="padding: 20px; font-family: sans-serif; color: var(--vscode-editor-foreground);">
                <h2>Loading Graph for ${atomId}...</h2>
            </div>`;

        try {
            // Fetch new trace data for the newly opened file
            const { stdout } = await execAsync(`atd trace ${atomId}`, { cwd: workspaceRoot });

            // Re-inject the HTML with the new data
            graphPanel.webview.html = `
                <!DOCTYPE html>
                <html lang="en">
                <head>
                    <script type="text/javascript" src="https://unpkg.com/vis-network/standalone/umd/vis-network.min.js"></script>
                    <style type="text/css">
                        body { margin: 0; padding: 0; background: var(--vscode-editor-background); color: var(--vscode-editor-foreground); font-family: sans-serif; }
                        #mynetwork { width: 100vw; height: 100vh; }
                        .header { position: absolute; top: 10px; left: 10px; z-index: 10; background: var(--vscode-editor-background); padding: 10px; border-radius: 5px; border: 1px solid var(--vscode-panel-border); }
                    </style>
                </head>
                <body>
                    <div class="header">
                        <h2>${atomId}</h2>
                        <p>Scroll to zoom. Drag to move.</p>
                    </div>
                    <div id="mynetwork"></div>
                    <script type="text/javascript">
                        const traceData = ${stdout};
                        
                        const nodes = new vis.DataSet();
                        const edges = new vis.DataSet();
                        
                        // Center Node
                        nodes.add({ 
                            id: traceData.target_id, 
                            label: traceData.target_id + "\\n(" + traceData.layer + ")", 
                            shape: 'box', 
                            color: { background: '#007acc', border: '#005a9e' }, 
                            font: { color: 'white', face: 'monospace' },
                            borderWidth: 2
                        });
                        
                        // Parents
                        if (traceData.graph_slice && traceData.graph_slice.parents) {
                            traceData.graph_slice.parents.forEach(p => {
                                if (!nodes.get(p)) {
                                    nodes.add({ id: p, label: p, shape: 'ellipse', color: '#4d4d4d', font: { color: 'white' } });
                                }
                                edges.add({ from: p, to: traceData.target_id, arrows: 'to', color: '#888888' });
                            });
                        }
                        
                        // Dependents
                        if (traceData.graph_slice && traceData.graph_slice.dependents) {
                            traceData.graph_slice.dependents.forEach(d => {
                                if (!nodes.get(d)) {
                                    nodes.add({ id: d, label: d, shape: 'ellipse', color: '#4d4d4d', font: { color: 'white' } });
                                }
                                edges.add({ from: traceData.target_id, to: d, arrows: 'to', color: '#888888' });
                            });
                        }

                        // Code Links
                        if (traceData.graph_slice && traceData.graph_slice.code_links) {
                            traceData.graph_slice.code_links.forEach(codeFile => {
                                if (!nodes.get(codeFile)) {
                                    nodes.add({ id: codeFile, label: codeFile, shape: 'text', font: { color: '#4EC9B0' } });
                                }
                                edges.add({ from: traceData.target_id, to: codeFile, arrows: 'to', color: '#4EC9B0', dashes: true });
                            });
                        }
                        
                        const container = document.getElementById('mynetwork');
                        const data = { nodes: nodes, edges: edges };
                        const options = {
                            physics: { solver: 'repulsion', repulsion: { nodeDistance: 150 } },
                            layout: { hierarchical: { direction: 'UD', sortMethod: 'directed' } }
                        };
                        
                        new vis.Network(container, data, options);
                    </script>
                </body>
                </html>
            `;
        } catch (e) {
            graphPanel.webview.html = `<div style="padding: 20px;"><h1>Error generating graph</h1><p>${e.message}</p></div>`;
        }
    };

    // The manual command now just triggers the update function
    const showGraphCommand = vscode.commands.registerCommand('atd.showFullGraph', () => {
        updateGraphPanel(vscode.window.activeTextEditor);
    });

    // 7. UNIFIED EDITOR CHANGE LISTENER
    // This watches for tab changes and updates both the Sidebar and the Webview!
    vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor && editor.document.fileName.endsWith('.atom.md')) {
            const atomId = path.basename(editor.document.fileName, '.atom.md');

            // 1. Update the Sidebar
            atdGraphProvider.refresh(atomId);

            // 2. Update the Webview Graph (Will auto-open if you want it to always appear)
            updateGraphPanel(editor);
        }
    });

    context.subscriptions.push(
        watcher, codeLensProvider, hoverProvider, linkProvider, definitionProvider, showDetailsCommand, showGraphCommand
    );
}

function deactivate() { }

module.exports = { activate, deactivate };