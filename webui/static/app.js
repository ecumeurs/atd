document.addEventListener('DOMContentLoaded', () => {
    // DOM Elements
    const treemapContainer = document.getElementById('treemap-container');
    const panelEmpty = document.getElementById('details-empty');
    const panelContent = document.getElementById('details-content');
    const closePanelBtn = document.getElementById('close-panel');
    const refreshBtn = document.getElementById('refresh-btn');
    const btnSummarize = document.getElementById('btn-summarize');
    const summaryBox = document.getElementById('summary-box');

    // Details DOM Elements
    const dType = document.getElementById('detail-type');
    const dStatus = document.getElementById('detail-status');
    const dTitle = document.getElementById('detail-title');
    const dId = document.getElementById('detail-id');
    const dMarkdown = document.getElementById('detail-markdown');
    const dCode = document.getElementById('detail-code');
    const dTestStatus = document.getElementById('detail-test-status');

    // Info DOM Elements
    const infoDocs = document.getElementById('info-docs');
    const infoProject = document.getElementById('info-project');
    const infoCount = document.getElementById('info-count');

    let currentTreeData = null;

    // Fetch initial data
    fetchData();

    refreshBtn.addEventListener('click', fetchData);

    closePanelBtn.addEventListener('click', () => {
        panelContent.style.display = 'none';
        panelEmpty.style.display = 'flex';
        // Clear active selection in treemap
        d3.selectAll('.node').style('opacity', 1);
    });

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
                renderTreemap(data);
            } else {
                treemapContainer.innerHTML = '<div style="padding: 24px; color: #9aa0a6;">No ATD data found. Try refreshing or check config.</div>';
            }
        } catch (error) {
            console.error("Failed to fetch tree data", error);
            treemapContainer.innerHTML = '<div style="padding: 24px; color: #d32f2f;">Error connecting to API.</div>';
        }
    }

    function buildHierarchy(flatData) {
        // D3 Treemap requires a single root usually.
        // If data has multiple roots, create a virtual root.

        let root = { id: "root", name: "Project Chimera", children: [] };

        // Map elements by ID for quick access
        const map = new Map();
        flatData.forEach(node => {
            map.set(node.id, { ...node, children: [], value: 1 }); // Give each block a nominal size
        });

        // Build tree based on parents array
        map.forEach(node => {
            if (node.parents && node.parents.length > 0) {
                // Try to attach to first parent found
                let parentId = node.parents[0];
                let parentNode = map.get(parentId);
                if (parentNode) {
                    parentNode.children.push(node);
                } else {
                    // Parent not found in map, attach to root
                    root.children.push(node);
                }
            } else {
                // No parents, attach to root
                root.children.push(node);
            }
        });

        return root;
    }

    function renderTreemap(rawData) {
        treemapContainer.innerHTML = ''; // Clear previous

        const hierarchicalData = buildHierarchy(rawData);

        const width = treemapContainer.clientWidth;
        const height = treemapContainer.clientHeight;

        // Create root node
        const root = d3.hierarchy(hierarchicalData)
            .sum(d => {
                // Add value based on whether it has children or is a leaf
                // To give decent sizes, we give everything a base value
                return d.children && d.children.length > 0 ? 0 : 1;
            })
            // Distribute sizing nicely
            .sort((a, b) => b.value - a.value);

        // Apply Treemap layout
        d3.treemap()
            .size([width, height])
            .paddingTop(20) // Provide room for headers of grouped blocks
            .paddingRight(4)
            .paddingBottom(4)
            .paddingLeft(4)
            .paddingInner(4)
            .round(true)
            (root);

        // Append div containers
        const nodes = d3.select('#treemap-container')
            .selectAll('.node')
            .data(root.leaves())
            .join('div')
            .attr('class', d => `node ${getStatusClass(d.data)}`)
            .style('left', d => `${d.x0}px`)
            .style('top', d => `${d.y0}px`)
            .style('width', d => `${d.x1 - d.x0}px`)
            .style('height', d => `${d.y1 - d.y0}px`)
            .on('click', (event, d) => showDetails(d.data, nodes));

        // Add node contents
        nodes.append('div')
            .attr('class', 'node-title')
            .text(d => d.data.human_name || d.data.id);

        nodes.append('div')
            .attr('class', 'node-meta')
            .text(d => d.data.type);

        // Handle window resize dynamically later if needed
    }

    function getStatusClass(atom) {
        if (!atom.id || atom.id === 'root') return 'status-grey';

        if (atom.is_green) {
            return 'status-green';
        } else if (atom.has_tests || (atom.linked_codes && atom.linked_codes.length > 0)) {
            // Has implementation but not fully green (missing spec parent or tests)
            return 'status-yellow';
        } else {
            // No implementation
            return 'status-grey';
        }
    }

    function showDetails(atom, allNodes) {
        if (atom.id === "root") return;

        // Visual selection
        allNodes.style('opacity', 0.4);
        d3.selectAll('.node').filter(d => d.data.id === atom.id)
            .style('opacity', 1);

        panelEmpty.style.display = 'none';
        panelContent.style.display = 'block';

        dType.textContent = atom.type;
        dTitle.textContent = atom.human_name || atom.id;
        dId.textContent = `@${atom.id}`;

        // Status Badge
        if (atom.is_green) {
            dStatus.textContent = "DONE";
            dStatus.style.backgroundColor = 'rgba(76, 175, 80, 0.2)';
            dStatus.style.color = 'var(--color-green-light)';
        } else if (atom.has_tests || (atom.linked_codes && atom.linked_codes.length > 0)) {
            dStatus.textContent = "IN PROGRESS";
            dStatus.style.backgroundColor = 'rgba(253, 216, 53, 0.2)';
            dStatus.style.color = 'var(--color-yellow-light)';
        } else {
            dStatus.textContent = "MACRO SPEC";
            dStatus.style.backgroundColor = 'rgba(120, 144, 156, 0.2)';
            dStatus.style.color = 'var(--text-muted)';
        }

        // Markdown content (raw text for now)
        dMarkdown.textContent = atom.content;

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
});
