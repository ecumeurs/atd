/**
 * State Module — Shared reactive state with simple event bus.
 * @spec-link [[module_webui]]
 */

const listeners = {};

export const state = {
    atoms: [],           // flat array of all atoms from /api/tree
    currentAtom: null,   // currently selected atom for detail panel
    selectedIds: new Set(),
    isSelectionMode: false,
    currentView: 'waterfall', // 'waterfall' | 'tree'
    foundationCollapsed: true,
    workspace: {
        inWorkspace: false,
        workspaceName: null,
        workspaceRoot: null,
        activeProject: null,
        projects: []
    }
};

export function setWorkspace(ws) {
    state.workspace = { ...state.workspace, ...ws };
    // Handle camelCase conversion from snake_case API response
    if (ws.in_workspace !== undefined) state.workspace.inWorkspace = ws.in_workspace;
    if (ws.workspace_name !== undefined) state.workspace.workspaceName = ws.workspace_name;
    if (ws.workspace_root !== undefined) state.workspace.workspaceRoot = ws.workspace_root;
    if (ws.active_project !== undefined) state.workspace.activeProject = ws.active_project;
    
    emit('workspace-updated', state.workspace);
}

export function setActiveProject(projectName) {
    state.workspace.activeProject = projectName;
    emit('workspace-updated', state.workspace);
}

export function on(event, fn) {
    if (!listeners[event]) listeners[event] = [];
    listeners[event].push(fn);
}

export function emit(event, data) {
    (listeners[event] || []).forEach(fn => fn(data));
}

export function setAtoms(data) {
    state.atoms = data || [];
    emit('atoms-changed', state.atoms);
}

export function setCurrentAtom(atom) {
    state.currentAtom = atom;
    emit('atom-selected', atom);
}

export function toggleSelectionMode() {
    state.isSelectionMode = !state.isSelectionMode;
    state.selectedIds.clear();
    emit('selection-mode-changed', state.isSelectionMode);
}

export function toggleSelection(id) {
    if (state.selectedIds.has(id)) state.selectedIds.delete(id);
    else state.selectedIds.add(id);
    emit('selection-changed', state.selectedIds);
}

export function setView(view) {
    state.currentView = view;
    emit('view-changed', view);
}

export function toggleFoundation() {
    state.foundationCollapsed = !state.foundationCollapsed;
    emit('foundation-toggled', state.foundationCollapsed);
}

// Build atom lookup map for O(1) access
export function getAtomMap() {
    const map = new Map();
    state.atoms.forEach(a => map.set(a.id, a));
    return map;
}
