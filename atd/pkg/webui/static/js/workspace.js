/**
 * Workspace Module — Handles workspace selection and project switching.
 * @spec-link [[module_webui_workspace]]
 */

import { fetchWorkspaceInfo, switchProject, fetchTree, fetchInfo } from './api.js';
import { setWorkspace, setActiveProject, setAtoms, on, state } from './state.js';

// DOM elements
let workspaceSelect = null;
let workspaceLabel = null;
let projectBadge = null;

export function initWorkspace() {
    // Cache DOM elements
    workspaceSelect = document.getElementById('workspace-select');
    workspaceLabel = document.getElementById('workspace-label');
    projectBadge = document.getElementById('project-badge');

    if (workspaceSelect) {
        workspaceSelect.addEventListener('change', handleProjectChange);
    }

    // Listen for workspace updates
    on('workspace-updated', renderWorkspace);

    // Initial load
    loadWorkspace();
}

async function loadWorkspace() {
    try {
        const info = await fetchInfo();
        if (info.workspace && info.workspace.in_workspace) {
            // Load full workspace info
            const wsInfo = await fetchWorkspaceInfo();
            setWorkspace(wsInfo);
        }
    } catch (error) {
        console.error('Failed to load workspace:', error);
    }
}

async function handleProjectChange(e) {
    const projectName = e.target.value;
    if (!projectName) return;

    try {
        // Show loading state
        workspaceSelect.disabled = true;
        
        const result = await switchProject(projectName);

        // Update state
        setWorkspace(result); // result should be the updated workspace info
        setActiveProject(projectName);

        // Reload data
        const treeData = await fetchTree();
        setAtoms(treeData);

        // Refresh UI display of paths
        const infoData = await fetchInfo();
        updateInfoDisplay(infoData);

    } catch (error) {
        console.error('Failed to switch project:', error);
        alert('Failed to switch project: ' + error.message);
        
        // Revert selection if failed
        renderWorkspace(state.workspace);
    } finally {
        workspaceSelect.disabled = false;
    }
}

function renderWorkspace(ws) {
    if (!ws.inWorkspace) {
        // Hide workspace UI
        if (workspaceLabel) workspaceLabel.style.display = 'none';
        if (workspaceSelect) workspaceSelect.style.display = 'none';
        if (projectBadge) projectBadge.style.display = 'none';
        return;
    }

    // Show workspace UI
    if (workspaceLabel) {
        workspaceLabel.style.display = 'inline-flex';
        const nameSpan = workspaceLabel.querySelector('span');
        if (nameSpan) nameSpan.textContent = ws.workspaceName;
    }

    if (projectBadge) {
        projectBadge.style.display = 'inline-flex';
        const nameSpan = projectBadge.querySelector('span');
        if (nameSpan) nameSpan.textContent = ws.activeProject;
    }

    if (workspaceSelect) {
        workspaceSelect.style.display = 'inline-block';
        workspaceSelect.innerHTML = '';

        // Add projects to select
        ws.projects.forEach(proj => {
            const option = document.createElement('option');
            option.value = proj.name;
            option.textContent = proj.name;
            if (proj.is_active || proj.name === ws.activeProject) {
                option.selected = true;
            }
            workspaceSelect.appendChild(option);
        });
    }
}

function updateInfoDisplay(info) {
    const infoDocs = document.getElementById('info-docs');
    const infoProject = document.getElementById('info-project');
    const infoCount = document.getElementById('info-count');

    if (infoDocs) infoDocs.textContent = info.docs_path || 'N/A';
    if (infoProject) infoProject.textContent = info.project_path || 'N/A';
    if (infoCount) infoCount.textContent = info.atd_count || '0';
}
