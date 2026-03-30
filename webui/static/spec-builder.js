/**
 * Spec Builder — Chat Logic
 * Manages conversation with Gemini API for ATD spec creation.
 */
(function () {
    'use strict';

    // ============================================
    // STATE
    // ============================================
    const state = {
        messages: [],       // { role: 'user'|'model', content: string }
        proposalGroups: {}, // { atom_id: [GeminiProposal, ...] }
        actions: [],        // { proposal_id, atom_id, action, summary }
        atdContext: [],      // Atom objects attached as context
        atdContextSent: new Set(),  // IDs already sent to avoid redundancy
        exchangeCount: 0,   // Number of user<->model roundtrips
        isLoading: false,
        debugMode: false,
    };

    // ============================================
    // DOM REFERENCES
    // ============================================
    const dom = {};

    function cacheDom() {
        dom.messagesContainer = document.getElementById('chat-messages');
        dom.chatInput = document.getElementById('chat-input');
        dom.sendBtn = document.getElementById('btn-send');
        dom.welcomeScreen = document.querySelector('.chat-welcome');
        dom.contextBar = document.getElementById('chat-context-bar');
        dom.contextChips = document.getElementById('context-chips');
        dom.addContextBtn = document.getElementById('btn-add-context');
        dom.proposalsList = document.getElementById('proposals-list');
        dom.proposalCount = document.getElementById('proposal-count');
        dom.proposalsActions = document.getElementById('proposals-actions');
        dom.newSessionBtn = document.getElementById('btn-new-session');
        dom.exportBtn = document.getElementById('btn-export-chat');
        dom.acceptAllBtn = document.getElementById('btn-accept-all');
        dom.rejectAllBtn = document.getElementById('btn-reject-all');
        dom.modelSelect = document.getElementById('model-select');
        dom.debugCheck = document.getElementById('check-debug-mode');
    }

    // ============================================
    // INITIALIZATION
    // ============================================
    function init() {
        cacheDom();
        if (!dom.chatInput) return; // Guard: spec-builder tab not in DOM yet

        // Send message
        dom.sendBtn.addEventListener('click', sendMessage);
        dom.chatInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && (e.ctrlKey || e.metaKey)) {
                e.preventDefault();
                sendMessage();
            }
        });

        // Auto-grow textarea
        dom.chatInput.addEventListener('input', autoGrowTextarea);

        // Suggestion chips
        document.querySelectorAll('.suggestion-chip').forEach(chip => {
            chip.addEventListener('click', () => {
                dom.chatInput.value = chip.dataset.prompt + ' ';
                dom.chatInput.focus();
                autoGrowTextarea();
            });
        });

        // Session controls (wired in Step 8, basic reset here)
        if (dom.newSessionBtn) {
            dom.newSessionBtn.addEventListener('click', resetSession);
        }

        // Load available models from API
        loadModels();

        // Wire bulk actions
        if (dom.acceptAllBtn) {
            dom.acceptAllBtn.addEventListener('click', async () => {
                const atomIds = Object.keys(state.proposalGroups);
                for (const id of atomIds) {
                    const group = state.proposalGroups[id];
                    const currentIdx = group[0]._currentDisplayIndex !== undefined ? group[0]._currentDisplayIndex : group.length - 1;
                    const card = dom.proposalsList.querySelector(`[data-atom-id="${id}"]`);
                    if (group[0]._status === 'pending') {
                        await handleProposalAction(id, currentIdx, 'accepted', card);
                    }
                }
            });
        }

        if (dom.rejectAllBtn) {
            dom.rejectAllBtn.addEventListener('click', () => {
                const atomIds = Object.keys(state.proposalGroups);
                atomIds.forEach(id => {
                    const group = state.proposalGroups[id];
                    const currentIdx = group[0]._currentDisplayIndex !== undefined ? group[0]._currentDisplayIndex : group.length - 1;
                    const card = dom.proposalsList.querySelector(`[data-atom-id="${id}"]`);
                    if (group[0]._status === 'pending') {
                        handleProposalAction(id, currentIdx, 'rejected', card);
                    }
                });
            });
        }
    }

    // ============================================
    // MODEL LOADING
    // ============================================
    async function loadModels() {
        try {
            const resp = await fetch('/api/gemini/models');
            if (!resp.ok) return;
            const data = await resp.json();
            if (!data.models || !dom.modelSelect) return;

            // Filter to gemini models and populate dropdown
            const currentValue = dom.modelSelect.value;
            dom.modelSelect.innerHTML = '';
            data.models.forEach(m => {
                const opt = document.createElement('option');
                opt.value = m.id;
                opt.textContent = m.display_name || m.id;
                if (m.id === (data.default || 'gemini-2.5-flash')) opt.selected = true;
                dom.modelSelect.appendChild(opt);
            });
            // Restore previous selection if it still exists
            if (currentValue) {
                const exists = Array.from(dom.modelSelect.options).some(o => o.value === currentValue);
                if (exists) dom.modelSelect.value = currentValue;
            }
        } catch (err) {
            console.warn('Failed to load models:', err);
        }
    }

    // ============================================
    // MESSAGE SENDING
    // ============================================
    async function sendMessage() {
        const text = dom.chatInput.value.trim();
        if (!text || state.isLoading) return;

        // Hide welcome screen
        if (dom.welcomeScreen) {
            dom.welcomeScreen.style.display = 'none';
        }

        // Add user message to state and render
        state.messages.push({ role: 'user', content: text });
        renderMessage('user', text);

        // Clear input
        dom.chatInput.value = '';
        autoGrowTextarea();

        // Show typing indicator
        state.isLoading = true;
        dom.sendBtn.disabled = true;
        const typingEl = showTypingIndicator();

        try {
            // Build request payload
            const payload = buildChatPayload();

            const response = await fetch('/api/gemini/chat', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload),
            });

            // Remove typing indicator
            typingEl.remove();

            if (!response.ok) {
                const err = await response.json();
                throw new Error(err.error || `HTTP ${response.status}`);
            }

            const data = await response.json();

            // Debug logging
            if (dom.debugCheck && dom.debugCheck.checked) {
                console.log("RAW GEMINI RESPONSE:", data);
            }

            // Add model message to state
            state.messages.push({ role: 'model', content: data.message });
            renderMessage('model', data.message);

            // Handle proposals
            if (data.proposals && data.proposals.length > 0) {
                handleProposals(data.proposals);
            }

            // Increment exchange count
            state.exchangeCount++;
            checkContextDrift();

            // Mark sent ATD context as already injected
            state.atdContext.forEach(atd => state.atdContextSent.add(atd.id));

        } catch (error) {
            typingEl.remove();
            renderErrorMessage(error.message);
        } finally {
            state.isLoading = false;
            dom.sendBtn.disabled = false;
            dom.chatInput.focus();
        }
    }

    function buildChatPayload() {
        // Only include ATD context not yet sent
        const newContext = state.atdContext.filter(atd => !state.atdContextSent.has(atd.id));

        return {
            messages: state.messages,
            model: dom.modelSelect ? dom.modelSelect.value : 'gemini-2.5-flash',
            atd_context: newContext,
            actions: state.actions,
        };
    }

    // ============================================
    // RENDERING
    // ============================================
    function renderMessage(role, content) {
        const msgEl = document.createElement('div');
        msgEl.className = `chat-msg ${role}`;

        const avatar = document.createElement('div');
        avatar.className = 'msg-avatar';
        avatar.textContent = role === 'user' ? '👤' : '🤖';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';

        if (role === 'model') {
            // Render as markdown
            const mdContainer = document.createElement('div');
            mdContainer.className = 'markdown-body';
            mdContainer.innerHTML = marked.parse(content);
            bubble.appendChild(mdContainer);
        } else {
            bubble.textContent = content;
        }

        msgEl.appendChild(avatar);
        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);
        scrollToBottom();
    }

    function renderErrorMessage(errorText) {
        const msgEl = document.createElement('div');
        msgEl.className = 'chat-msg model';

        const avatar = document.createElement('div');
        avatar.className = 'msg-avatar';
        avatar.textContent = '⚠️';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';
        bubble.style.borderColor = 'var(--color-red)';
        bubble.innerHTML = `<span style="color: var(--color-red-light)">Error: ${escapeHtml(errorText)}</span>`;

        msgEl.appendChild(avatar);
        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);
        scrollToBottom();
    }

    function showTypingIndicator() {
        const msgEl = document.createElement('div');
        msgEl.className = 'chat-msg model';
        msgEl.id = 'typing-indicator';

        const avatar = document.createElement('div');
        avatar.className = 'msg-avatar';
        avatar.textContent = '🤖';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble typing-indicator';
        bubble.innerHTML = '<div class="typing-dot"></div><div class="typing-dot"></div><div class="typing-dot"></div>';

        msgEl.appendChild(avatar);
        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);
        scrollToBottom();
        return msgEl;
    }

    // ============================================
    // PROPOSALS — Full Implementation with Versioning & Grouping
    // ============================================
    function handleProposals(proposals) {
        proposals.forEach(p => {
            p._id = `proposal_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
            p._status = 'pending';
            
            if (!state.proposalGroups[p.atom_id]) {
                state.proposalGroups[p.atom_id] = [];
            }
            // Add new version
            state.proposalGroups[p.atom_id].push(p);
            
            // Set current version pointer to the latest one
            p._versionIndex = state.proposalGroups[p.atom_id].length - 1;
            
            renderProposalCard(p.atom_id);
        });
        updateProposalCount();
    }

    async function renderProposalCard(atomId) {
        const group = state.proposalGroups[atomId];
        if (!group || group.length === 0) return;
        
        // Find existing card or create new
        let card = dom.proposalsList.querySelector(`[data-atom-id="${atomId}"]`);
        const isNew = !card;
        
        if (isNew) {
            // Remove empty state if present
            const emptyEl = dom.proposalsList.querySelector('.proposals-empty');
            if (emptyEl) emptyEl.remove();

            card = document.createElement('div');
            card.className = 'proposal-card';
            card.dataset.atomId = atomId;
            dom.proposalsList.appendChild(card);
        }

        // Use the latest version for rendering by default, or the one previously selected
        const currentIdx = group[0]._currentDisplayIndex !== undefined ? group[0]._currentDisplayIndex : group.length - 1;
        const p = group[currentIdx];

        // Check if atom exists for Smart Matching
        const existsLocally = window.Atoms ? window.Atoms[p.atom_id] : null;
        // Don't require matching if the first version in this card is a CREATE or if it exists locally
        const isPendingCreate = group.some(prop => prop.action === 'CREATE');
        const matchingRequired = p.action === 'UPDATE' && !existsLocally && !isPendingCreate;

        card.innerHTML = '';
        
        // Header with action badge and buttons
        const header = document.createElement('div');
        header.className = 'proposal-card-header';

        const badge = document.createElement('span');
        badge.className = `proposal-action-badge ${p.action}`;
        badge.textContent = p.action;
        
        // Version Indicator
        if (group.length > 1) {
            const vIndicator = document.createElement('span');
            vIndicator.className = 'proposal-version-badge';
            vIndicator.textContent = `v${currentIdx + 1}/${group.length}`;
            header.appendChild(vIndicator);
        }

        const actions = document.createElement('div');
        actions.className = 'proposal-card-actions';

        if (group.length > 1) {
            const prevBtn = createProposalBtn('prev', '←', () => {
                group[0]._currentDisplayIndex = Math.max(0, currentIdx - 1);
                renderProposalCard(atomId);
            });
            const nextBtn = createProposalBtn('next', '→', () => {
                group[0]._currentDisplayIndex = Math.min(group.length - 1, currentIdx + 1);
                renderProposalCard(atomId);
            });
            actions.append(prevBtn, nextBtn);
        }

        const acceptBtn = createProposalBtn('accept', '✓', () => handleProposalAction(atomId, currentIdx, 'accepted', card));
        const rejectBtn = createProposalBtn('reject', '✗', () => handleProposalAction(atomId, currentIdx, 'rejected', card));
        const previewBtn = createProposalBtn('preview', '👁', () => toggleProposalDetail(card, p));

        actions.append(acceptBtn, rejectBtn, previewBtn);
        header.append(badge, actions);

        // Body
        const atomIdEl = document.createElement('div');
        atomIdEl.className = 'proposal-atom-id';
        atomIdEl.textContent = `@${p.atom_id}`;

        const atomName = document.createElement('div');
        atomName.className = 'proposal-atom-name';
        atomName.textContent = p.content?.human_name || p.atom_id;

        const impact = document.createElement('div');
        impact.className = 'proposal-impact';
        impact.textContent = p.impact_summary || '';

        card.append(header, atomIdEl, atomName, impact);

        // Smart Matching UI if needed
        if (matchingRequired) {
            const matchingPanel = document.createElement('div');
            matchingPanel.className = 'matching-panel';
            matchingPanel.innerHTML = `
                <div class="matching-warning">⚠️ No atom found with ID: ${p.atom_id}</div>
                <div class="matching-actions">
                    <button class="btn btn-outline btn-xs" id="btn-search-match-${p._id}">Search & Link</button>
                    <button class="btn btn-outline btn-xs" id="btn-create-anyway-${p._id}">Create as New</button>
                </div>
                <div class="matching-manual">
                    <div class="manual-id-input-wrapper">
                        <input type="text" id="input-manual-id-${p._id}" placeholder="Enter existing atom_id...">
                        <button class="btn btn-primary btn-xs btn-link" id="btn-manual-link-${p._id}">Link</button>
                    </div>
                </div>
            `;
            card.appendChild(matchingPanel);
            
            card.querySelector(`#btn-search-match-${p._id}`).addEventListener('click', () => {
                showMatchingSearch(p, (selectedAtom) => {
                    p._original_atom_id = p.atom_id;
                    p.atom_id = selectedAtom.id;
                    p.action = 'UPDATE';
                    renderProposalCard(atomId); // Re-render this group's card
                });
            });

            card.querySelector(`#btn-manual-link-${p._id}`).addEventListener('click', () => {
                const manualId = card.querySelector(`#input-manual-id-${p._id}`).value.trim();
                if (!manualId) return;
                p._original_atom_id = p.atom_id;
                p.atom_id = manualId;
                p.action = 'UPDATE';
                renderProposalCard(atomId);
            });
            
            card.querySelector(`#btn-create-anyway-${p._id}`).addEventListener('click', () => {
                p.action = 'CREATE';
                renderProposalCard(atomId);
            });
        }

        // Show bulk actions
        if (dom.proposalsActions) dom.proposalsActions.style.display = 'flex';
    }

    function createProposalBtn(type, icon, handler) {
        const btn = document.createElement('button');
        btn.className = `proposal-btn ${type}`;
        btn.title = type.charAt(0).toUpperCase() + type.slice(1);
        btn.textContent = icon;
        btn.addEventListener('click', handler);
        return btn;
    }

    async function showMatchingSearch(proposal, onSelect) {
        // Simple search UI: show a list of top 5 matches
        const query = proposal.atom_id.split('_').slice(1).join(' ') || proposal.atom_id;
        try {
            const resp = await fetch(`/api/gemini/atoms?q=${encodeURIComponent(query)}`);
            if (!resp.ok) throw new Error('Search failed');
            const matches = await resp.json();
            
            const dropdown = document.createElement('div');
            dropdown.className = 'matching-results';
            if (matches.length === 0) {
                dropdown.innerHTML = '<div class="matching-no-results">No similar atoms found.</div>';
            } else {
                matches.slice(0, 5).forEach(m => {
                    const item = document.createElement('div');
                    item.className = 'matching-item';
                    item.innerHTML = `<strong>${m.id}</strong><br><small>${m.human_name}</small>`;
                    item.onclick = () => {
                        onSelect(m);
                        dropdown.remove();
                    };
                    dropdown.appendChild(item);
                });
            }
            
            // Append to the specific card
            const card = dom.proposalsList.querySelector(`[data-atom-id="${proposal.atom_id}"]`);
            if (card) {
                const existing = card.querySelector('.matching-results');
                if (existing) existing.remove();
                card.appendChild(dropdown);
            }
        } catch (err) {
            console.error('Matching search failed:', err);
        }
    }

    async function handleProposalAction(atomId, versionIdx, action, cardEl) {
        const group = state.proposalGroups[atomId];
        const proposal = group[versionIdx];
        
        if (proposal._status !== 'pending') return;
        
        // Mark ALL versions in this group as actioned
        group.forEach(p => p._status = action);

        // Record in action history for forwarding to Gemini
        state.actions.push({
            proposal_id: proposal._id,
            atom_id: proposal.atom_id,
            action: action.toUpperCase(),
            summary: `${proposal.action} ${proposal.atom_id}: ${proposal.impact_summary || ''}`,
        });

        // Update card UI
        cardEl.className = `proposal-card ${action}`;

        // If accepted, apply the proposal
        if (action === 'accepted') {
            try {
                const resp = await fetch('/api/gemini/apply-proposal', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        action: proposal.action,
                        atom_id: proposal.atom_id,
                        content: proposal.content || {},
                    }),
                });
                if (!resp.ok) {
                    const err = await resp.json();
                    alert(`Failed to apply proposal: ${err.error}\n${err.details || ''}`);
                    group.forEach(p => p._status = 'pending');
                    cardEl.className = 'proposal-card';
                    return;
                }
                // Show success indicator on card
                const successTag = document.createElement('div');
                successTag.style.cssText = 'color: var(--color-green-light); font-size: 11px; margin-top: 8px;';
                successTag.textContent = '✓ Applied successfully';
                cardEl.appendChild(successTag);
                
                // Remove from UI after a short delay
                setTimeout(() => {
                    cardEl.style.opacity = '0';
                    cardEl.style.transform = 'translateX(20px)';
                    setTimeout(() => {
                        cardEl.remove();
                        delete state.proposalGroups[atomId];
                        updateProposalCount();
                    }, 300);
                }, 1500);
            } catch (err) {
                alert('Network error applying proposal: ' + err.message);
                group.forEach(p => p._status = 'pending');
                cardEl.className = 'proposal-card';
                return;
            }
        } else if (action === 'rejected') {
            // Remove from UI after delay
            setTimeout(() => {
                cardEl.remove();
                delete state.proposalGroups[atomId];
                updateProposalCount();
            }, 1000);
        }

        // Disable action buttons
        cardEl.querySelectorAll('.proposal-btn').forEach(btn => {
            btn.disabled = true;
            btn.style.opacity = '0.3';
        });

        updateProposalCount();
    }

    function toggleProposalDetail(cardEl, proposal) {
        const existing = cardEl.querySelector('.proposal-detail');
        if (existing) {
            existing.remove();
            return;
        }

        const detail = document.createElement('div');
        detail.className = 'proposal-detail';

        const fields = proposal.content || {};
        const rows = [];

        if (fields.human_name) rows.push(['Name', fields.human_name]);
        if (fields.type) rows.push(['Type', fields.type]);
        if (fields.layer) rows.push(['Layer', fields.layer]);
        if (fields.status) rows.push(['Status', fields.status]);
        if (fields.intent) rows.push(['Intent', fields.intent]);
        if (fields.logic) rows.push(['Logic', fields.logic]);
        if (fields.technical_interface) rows.push(['Interface', fields.technical_interface]);
        if (fields.expectation) rows.push(['Expect', fields.expectation]);
        if (fields.tags) rows.push(['Tags', Array.isArray(fields.tags) ? fields.tags.join(', ') : fields.tags]);
        if (fields.parents) rows.push(['Parents', Array.isArray(fields.parents) ? fields.parents.join(', ') : fields.parents]);

        rows.forEach(([label, value]) => {
            const row = document.createElement('div');
            row.className = 'field-row';
            row.innerHTML = `<span class="field-label">${label}</span><span class="field-value">${escapeHtml(value)}</span>`;
            detail.appendChild(row);
        });

        cardEl.appendChild(detail);
    }

    function updateProposalCount() {
        const atomIds = Object.keys(state.proposalGroups);
        const pending = atomIds.filter(id => state.proposalGroups[id][0]._status === 'pending').length;
        
        dom.proposalCount.textContent = `${pending} pending`;

        if (pending === 0 && dom.proposalsActions) {
            dom.proposalsActions.style.display = 'none';
        }
    }

    // ============================================
    // CONTEXT DRIFT CHECK
    // ============================================
    function checkContextDrift() {
        if (state.exchangeCount === 10) {
            const warning = document.createElement('div');
            warning.className = 'drift-warning';
            warning.innerHTML = `
                <span class="drift-icon">⚠️</span>
                <span><strong>Context drift warning:</strong> You've had ${state.exchangeCount} exchanges. Consider starting a new session to maintain accuracy.</span>
                <button class="btn btn-outline btn-sm" onclick="document.getElementById('btn-new-session').click()">Restart</button>
            `;
            dom.messagesContainer.appendChild(warning);
            scrollToBottom();
        }
    }

    // ============================================
    // SESSION MANAGEMENT
    // ============================================
    function resetSession() {
        if (!confirm('Start a new session? All conversation history will be cleared.')) return;

        state.messages = [];
        state.proposalGroups = {};
        state.actions = [];
        state.atdContext = [];
        state.atdContextSent.clear();
        state.exchangeCount = 0;

        // Reset UI
        dom.messagesContainer.innerHTML = '';
        if (dom.welcomeScreen) {
            dom.messagesContainer.appendChild(dom.welcomeScreen);
            dom.welcomeScreen.style.display = 'flex';
        }

        // Reset proposals
        dom.proposalsList.innerHTML = '<div class="proposals-empty"><p>Proposals from the AI will appear here for your review.</p><p class="proposals-hint">You can accept ✓ or reject ✗ each proposal individually.</p></div>';
        dom.proposalCount.textContent = '0 pending';
        if (dom.proposalsActions) dom.proposalsActions.style.display = 'none';
    }

    // ============================================
    // UTILITIES
    // ============================================
    function autoGrowTextarea() {
        dom.chatInput.style.height = 'auto';
        dom.chatInput.style.height = Math.min(dom.chatInput.scrollHeight, 150) + 'px';
    }

    function scrollToBottom() {
        dom.messagesContainer.scrollTop = dom.messagesContainer.scrollHeight;
    }

    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.appendChild(document.createTextNode(text));
        return div.innerHTML;
    }

    // ============================================
    // EXPOSE FOR OTHER MODULES
    // ============================================
    window.SpecBuilder = {
        state,
        resetSession,
        getProposals: () => Object.values(state.proposalGroups).flat(),
        getActions: () => state.actions,
    };

    // ============================================
    // BOOT
    // ============================================
    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }

})();
