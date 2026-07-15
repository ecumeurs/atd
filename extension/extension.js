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

    async function runAtd(args) {
        const { stdout, stderr } = await cp.spawn('atd', args, {
            cwd: workspaceRoot
        });
        const output = await new Promise((resolve) => {
            stdout.on('data', (data) => resolve(data.toString()));
            stderr.on('data', () => resolve(''));
        });
        if (stderr) outputChannel.appendLine(`stderr: ${stderr}`);
        return JSON.parse(output);
    }

    let docsPath = '';
    let workspaceConfig = null;
    let activeProject = null;

    // Status bar item showing the active project (workspace mode only)
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.command = 'atd.switchProject';
    statusBarItem.tooltip = 'ATD: click to switch active project';
    context.subscriptions.push(statusBarItem);

    // Sets the active project and updates docsPath + status bar accordingly.
    const setActiveProject = (proj) => {
        activeProject = proj;
        const projAtdFile = path.join(workspaceRoot, proj.path, '.atd');
        let projDocsPath = proj.docs_path || 'docs/';
        try {
            if (fs.existsSync(projAtdFile)) {
                const cfg = JSON.parse(fs.readFileSync(projAtdFile, 'utf8'));
                projDocsPath = cfg.docs_path || projDocsPath;
            }
        } catch (e) { /* keep default */ }
        docsPath = path.join(proj.path, projDocsPath);
        statusBarItem.text = `$(project) ATD: ${proj.name}`;
        statusBarItem.show();
        outputChannel.appendLine(`[ATD Linker] Active project: ${proj.name} (docs: ${docsPath})`);
    };

    // Get the first open workspace folder
    const workspaceFolder = vscode.workspace.workspaceFolders?.[0];
    if (!workspaceFolder) {
        outputChannel.appendLine("[ATD Linker] No workspace folder found.");
        return;
    }
    const workspaceRoot = workspaceFolder.uri.fsPath;

    const loadWorkspace = () => {
        const wsFile = path.join(workspaceRoot, '.atd.workspace');
        try {
            if (fs.existsSync(wsFile)) {
                const data = fs.readFileSync(wsFile, 'utf8');
                workspaceConfig = JSON.parse(data);
                outputChannel.appendLine(`[ATD Linker] Loaded workspace with ${workspaceConfig.projects?.length || 0} projects.`);
                // Auto-select the first project if none is active yet
                if (!activeProject && workspaceConfig.projects?.length > 0) {
                    setActiveProject(workspaceConfig.projects[0]);
                }
            }
        } catch (err) {
            outputChannel.appendLine(`[ATD Linker ERROR] Failed to parse .atd.workspace: ${err}`);
        }
    };

    // @spec-link [[mechanic_vscode_atd_config]]
    // 1. Function to read the .atd configuration file
    const loadConfig = () => {
        const atdFile = path.join(workspaceRoot, '.atd');
        try {
            if (fs.existsSync(atdFile)) {
                const data = fs.readFileSync(atdFile, 'utf8');
                const config = JSON.parse(data);
                docsPath = config.docs_path || 'docs/';
                outputChannel.appendLine(`[ATD Linker] Loaded docs_path: ${docsPath}`);
            }
        } catch (err) {
            outputChannel.appendLine(`[ATD Linker ERROR] Failed to parse .atd config: ${err}`);
        }
    };

    loadWorkspace();
    loadConfig();

    const watcher = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(workspaceRoot, '.atd'));
    watcher.onDidChange(loadConfig);
    watcher.onDidCreate(loadConfig);

    const wsWatcher = vscode.workspace.createFileSystemWatcher(new vscode.RelativePattern(workspaceRoot, '.atd.workspace'));
    wsWatcher.onDidChange(loadWorkspace);
    wsWatcher.onDidCreate(loadWorkspace);

    const getTargetUri = (ATDId) => {
        // Handle cross-project references [[project:atom_id]]
        if (ATDId.includes(':') && workspaceConfig) {
            const [projName, atomId] = ATDId.split(':');
            const proj = workspaceConfig.projects.find(p => p.name === projName);
            if (proj) {
                const projDocs = proj.docs_path || 'docs/';
                const targetPath = path.join(workspaceRoot, proj.path, projDocs, `${atomId}.atom.md`);
                return vscode.Uri.file(targetPath);
            }
        }

        // Default to local project
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
                const sectionRegex = new RegExp(`${section}:\\s*\\n(?:\\s*-\\s*\\[\\[([a-zA-Z0-9_\\-\\:]+)\\]\\]\\n?)*`, 'm');
                const match = content.match(sectionRegex);
                if (!match) return [];
                return [...match[0].matchAll(/\[\[([a-zA-Z0-9_\\-\\:]+)\]\]/g)].map(m => m[1]);
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
                const traceData = await runAtd(['trace', atomId]);

                const implPercent = Math.round(traceData.health_summary.implementation_rate * 100);
                const testPercent = Math.round(traceData.health_summary.test_coverage_rate * 100);
                const warningCount = traceData.warnings ? traceData.warnings.length : 0;

                return [
                    new vscode.CodeLens(new vscode.Range(0, 0, 0, 0), {
                        title: `✅ Ancestry | ⚙️ Impl: ${implPercent}% | 🧪 Tests: ${testPercent}% ${warningCount > 0 ? `| ⚠️ ${warningCount}` : ''}`,
                        command: 'atd.showDetails',
                        arguments: [traceData]
                    }),
                    new vscode.CodeLens(new vscode.Range(0, 0, 0, 0), {
                        title: '✏️ Rename Atom',
                        command: 'atd.renameAtom'
                    })
                ];
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

    // @spec-link [[mechanic_vscode_rename_atom]]
    const renameAtomCommand = vscode.commands.registerCommand('atd.renameAtom', async () => {
        const activeEditor = vscode.window.activeTextEditor;
        if (!activeEditor || !activeEditor.document.fileName.endsWith('.atom.md')) {
            vscode.window.showWarningMessage('Please open an .atom.md file to rename it.');
            return;
        }

        const fileName = path.basename(activeEditor.document.fileName);
        const currentID = fileName.replace('.atom.md', '');

        const newID = await vscode.window.showInputBox({
            prompt: 'Enter new atom ID',
            value: currentID,
            placeHolder: 'new_atom_id',
            validateInput: (value) => {
                if (!value || value.trim() === '') {
                    return 'Atom ID cannot be empty';
                }
                if (value.includes(' ') || value.includes('(') || value.includes(')')) {
                    return 'Atom ID cannot contain spaces or parentheses';
                }
                return null;
            }
        });

        if (!newID || newID === currentID) {
            return;
        }

        const trimmedID = newID.trim();
        outputChannel.appendLine(`[ATD Rename] Renaming ${currentID} to ${trimmedID}`);

        try {
            await cp.spawn('atd', ['update', '--file', activeEditor.document.fileName, '--set', `id=${trimmedID}`], {
                cwd: workspaceRoot
            });
            outputChannel.appendLine(`[ATD Rename] Success: atom file updated`);

            // Trigger weave to update parent/dependent relationships
            await execAsync(`atd weave`, { cwd: workspaceRoot });
            outputChannel.appendLine(`[ATD Rename] Dependencies updated via atd weave`);

            vscode.window.showInformationMessage(`Atom renamed from "${currentID}" to "${trimmedID}"`);

            // Open the renamed file
            const newFilePath = path.join(workspaceRoot, docsPath, `${trimmedID}.atom.md`);
            const doc = await vscode.workspace.openTextDocument(vscode.Uri.file(newFilePath));
            await vscode.window.showTextDocument(doc);
        } catch (e) {
            outputChannel.appendLine(`[ATD Rename ERROR] ${e.message}`);
            vscode.window.showErrorMessage(`Failed to rename atom: ${e.message}`);
        }
    });

    // @spec-link [[mechanic_vscode_switch_project]]
    const switchProjectCommand = vscode.commands.registerCommand('atd.switchProject', async () => {
        if (!workspaceConfig || !workspaceConfig.projects?.length) {
            vscode.window.showInformationMessage('ATD: Not in a multi-project workspace.');
            return;
        }
        const items = workspaceConfig.projects.map(p => ({
            label: p.name,
            description: p.path,
            detail: activeProject?.name === p.name ? '$(check) currently active' : '',
            project: p
        }));
        const selected = await vscode.window.showQuickPick(items, {
            placeHolder: 'Select active ATD project',
            matchOnDescription: true
        });
        if (selected) {
            setActiveProject(selected.project);
            vscode.window.showInformationMessage(`ATD: Switched to project "${selected.label}"`);
        }
    });

    // 4. THE NEW HOVER PROVIDER FOR SOURCE CODE
    // @spec-link [[mechanic_vscode_hover_provider]]
    const hoverProvider = vscode.languages.registerHoverProvider('*', {
        async provideHover(document, position) {
            const range = document.getWordRangeAtPosition(position, /@spec-link\s+\[\[([a-zA-Z0-9_\-\:]+)\]\]/);
            if (!range) return null;

            const text = document.getText(range);
            const match = text.match(/\[\[([a-zA-Z0-9_\-\:]+)\]\]/);
            if (!match) return null;

            const atomId = match[1];
            const meta = parseAtomMetadata(atomId);

            let traceData = null;
            try {
                traceData = await runAtd(['trace', atomId]);
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
                    const trace = await runAtd(['trace', id]);
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
    const treeDisposable = vscode.window.registerTreeDataProvider('atdGraphExplorer', atdGraphProvider);

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
                    retainContextWhenHidden: true, // Keeps the graph loaded when switching tabs
                    localResourceRoots: [vscode.Uri.joinPath(context.extensionUri, 'node_modules', 'vis-network', 'standalone', 'umd')]
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
            const traceData = await runAtd(['trace', atomId]);

            // Re-inject the HTML with the new data
            const visNetworkUri = graphPanel.webview.asWebviewUri(
                vscode.Uri.joinPath(context.extensionUri, 'node_modules', 'vis-network', 'standalone', 'umd', 'vis-network.min.js')
            );

            graphPanel.webview.html = `
                <!DOCTYPE html>
                <html lang="en">
                <head>
                    <script type="text/javascript" src="${visNetworkUri}"></script>
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
                        const traceData = ${JSON.stringify(traceData)};
                        
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
            graphPanel.webview.html = `<div style="padding: 20px;"><h1>Error generating graph</h1><p>${e.message.replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p></div>`;
        }
    };

    // The manual command now just triggers the update function
    const showGraphCommand = vscode.commands.registerCommand('atd.showFullGraph', () => {
        updateGraphPanel(vscode.window.activeTextEditor);
    });

    // 7. UNIFIED EDITOR CHANGE LISTENER
    // This watches for tab changes and updates both the Sidebar and the Webview!
    const editorListenerDisposable = vscode.window.onDidChangeActiveTextEditor(editor => {
        if (editor && editor.document.fileName.endsWith('.atom.md')) {
            const atomId = path.basename(editor.document.fileName, '.atom.md');

            // 1. Update the Sidebar
            atdGraphProvider.refresh(atomId);

            // 2. Update the Webview Graph (Will auto-open if enabled)
            const config = vscode.workspace.getConfiguration('atd');
            const autoShow = config.get('autoShowGraph', true);
            
            if (autoShow || graphPanel) {
                updateGraphPanel(editor);
            }
        }
    });


    context.subscriptions.push(
        treeDisposable, editorListenerDisposable, watcher, wsWatcher, codeLensProvider, hoverProvider, linkProvider, definitionProvider,
        showDetailsCommand, renameAtomCommand, switchProjectCommand, showGraphCommand
    );
}

function deactivate() { }

module.exports = { activate, deactivate };