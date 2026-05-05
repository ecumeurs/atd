/**
 * Spec Builder — Chat Logic
 * Manages conversation with various LLM APIs for ATD spec creation.
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
        exhaustedModels: new Set(), // Model IDs that reached quota (429)
        totalTokens: 0,
        chatModeEnabled: true,
        providers: [],      // Available providers from backend
        keys: {},           // { providerName: apiKey } loaded from localStorage
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
        dom.chatModeCheck = document.getElementById('check-chat-mode');
        dom.totalTokensDisplay = document.getElementById('total-tokens');
        
        // Settings Modal
        dom.settingsBtn = document.getElementById('btn-llm-settings');
        dom.settingsModal = document.getElementById('llm-settings-modal');
        dom.closeSettingsBtn = document.getElementById('close-llm-settings');
        dom.saveSettingsBtn = document.getElementById('btn-save-llm-settings');
        
        dom.accessMethod = document.getElementById('llm-access-method');
        dom.settingsGemini = document.getElementById('settings-gemini');
        dom.settingsOpenAI = document.getElementById('settings-openai');
        dom.geminiKeyInput = document.getElementById('input-gemini-key');
        dom.openaiKeyInput = document.getElementById('input-openai-key');
        dom.openaiUrlInput = document.getElementById('input-openai-url');
    }

    // ============================================
    // INITIALIZATION
    // ============================================
    // @spec-link [[ui_webui_spec_builder]]
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

        // Add Context Button
        if (dom.addContextBtn) {
            dom.addContextBtn.addEventListener('click', openContextSearch);
        }

        // Session controls (wired in Step 8, basic reset here)
        if (dom.newSessionBtn) {
            dom.newSessionBtn.addEventListener('click', resetSession);
        }

        // Load keys from localStorage
        loadKeys();

        // Load available models from API
        loadModels();

        // Settings Modal Events
        if (dom.settingsBtn) {
            dom.settingsBtn.addEventListener('click', openSettings);
        }
        if (dom.closeSettingsBtn) {
            dom.closeSettingsBtn.addEventListener('click', () => dom.settingsModal.style.display = 'none');
        }
        if (dom.saveSettingsBtn) {
            dom.saveSettingsBtn.addEventListener('click', saveSettings);
        }

        if (dom.accessMethod) {
            dom.accessMethod.addEventListener('change', () => {
                const method = dom.accessMethod.value;
                dom.settingsGemini.style.display = method === 'gemini' ? 'block' : 'none';
                dom.settingsOpenAI.style.display = method === 'openai' ? 'block' : 'none';
            });
        }

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

        // Chat Mode Toggle
        if (dom.chatModeCheck) {
            dom.chatModeCheck.addEventListener('change', handleChatModeChange);
        }

        // Export button
        if (dom.exportBtn) {
            dom.exportBtn.addEventListener('click', exportConversation);
        }

        // Global keyboard shortcuts
        document.addEventListener('keydown', (e) => {
            // Only active when Spec Builder tab is visible
            const specTab = document.getElementById('content-spec-builder');
            if (!specTab || !specTab.classList.contains('active')) return;

            // Ctrl+Enter → Send message handled in textarea listener

            // Escape → Close any open modal
            if (e.key === 'Escape') {
                closeContextSearch();
                const contextConfirm = document.querySelector('.context-search-modal');
                if (contextConfirm) contextConfirm.remove();
            }

            // Ctrl+Shift+N → New Session
            if (e.key === 'N' && e.ctrlKey && e.shiftKey) {
                e.preventDefault();
                resetSession();
            }

            // Ctrl+Shift+E → Export
            if (e.key === 'E' && e.ctrlKey && e.shiftKey) {
                e.preventDefault();
                exportConversation();
            }

            // Ctrl+Shift+A → Add context
            if (e.key === 'A' && e.ctrlKey && e.shiftKey) {
                e.preventDefault();
                openContextSearch();
            }

            // Focus chat input on any printable key if not in a modal
            if (!e.ctrlKey && !e.metaKey && !e.altKey && e.key.length === 1) {
                const modalOpen = document.getElementById('context-search-modal') || 
                                 document.querySelector('.context-search-modal') ||
                                 (dom.settingsModal && dom.settingsModal.style.display === 'flex');
                if (document.activeElement !== dom.chatInput && !modalOpen) {
                    dom.chatInput.focus();
                }
            }
        });
    }

    // ============================================
    // MODEL LOADING
    // ============================================
    // @spec-link [[mechanic_webui_llm_model_list]]
    async function loadModels() {
        try {
            const headers = {};
            const method = state.keys.access_method || 'gemini';
            if (method === 'openai' && state.keys.openai_url) {
                headers['X-LLM-Base-URL'] = state.keys.openai_url;
                if (state.keys.openai) headers['X-LLM-Key'] = state.keys.openai;
            }

            const resp = await fetch('/api/llm/models', { headers });
            if (!resp.ok) return;
            const data = await resp.json();
            if (!data.models || !dom.modelSelect) return;

            // Track providers
            const providersSet = new Set();
            data.models.forEach(m => {
                if (m.provider) providersSet.add(m.provider);
            });
            state.providers = Array.from(providersSet);

            // Populate dropdown
            const currentValue = dom.modelSelect.value;
            dom.modelSelect.innerHTML = '';
            
            // Group by provider
            const grouped = {};
            data.models.forEach(m => {
                if (!grouped[m.provider]) grouped[m.provider] = [];
                grouped[m.provider].push(m);
            });

            Object.keys(grouped).forEach(provider => {
                const group = document.createElement('optgroup');
                group.label = provider.charAt(0).toUpperCase() + provider.slice(1);
                
                grouped[provider].forEach(m => {
                    const opt = document.createElement('option');
                    opt.value = m.id;
                    opt.dataset.provider = provider;
                    
                    const isExhausted = state.exhaustedModels.has(m.id);
                    opt.textContent = (m.display_name || m.id) + (isExhausted ? ' ⚠️ (Quota Exceeded)' : '');
                    
                    if (m.id === data.default) {
                        opt.selected = true;
                    }
                    group.appendChild(opt);
                });
                dom.modelSelect.appendChild(group);
            });

            if (currentValue) {
                const exists = Array.from(dom.modelSelect.options).some(o => o.value === currentValue);
                if (exists) dom.modelSelect.value = currentValue;
            }
            
            // Refresh settings UI if open
            if (dom.settingsModal && dom.settingsModal.style.display === 'flex') {
                renderSettingsFields();
            }
        } catch (err) {
            console.warn('Failed to load models:', err);
        }
    }

    // ============================================
    // KEY MANAGEMENT
    // ============================================
    function loadKeys() {
        const saved = localStorage.getItem('atd_llm_keys');
        if (saved) {
            try {
                state.keys = JSON.parse(saved);
            } catch (e) {
                console.error('Failed to parse saved keys');
            }
        }
    }

    function openSettings() {
        if (!state.keys.access_method) state.keys.access_method = 'gemini';
        
        dom.accessMethod.value = state.keys.access_method;
        dom.geminiKeyInput.value = state.keys.gemini || '';
        dom.openaiKeyInput.value = state.keys.openai || '';
        dom.openaiUrlInput.value = state.keys.openai_url || '';
        
        // Trigger toggle
        dom.accessMethod.dispatchEvent(new Event('change'));
        
        dom.settingsModal.style.display = 'flex';
    }

    function saveSettings() {
        state.keys.access_method = dom.accessMethod.value;
        state.keys.gemini = dom.geminiKeyInput.value.trim();
        state.keys.openai = dom.openaiKeyInput.value.trim();
        state.keys.openai_url = dom.openaiUrlInput.value.trim();
        
        localStorage.setItem('atd_llm_keys', JSON.stringify(state.keys));
        dom.settingsModal.style.display = 'none';
        
        // Refresh models with new keys/URL
        loadModels();

        // Show a small success notice in chat
        renderActionInChat({ action: 'SETTINGS', atom_id: 'keys', impact_summary: 'LLM settings updated' }, 'accepted');
    }

    // ============================================
    // MESSAGE SENDING
    // ============================================
    // @spec-link [[ui_webui_spec_builder]]
    async function sendMessage() {
        const text = dom.chatInput.value.trim();
        if (!text || state.isLoading) return;

        // Check if there's new ATD context to confirm
        const newContext = state.atdContext.filter(atd => !state.atdContextSent.has(atd.id));
        if (newContext.length > 0) {
            const confirmed = await showContextConfirmation(newContext);
            if (!confirmed) return;
        }

        // Hide welcome screen
        if (dom.welcomeScreen) {
            dom.welcomeScreen.style.display = 'none';
        }

        // Add user message to state and render
        state.messages.push({ role: 'user', content: text });
        renderMessage('user', text, newContext);

        // Clear input
        dom.chatInput.value = '';
        autoGrowTextarea();

        // Show typing indicator
        state.isLoading = true;
        dom.sendBtn.disabled = true;
        dom.sendBtn.classList.add('loading');
        const typingEl = showTypingIndicator();

        try {
            // Build request payload
            // Note: sendMessage already filtered newContext above, but we can re-filter or just use the same logic
            const contextToAttach = state.atdContext.filter(atd => !state.atdContextSent.has(atd.id));

            const payload = {
                messages: state.messages,
                model: dom.modelSelect ? dom.modelSelect.value : 'gemini-3.1-flash-lite-preview',
                atd_context: contextToAttach,
                actions: state.actions,
                omit_history: !state.chatModeEnabled,
            };

            const selectedOption = dom.modelSelect.options[dom.modelSelect.selectedIndex];
            const provider = selectedOption ? selectedOption.dataset.provider : 'gemini';
            const headers = { 'Content-Type': 'application/json' };
            
            // Inject key from settings
            const method = state.keys.access_method || 'gemini';
            if (method === 'gemini' && state.keys.gemini) {
                headers['X-LLM-Key-Gemini'] = state.keys.gemini;
            } else if (method === 'openai' && provider !== 'gemini') {
                // For any non-gemini model, use the custom OpenAI settings if in OpenAI mode
                if (state.keys.openai) headers['X-LLM-Key'] = state.keys.openai;
                if (state.keys.openai_url) headers['X-LLM-Base-URL'] = state.keys.openai_url;
            } else if (state.keys[provider]) {
                // Fallback to provider-specific key if available
                headers[`X-LLM-Key-${provider.charAt(0).toUpperCase() + provider.slice(1)}`] = state.keys[provider];
            }

            const response = await fetch('/api/llm/chat', {
                method: 'POST',
                headers: headers,
                body: JSON.stringify(payload),
            });

            // Remove typing indicator
            typingEl.remove();

            if (!response.ok) {
                const err = await response.json();
                if (response.status === 429) {
                    // Track exhausted model
                    const modelId = payload.model;
                    state.exhaustedModels.add(modelId);
                    // Refresh dropdown to show warning
                    loadModels();
                    
                    renderQuotaErrorMessage(err.error || 'Quota exceeded');
                } else {
                    renderErrorMessage(err.error || `HTTP ${response.status}`, text);
                }
                
                state.isLoading = false;
                dom.sendBtn.disabled = false;
                return;
            }

            const data = await response.json();

            // Debug logging
            if (dom.debugCheck && dom.debugCheck.checked) {
                console.log("RAW GEMINI RESPONSE:", data);
            }

            // Add model message to state
            state.messages.push({ role: 'model', content: data.message });
            
            // Update total tokens
            if (data.usage) {
                state.totalTokens += (data.usage.total_tokens || 0);
                updateTotalTokensDisplay();
            }

            renderMessage('model', data.message, [], data.usage);

            // Handle proposals
            if (data.proposals && data.proposals.length > 0) {
                handleProposals(data.proposals);
            }

            // Increment exchange count
            state.exchangeCount++;
            checkContextDrift();

            // Mark sent ATD context as already injected
            state.atdContext.forEach(atd => state.atdContextSent.add(atd.id));
            renderContextBar(); // Re-render to show checkmarks

            // Recommend manual attachment if model mentioned IDs
            autoRecommendContext(data.message);

        } catch (error) {
            typingEl.remove();
            renderErrorMessage(error.message, text);
        } finally {
            state.isLoading = false;
            dom.sendBtn.disabled = false;
            dom.sendBtn.classList.remove('loading');
            dom.chatInput.focus();
        }
    }

    function buildChatPayload() {
        // Only include ATD context not yet sent
        const newContext = state.atdContext.filter(atd => !state.atdContextSent.has(atd.id));

        return {
            messages: state.messages,
            model: dom.modelSelect ? dom.modelSelect.value : 'gemini-3.1-flash-lite-preview',
            atd_context: newContext,
            actions: state.actions,
            omit_history: !state.chatModeEnabled,
        };
    }

    // ============================================
    // RENDERING
    // ============================================
    // @spec-link [[ui_webui_spec_builder]]
    function renderActionInChat(proposal, action) {
        const actionEl = document.createElement('div');
        actionEl.className = 'chat-action-notice';

        const actionVerb = action === 'accepted' ? 'Accepted' : 'Rejected';
        const actionIcon = action === 'accepted' ? '✅' : '❌';
        const actionClass = action === 'accepted' ? 'action-accept' : 'action-reject';

        actionEl.innerHTML = `
            <span class="${actionClass}">${actionIcon} ${actionVerb}:</span>
            <span class="action-detail">
                <strong>${proposal.action}</strong> @${proposal.atom_id}
                ${proposal.content?.human_name ? `— ${escapeHtml(proposal.content.human_name)}` : ''}
            </span>
        `;

        dom.messagesContainer.appendChild(actionEl);
        scrollToBottom();
    }

    // @spec-link [[ui_webui_spec_builder]]
    function renderMessage(role, content, contextInfo = [], usage = null) {
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

            // @spec-link [[requirement_webui_token_transparency]]
            // Add usage info
            if (usage) {
                const usageDiv = document.createElement('div');
                usageDiv.className = 'chat-msg-usage';
                usageDiv.innerHTML = `
                    <span class="usage-item"><strong>Tokens:</strong> ${usage.total_tokens || 0}</span>
                    <span class="usage-item">(${usage.prompt_tokens || 0} prompt / ${usage.candidates_tokens || 0} resp)</span>
                `;
                bubble.appendChild(usageDiv);
            }
        } else {
            bubble.textContent = content;

            // Add context attachment links if any
            if (contextInfo.length > 0) {
                const badge = document.createElement('div');
                badge.className = 'msg-context-badge';
                badge.innerHTML = `📎 Attached: ${contextInfo.map(a => 
                    `<a href="/?atom=${a.id}" target="_blank" class="context-link">${escapeHtml(a.id)}</a>`
                ).join(', ')}`;
                bubble.appendChild(badge);
            }
        }
        
        // Add timestamp
        const time = document.createElement('div');
        time.className = 'msg-time';
        time.textContent = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        bubble.appendChild(time);

        msgEl.appendChild(avatar);
        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);
        scrollToBottom();
    }

    function renderQuotaErrorMessage(rawError) {
        // Example: Quota exceeded for metric: ..., limit: 20, model: gemini-2.5-flash
        const modelMatch = rawError.match(/model: ([^\s,\]]+)/);
        const limitMatch = rawError.match(/limit: (\d+)/);
        
        const currentModel = modelMatch ? modelMatch[1] : (dom.modelSelect.value || 'current model');
        const limit = limitMatch ? limitMatch[1] : 'unknown';

        const msgEl = document.createElement('div');
        msgEl.className = 'chat-msg error system-error';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';
        bubble.style.borderColor = 'var(--accent)';
        
        let html = `
            <div style="font-weight: 600; color: var(--color-red-light); margin-bottom: 8px;">⚠️ Quota Exceeded</div>
            <div style="margin-bottom: 12px; font-size: 13px;">
                Quota exceeded for the model <strong>${currentModel}</strong>, limit: <strong>${limit}</strong>.
            </div>
        `;

        // Suggest alternatives
        const alternatives = Array.from(dom.modelSelect.options)
            .map(opt => ({ id: opt.value, name: opt.textContent }))
            .filter(m => m.id !== currentModel && !m.id.includes(currentModel));

        if (alternatives.length > 0) {
            html += `
                <div style="font-size: 12px; margin-top: 10px;">
                    <div style="color: var(--text-muted); margin-bottom: 6px;">Try switching to an alternative model:</div>
                    <div style="display: flex; flex-wrap: wrap; gap: 6px;" id="error-suggestions"></div>
                </div>
            `;
        }

        bubble.innerHTML = html;
        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);

        // Add suggestion chips
        const suggestions = msgEl.querySelector('#error-suggestions');
        if (suggestions) {
            alternatives.slice(0, 3).forEach(alt => {
                const chip = document.createElement('button');
                chip.className = 'btn btn-outline btn-xs';
                chip.style.fontSize = '10px';
                chip.style.padding = '2px 8px';
                chip.textContent = alt.name;
                chip.onclick = () => {
                    dom.modelSelect.value = alt.id;
                    dom.chatInput.focus();
                    msgEl.remove();
                };
                suggestions.appendChild(chip);
            });
        }

        scrollToBottom();
        state.isLoading = false;
        if (dom.sendBtn) dom.sendBtn.disabled = false;
    }

    function renderErrorMessage(errorText, retryPayload) {
        const msgEl = document.createElement('div');
        msgEl.className = 'chat-msg model';

        const avatar = document.createElement('div');
        avatar.className = 'msg-avatar';
        avatar.style.background = 'var(--color-red)';
        avatar.textContent = '⚠️';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';
        bubble.style.borderColor = 'rgba(211, 47, 47, 0.3)';

        bubble.innerHTML = `
            <div style="color: var(--color-red-light); font-weight: 500; margin-bottom: 4px;">Error</div>
            <div style="font-size: 13px; color: var(--text-muted);">${escapeHtml(errorText)}</div>
        `;

        if (retryPayload) {
            const retryBtn = document.createElement('button');
            retryBtn.className = 'btn btn-outline btn-sm';
            retryBtn.style.marginTop = '8px';
            retryBtn.textContent = '🔄 Retry';
            retryBtn.addEventListener('click', () => {
                msgEl.remove();
                // We don't remove from state.messages here because it failed to reach the model
                dom.chatInput.value = retryPayload;
                sendMessage();
            });
            bubble.appendChild(retryBtn);
        }

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
    // @spec-link [[ui_webui_spec_builder]]
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

    // @spec-link [[ui_webui_spec_builder]]
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

    // @spec-link [[ui_webui_spec_builder]]
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

        // Show visual notice in chat
        renderActionInChat(proposal, action);

        // Update card UI
        cardEl.className = `proposal-card ${action}`;

        // If accepted, apply the proposal
        if (action === 'accepted') {
            try {
                const resp = await fetch('/api/llm/apply-proposal', {
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
    // ATD CONTEXT MANAGEMENT
    // ============================================

    // @spec-link [[rule_webui_context_history_management]]
    function openContextSearch() {
        const overlay = document.createElement('div');
        overlay.className = 'context-search-modal';
        overlay.id = 'context-search-modal';

        const panel = document.createElement('div');
        panel.className = 'context-search-panel';

        const input = document.createElement('input');
        input.className = 'context-search-input';
        input.placeholder = 'Search atoms by name or ID...';
        input.autofocus = true;

        const results = document.createElement('div');
        results.className = 'context-search-results';

        panel.append(input, results);
        overlay.appendChild(panel);
        document.body.appendChild(overlay);

        // Close on background click
        overlay.addEventListener('click', (e) => {
            if (e.target === overlay) closeContextSearch();
        });

        // Close on Escape
        const escHandler = (e) => {
            if (e.key === 'Escape') {
                closeContextSearch();
                document.removeEventListener('keydown', escHandler);
            }
        };
        document.addEventListener('keydown', escHandler);

        // Search as user types
        let debounceTimer;
        input.addEventListener('input', () => {
            clearTimeout(debounceTimer);
            debounceTimer = setTimeout(() => searchAtoms(input.value, results), 200);
        });

        // Load initial results
        searchAtoms('', results);

        // Focus after a tick (ensure overlay is rendered)
        setTimeout(() => input.focus(), 50);
    }

    function closeContextSearch() {
        const modal = document.getElementById('context-search-modal');
        if (modal) modal.remove();
    }

    // @spec-link [[rule_webui_context_history_management]]
    async function searchAtoms(query, resultsContainer) {
        try {
            const resp = await fetch(`/api/gemini/atoms?q=${encodeURIComponent(query)}`);
            const atoms = await resp.json();

            resultsContainer.innerHTML = '';

            if (!atoms || atoms.length === 0) {
                resultsContainer.innerHTML = '<div style="padding: 20px; text-align: center; color: var(--text-muted);">No atoms found</div>';
                return;
            }

            atoms.forEach(atom => {
                // Skip already attached atoms
                const alreadyAttached = state.atdContext.some(a => a.id === atom.id);

                const item = document.createElement('div');
                item.className = 'context-search-item';
                if (alreadyAttached) item.style.opacity = '0.4';

                item.innerHTML = `
                    <span class="item-type">${atom.type || '?'}</span>
                    <span class="item-name">${escapeHtml(atom.human_name || atom.id)}</span>
                    <span class="item-id">${atom.id}</span>
                `;

                if (!alreadyAttached) {
                    item.addEventListener('click', async () => {
                        // For adding to context, we might want full content now or later.
                        // Let's fetch full atom data to be ready.
                        try {
                            const fullResp = await fetch(`/api/gemini/atom/${atom.id}`);
                            const fullAtom = await fullResp.json();
                            addAtdContext(fullAtom);
                            closeContextSearch();
                        } catch (err) {
                            alert("Failed to load atom details");
                        }
                    });
                }

                resultsContainer.appendChild(item);
            });
        } catch (err) {
            resultsContainer.innerHTML = '<div style="padding: 20px; color: var(--color-red-light);">Failed to search atoms</div>';
        }
    }

    // @spec-link [[rule_webui_context_history_management]]
    function addAtdContext(atom) {
        // Don't add duplicates
        if (state.atdContext.some(a => a.id === atom.id)) return;

        state.atdContext.push(atom);
        renderContextBar();
    }

    function removeAtdContext(atomId) {
        state.atdContext = state.atdContext.filter(a => a.id !== atomId);
        renderContextBar();
    }

    function renderContextBar() {
        // ALWAYS show the bar if the button is needed
        dom.contextBar.style.display = 'flex';
        dom.contextChips.innerHTML = '';

        if (state.atdContext.length === 0) {
            dom.contextChips.innerHTML = '<span style="color:var(--text-muted); font-size:11px; opacity:0.5;">No context attached</span>';
        }

        state.atdContext.forEach(atom => {
            const chip = document.createElement('span');
            chip.className = 'context-chip';

            const alreadySent = state.atdContextSent.has(atom.id);
            if (alreadySent) chip.style.opacity = '0.5';

            chip.innerHTML = `
                <span>${escapeHtml(atom.human_name || atom.id)}</span>
                ${!alreadySent ? '<span class="chip-remove" title="Remove">×</span>' : '<span title="Already sent" style="font-size:10px; margin-left:4px;">✓</span>'}
            `;

            if (!alreadySent) {
                chip.querySelector('.chip-remove').addEventListener('click', (e) => {
                    e.stopPropagation();
                    removeAtdContext(atom.id);
                });
            }

            dom.contextChips.appendChild(chip);
        });
    }

    // ============================================
    // CONTEXT CONFIRMATION
    // ============================================

    function showContextConfirmation(newContext) {
        return new Promise((resolve) => {
            const overlay = document.createElement('div');
            overlay.className = 'context-search-modal'; // Reuse modal styles
            overlay.style.zIndex = '101';

            const panel = document.createElement('div');
            panel.className = 'context-search-panel';
            panel.style.width = '420px';

            panel.innerHTML = `
                <div style="padding: 24px;">
                    <h3 style="font-size: 16px; margin-bottom: 12px; color: var(--text-main);">📎 ATD Context Injection</h3>
                    <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 16px; line-height: 1.5;">
                        The following atoms will be sent as context with your message. Their full content will be included once.
                    </p>
                    <div id="confirm-context-list" style="max-height: 250px; overflow-y: auto; margin-bottom: 12px;"></div>
                    <div style="display: flex; gap: 12px; margin-top: 20px;">
                        <button class="btn btn-primary" id="confirm-send" style="flex:1;">Send with Context</button>
                        <button class="btn btn-outline" id="confirm-cancel" style="flex:1;">Cancel</button>
                    </div>
                </div>
            `;

            overlay.appendChild(panel);
            document.body.appendChild(overlay);

            // Render context items
            const list = panel.querySelector('#confirm-context-list');
            newContext.forEach(atom => {
                const item = document.createElement('div');
                item.style.cssText = 'display: flex; align-items: center; gap: 12px; padding: 10px 12px; border-radius: 8px; background: var(--bg-card); border: 1px solid var(--border); margin-bottom: 8px;';
                item.innerHTML = `
                    <input type="checkbox" checked data-atom-id="${atom.id}" style="accent-color: var(--accent); width: 16px; height: 16px; cursor: pointer;">
                    <div style="display: flex; flex-direction: column;">
                        <span style="font-size: 10px; color: var(--accent); font-weight: 700; text-transform: uppercase;">${atom.type || ''}</span>
                        <span style="font-size: 13px; font-weight: 500;">${escapeHtml(atom.human_name || atom.id)}</span>
                    </div>
                `;
                list.appendChild(item);
            });

            panel.querySelector('#confirm-send').addEventListener('click', () => {
                const unchecked = list.querySelectorAll('input:not(:checked)');
                unchecked.forEach(cb => {
                    const id = cb.dataset.atomId;
                    removeAtdContext(id);
                });
                overlay.remove();
                resolve(true);
            });

            panel.querySelector('#confirm-cancel').addEventListener('click', () => {
                overlay.remove();
                resolve(false);
            });

            overlay.addEventListener('click', (e) => {
                if (e.target === overlay) {
                    overlay.remove();
                    resolve(false);
                }
            });
        });
    }

    function autoRecommendContext(modelMessage) {
        const atomIdPattern = /\b([a-z]+_[a-z_]+)\b/g;
        const matches = [...new Set(modelMessage.match(atomIdPattern) || [])];

        const candidates = matches.filter(id => id.length > 5);
        if (candidates.length === 0) return;

        candidates.forEach(async id => {
            if (state.atdContext.some(a => a.id === id)) return;
            if (state.atdContextSent.has(id)) return;

            if (window.Atoms && window.Atoms[id]) {
                const atom = window.Atoms[id];
                showRecommendationChip(atom);
            }
        });
    }

    function showRecommendationChip(atom) {
        const msgEl = document.createElement('div');
        msgEl.className = 'chat-msg model recommendation';
        msgEl.style.animation = 'msgFadeIn 0.5s ease';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';
        bubble.style.background = 'rgba(76, 139, 245, 0.05)';
        bubble.style.border = '1px dashed var(--accent)';
        bubble.style.fontSize = '12px';
        bubble.innerHTML = `
            <div style="display: flex; align-items: center; gap: 8px;">
                <span>💡 Suggested context: <strong>${escapeHtml(atom.human_name || atom.id)}</strong></span>
                <button class="btn btn-primary btn-xs" style="padding: 2px 8px; font-size: 10px;">Attach</button>
            </div>
        `;

        bubble.querySelector('button').addEventListener('click', async () => {
            try {
                const resp = await fetch(`/api/gemini/atom/${atom.id}`);
                const fullAtom = await resp.json();
                addAtdContext(fullAtom);
                msgEl.remove();
            } catch (err) {
                alert("Failed to load suggested atom");
            }
        });

        msgEl.appendChild(bubble);
        dom.messagesContainer.appendChild(msgEl);
        scrollToBottom();
    }

    // ============================================
    // CONTEXT DRIFT CHECK
    // ============================================
    // @spec-link [[rule_webui_context_history_management]]
    function checkContextDrift() {
        const count = state.exchangeCount;

        // First warning at 10 exchanges
        if (count === 10) {
            renderDriftWarning(
                'caution',
                `You've had ${count} exchanges in this session. Context drift may reduce specification quality.`,
                'Consider starting a new session to maintain accuracy.'
            );
        }

        // Stronger warning at 15
        if (count === 15) {
            renderDriftWarning(
                'warning',
                `${count} exchanges — context window is getting crowded.`,
                'Model accuracy may degrade. A new session is strongly recommended.'
            );
        }

        // Hard warning at 20
        if (count === 20) {
            renderDriftWarning(
                'critical',
                `${count} exchanges — approaching context limits.`,
                'Starting a new session is essential to avoid specification errors.'
            );
        }

        // Update exchange counter badge (if visible)
        updateExchangeBadge();
    }

    function renderDriftWarning(severity, mainText, subText) {
        const warning = document.createElement('div');
        warning.className = `drift-warning drift-${severity}`;

        const icon = severity === 'critical' ? '🔴' : severity === 'warning' ? '🟡' : '⚠️';

        warning.innerHTML = `
            <span class="drift-icon">${icon}</span>
            <div class="drift-text" style="flex: 1;">
                <strong>${mainText}</strong>
                <div style="font-size: 12px; margin-top: 2px; opacity: 0.8;">${subText}</div>
            </div>
            <button class="btn btn-outline btn-sm drift-restart-btn" style="white-space: nowrap;">🔄 New Session</button>
        `;

        warning.querySelector('.drift-restart-btn').addEventListener('click', () => {
            dom.newSessionBtn.click();
        });

        dom.messagesContainer.appendChild(warning);
        scrollToBottom();
    }

    function updateExchangeBadge() {
        let badge = document.getElementById('exchange-badge');
        if (!badge) {
            badge = document.createElement('span');
            badge.id = 'exchange-badge';
            badge.className = 'exchange-badge';
            const header = document.querySelector('.chat-header .chat-controls');
            if (header) header.prepend(badge);
        }

        const count = state.exchangeCount;
        badge.textContent = `${count}/10`;

        if (count >= 15) {
            badge.className = 'exchange-badge critical';
        } else if (count >= 10) {
            badge.className = 'exchange-badge warning';
        } else {
            badge.className = 'exchange-badge';
        }
    }

    // ============================================
    // UI HELPERS
    // ============================================

    function handleChatModeChange(e) {
        const enabled = e.target.checked;
        
        // If we have history, warn the user
        if (state.messages.length > 0) {
            const confirmed = confirm(`Changing to ${enabled ? 'Chat' : 'Single-shot'} Mode will clear the current session history. Proceed?`);
            if (!confirmed) {
                // Revert toggle
                dom.chatModeCheck.checked = !enabled;
                return;
            }
            // If confirmed, reset
            resetSession(true); // Skip confirm inside resetSession
        }
        
        state.chatModeEnabled = enabled;
        console.log(`Chat Mode ${enabled ? 'Enabled' : 'Disabled'} (Single-shot)`);
    }

    // @spec-link [[ui_webui_spec_builder]]
    function exportConversation() {
        if (state.messages.length === 0) {
            alert('No conversation to export.');
            return;
        }

        let md = `# ATD Spec Builder Session\n`;
        md += `**Date:** ${new Date().toISOString()}\n`;
        md += `**Exchanges:** ${state.exchangeCount}\n`;
        md += `**Total Tokens:** ${state.totalTokens.toLocaleString()}\n`;
        
        const accepted = state.actions.filter(a => a.action === 'ACCEPTED').length;
        const rejected = state.actions.filter(a => a.action === 'REJECTED').length;
        md += `**Proposals:** ${state.actions.length} (${accepted} accepted, ${rejected} rejected)\n\n`;
        md += `---\n\n`;

        // Messages
        md += `## Conversation\n\n`;
        state.messages.forEach(msg => {
            const role = msg.role === 'user' ? '**User**' : '**Spec Builder**';
            md += `### ${role}\n\n${msg.content}\n\n`;
        });

        // Proposals summary
        if (state.actions.length > 0) {
            md += `---\n\n## Actions History\n\n`;
            md += `| Action | Atom ID | Summary |\n`;
            md += `|--------|---------|---------|\n`;
            state.actions.forEach(a => {
                md += `| ${a.action} | \`${a.atom_id}\` | ${a.summary} |\n`;
            });
        }

        // ATD Context
        if (state.atdContext.length > 0) {
            md += `\n---\n\n## ATD Context Used\n\n`;
            state.atdContext.forEach(a => {
                md += `- \`${a.id}\` — ${a.human_name || a.id}\n`;
            });
        }

        // Download
        const blob = new Blob([md], { type: 'text/markdown' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `spec-session-${new Date().toISOString().slice(0, 10)}.md`;
        a.click();
        URL.revokeObjectURL(url);
    }

    function updateTotalTokensDisplay() {
        if (dom.totalTokensDisplay) {
            dom.totalTokensDisplay.textContent = state.totalTokens.toLocaleString();
        }
    }

    // ============================================
    // SESSION MANAGEMENT
    // ============================================
    // @spec-link [[ui_webui_spec_builder]]
    function resetSession(skipConfirm = false) {
        if (!skipConfirm && state.messages.length > 0 && !confirm('Start a new session? All conversation history will be cleared.')) return;

        state.messages = [];
        state.proposalGroups = {};
        state.actions = [];
        state.atdContext = [];
        state.atdContextSent.clear();
        state.exchangeCount = 0;
        state.totalTokens = 0;
        state.sessionStart = new Date();
        updateTotalTokensDisplay();

        // Reset UI
        dom.messagesContainer.innerHTML = '';
        
        // Re-create welcome screen
        const welcome = document.createElement('div');
        welcome.className = 'chat-welcome';
        welcome.innerHTML = `
            <div class="welcome-icon">🏗️</div>
            <h3>ATD Spec Builder</h3>
            <p>Describe the feature or system you want to specify. I'll help you create properly structured ATD atoms following the atomic documentation rules.</p>
            <div class="welcome-suggestions">
                <button class="suggestion-chip" data-prompt="I want to define a new feature for">Define a feature</button>
                <button class="suggestion-chip" data-prompt="Help me decompose this requirement:">Decompose a requirement</button>
                <button class="suggestion-chip" data-prompt="Review and improve these existing atoms:">Review existing atoms</button>
            </div>
        `;
        dom.messagesContainer.appendChild(welcome);

        // Re-wire suggestion chips
        welcome.querySelectorAll('.suggestion-chip').forEach(chip => {
            chip.addEventListener('click', () => {
                dom.chatInput.value = chip.dataset.prompt + ' ';
                dom.chatInput.focus();
                autoGrowTextarea();
                welcome.style.display = 'none';
            });
        });

        // Reset proposals
        dom.proposalsList.innerHTML = '<div class="proposals-empty"><p>Proposals from the AI will appear here for your review.</p><p class="proposals-hint">You can accept ✓ or reject ✗ each proposal individually.</p></div>';
        dom.proposalCount.textContent = '0 pending';
        if (dom.proposalsActions) dom.proposalsActions.style.display = 'none';

        // Reset context bar
        dom.contextBar.style.display = 'none';
        dom.contextChips.innerHTML = '';

        // Reset exchange badge
        const badge = document.getElementById('exchange-badge');
        if (badge) badge.remove();
    }

    // @spec-link [[api_webui_health_stats]]
    async function checkBackendHealth() {
        try {
            const resp = await fetch('/api/info');
            if (!resp.ok) throw new Error('Backend unreachable');
            return true;
        } catch (err) {
            renderErrorMessage('Cannot connect to WebUI backend. Please ensure the server is running.');
            return false;
        }
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
        checkBackendHealth,
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
