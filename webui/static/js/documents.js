/**
 * Documents Module — Handles Context Setup and Document Assembly
 * @spec-link [[ui_webui_document_viewer]]
 */
import { on, state, emit } from './state.js';
import { searchDocumentContext, generateDocument, fetchRecentDocuments } from './api.js';

let setupModal, viewerModal;
let intentInput, atomList, addAtomBtn, generateBtn, loadingSpinner;
let viewerTitle, viewerMeta, viewerContent, downloadBtn;
let recentDocsBtn, recentDocsDropdown, recentDocsList;

export function initDocuments() {
    setupModal = document.getElementById('document-setup-modal');
    viewerModal = document.getElementById('document-modal');
    
    intentInput = document.getElementById('doc-setup-intent');
    atomList = document.getElementById('doc-setup-atoms-list');
    addAtomBtn = document.getElementById('btn-doc-setup-add');
    generateBtn = document.getElementById('btn-doc-setup-generate');
    loadingSpinner = document.getElementById('doc-setup-loading');
    
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

    addAtomBtn?.addEventListener('click', openAtomPicker);
    
    generateBtn?.addEventListener('click', handleGenerate);

    on('document-setup-requested', async (query) => {
        intentInput.value = query;
        atomList.innerHTML = '<li style="padding: 10px;">Searching context...</li>';
        setupModal.style.display = 'flex';
        loadingSpinner.style.display = 'none';
        generateBtn.disabled = false;

        try {
            const context = await searchDocumentContext(query);
            atomList.innerHTML = '';
            if (!context || context.length === 0) {
                atomList.innerHTML = '<li style="padding: 10px;">No contextual ATDs found. Please add manually.</li>';
            } else {
                context.forEach(atom => addAtomToChecklist(atom));
            }
        } catch (e) {
            atomList.innerHTML = `<li style="padding: 10px; color: var(--color-red);">Error: ${e.message}</li>`;
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

function addAtomToChecklist(atom, checked = true) {
    if (atomList.querySelector(`input[value="${atom.id}"]`)) return;

    const li = document.createElement('li');
    li.style.padding = '8px 10px';
    li.style.borderBottom = '1px solid var(--border)';
    li.style.display = 'flex';
    li.style.alignItems = 'center';
    li.style.gap = '10px';

    const checkbox = document.createElement('input');
    checkbox.type = 'checkbox';
    checkbox.value = atom.id;
    checkbox.checked = checked;

    const info = document.createElement('div');
    info.style.flex = '1';
    
    const layerColors = {
        CUSTOMER: 'var(--color-customer)',
        ARCHITECTURE: 'var(--color-architecture)',
        IMPLEMENTATION: 'var(--color-implementation)',
    };
    
    info.innerHTML = `
        <span style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}; font-size: 11px; margin-right: 5px; font-weight: bold;">[${atom.layer}]</span>
        <span style="font-weight: 500;">${atom.id}</span>
        <div style="font-size: 11px; color: var(--text-muted); margin-top: 2px;">${atom.human_name || ''}</div>
    `;

    li.appendChild(checkbox);
    li.appendChild(info);
    atomList.appendChild(li);
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

        const matches = window.state?.atoms?.filter(a => 
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
    const intent = intentInput.value.trim();
    if (!intent) {
        alert('Please provide an intent or narrative request.');
        return;
    }

    const startInputs = atomList.querySelectorAll('input:checked');
    const starts = Array.from(startInputs).map(i => i.value);
    
    if (starts.length === 0) {
        alert('Please select at least one graph node to begin assembly.');
        return;
    }

    generateBtn.disabled = true;
    loadingSpinner.style.display = 'flex';

    try {
        const doc = await generateDocument(intent, starts);
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
