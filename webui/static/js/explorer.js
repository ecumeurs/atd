/**
 * Explorer Module — "Waterfall of Intent" three-column lane visualization.
 * @spec-link [[ui_webui_waterfall_explorer]]
 */
import { state, on, setCurrentAtom, getAtomMap, emit, toggleFoundation } from './state.js';

let waterfallContainer;
let foundationContainer;

const LANE_LAYERS = ['CUSTOMER', 'ARCHITECTURE', 'IMPLEMENTATION'];

// Health classification helpers
// @spec-link [[mechanic_webui_health_categorization]]
function classifyAtom(atom, atomMap) {
    const isStable = atom.status === 'STABLE';
    const hasImpl = atom.linked_codes && atom.linked_codes.length > 0;
    const hasTests = atom.has_tests;

    if (isStable && hasImpl && hasTests) return 'done';
    if (isStable && hasImpl && !hasTests) return 'almost-done';
    if (!hasImpl && (atom.layer === 'IMPLEMENTATION')) return 'needs-impl';
    if (atom.status === 'DRAFT') return 'draft';
    if (!isStable) return 'wip';
    return 'informational';
}

function getHealthBorderClass(health) {
    switch (health) {
        case 'done': return 'card-done';
        case 'almost-done': return 'card-almost-done';
        case 'needs-impl': return 'card-needs-impl';
        case 'draft': return 'card-draft';
        case 'wip': return 'card-wip';
        default: return 'card-info';
    }
}

export function initExplorer() {
    waterfallContainer = document.getElementById('waterfall-container');
    foundationContainer = document.getElementById('foundation-container');

    on('atoms-changed', renderWaterfall);
    on('foundation-toggled', renderWaterfall);
    on('navigate-to-atom', highlightAtomPath);

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            highlightAtomPath(null);
            setCurrentAtom(null);
            const closeBtn = document.getElementById('close-panel');
            if (closeBtn) closeBtn.click();
        }
    });

    // TODO: Re-enable connector lines when scroll-sync is implemented
    // @spec-link [[mechanic_webui_connector_lines]]
}

// @spec-link [[ui_webui_waterfall_explorer]]
export function renderWaterfall(atomsData) {
    const atoms = atomsData || state.atoms;
    if (!atoms || !waterfallContainer) return;

    const atomMap = getAtomMap();

    // Separate foundation atoms (done) from active atoms
    const foundationAtoms = [];
    const activeAtoms = [];

    atoms.forEach(atom => {
        const health = classifyAtom(atom, atomMap);
        if (health === 'done') {
            foundationAtoms.push(atom);
        } else {
            activeAtoms.push(atom);
        }
    });

    // Group active atoms by layer
    const lanes = {
        CUSTOMER: [],
        ARCHITECTURE: [],
        IMPLEMENTATION: [],
    };

    activeAtoms.forEach(atom => {
        const lane = lanes[atom.layer];
        if (lane) lane.push(atom);
        else lanes.IMPLEMENTATION.push(atom); // fallback
    });

    // Render
    waterfallContainer.innerHTML = '';

    // Create lane containers
    LANE_LAYERS.forEach(layer => {
        const laneEl = document.createElement('div');
        laneEl.className = `lane lane-${layer.toLowerCase()}`;
        laneEl.dataset.layer = layer;

        const header = document.createElement('div');
        header.className = 'lane-header';
        header.innerHTML = `
            <h3>${layer}</h3>
            <span class="lane-count">${lanes[layer].length}</span>
        `;
        laneEl.appendChild(header);

        const content = document.createElement('div');
        content.className = 'lane-content';

        // Sort atoms by type, then name
        const sortedAtoms = lanes[layer].sort((a, b) => {
            if (a.type !== b.type) {
                return (a.type || '').localeCompare(b.type || '');
            }
            const nameA = a.human_name || a.id || '';
            const nameB = b.human_name || b.id || '';
            return nameA.localeCompare(nameB);
        });

        // Render all atoms flat
        sortedAtoms.forEach(atom => {
            content.appendChild(createAtomCard(atom, atomMap));
        });

        if (lanes[layer].length === 0) {
            const empty = document.createElement('div');
            empty.className = 'lane-empty';
            empty.textContent = 'No atoms in this layer';
            content.appendChild(empty);
        }

        laneEl.appendChild(content);
        waterfallContainer.appendChild(laneEl);
    });

    // Render Foundation section
    renderFoundation(foundationAtoms);

    // Connector lines disabled — see mechanic_webui_connector_lines
}

