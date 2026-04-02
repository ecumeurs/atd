/**
 * Search Overlay Module — Ctrl+K command palette for atom lookup.
 * @spec-link [[ui_webui_search_overlay]]
 */
import { state, setCurrentAtom, emit } from './state.js';
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

    const results = document.createElement('div');
    results.className = 'search-results';

    const hint = document.createElement('div');
    hint.className = 'search-hint';
    hint.textContent = 'Type to search across all atoms';

    panel.appendChild(input);
    panel.appendChild(results);
    panel.appendChild(hint);
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
            return;
        }
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
        const atoms = await searchAtoms(query);
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
                CUSTOMER: 'var(--color-customer)',
                ARCHITECTURE: 'var(--color-architecture)',
                IMPLEMENTATION: 'var(--color-implementation)',
            };

            item.innerHTML = `
                <div class="search-result-left">
                    <span class="search-result-type" style="color: ${layerColors[atom.layer] || 'var(--text-muted)'}">${atom.type}</span>
                    <span class="search-result-name">${escapeHtml(atom.human_name || atom.id)}</span>
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
