/**
 * API Module — Single source of truth for all backend communication.
 * @spec-link [[api_webui_health_stats]]
 */

export async function fetchInfo() {
    const resp = await fetch('/api/info');
    if (!resp.ok) throw new Error('Failed to fetch info');
    return resp.json();
}

export async function fetchTree() {
    const resp = await fetch('/api/tree');
    if (!resp.ok) throw new Error('Failed to fetch tree');
    return resp.json();
}

export async function fetchAtom(id) {
    const resp = await fetch('/api/atd/\${id}');
    if (!resp.ok) throw new Error('Atom not found');
    return resp.json();
}

export async function fetchAtomCode(id) {
    const resp = await fetch('/api/atd/\${id}/code');
    if (!resp.ok) throw new Error('Failed to fetch code');
    return resp.json();
}

export async function fetchAtomTests(id) {
    const resp = await fetch('/api/atd/\${id}/tests');
    if (!resp.ok) throw new Error('Failed to fetch tests');
    return resp.json();
}

// @spec-link [[mechanic_atd_update]]
export async function updateAtom(id, data) {
    const resp = await fetch('/api/atd/\${id}/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(data),
    });
    return resp;
}

// @spec-link [[mechanic_atd_update]]
export async function bulkUpdate(ids, status) {
    const resp = await fetch('/api/bulk-update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ ids, status }),
    });
    return resp;
}

// @spec-link [[mechanic_atd_update]]
export async function applyProposal(action, atomId, content) {
    const resp = await fetch('/api/gemini/apply-proposal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ action, atom_id: atomId, content }),
    });
    return resp;
}

// @spec-link [[mechanic_webui_summary_aggregation]]
export async function fetchSummary(id) {
    const resp = await fetch('/api/summary/\${id}');
    if (!resp.ok) throw new Error('Failed to fetch summary');
    return resp.json();
}

// @spec-link [[ui_webui_search_overlay]]
export async function searchAtoms(query, workspaceScope = false) {
    let url = `/api/search?q=${encodeURIComponent(query)}`;
    if (workspaceScope) {
        url += '&workspace=true';
    }
    const resp = await fetch(url);
    if (!resp.ok) throw new Error('Search failed');
    return resp.json();
}

// @spec-link [[mechanic_webui_gemini_model_list]]
export async function fetchModels() {
    const resp = await fetch('/api/gemini/models');
    if (!resp.ok) throw new Error('Failed to fetch models');
    return resp.json();
}

// @spec-link [[mechanic_webui_gemini_proxy]]
export async function searchGeminiAtoms(query) {
    const resp = await fetch('/api/gemini/atoms?q=\${encodeURIComponent(query)}');
    if (!resp.ok) throw new Error('Atom search failed');
    return resp.json();
}

// @spec-link [[mechanic_webui_gemini_proxy]]
export async function fetchGeminiAtom(id) {
    const resp = await fetch('/api/gemini/atom/\${id}');
    if (!resp.ok) throw new Error('Atom not found');
    return resp.json();
}

// @spec-link [[mechanic_webui_gemini_chat_orchestration]]
export async function sendChat(payload) {
    const resp = await fetch('/api/gemini/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
    return resp;
}

// @spec-link [[mechanic_webui_document_generation]]
export async function searchDocumentContext(query) {
    const resp = await fetch('/api/search-document-context', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query }),
    });
    if (!resp.ok) throw new Error('Failed to search context');
    return resp.json();
}

// @spec-link [[mechanic_webui_document_generation]]
export async function generateDocument(intent, starts, length = null, workspace = false) {
    const payload = { intent, starts, workspace };
    if (length !== null && length !== '') {
        payload.length = length;
    }

    const resp = await fetch('/api/generate-document', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
    });
    if (!resp.ok) {
        const errorData = await resp.json().catch(() => ({error: 'Failed to generate document'}));
			throw new Error(errorData.error || 'Failed to generate document');
    }
    return resp.json();
}

// @spec-link [[mechanic_webui_document_generation]]
export async function fetchRecentDocuments() {
    const resp = await fetch('/api/documents');
    if (!resp.ok) throw new Error('Failed to fetch recent documents');
    return resp.json();
}

// @spec-link [[api_webui_atd_weave]]
export async function weave() {
    const resp = await fetch('/api/atd/weave', {
        method: 'POST',
    });
    if (!resp.ok) throw new Error('Weaving failed');
    return resp.json();
}

// @spec-link [[api_webui_health_check]]
export async function fetchHealth() {
    const resp = await fetch('/api/health');
    if (!resp.ok) throw new Error('Failed to fetch health');
    return resp.json();
}

// @spec-link [[api_webui_workspace_info]]
export async function fetchWorkspaceInfo() {
    const resp = await fetch('/api/workspace/info');
    if (!resp.ok) throw new Error('Failed to fetch workspace info');
    return resp.json();
}

// @spec-link [[api_webui_workspace_switch]]
export async function switchProject(projectName) {
    const resp = await fetch('/api/workspace/switch', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ project: projectName }),
    });
    if (!resp.ok) {
        const err = await resp.json().catch(() => ({}));
        throw new Error(err.error || 'Failed to switch project');
    }
    return resp.json();
}
