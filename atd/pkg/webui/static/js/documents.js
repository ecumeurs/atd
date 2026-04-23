/**
 * Documents Module — Handles Context Setup and Document Assembly
 * @spec-link [[ui_webui_document_viewer]]
 */
import { on, state, emit } from './state.js';
import { searchDocumentContext, generateDocument, fetchRecentDocuments } from './api.js';

let setupModal, viewerModal;
let promptInput, searchInput, searchBtn, foundList, selectedContainer, generateBtn, loadingSpinner, clearBtn, lengthSelect;
let viewerTitle, viewerMeta, viewerContent, downloadBtn;
let recentDocsBtn, recentDocsDropdown, recentDocsList;
let selectedAtoms = new Set();
let foundAtoms = [];

export function initDocuments() {
    setupModal = document.getElementById('document-setup-modal');
    viewerModal = document.getElementById('document-modal');

    promptInput = document.getElementById('doc-setup-prompt');
    searchInput = document.getElementById('doc-setup-search');
    searchBtn = document.getElementById('btn-doc-setup-search');
    foundList = document.getElementById('doc-setup-found-list');
    selectedContainer = document.getElementById('doc-setup-selected-container');
    clearBtn = document.getElementById('btn-doc-setup-clear');
    lengthSelect = document.getElementById('doc-setup-length');
    generateBtn = document.getElementById('btn-doc-setup-generate');
    loadingSpinner = document.getElementById('doc-setup-loading');
    
    // NEW: Workspace toggle
    const workspaceGroup = document.getElementById('doc-setup-workspace-group');

    viewerTitle = document.getElementById('document-modal-title');
    viewerMeta = document.getElementById('document-modal-meta');
    viewerContent = document.getElementById('document-modal-content');
    downloadBtn = document.getElementById('btn-download-document');

    recentDocsBtn = document.getElementById('btn-recent-docs');
    recentDocsDropdown = document.getElementById('recent-docs-dropdown');
    recentDocsList = document.getElementById('recent-docs-list');

    // Setup Modal Controls
    document.getElementById('close-document-setup')?.addEventListener('click', () => {
        setupModal.style.display = 'none';
    });

    document.getElementById('close-document-modal')?.addEventListener('click', () => {
        viewerModal.style.display = 'none';
    });

    searchBtn?.addEventListener('click', () => handleSearch(searchInput.value.trim()));

    searchInput?.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            handleSearch(searchInput.value.trim());
        }
    });

    clearBtn?.addEventListener('click', () => {
        selectedAtoms.clear();
        updateSelectedDisplay();
    });

    generateBtn?.addEventListener('click', handleGenerate);

    on('document-setup-requested', async (query) => {
        promptInput.value = query;
        setupModal.style.display = 'flex';
        loadingSpinner.style.display = 'none';
        generateBtn.disabled = false;
        selectedAtoms.clear();
        updateSelectedDisplay();
        foundList.innerHTML = '<li style="padding: 10px;">Search for ATDs above</li>';

        // NEW: Show/hide workspace toggle based on state
        const workspaceGroup = document.getElementById('doc-setup-workspace-group');
        if (workspaceGroup) {
            workspaceGroup.style.display = state.workspace.inWorkspace ? 'block' : 'none';
        }

        // Auto-search if query provided
        if (query.trim().length >= 2) {
            await handleSearch(query);
        }
    });

    // Recent Docs Dropdown
    recentDocsBtn?.addEventListener('click', async (e) => {
        e.stopPropagation();
        const disp = recentDocsDropdown.style.display;
        if (disp === 'block') {
            recentDocsDropdown.style.display = 'none';
        } else {
            recentDocsDropdown.style.display = 'block';
            await loadRecentDocs();
        }
    });

    document.addEventListener('click', () => {
        if (recentDocsDropdown) recentDocsDropdown.style.display = 'none';
    });
}

async function handleSearch(query) {
    if (query.length < 2) {
        foundList.innerHTML = '<li style="padding: 10px; text-align: center;">Type at least 2 characters</li>';
        return;
    }

    foundList.innerHTML = '<li style="padding: 10px; text-align: center;">Searching...</li>';

    try {
        const atoms = await searchAtoms(query);
        foundAtoms = atoms || [];
        updateFoundDisplay();
    } catch (err) {
        foundList.innerHTML = `<li style="padding: 10px; color: var(--color-red);">Search error: ${err.message}</li>`;
    }
}

