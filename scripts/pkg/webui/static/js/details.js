/**
 * Details Module — Atom detail panel rendering, edit form, and navigation.
 * @spec-link [[ui_webui_traceability_explorer]]
 */
import { state, setCurrentAtom, on, emit } from './state.js';
import { updateAtom, fetchSummary } from './api.js';

let dom = {};

export function initDetails() {
    dom = {
        panelEmpty: document.getElementById('details-empty'),
        panelContent: document.getElementById('details-content'),
        closeBtn: document.getElementById('close-panel'),
        dType: document.getElementById('detail-type'),
        dLayer: document.getElementById('detail-layer'),
        dStatus: document.getElementById('detail-status'),
        dTitle: document.getElementById('detail-title'),
        dId: document.getElementById('detail-id'),
        dMarkdown: document.getElementById('detail-markdown'),
        dCode: document.getElementById('detail-code'),
        dTestStatus: document.getElementById('detail-test-status'),
        dParents: document.getElementById('detail-parents'),
        dDependents: document.getElementById('detail-dependents'),
        editForm: document.getElementById('edit-form'),
        btnEdit: document.getElementById('btn-edit'),
        btnSave: document.getElementById('btn-save'),
        btnCancelEdit: document.getElementById('btn-cancel-edit'),
        btnSummarize: document.getElementById('btn-summarize'),
        summaryBox: document.getElementById('summary-box'),
        editHumanName: document.getElementById('edit-human-name'),
        editType: document.getElementById('edit-type'),
        editStatus: document.getElementById('edit-status'),
        editPriority: document.getElementById('edit-priority'),
        editTags: document.getElementById('edit-tags'),
        editContent: document.getElementById('edit-content'),
        detailActions: document.querySelector('.detail-actions'),
    };

    dom.closeBtn.addEventListener('click', closePanel);
    dom.btnEdit.addEventListener('click', enterEditMode);
    dom.btnCancelEdit.addEventListener('click', exitEditMode);
    dom.btnSave.addEventListener('click', saveAtomChanges);

    on('atom-selected', showDetails);
}

function closePanel() {
    dom.panelContent.style.display = 'none';
    dom.panelEmpty.style.display = 'flex';
    exitEditMode();
    setCurrentAtom(null);
}

export function showDetails(atom) {
    if (!atom || atom.id === 'root') return;

    dom.panelEmpty.style.display = 'none';
    dom.panelContent.style.display = 'block';
    exitEditMode();

    dom.dType.textContent = atom.type;
    dom.dLayer.textContent = atom.layer || 'UNKNOWN';

    // Layer Badge color
    if (atom.layer === 'CUSTOMER') {
        dom.dLayer.style.backgroundColor = 'rgba(232, 121, 249, 0.2)';
        dom.dLayer.style.color = 'var(--color-customer)';
    } else if (atom.layer === 'ARCHITECTURE') {
        dom.dLayer.style.backgroundColor = 'rgba(96, 165, 250, 0.2)';
        dom.dLayer.style.color = 'var(--color-architecture)';
    } else if (atom.layer === 'IMPLEMENTATION') {
        dom.dLayer.style.backgroundColor = 'rgba(52, 211, 153, 0.2)';
        dom.dLayer.style.color = 'var(--color-implementation)';
    } else {
        dom.dLayer.style.backgroundColor = 'rgba(120, 144, 156, 0.2)';
        dom.dLayer.style.color = 'var(--color-grey-light)';
    }
    dom.dTitle.textContent = atom.human_name || atom.id;
    dom.dId.textContent = `@${atom.id}`;

    // Status Badge
    dom.dStatus.textContent = atom.computed_status || 'UNKNOWN';
    const colorMap = {
        green: { bg: 'rgba(76, 175, 80, 0.2)', fg: 'var(--color-green-light)' },
        yellow: { bg: 'rgba(253, 216, 53, 0.2)', fg: 'var(--color-yellow-light)' },
        red: { bg: 'rgba(211, 47, 47, 0.2)', fg: 'var(--color-red-light)' },
    };
    const colors = colorMap[atom.computed_color] || { bg: 'rgba(120, 144, 156, 0.2)', fg: 'var(--text-muted)' };
    dom.dStatus.style.backgroundColor = colors.bg;
    dom.dStatus.style.color = colors.fg;

    // Content Sections Concatenation
    const fullContent = (atom.intent || "") + "\n\n" + 
                       (atom.logic ? `## THE RULE / LOGIC\n${atom.logic}\n\n` : "") +
                       (atom.interface ? `## TECHNICAL INTERFACE\n${atom.interface}\n\n` : "") +
                       (atom.expectation ? `## EXPECTATION\n${atom.expectation}\n\n` : "");
    dom.dMarkdown.innerHTML = marked.parse(fullContent || '');

    // Parents & Dependents (clickable navigation) — ISS-004
    renderLinks(dom.dParents, 'Parents', atom.parents);
    renderLinks(dom.dDependents, 'Dependents', atom.dependents);

    // Code Links
    dom.dCode.innerHTML = '';
    if (atom.linked_codes && atom.linked_codes.length > 0) {
        atom.linked_codes.forEach(link => {
            const li = document.createElement('li');
            li.textContent = link;
            dom.dCode.appendChild(li);
        });
    } else {
        dom.dCode.innerHTML = '<li style="color:var(--text-muted);border:none;background:transparent;">No code references found.</li>';
    }

    // Testing
    if (atom.has_tests) {
        dom.dTestStatus.innerHTML = '<span class="test-pass">●</span> Tests detected on implementation.';
    } else if (atom.linked_codes && atom.linked_codes.length > 0) {
        dom.dTestStatus.innerHTML = '<span class="test-missing">●</span> Implementation found, but testing missing.';
    } else {
        dom.dTestStatus.innerHTML = '<span style="color:var(--text-muted)">No implementation to test.</span>';
    }

    // Summary
    dom.summaryBox.style.display = 'none';
    dom.btnSummarize.onclick = async () => {
        dom.summaryBox.style.display = 'block';
        dom.summaryBox.innerHTML = '<i>Generating summary via local LLM...</i>';
        try {
            const data = await fetchSummary(atom.id);
            dom.summaryBox.innerHTML = marked.parse(data.summary || 'No summary available.');
            if (data.llm_used) {
                dom.summaryBox.innerHTML += '<div class="summary-source">✨ Enhanced by local LLM</div>';
            }
        } catch (err) {
            dom.summaryBox.innerHTML = '<span style="color:var(--color-red)">Failed to generate summary.</span>';
        }
    };
}

