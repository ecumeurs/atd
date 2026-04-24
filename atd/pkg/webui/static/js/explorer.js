/**
 * Explorer Module — "Waterfall of Intent" three-column lane visualization.
 * @spec-link [[ui_webui_waterfall_explorer]]
 */
import { state, on, setCurrentAtom, getAtomMap, emit, toggleFoundation } from './state.js';
import { fetchHeatMap } from './api.js';

let waterfallContainer;
let foundationContainer;
let heatMapData = null;
let activeHeatLayer = 'none';

const LANE_LAYERS = ['CUSTOMER', 'ARCHITECTURE', 'IMPLEMENTATION'];

// Heat state mapping to classes
const HEAT_CLASSES = {
    'cold': 'heat-cold',
    'optimal': 'heat-optimal',
    'warm': 'heat-warm',
    'hot': 'heat-hot',
    'stable': 'heat-stable'
};

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

    initHeatMapToggles();
}

function initHeatMapToggles() {
    const container = document.getElementById('heatmap-toggles');
    if (!container) return;

    container.addEventListener('click', async (e) => {
        const btn = e.target.closest('button');
        if (!btn) return;

        const layer = btn.dataset.heat;
        if (layer === activeHeatLayer) return;

        // Update active class
        container.querySelectorAll('button').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');

        activeHeatLayer = layer;

        if (layer !== 'none' && !heatMapData) {
            try {
                heatMapData = await fetchHeatMap();
            } catch (err) {
                console.error('Failed to fetch heatmap:', err);
                activeHeatLayer = 'none';
                container.querySelector('[data-heat="none"]').classList.add('active');
                btn.classList.remove('active');
                return;
            }
        }

        renderWaterfall();
    });
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
        // Customer atoms stay in their lane even if 'done' (ISS-087)
        if (health === 'done' && atom.layer !== 'CUSTOMER') {
            foundationAtoms.push(atom);
        } else {
            activeAtoms.push(atom);
            // Also keep them in foundation for the count/dots if they are done
            if (health === 'done') foundationAtoms.push(atom);
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

    // Special style for stable Customer atoms in the lane (ISS-087)
    if (health === 'done' && atom.layer === 'CUSTOMER') {
        card.classList.add('card-stable-lane');
    }

    // Determine if vaporware (no impl linked)
    const isVaporware = !atom.linked_codes || atom.linked_codes.length === 0;
    if (isVaporware && atom.layer === 'IMPLEMENTATION') {
        card.classList.add('card-vaporware');
    }

    // Heat Map Integration
    if (activeHeatLayer !== 'none' && heatMapData && heatMapData[atom.id]) {
        const heatResult = heatMapData[atom.id];
        let heatState = 'optimal';
        let badge = '';

        switch (activeHeatLayer) {
            case 'dependency':
                heatState = heatResult.dependency_state;
                if (heatState === 'hot') badge = '🔴';
                else if (heatState === 'warm') badge = '⚡';
                break;
            case 'code':
                heatState = heatResult.code_state;
                if (heatState === 'hot') badge = '🔴';
                else if (heatState === 'warm') badge = '⚡';
                break;
            case 'updates':
                heatState = heatResult.update_state;
                if (heatState === 'hot') badge = '🔴';
                else if (heatState === 'warm') badge = '⚡';
                break;
        }

        if (heatState) {
            card.classList.add(HEAT_CLASSES[heatState] || 'heat-optimal');
            if (badge) {
                const badgeEl = document.createElement('div');
                badgeEl.className = 'heat-badge';
                badgeEl.textContent = badge;
                card.appendChild(badgeEl);
            }

            const metrics = heatResult.metrics;
            const tooltip = document.createElement('div');
            tooltip.className = 'heat-metrics-tooltip';

            if (activeHeatLayer === 'dependency') {
                tooltip.innerHTML = `
                    <div class="heat-metric-item"><span>Parents:</span> <span class="heat-metric-value">${metrics.parents}</span></div>
                    <div class="heat-metric-item"><span>Dependents:</span> <span class="heat-metric-value">${metrics.dependents}</span></div>
                `;
            } else if (activeHeatLayer === 'code') {
                tooltip.innerHTML = `
                    <div class="heat-metric-item"><span>Files:</span> <span class="heat-metric-value">${metrics.codeFilesLinked}</span></div>
                `;
            } else if (activeHeatLayer === 'updates') {
                tooltip.innerHTML = `
                    <div class="heat-metric-item"><span>Commits:</span> <span class="heat-metric-value">${metrics.recentUpdates}</span></div>
                    <div class="heat-metric-item"><span>Last:</span> <span class="heat-metric-value">${metrics.lastUpdated || 'N/A'}</span></div>
                `;
            }
            card.appendChild(tooltip);
        }
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
    const intentText = atom.intent || '';
    intent.textContent = intentText;
    intent.title = intentText;

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
        // Clear all highlights and reset order
        waterfallContainer.querySelectorAll('.atom-card').forEach(c => {
            c.classList.remove('card-selected', 'card-ancestor', 'card-descendant', 'card-highlighted', 'card-dimmed');
            c.style.order = '';
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
    let selectedEl = null;

    waterfallContainer.querySelectorAll('.atom-card').forEach(c => {
        const id = c.dataset.atomId;
        c.classList.remove('card-selected', 'card-ancestor', 'card-descendant', 'card-highlighted', 'card-dimmed');
        c.style.order = ''; // Reset order

        if (id === atom.id) {
            c.classList.add('card-selected');
            c.style.order = '-10'; // Move to top of lane (ISS-088)
            selectedEl = c;
        } else if (ancestorIds.has(id)) {
            c.classList.add('card-ancestor');
            c.style.order = '-5'; // Near top (ISS-088)
        } else if (descendantIds.has(id)) {
            c.classList.add('card-descendant');
            c.style.order = '-1'; // Below ancestors (ISS-088)
        } else {
            c.classList.add('card-dimmed');
        }
    });

    // Scroll all lanes to top to show bubbled items (ISS-088)
    // We use requestAnimationFrame to ensure the 'order' change has been processed by the browser
    requestAnimationFrame(() => {
        const laneContents = waterfallContainer.querySelectorAll('.lane-content');
        laneContents.forEach(lane => {
            lane.scrollTo({ top: 0, behavior: 'smooth' });
        });
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
