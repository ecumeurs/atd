/**
 * Tree View Module — Hierarchical list view with selection mode.
 * @spec-link [[ui_webui_explorer_treeview]]
 */
import { state, on, setCurrentAtom, toggleSelection } from './state.js';
import { bulkUpdate } from './api.js';

let treeviewContainer;
let bulkBar, selectedCountEl, bulkStatusSelect, btnBulkApply, btnBulkCancel;

export function initTreeView() {
    treeviewContainer = document.getElementById('treeview-container');
    bulkBar = document.getElementById('bulk-bar');
    selectedCountEl = document.getElementById('selected-count');
    bulkStatusSelect = document.getElementById('bulk-status-select');
    btnBulkApply = document.getElementById('btn-bulk-apply');
    btnBulkCancel = document.getElementById('btn-bulk-cancel');

    bulkStatusSelect.addEventListener('change', () => {
        btnBulkApply.disabled = !bulkStatusSelect.value || state.selectedIds.size === 0;
    });

    btnBulkApply.addEventListener('click', applyBulkUpdate);

    on('atoms-changed', renderTreeView);
    on('selection-mode-changed', (isActive) => {
        bulkBar.style.display = isActive ? 'flex' : 'none';
        renderTreeView(state.atoms);
    });
    on('selection-changed', updateBulkUI);
}

function updateBulkUI() {
    selectedCountEl.textContent = state.selectedIds.size;
    btnBulkApply.disabled = state.selectedIds.size === 0 || !bulkStatusSelect.value;
}

async function applyBulkUpdate() {
    const status = bulkStatusSelect.value;
    const ids = Array.from(state.selectedIds);
    btnBulkApply.disabled = true;
    btnBulkApply.textContent = 'Updating...';

    try {
        const resp = await bulkUpdate(ids, status);
        if (resp.ok) {
            const { emit } = await import('./state.js');
            emit('data-refresh-needed');
        } else {
            const err = await resp.json();
            alert('Update failed: ' + (err.error || 'Unknown error'));
        }
    } catch (error) {
        alert('Failed to send bulk update request.');
    } finally {
        btnBulkApply.textContent = 'Apply Change';
        updateBulkUI();
    }
}

function buildHierarchy(flatData) {
    const root = { id: 'root', name: 'All Atoms', children: [] };
    const map = new Map();
    flatData.forEach(node => map.set(node.id, { ...node, children: [], value: 1 }));
    map.forEach(node => {
        if (node.parents && node.parents.length > 0) {
            let parentId = node.parents[0];
            let parentNode = map.get(parentId);
            if (parentNode) parentNode.children.push(node);
            else root.children.push(node);
        } else {
            root.children.push(node);
        }
    });
    return root;
}

export function renderTreeView(rawData) {
    if (!rawData || !treeviewContainer) return;
    treeviewContainer.innerHTML = '';
    const hierarchicalData = buildHierarchy(rawData);
    const rootList = document.createElement('div');
    hierarchicalData.children.forEach(child => {
        rootList.appendChild(buildTreeElement(child));
    });
    treeviewContainer.appendChild(rootList);
}

function buildTreeElement(node) {
    const container = document.createElement('div');
    container.className = 'tree-node-container';

    const item = document.createElement('div');
    const isRoot = node.id === 'root';
    item.className = `tree-item ${getStatusClass(node)}${isRoot ? '' : ' selectable'}`;
    item.dataset.id = node.id;

    if (state.isSelectionMode && !isRoot) {
        const cb = document.createElement('input');
        cb.type = 'checkbox';
        cb.className = 'tree-checkbox';
        cb.checked = state.selectedIds.has(node.id);
        cb.addEventListener('click', (e) => e.stopPropagation());
        cb.addEventListener('change', () => {
            toggleSelection(node.id);
        });
        item.appendChild(cb);
    }

    if (node.version && !isRoot) {
        const ver = document.createElement('span');
        ver.className = 'version-badge';
        ver.textContent = `v${node.version}`;
        item.appendChild(ver);
    }

    if (node.status && !isRoot) {
        const stat = document.createElement('span');
        stat.className = `status-pill ${node.status}`;
        stat.textContent = node.status;
        item.appendChild(stat);
    }

    const title = document.createElement('span');
    title.textContent = node.human_name || node.id;
    item.appendChild(title);

    const type = document.createElement('span');
    type.className = 'type-badge';
    type.textContent = node.type;
    item.appendChild(type);

    item.addEventListener('click', (e) => {
        e.stopPropagation();
        if (state.isSelectionMode && !isRoot) {
            toggleSelection(node.id);
            const cb = item.querySelector('.tree-checkbox');
            if (cb) cb.checked = state.selectedIds.has(node.id);
        } else {
            setCurrentAtom(node);
        }
    });

    container.appendChild(item);

    if (node.children && node.children.length > 0) {
        const childrenContainer = document.createElement('div');
        childrenContainer.className = 'tree-node';
        node.children.forEach(child => {
            childrenContainer.appendChild(buildTreeElement(child));
        });
        container.appendChild(childrenContainer);
    }

    return container;
}

function getStatusClass(atom) {
    if (!atom.id || atom.id === 'root') return 'status-grey';
    return `status-${atom.computed_color || 'grey'}`;
}