// @spec-link [[ui_webui_waterfall_explorer]]
function renderLinks(container, label, ids) {
    if (!container) return;
    container.innerHTML = '';
    if (!ids || ids.length === 0) {
        container.innerHTML = `<span class="link-empty">None</span>`;
        return;
    }
    ids.forEach(id => {
        const link = document.createElement('a');
        link.className = 'atom-nav-link';
        link.href = `?atom=${id}`;
        link.textContent = id;
        link.addEventListener('click', (e) => {
            e.preventDefault();
            const atom = state.atoms.find(a => a.id === id);
            if (atom) {
                setCurrentAtom(atom);
                emit('navigate-to-atom', atom);
                // Update URL without reload
                window.history.pushState({}, '', `?atom=${id}`);
            }
        });
        container.appendChild(link);
    });
}

function enterEditMode() {
    const atom = state.currentAtom;
    if (!atom) return;

    dom.editHumanName.value = atom.human_name || '';
    dom.editType.value = atom.type || 'MECHANIC';
    dom.editStatus.value = atom.status || 'DRAFT';
    dom.editPriority.value = atom.priority || 3;
    dom.editTags.value = (atom.tags || []).join(', ');
    const fullContent = (atom.intent || "") + "\n\n" + 
                       (atom.logic ? `## THE RULE / LOGIC\n${atom.logic}\n\n` : "") +
                       (atom.interface ? `## TECHNICAL INTERFACE\n${atom.interface}\n\n` : "") +
                       (atom.expectation ? `## EXPECTATION\n${atom.expectation}\n\n` : "");
    dom.editContent.value = fullContent || '';

    dom.editForm.style.display = 'block';
    dom.detailActions.style.display = 'none';
    document.querySelectorAll('#details-content .detail-section').forEach(el => el.style.display = 'none');
}

function exitEditMode() {
    dom.editForm.style.display = 'none';
    dom.detailActions.style.display = 'flex';
    document.querySelectorAll('#details-content .detail-section').forEach(el => el.style.display = 'block');
}

function slugify(text) {
    return text.toString().toLowerCase()
        .replace(/\s+/g, '_')
        .replace(/[^\w-]+/g, '')
        .replace(/--+/g, '_')
        .replace(/^-+/, '')
        .replace(/-+$/, '');
}

// @spec-link [[mechanic_atd_update]]
async function saveAtomChanges() {
    const atom = state.currentAtom;
    if (!atom) return;

    const newType = dom.editType.value;
    const newHumanName = dom.editHumanName.value;
    const newId = `${newType.toLowerCase()}_${slugify(newHumanName)}`;

    const updatedData = {
        id: newId,
        human_name: newHumanName,
        type: newType,
        status: dom.editStatus.value,
        priority: parseInt(dom.editPriority.value, 10),
        tags: dom.editTags.value.split(',').map(t => t.trim()).filter(t => t !== ''),
        intent: dom.editContent.value,
    };

    dom.btnSave.disabled = true;
    dom.btnSave.textContent = 'Saving...';

    try {
        const resp = await updateAtom(atom.id, updatedData);
        if (resp.ok) {
            emit('data-refresh-needed');
            exitEditMode();
        } else {
            const err = await resp.json();
            alert('Save failed: ' + (err.error || 'Unknown error'));
        }
    } catch (error) {
        alert('Failed to send update request.');
    } finally {
        dom.btnSave.disabled = false;
        dom.btnSave.textContent = 'Save Changes';
    }
}
