/**
 * App Entry Point — Initializes all modules, handles tabs and deep-linking.
 * @spec-link [[module_webui]]
 */
import { state, on, setAtoms, setCurrentAtom, setView, toggleSelectionMode } from './state.js';
import { fetchTree, fetchInfo } from './api.js';
import { initExplorer, renderWaterfall } from './explorer.js';
import { initDetails } from './details.js';
import { initTreeView, renderTreeView } from './treeview.js';
import { initSearch } from './search.js';
import { initDocuments } from './documents.js';

document.addEventListener('DOMContentLoaded', () => {
    // DOM elements
    const infoDocs = document.getElementById('info-docs');
    const infoProject = document.getElementById('info-project');
    const infoCount = document.getElementById('info-count');
    const refreshBtn = document.getElementById('refresh-btn');
    const toggleLayoutBtn = document.getElementById('toggle-layout');
    const toggleSelectModeBtn = document.getElementById('toggle-select-mode');
    const viewWaterfallBtn = document.getElementById('view-waterfall');
    const viewTreeBtn = document.getElementById('view-tree');
    const waterfallContainer = document.getElementById('waterfall-container');
    const treeviewContainer = document.getElementById('treeview-container');

    // Initialize modules
    initExplorer();
    initDetails();
    initTreeView();
    initSearch();
    initDocuments();

    // Load Spec Builder lazily when its tab is opened
    let specBuilderLoaded = false;

    // --- Tab Switching ---
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const tab = btn.dataset.tab;
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            const target = document.getElementById(`content-${tab}`);
            if (target) target.classList.add('active');

            // Show/hide explorer-only header tools
            const explorerTools = document.querySelector('.header-tools');
            if (explorerTools) {
                explorerTools.style.display = tab === 'explorer' ? 'flex' : 'none';
            }

            // Redraw waterfall when switching back
            if (tab === 'explorer' && state.atoms.length > 0) {
                setTimeout(() => {
                    if (state.currentView === 'waterfall') renderWaterfall();
                    else renderTreeView(state.atoms);
                }, 10);
            }

            // Load spec builder on first switch
            if (tab === 'spec-builder' && !specBuilderLoaded) {
                import('./spec-builder.js').then(mod => {
                    if (mod.init) mod.init();
                    specBuilderLoaded = true;
                });
            }

            if (tab === 'spec-builder' && window.SpecBuilder) {
                window.SpecBuilder.checkBackendHealth();
            }
        });
    });

    // --- View Switching ---
    if (viewWaterfallBtn) {
        viewWaterfallBtn.addEventListener('click', () => {
            setView('waterfall');
            viewWaterfallBtn.classList.add('active');
            viewTreeBtn.classList.remove('active');
            waterfallContainer.style.display = 'flex';
            treeviewContainer.style.display = 'none';
            renderWaterfall();
        });
    }

    if (viewTreeBtn) {
        viewTreeBtn.addEventListener('click', () => {
            setView('tree');
            viewTreeBtn.classList.add('active');
            viewWaterfallBtn.classList.remove('active');
            waterfallContainer.style.display = 'none';
            treeviewContainer.style.display = 'block';
            renderTreeView(state.atoms);
        });
    }

    // --- Header Buttons ---
    refreshBtn.addEventListener('click', loadData);

    if (toggleLayoutBtn) {
        toggleLayoutBtn.addEventListener('click', () => {
            const detailsPanel = document.getElementById('details-panel');
            detailsPanel.classList.toggle('expanded');
            if (state.currentView === 'waterfall') {
                setTimeout(() => renderWaterfall(), 300);
            }
        });
    }

    if (toggleSelectModeBtn) {
        toggleSelectModeBtn.addEventListener('click', () => {
            toggleSelectionMode();
            toggleSelectModeBtn.textContent = state.isSelectionMode ? 'Exit Select Mode' : 'Select Mode';
            if (state.isSelectionMode) {
                // Force tree view in selection mode
                viewTreeBtn.click();
            }
        });
    }

    // --- Data Refresh Events ---
    on('data-refresh-needed', loadData);

    // --- Initial Load ---
    loadData();

    // @spec-link [[mechanic_webui_explorer_workflow]]
    async function loadData() {
        try {
            const [treeData, infoData] = await Promise.all([
                fetchTree(),
                fetchInfo(),
            ]);

            if (infoData) {
                infoDocs.textContent = infoData.atd_path || 'N/A';
                infoProject.textContent = infoData.project_path || 'N/A';
                infoCount.textContent = infoData.atd_count || '0';
            }

            if (treeData && treeData.length > 0) {
                setAtoms(treeData);

                // Deep-link: ?atom=id
                const params = new URLSearchParams(window.location.search);
                const atomId = params.get('atom');
                if (atomId) {
                    const atom = treeData.find(a => a.id === atomId);
                    if (atom) {
                        setTimeout(() => setCurrentAtom(atom), 50);
                    }
                }
            } else {
                waterfallContainer.innerHTML = '<div class="explorer-empty">No ATD data found. Try refreshing or check config.</div>';
                treeviewContainer.innerHTML = '<div class="explorer-empty">No ATD data found.</div>';
            }
        } catch (error) {
            console.error('Failed to load data:', error);
            waterfallContainer.innerHTML = '<div class="explorer-error">Error connecting to API.</div>';
        }
    }
});