function createAtomCard(atom, atomMap) {
    const health = classifyAtom(atom, atomMap);
    const card = document.createElement('div');
    card.className = `atom-card ${getHealthBorderClass(health)}`;
    card.dataset.atomId = atom.id;

    // Determine if vaporware (no impl linked)
    const isVaporware = !atom.linked_codes || atom.linked_codes.length === 0;
    if (isVaporware && atom.layer === 'IMPLEMENTATION') {
        card.classList.add('card-vaporware');
    }

    // Type badge
    const typeBadge = document.createElement('span');
    typeBadge.className = 'card-type-badge';
    typeBadge.textContent = atom.type;

    // Status pill
    const statusPill = document.createElement('span');
    statusPill.className = `card-status-pill ${atom.status || 'DRAFT'}`;
    statusPill.textContent = atom.status || 'DRAFT';

    // Name
    const name = document.createElement('div');
    name.className = 'card-name';
    name.textContent = atom.human_name || atom.id;
    name.title = atom.human_name || atom.id;

    // Intent preview
    const intent = document.createElement('div');
    intent.className = 'card-intent';
    const intentText = extractIntentFromContent(atom.content);
    intent.textContent = intentText || '';
    intent.title = intentText || '';

    // Coverage indicator
    const coverage = document.createElement('div');
    coverage.className = 'card-coverage';
    const hasImpl = atom.linked_codes && atom.linked_codes.length > 0;
    const hasTests = atom.has_tests;
    coverage.innerHTML = `
        <span class="cov-block ${hasImpl ? 'cov-filled' : ''}" title="Implementation"></span>
        <span class="cov-block ${hasTests ? 'cov-filled' : ''}" title="Tests"></span>
    `;

    // Header row
    const headerRow = document.createElement('div');
    headerRow.className = 'card-header-row';
    headerRow.appendChild(typeBadge);
    headerRow.appendChild(statusPill);
    headerRow.appendChild(coverage);

    card.appendChild(headerRow);
    card.appendChild(name);
    if (intentText) card.appendChild(intent);

    // Click handler
    card.addEventListener('click', () => {
        if (state.currentAtom && state.currentAtom.id === atom.id) {
            highlightAtomPath(null);
            setCurrentAtom(null);
            const closeBtn = document.getElementById('close-panel');
            if (closeBtn) closeBtn.click();
        } else {
            setCurrentAtom(atom);
            highlightAtomPath(atom);
        }
    });

    return card;
}

function extractIntentFromContent(content) {
    if (!content) return '';
    const lines = content.split('\n');
    let inIntent = false;
    for (const line of lines) {
        const trimmed = line.trim();
        if (trimmed.startsWith('## INTENT')) { inIntent = true; continue; }
        if (inIntent && trimmed.startsWith('## ')) break;
        if (inIntent && trimmed) return trimmed;
    }
    return '';
}

// Connector lines disabled — will be re-enabled with scroll-sync
// @spec-link [[mechanic_webui_connector_lines]]

function highlightAtomPath(atom) {
    if (!waterfallContainer) return;

    if (!atom) {
        // Clear all highlights
        waterfallContainer.querySelectorAll('.atom-card').forEach(c => {
            c.classList.remove('card-selected', 'card-ancestor', 'card-descendant', 'card-highlighted', 'card-dimmed');
        });
        return;
    }

    // Clear all highlights
    waterfallContainer.querySelectorAll('.atom-card').forEach(c => {
        c.classList.remove('card-selected', 'card-ancestor', 'card-descendant', 'card-highlighted', 'card-dimmed');
    });

    const atomMap = getAtomMap();
    const ancestorIds = new Set();
    const descendantIds = new Set();

    // Walk parents upward
    const walkUp = (id) => {
        const a = atomMap.get(id);
        if (!a || !a.parents) return;
        a.parents.forEach(pid => {
            if (!ancestorIds.has(pid)) {
                ancestorIds.add(pid);
                walkUp(pid);
            }
        });
    };
    walkUp(atom.id);

    // Walk dependents downward
    const walkDown = (id) => {
        const a = atomMap.get(id);
        if (!a || !a.dependents) return;
        a.dependents.forEach(did => {
            if (!descendantIds.has(did)) {
                descendantIds.add(did);
                walkDown(did);
            }
        });
    };
    walkDown(atom.id);

    // Apply highlights to cards only (connector lines disabled)
    waterfallContainer.querySelectorAll('.atom-card').forEach(c => {
        const id = c.dataset.atomId;
        if (id === atom.id) {
            c.classList.add('card-selected');
        } else if (ancestorIds.has(id)) {
            c.classList.add('card-ancestor');
        } else if (descendantIds.has(id)) {
            c.classList.add('card-descendant');
        } else {
            c.classList.add('card-dimmed');
        }
    });
}

function renderFoundation(foundationAtoms) {
    if (!foundationContainer) return;
    foundationContainer.innerHTML = '';

    const header = document.createElement('div');
    header.className = 'foundation-header';
    header.innerHTML = `
        <span class="foundation-icon">🏛️</span>
        <span class="foundation-title">Foundation</span>
        <span class="foundation-badge">${foundationAtoms.length} stable</span>
        <span class="foundation-toggle">${state.foundationCollapsed ? '▸' : '▾'}</span>
    `;
    header.addEventListener('click', () => {
        toggleFoundation();
    });
    foundationContainer.appendChild(header);

    if (!state.foundationCollapsed && foundationAtoms.length > 0) {
        const grid = document.createElement('div');
        grid.className = 'foundation-grid';

        foundationAtoms.forEach(atom => {
            const dot = document.createElement('div');
            dot.className = 'foundation-dot';
            dot.title = `${atom.human_name || atom.id} (${atom.type})`;
            dot.dataset.atomId = atom.id;
            dot.addEventListener('click', () => setCurrentAtom(atom));

            const label = document.createElement('span');
            label.className = 'foundation-dot-label';
            label.textContent = atom.human_name || atom.id;
            dot.appendChild(label);

            grid.appendChild(dot);
        });

        foundationContainer.appendChild(grid);
    }
}
