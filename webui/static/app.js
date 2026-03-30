document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const treemapContainer = document.getElementById('treemap-container');
    const treeviewContainer = document.getElementById('treeview-container');
    const panelEmpty = document.getElementById('details-empty');
    const panelContent = document.getElementById('details-content');
    const closePanelBtn = document.getElementById('close-panel');
    const refreshBtn = document.getElementById('refresh-btn');
    const btnSummarize = document.getElementById('btn-summarize');
    const summaryBox = document.getElementById('summary-box');

    // View Switcher Elements
    const viewTreemapBtn = document.getElementById('view-treemap');
    const viewTreeBtn = document.getElementById('view-tree');
    const toggleLayoutBtn = document.getElementById('toggle-layout');

    // Details DOM Elements
    const dType = document.getElementById('detail-type');
    const dStatus = document.getElementById('detail-status');
    const dTitle = document.getElementById('detail-title');
    const dId = document.getElementById('detail-id');
    const dMarkdown = document.getElementById('detail-markdown');
    const dCode = document.getElementById('detail-code');
    const dTestStatus = document.getElementById('detail-test-status');

    // Edit Form Elements
    const editForm = document.getElementById('edit-form');
    const btnEdit = document.getElementById('btn-edit');
    const btnSave = document.getElementById('btn-save');
    const btnCancelEdit = document.getElementById('btn-cancel-edit');
    const editHumanName = document.getElementById('edit-human-name');
    const editType = document.getElementById('edit-type');
    const editStatus = document.getElementById('edit-status');
    const editPriority = document.getElementById('edit-priority');
    const editTags = document.getElementById('edit-tags');
    const editContent = document.getElementById('edit-content');
    const detailActions = document.querySelector('.detail-actions');

    // Selection Mode Elements
    const toggleSelectModeBtn = document.getElementById('toggle-select-mode');
    const bulkBar = document.getElementById('bulk-bar');
    const selectedCountEl = document.getElementById('selected-count');
    const bulkStatusSelect = document.getElementById('bulk-status-select');
    const btnBulkApply = document.getElementById('btn-bulk-apply');
    const btnBulkCancel = document.getElementById('btn-bulk-cancel');

    // Info DOM Elements
    const infoDocs = document.getElementById('info-docs');
    const infoProject = document.getElementById('info-project');
    const infoCount = document.getElementById('info-count');

    let currentFlatData = null;
    let currentAtom = null;
    let isSelectionMode = false;
    let selectedIds = new Set();
    let currentViewScope = 'all'; // 'all', 'done', 'almost_done', 'wip', 'jungle'

    const colorScheme = {
        done: { main: '#2e7d32', light: '#4caf50', label: 'Done' },
        almost_done: { main: '#fbc02d', light: '#fdd835', label: 'Almost Done' },
        wip: { main: '#0288d1', light: '#03a9f4', label: 'WIP' },
        jungle: { main: '#546e7a', light: '#78909c', label: 'Doc Jungle' }
    };

    // Fetch initial data
    fetchData();

    refreshBtn.addEventListener('click', fetchData);

    viewTreemapBtn.addEventListener('click', () => switchView('treemap'));
    viewTreeBtn.addEventListener('click', () => switchView('tree'));

    toggleSelectModeBtn.addEventListener('click', toggleSelectMode);
    btnBulkCancel.addEventListener('click', toggleSelectMode);

    bulkStatusSelect.addEventListener('change', () => {
        btnBulkApply.disabled = !bulkStatusSelect.value || selectedIds.size === 0;
    });

    btnBulkApply.addEventListener('click', applyBulkUpdate);

    toggleLayoutBtn.addEventListener('click', () => {
        const detailsPanel = document.getElementById('details-panel');
        detailsPanel.classList.toggle('expanded');
        // Redraw treemap if needed to fit new size
        if (currentFlatData && viewTreemapBtn.classList.contains('active')) {
            renderTreemap(currentFlatData);
        }
    });

    btnEdit.addEventListener('click', enterEditMode);
    btnCancelEdit.addEventListener('click', exitEditMode);
    btnSave.addEventListener('click', saveAtomChanges);

    closePanelBtn.addEventListener('click', () => {
        panelContent.style.display = 'none';
        panelEmpty.style.display = 'flex';
        exitEditMode();
        // Clear active selection
        d3.selectAll('.node').style('opacity', 1);
        document.querySelectorAll('.tree-item').forEach(el => el.classList.remove('active'));
    });

    // @spec-link [[mechanic_atd_update]]
    function toggleSelectMode() {
        isSelectionMode = !isSelectionMode;
        document.body.classList.toggle('selection-mode', isSelectionMode);

        if (isSelectionMode) {
            toggleSelectModeBtn.textContent = 'Exit Select Mode';
            bulkBar.style.display = 'flex';
            // Force Tree View in select mode for now as treemap is harder to multi-select visually
            switchView('tree');
        } else {
            toggleSelectModeBtn.textContent = 'Select Mode';
            bulkBar.style.display = 'none';
            selectedIds.clear();
            updateBulkUI();
        }

        renderTreeView(currentFlatData);
    }

    function updateBulkUI() {
        selectedCountEl.textContent = selectedIds.size;
        btnBulkApply.disabled = selectedIds.size === 0 || !bulkStatusSelect.value;
    }

    // @spec-link [[mechanic_atd_update]]
    async function applyBulkUpdate() {
        const status = bulkStatusSelect.value;
        const ids = Array.from(selectedIds);

        btnBulkApply.disabled = true;
        btnBulkApply.textContent = 'Updating...';

        try {
            const resp = await fetch('/api/bulk-update', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ ids, status })
            });

            if (resp.ok) {
                toggleSelectMode(); // Exit selection mode on success
                await fetchData();  // Refresh data
            } else {
                const err = await resp.json();
                alert('Update failed: ' + (err.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Bulk update error:', error);
            alert('Failed to send bulk update request.');
        } finally {
            btnBulkApply.textContent = 'Apply Change';
            updateBulkUI();
        }
    }

    function switchView(view) {
        if (view === 'treemap') {
            viewTreemapBtn.classList.add('active');
            viewTreeBtn.classList.remove('active');
            treemapContainer.style.display = 'block';
            treeviewContainer.style.display = 'none';
            // Render breadcrumb or back button if needed
            renderBreadcrumb();
        } else {
            viewTreemapBtn.classList.remove('active');
            viewTreeBtn.classList.add('active');
            treemapContainer.style.display = 'none';
            treeviewContainer.style.display = 'block';
            // Hide breadcrumbs in tree view for now
            const bc = document.getElementById('breadcrumb-container');
            if (bc) bc.style.display = 'none';
        }
    }

    // @spec-link [[mechanic_webui_breadcrumb_navigation]]
    function renderBreadcrumb() {
        let bcContainer = document.getElementById('breadcrumb-container');
        if (!bcContainer) {
            bcContainer = document.createElement('div');
            bcContainer.id = 'breadcrumb-container';
            bcContainer.className = 'breadcrumb-container';
            treemapContainer.parentNode.insertBefore(bcContainer, treemapContainer);
        }

        if (currentViewScope === 'all') {
            bcContainer.style.display = 'none';
            return;
        }

        bcContainer.style.display = 'flex';
        bcContainer.innerHTML = '';

        const rootLink = document.createElement('span');
        rootLink.className = 'breadcrumb-link';
        rootLink.textContent = 'All Atoms';
        rootLink.onclick = () => {
            currentViewScope = 'all';
            renderTreemap(currentFlatData);
            renderBreadcrumb();
        };

        const separator = document.createElement('span');
        separator.className = 'breadcrumb-separator';
        separator.textContent = ' > ';

        const currentBucket = document.createElement('span');
        currentBucket.className = 'breadcrumb-current';
        currentBucket.textContent = colorScheme[currentViewScope].label;

        bcContainer.appendChild(rootLink);
        bcContainer.appendChild(separator);
        bcContainer.appendChild(currentBucket);
    }

    // @spec-link [[mechanic_webui_explorer_workflow]]
    async function fetchData() {
        try {
            const [treeResp, infoResp] = await Promise.all([
                fetch('/api/tree'),
                fetch('/api/info')
            ]);

            if (infoResp.ok) {
                const infoData = await infoResp.json();
                infoDocs.textContent = infoData.atd_path || 'N/A';
                infoProject.textContent = infoData.project_path || 'N/A';
                infoCount.textContent = infoData.atd_count || '0';
            }

            const data = await treeResp.json();
            if (data && data.length > 0) {
                currentFlatData = data;
                renderTreemap(data);
                renderTreeView(data);

                // Handle deep-linking via ?atom=id
                const params = new URLSearchParams(window.location.search);
                const atomId = params.get('atom');
                if (atomId) {
                    const atom = data.find(a => a.id === atomId);
                    if (atom) {
                        // Switch to explorer tab if not already there
                        const explorerTabBtn = document.querySelector('.tab-btn[data-tab="explorer"]');
                        if (explorerTabBtn) explorerTabBtn.click();
                        
                        // Show details after a small delay to ensure UI is ready
                        setTimeout(() => showDetails(atom), 50);
                    }
                }
            } else {
                treemapContainer.innerHTML = '<div style="padding: 24px; color: #9aa0a6;">No ATD data found. Try refreshing or check config.</div>';
                treeviewContainer.innerHTML = '<div style="padding: 24px; color: #9aa0a6;">No ATD data found.</div>';
            }
        } catch (error) {
            console.error("Failed to fetch tree data", error);
            treemapContainer.innerHTML = '<div style="padding: 24px; color: #d32f2f;">Error connecting to API.</div>';
            treeviewContainer.innerHTML = '<div style="padding: 24px; color: #d32f2f;">Error connecting to API.</div>';
        }
    }

    // @spec-link [[mechanic_webui_health_categorization]]
    function categorizeAtoms(flatData) {
        const atomMap = new Map();
        flatData.forEach(a => atomMap.set(a.id, a));

        const hasAncestorType = (atom, targetType) => {
            const visited = new Set();
            const queue = [...(atom.parents || [])];
            while (queue.length > 0) {
                const parentId = queue.shift();
                if (visited.has(parentId)) continue;
                visited.add(parentId);
                const parent = atomMap.get(parentId);
                if (parent) {
                    if (parent.layer === targetType) return true;
                    queue.push(...(parent.parents || []));
                }
            }
            return false;
        };

        const buckets = {
            done: [],
            almost_done: [],
            wip: [],
            jungle: []
        };

        flatData.forEach(atom => {
            const isStable = atom.status === 'STABLE';
            const hasImpl = atom.linked_codes && atom.linked_codes.length > 0;
            const hasTests = atom.has_tests;

            if (isStable && hasImpl && hasTests) {
                buckets.done.push(atom);
            } else if (isStable && hasImpl && !hasTests) {
                buckets.almost_done.push(atom);
            } else if (isStable && !hasImpl) {
                // WIP: Full doc (implementation ATD has Arch + Customer ancestor, but no impl, and all must be STABLE)
                // Note: The "all must be STABLE" part is complex to verify here, 
                // so we focus on the ancestors' layer and current atom being STABLE.
                const hasArchAncestor = hasAncestorType(atom, 'ARCHITECTURE');
                const hasCustomerAncestor = hasAncestorType(atom, 'CUSTOMER');

                if (hasArchAncestor && hasCustomerAncestor) {
                    buckets.wip.push(atom);
                } else {
                    buckets.jungle.push(atom);
                }
            } else {
                buckets.jungle.push(atom);
            }
        });

        return buckets;
    }

    // @spec-link [[rule_webui_blob_sizing]]
    function buildHierarchy(flatData, useCategories = true) {
        if (!useCategories) {
            // Build simple hierarchy for within-blob view
            const root = { id: "root", name: "Cluster", children: [] };
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

        // Build 2nd level hierarchy for Category view
        const buckets = categorizeAtoms(flatData);
        let root = { id: "root", name: "Health Map", children: [] };

        Object.keys(buckets).forEach(key => {
            if (buckets[key].length > 0 || true) { // Show empty buckets? 
                root.children.push({
                    id: key,
                    name: colorScheme[key].label,
                    type: 'CATEGORY',
                    children: buckets[key].map(a => ({ ...a, value: 1 }))
                    // To maintain internal hierarchy within buckets, call buildHierarchy recursively:
                    // children: buildHierarchy(buckets[key], false).children
                });
            }
        });

        return root;
    }

    // @spec-link [[ui_webui_explorer_layout]]
    // @spec-link [[ui_webui_traceability_explorer]]
    // @spec-link [[mechanic_webui_explorer_workflow]]
    function renderTreemap(rawData) {
        treemapContainer.innerHTML = '';

        let hierarchicalData;
        if (currentViewScope === 'all') {
            hierarchicalData = buildHierarchy(rawData, true);
        } else {
            const buckets = categorizeAtoms(rawData);
            hierarchicalData = buildHierarchy(buckets[currentViewScope], false);
        }

        const width = treemapContainer.clientWidth;
        const height = treemapContainer.clientHeight;

        const root = d3.hierarchy(hierarchicalData)
            .sum(d => d.value || 0)
            .sort((a, b) => b.value - a.value);

        d3.treemap()
            .size([width, height])
            .paddingTop(d => d.depth === 1 ? 40 : 10)
            .paddingRight(4)
            .paddingBottom(4)
            .paddingLeft(4)
            .paddingInner(4)
            .round(true)
            (root);

        const svg = d3.select('#treemap-container')
            .append('svg')
            .attr('width', width)
            .attr('height', height)
            .style('border-radius', '12px');

        const nodes = svg.selectAll('g')
            .data(currentViewScope === 'all' ? root.descendants().filter(d => d.depth > 0) : root.leaves())
            .join('g')
            .attr('transform', d => `translate(${d.x0},${d.y0})`)
            .on('click', (event, d) => {
                event.stopPropagation();
                if (d.data.type === 'CATEGORY') {
                    currentViewScope = d.data.id;
                    renderTreemap(rawData);
                    renderBreadcrumb();
                } else {
                    showDetails(d.data);
                }
            });

        nodes.append('rect')
            .attr('width', d => d.x1 - d.x0)
            .attr('height', d => d.y1 - d.y0)
            .attr('class', d => `node-rect status-${d.data.computed_color || 'grey'} ${d.data.type === 'CATEGORY' ? 'category-blob' : ''}`)
            .attr('fill', d => {
                if (d.data.type === 'CATEGORY') return colorScheme[d.data.id].main;
                // Atoms in category view might need generic coloring or specific
                return getCategoryColor(d.data);
            })
            .attr('rx', 8)
            .style('cursor', 'pointer');

        nodes.append('text')
            .attr('x', 12)
            .attr('y', 24)
            .attr('class', d => d.data.type === 'CATEGORY' ? 'label-category' : 'label-atom')
            .text(d => d.data.human_name || d.data.id)
            .attr('fill', d => d.data.type === 'CATEGORY' || d.data.computed_color === 'yellow' ? '#fff' : '#fff');

        if (currentViewScope === 'all') {
            nodes.filter(d => d.data.type === 'CATEGORY')
                .append('text')
                .attr('x', 12)
                .attr('y', 45)
                .attr('class', 'label-count')
                .text(d => `${d.data.children.length} atoms`)
                .attr('fill', 'rgba(255,255,255,0.7)');
        }
    }

    function getCategoryColor(atom) {
        // Fallback or specific logic
        const buckets = categorizeAtoms(currentFlatData);
        if (buckets.done.some(a => a.id === atom.id)) return colorScheme.done.main;
        if (buckets.almost_done.some(a => a.id === atom.id)) return colorScheme.almost_done.main;
        if (buckets.wip.some(a => a.id === atom.id)) return colorScheme.wip.main;
        return colorScheme.jungle.main;
    }

    // @spec-link [[ui_webui_traceability_explorer]]
    function renderTreeView(rawData) {
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

        if (isSelectionMode && !isRoot) {
            const cb = document.createElement('input');
            cb.type = 'checkbox';
            cb.className = 'tree-checkbox';
            cb.checked = selectedIds.has(node.id);
            cb.addEventListener('click', (e) => e.stopPropagation());
            cb.addEventListener('change', (e) => {
                if (cb.checked) selectedIds.add(node.id);
                else selectedIds.delete(node.id);
                updateBulkUI();
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
            if (isSelectionMode && !isRoot) {
                const cb = item.querySelector('.tree-checkbox');
                cb.checked = !cb.checked;
                cb.dispatchEvent(new Event('change'));
            } else {
                showDetails(node);
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

    // @spec-link [[ui_webui_traceability_explorer]]
    function showDetails(atom) {
        if (!atom || atom.id === "root") return;

        // Visual selection in Treemap
        d3.selectAll('.node').style('opacity', 0.4);
        d3.selectAll('.node').filter(d => d.data.id === atom.id)
            .style('opacity', 1);

        // Visual selection in TreeView
        document.querySelectorAll('.tree-item').forEach(el => {
            el.classList.remove('active');
            if (el.dataset.id === atom.id) el.classList.add('active');
        });

        panelEmpty.style.display = 'none';
        panelContent.style.display = 'block';
        currentAtom = atom;
        exitEditMode();

        dType.textContent = atom.type;
        dTitle.textContent = atom.human_name || atom.id;
        dId.textContent = `@${atom.id}`;

        // Status Badge
        dStatus.textContent = atom.computed_status || "UNKNOWN";

        switch (atom.computed_color) {
            case 'green':
                dStatus.style.backgroundColor = 'rgba(76, 175, 80, 0.2)';
                dStatus.style.color = 'var(--color-green-light)';
                break;
            case 'yellow':
                dStatus.style.backgroundColor = 'rgba(253, 216, 53, 0.2)';
                dStatus.style.color = 'var(--color-yellow-light)';
                break;
            case 'red':
                dStatus.style.backgroundColor = 'rgba(211, 47, 47, 0.2)';
                dStatus.style.color = 'var(--color-red-light)';
                break;
            default:
                dStatus.style.backgroundColor = 'rgba(120, 144, 156, 0.2)';
                dStatus.style.color = 'var(--text-muted)';
        }

        // Markdown content
        dMarkdown.innerHTML = marked.parse(atom.content || '');

        // Code Links
        dCode.innerHTML = '';
        if (atom.linked_codes && atom.linked_codes.length > 0) {
            atom.linked_codes.forEach(link => {
                const li = document.createElement('li');
                li.textContent = link;
                dCode.appendChild(li);
            });
        } else {
            dCode.innerHTML = '<li style="color:var(--text-muted);border:none;background:transparent;">No code references found.</li>';
        }

        // Testing
        if (atom.has_tests) {
            dTestStatus.innerHTML = '<span class="test-pass">●</span> Tests detected on implementation.';
        } else if (atom.linked_codes && atom.linked_codes.length > 0) {
            dTestStatus.innerHTML = '<span class="test-missing">●</span> Implementation found, but testing missing.';
        } else {
            dTestStatus.innerHTML = '<span style="color:var(--text-muted)">No implementation to test.</span>';
        }

        // Summarize logic
        summaryBox.style.display = 'none';
        btnSummarize.onclick = async () => {
            summaryBox.style.display = 'block';
            summaryBox.innerHTML = '<i>Generating summary via local LLM...</i>';
            try {
                const res = await fetch(`/api/summary/${atom.id}`);
                const data = await res.json();
                summaryBox.innerHTML = data.summary;
            } catch (err) {
                summaryBox.innerHTML = '<span style="color:var(--color-red)">Failed to reach Ollama endpoint.</span>';
            }
        };
    }

    function enterEditMode() {
        if (!currentAtom) return;

        editHumanName.value = currentAtom.human_name || '';
        editType.value = currentAtom.type || 'MECHANIC';
        editStatus.value = currentAtom.status || 'DRAFT';
        editPriority.value = currentAtom.priority || 'CORE';
        editTags.value = (currentAtom.tags || []).join(', ');
        editContent.value = currentAtom.content || '';

        editForm.style.display = 'block';
        detailActions.style.display = 'none';
        document.querySelectorAll('.detail-section').forEach(el => el.style.display = 'none');
    }

    function slugify(text) {
        return text.toString().toLowerCase()
            .replace(/\s+/g, '_')           // Replace spaces with _
            .replace(/[^\w-]+/g, '')       // Remove all non-word chars
            .replace(/--+/g, '_')           // Replace multiple - with single _
            .replace(/^-+/, '')             // Trim - from start of text
            .replace(/-+$/, '');            // Trim - from end of text
    }

    function exitEditMode() {
        editForm.style.display = 'none';
        detailActions.style.display = 'flex';
        document.querySelectorAll('.detail-section').forEach(el => el.style.display = 'block');
    }

    // @spec-link [[mechanic_atd_update]]
    async function saveAtomChanges() {
        if (!currentAtom) return;

        const newType = editType.value;
        const newHumanName = editHumanName.value;
        const newId = `${newType.toLowerCase()}_${slugify(newHumanName)}`;

        const updatedData = {
            id: newId,
            human_name: newHumanName,
            type: newType,
            status: editStatus.value,
            priority: editPriority.value,
            tags: editTags.value.split(',').map(t => t.trim()).filter(t => t !== ''),
            content: editContent.value
        };

        btnSave.disabled = true;
        btnSave.textContent = 'Saving...';

        try {
            const resp = await fetch(`/api/atd/${currentAtom.id}/update`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(updatedData)
            });

            if (resp.ok) {
                await fetchData();
                // Find the updated atom in the new data to refresh details
                const updatedAtom = currentFlatData.find(a => a.id === currentAtom.id);
                if (updatedAtom) showDetails(updatedAtom);
                exitEditMode();
            } else {
                const err = await resp.json();
                alert('Save failed: ' + (err.error || 'Unknown error'));
            }
        } catch (error) {
            console.error('Save error:', error);
            alert('Failed to send update request.');
        } finally {
            btnSave.disabled = false;
            btnSave.textContent = 'Save Changes';
        }
    }
    // @spec-link [[module_webui]]
    // --- Tab Switching ---
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const tab = btn.dataset.tab;
            // Update tab buttons
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            // Update tab content
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            const target = document.getElementById(`content-${tab}`);
            if (target) target.classList.add('active');
            // Show/hide explorer-only header tools
            const explorerTools = document.querySelector('.header-tools');
            if (explorerTools) {
                explorerTools.style.display = tab === 'explorer' ? 'flex' : 'none';
            }
            // Fix: Redraw treemap when switching back to explorer to avoid black screen
            if (tab === 'explorer' && currentFlatData) {
                // Use setTimeout to ensure container is visible before measuring
                setTimeout(() => renderTreemap(currentFlatData), 10);
            }

            // Step 09: Health check on Spec Builder entry
            if (tab === 'spec-builder' && window.SpecBuilder) {
                window.SpecBuilder.checkBackendHealth();
            }
        });
    });
});
