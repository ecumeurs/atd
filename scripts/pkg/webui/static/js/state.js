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
};

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