function updateFoundDisplay() {
    foundList.innerHTML = '';

    if (!foundAtoms || foundAtoms.length === 0) {
        foundList.innerHTML = '<li style="padding: 10px; text-align: center;">No ATDs found</li>';
        document.getElementById('doc-setup-found-count').textContent = '0 found';
        return;
    }

    document.getElementById('doc-setup-found-count').textContent = `${foundAtoms.length} found`;

    const layerColors = {
        CUSTOMER: 'var(--color-customer)',
        ARCHITECTURE: 'var(--color-architecture)',
        IMPLEMENTATION: 'var(--color-implementation)',
    };

    foundAtoms.forEach(atom => {
        const li = document.createElement('li');
        li.style.padding = '8px 10px';
        li.style.borderBottom = '1px solid var(--border)';
        li.style.display = 'flex';
        li.style.alignItems = 'center';
        li.style.gap = '10px';
        li.style.cursor = 'pointer';
        li.style.transition = 'background 0.2s';

        const isSelected = selectedAtoms.has(atom.id);
        if (isSelected) {
            li.style.background = 'var(--color-selected)';
        }

        const info = document.createElement('div');
        info.style.flex = '1';
        info.innerHTML = `
            <span style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}; font-size: 11px; margin-right: 5px; font-weight: bold;">[${atom.layer}]</span>
            <span style="font-weight: 500;">${atom.id}</span>
            <div style="font-size: 11px; color: var(--text-muted); margin-top: 2px;">${atom.human_name || ''}</div>
        `;

        li.appendChild(info);
        li.addEventListener('click', () => toggleAtomSelection(atom));
        li.addEventListener('mouseenter', () => {
            if (!selectedAtoms.has(atom.id)) li.style.background = 'var(--color-hover)';
        });
        li.addEventListener('mouseleave', () => {
            if (!selectedAtoms.has(atom.id)) li.style.background = '';
        });

        foundList.appendChild(li);
    });
}

function toggleAtomSelection(atom) {
    if (selectedAtoms.has(atom.id)) {
        selectedAtoms.delete(atom.id);
    } else {
        selectedAtoms.add(atom.id);
    }
    updateFoundDisplay();
    updateSelectedDisplay();
}

function updateSelectedDisplay() {
    // Clear current display
    selectedContainer.innerHTML = '';

    if (selectedAtoms.size === 0) {
        const placeholder = document.createElement('div');
        placeholder.id = 'doc-setup-selected-placeholder';
        placeholder.style.color = 'var(--text-muted)';
        placeholder.style.fontSize = '13px';
        placeholder.textContent = 'No ATDs selected';
        selectedContainer.appendChild(placeholder);
        return;
    }

    // Create tags for selected atoms
    selectedAtoms.forEach(atomId => {
        const atom = foundAtoms.find(a => a.id === atomId) || state.atoms.find(a => a.id === atomId);
        if (!atom) return;

        const tag = document.createElement('div');
        tag.className = 'atom-tag';
        tag.style.display = 'flex';
        tag.style.alignItems = 'center';
        tag.style.gap = '5px';
        tag.style.background = 'var(--color-primary)';
        tag.style.color = 'white';
        tag.style.padding = '5px 10px';
        tag.style.borderRadius = '20px';
        tag.style.fontSize = '12px';
        tag.style.fontWeight = '500';
        tag.style.transition = 'transform 0.2s, background 0.2s';

        const layerColors = {
            CUSTOMER: '#4CAF50',
            ARCHITECTURE: '#2196F3',
            IMPLEMENTATION: '#FF9800',
        };

        tag.innerHTML = `
            <span style="color: ${layerColors[atom.layer] || '#666'}; font-weight: bold; margin-right: 2px;">[${atom.layer}]</span>
            <span>${atom.id}</span>
            <button class="tag-remove" style="background: none; border: none; color: white; cursor: pointer; padding: 0; font-size: 14px; line-height: 1; opacity: 0.7;" title="Remove">×</button>
        `;

        tag.querySelector('.tag-remove').addEventListener('click', (e) => {
            e.stopPropagation();
            selectedAtoms.delete(atomId);
            updateFoundDisplay();
            updateSelectedDisplay();
        });

        tag.addEventListener('mouseenter', () => {
            tag.style.transform = 'scale(1.05)';
            tag.querySelector('.tag-remove').style.opacity = '1';
        });
        tag.addEventListener('mouseleave', () => {
            tag.style.transform = 'scale(1)';
            tag.querySelector('.tag-remove').style.opacity = '0.7';
        });

        selectedContainer.appendChild(tag);
    });
}

