/**
 * Search Overlay Module — Ctrl+K command palette for atom lookup.
 * @spec-link [[ui_webui_search_overlay]]
 */
import { state, setCurrentAtom, emit, isInWorkspace, getWorkspace } from './state.js';
import { searchAtoms } from './api.js';

let overlay = null;

export function initSearch() {
    document.addEventListener('keydown', (e) => {
        if ((e.ctrlKey || e.metaKey) && e.key === 'k') {
            e.preventDefault();
            openSearchOverlay();
        }
        if (e.key === 'Escape' && overlay) {
            closeSearchOverlay();
        }
    });
}

function openSearchOverlay() {
    if (overlay) return;

    overlay = document.createElement('div');
    overlay.className = 'search-overlay';
    overlay.addEventListener('click', (e) => {
        if (e.target === overlay) closeSearchOverlay();
    });

    const panel = document.createElement('div');
    panel.className = 'search-panel';

    const input = document.createElement('input');
    input.className = 'search-input';
    input.placeholder = 'Search atoms by name, ID, or content... (Ctrl+K)';
    input.autofocus = true;

    // NEW: Workspace scope toggle
    if (isInWorkspace()) {
        const workspace = getWorkspace();
        const scopeContainer = document.createElement('div');
        scopeContainer.className = 'search-scope-container';

        const scopeLabel = document.createElement('span');
        scopeLabel.className = 'search-scope-label';
        scopeLabel.textContent = 'Search scope:';

        const scopeToggle = document.createElement('div');
        scopeToggle.className = 'search-scope-toggle';

        const projectRadio = document.createElement('label');
        projectRadio.className = 'scope-option';
        projectRadio.innerHTML = `
            <input type="radio" name="search-scope" value="project" checked>
            <span>${escapeHtml(workspace.activeProject || 'Current Project')}</span>
        `;

        const workspaceRadio = document.createElement('label');
        workspaceRadio.className = 'scope-option';
        workspaceRadio.innerHTML = `
            <input type="radio" name="search-scope" value="workspace">
            <span>All Projects</span>
        `;

        scopeToggle.appendChild(projectRadio);
        scopeToggle.appendChild(workspaceRadio);
        scopeContainer.appendChild(scopeLabel);
        scopeContainer.appendChild(scopeToggle);
        panel.appendChild(scopeContainer);

        // Re-trigger search on toggle
        scopeToggle.addEventListener('change', () => {
            const q = input.value.trim();
            if (q.length >= 2) performSearch(q, results, hint);
        });
    }

    const results = document.createElement('div');
    results.className = 'search-results';

    const hint = document.createElement('div');
    hint.className = 'search-hint';
    hint.textContent = 'Type to search across all atoms';

    const generateBtn = document.createElement('button');
    generateBtn.className = 'btn btn-outline btn-sm';
    generateBtn.id = 'btn-generate-doc';
    generateBtn.style.display = 'none';
    generateBtn.style.marginTop = '10px';
    generateBtn.style.width = '100%';
    generateBtn.style.justifyContent = 'center';

    panel.appendChild(input);
    panel.appendChild(results);
    panel.appendChild(hint);
    panel.appendChild(generateBtn);
    overlay.appendChild(panel);
    document.body.appendChild(overlay);

    setTimeout(() => input.focus(), 50);

    let debounceTimer;
    input.addEventListener('input', () => {
        clearTimeout(debounceTimer);
        const q = input.value.trim();
        if (q.length < 2) {
            results.innerHTML = '';
            hint.textContent = 'Type at least 2 characters to search';
            generateBtn.style.display = 'none';
            return;
        }
        generateBtn.style.display = 'flex';
        generateBtn.innerHTML = `📄 Generate Document for "<span style="font-weight:bold">${escapeHtml(q)}</span>"`;
        generateBtn.onclick = () => {
            closeSearchOverlay();
            emit('document-setup-requested', q);
        };
        
        hint.textContent = 'Searching...';
        debounceTimer = setTimeout(() => performSearch(q, results, hint), 250);
    });

    // Keyboard navigation
    let selectedIndex = -1;
    input.addEventListener('keydown', (e) => {
        const items = results.querySelectorAll('.search-result-item');
        if (e.key === 'ArrowDown') {
            e.preventDefault();
            selectedIndex = Math.min(selectedIndex + 1, items.length - 1);
            updateSelection(items, selectedIndex);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            selectedIndex = Math.max(selectedIndex - 1, 0);
            updateSelection(items, selectedIndex);
        } else if (e.key === 'Enter' && selectedIndex >= 0 && items[selectedIndex]) {
            e.preventDefault();
            items[selectedIndex].click();
        }
    });
}

function updateSelection(items, index) {
    items.forEach((item, i) => {
        item.classList.toggle('search-selected', i === index);
    });
}

async function performSearch(query, resultsContainer, hint) {
    try {
        // NEW: Check workspace scope selection
        const scopeRadio = document.querySelector('input[name="search-scope"]:checked');
        const useWorkspace = scopeRadio?.value === 'workspace';

        const atoms = await searchAtoms(query, useWorkspace);
        resultsContainer.innerHTML = '';

        if (!atoms || atoms.length === 0) {
            hint.textContent = 'No results found';
            return;
        }

        hint.textContent = `${atoms.length} result${atoms.length > 1 ? 's' : ''}`;

        atoms.slice(0, 20).forEach(atom => {
            const item = document.createElement('div');
            item.className = 'search-result-item';

            const layerColors = {
                BUSINESS: 'var(--color-business)',
                ARCHITECTURE: 'var(--color-architecture)',
                IMPLEMENTATION: 'var(--color-implementation)',
            };

            item.innerHTML = `
                <div class="search-result-left">
                    <span class="search-result-type" style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}">${atom.type}</span>
                    <span class="search-result-name">${escapeHtml(atom.human_name || atom.id)}</span>
                    ${atom.project && atom.project !== state.workspace.activeProject 
                        ? `<span class="search-result-project" title="From project: ${atom.project}">${escapeHtml(atom.project)}</span>` 
                        : ''}
                </div>
                <div class="search-result-right">
                    <span class="search-result-layer">${atom.layer}</span>
                    <span class="search-result-status ${atom.status}">${atom.status}</span>
                </div>
            `;

            if (atom.intent) {
                const intentEl = document.createElement('div');
                intentEl.className = 'search-result-intent';
                intentEl.textContent = atom.intent;
                item.appendChild(intentEl);
            }

            item.addEventListener('click', () => {
                // Find full atom data from state
                const fullAtom = state.atoms.find(a => a.id === atom.id);
                if (fullAtom) {
                    setCurrentAtom(fullAtom);
                    emit('navigate-to-atom', fullAtom);
                    window.history.pushState({}, '', `?atom=${atom.id}`);
                }
                closeSearchOverlay();
            });

            resultsContainer.appendChild(item);
        });
    } catch (err) {
        hint.textContent = 'Search error: ' + err.message;
    }
}

function closeSearchOverlay() {
    if (overlay) {
        overlay.remove();
        overlay = null;
    }
}

function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.appendChild(document.createTextNode(text));
    return div.innerHTML;
}