function openAtomPicker(e) {
    e.stopPropagation();
    const existing = document.querySelector('.inline-atom-picker');
    if (existing) existing.remove();

    const picker = document.createElement('div');
    picker.className = 'inline-atom-picker';
    picker.style.position = 'absolute';
    picker.style.background = 'var(--bg-panel)';
    picker.style.border = '1px solid var(--border)';
    picker.style.padding = '10px';
    picker.style.borderRadius = '4px';
    picker.style.boxShadow = '0 4px 6px rgba(0,0,0,0.3)';
    picker.style.zIndex = '1001';
    
    const rect = addAtomBtn.getBoundingClientRect();
    picker.style.top = `${rect.bottom + 5}px`;
    picker.style.left = `${rect.left}px`;
    picker.style.width = '300px';

    const input = document.createElement('input');
    input.type = 'text';
    input.placeholder = 'Search globally...';
    input.className = 'search-input';
    input.style.width = '100%';
    input.style.marginBottom = '10px';

    const results = document.createElement('ul');
    results.style.listStyle = 'none';
    results.style.padding = '0';
    results.style.margin = '0';
    results.style.maxHeight = '150px';
    results.style.overflowY = 'auto';

    picker.appendChild(input);
    picker.appendChild(results);
    document.body.appendChild(picker);

    input.focus();

    input.addEventListener('input', () => {
        const query = input.value.toLowerCase();
        results.innerHTML = '';
        if (query.length < 2) return;

        const matches = state?.atoms?.filter(a =>
            a.id.toLowerCase().includes(query) ||
            (a.human_name && a.human_name.toLowerCase().includes(query))
        ) || [];

        matches.slice(0, 10).forEach(atom => {
            const li = document.createElement('li');
            li.style.padding = '5px 0';
            li.style.borderBottom = '1px dashed var(--border)';
            li.style.cursor = 'pointer';
            li.style.fontSize = '12px';
            li.innerHTML = `<strong>${atom.id}</strong><br><span style="color:var(--text-muted)">${atom.layer}</span>`;
            li.addEventListener('click', () => {
                addAtomToChecklist(atom, true);
                picker.remove();
            });
            results.appendChild(li);
        });
    });

    document.addEventListener('click', function closePicker(event) {
        if (!picker.contains(event.target) && event.target !== addAtomBtn) {
            picker.remove();
            document.removeEventListener('click', closePicker);
        }
    });
}

async function handleGenerate() {
    const prompt = promptInput.value.trim();
    if (!prompt) {
        alert('Please provide a prompt or narrative request.');
        return;
    }

    const starts = Array.from(selectedAtoms);
    if (starts.length === 0) {
        alert('Please select at least one ATD to begin assembly.');
        return;
    }

    generateBtn.disabled = true;
    loadingSpinner.style.display = 'flex';

    try {
        // Get length parameter if provided
        let length = null;
        const lengthValue = lengthSelect.value.trim();
        if (lengthValue) {
            length = lengthValue;
        }

        const useWorkspace = document.getElementById('doc-setup-workspace')?.checked || false;
        const doc = await generateDocument(prompt, starts, length, useWorkspace);
        setupModal.style.display = 'none';
        showDocumentViewer(doc);
    } catch(err) {
        alert('Generation failed: ' + err.message);
    } finally {
        generateBtn.disabled = false;
        loadingSpinner.style.display = 'none';
    }
}

function showDocumentViewer(doc) {
    viewerTitle.textContent = "Generated: " + doc.intent;
    viewerMeta.textContent = `Generated on: ${new Date(doc.timestamp).toLocaleString()} • Nodes involved: ${doc.atoms_involved.length}`;
    
    viewerContent.innerHTML = marked.parse(doc.content || "*No content generated.*");
    
    downloadBtn.onclick = () => {
        const blob = new Blob([doc.content], { type: 'text/markdown' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        const slug = doc.intent.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/(^-|-$)/g, '');
        a.download = `atd-doc-${slug}.md`;
        a.click();
        URL.revokeObjectURL(url);
    };

    viewerModal.style.display = 'flex';
}

async function loadRecentDocs() {
    recentDocsList.innerHTML = '<li style="padding: 10px;">Loading...</li>';
    try {
        const docs = await fetchRecentDocuments();
        recentDocsList.innerHTML = '';
        if (!docs || docs.length === 0) {
            recentDocsList.innerHTML = '<li style="padding: 10px;">No recent docs in cache.</li>';
            return;
        }

        docs.forEach(doc => {
            const li = document.createElement('li');
            li.style.padding = '8px 10px';
            li.style.borderBottom = '1px solid var(--border)';
            li.style.cursor = 'pointer';
            
            li.innerHTML = `
                <div style="font-weight: 500; font-size: 13px;">${escapeHtml(doc.intent)}</div>
                <div style="font-size: 11px; color: var(--text-muted);">${new Date(doc.timestamp).toLocaleString()}</div>
            `;
            
            li.addEventListener('click', async () => {
                const fullResp = await fetch(`/api/documents/${doc.id}`);
                const fullDoc = await fullResp.json();
                showDocumentViewer(fullDoc);
            });
            recentDocsList.appendChild(li);
        });
    } catch (e) {
        recentDocsList.innerHTML = `<li style="padding: 10px; color: var(--color-red);">Failed to load</li>`;
    }
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
}
