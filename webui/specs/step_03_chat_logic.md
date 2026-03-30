# Step 3: Chat Logic & Message Flow

## Objective
Create `spec-builder.js` with the core chat logic: sending messages to `/api/gemini/chat`, rendering responses, managing conversation history, and auto-growing the textarea.

## Prerequisites
- Step 1: `/api/gemini/chat` endpoint exists
- Step 2: HTML shell exists with `#chat-messages`, `#chat-input`, `#btn-send`, `.suggestion-chip`, `#model-select` elements
- `marked.js` is loaded globally (exists in index.html)

## File to Create

### `webui/static/spec-builder.js`

```javascript
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
        proposals: [],      // GeminiProposal[] from all responses
        actions: [],        // { proposal_id, atom_id, action, summary }
        atdContext: [],      // Atom objects attached as context
        atdContextSent: new Set(),  // IDs already sent to avoid redundancy
        exchangeCount: 0,   // Number of user<->model roundtrips
        isLoading: false,
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
    // PROPOSALS (basic rendering — expanded in Step 4)
    // ============================================
    function handleProposals(proposals) {
        proposals.forEach(p => {
            p._id = `proposal_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
            p._status = 'pending'; // pending | accepted | rejected
            state.proposals.push(p);
        });
        renderProposalsList();
    }

    function renderProposalsList() {
        // Placeholder — fully implemented in Step 4
        const pending = state.proposals.filter(p => p._status === 'pending');
        dom.proposalCount.textContent = `${pending.length} pending`;

        if (state.proposals.length === 0) return;

        // Clear empty state
        const emptyEl = dom.proposalsList.querySelector('.proposals-empty');
        if (emptyEl) emptyEl.remove();

        // Show action buttons
        if (pending.length > 0 && dom.proposalsActions) {
            dom.proposalsActions.style.display = 'flex';
        }

        // Render each new proposal card (basic version)
        proposals_to_render: for (const p of state.proposals) {
            if (document.querySelector(`[data-proposal-id="${p._id}"]`)) continue;

            const card = document.createElement('div');
            card.className = 'proposal-card';
            card.dataset.proposalId = p._id;

            card.innerHTML = `
                <div class="proposal-card-header">
                    <span class="proposal-action-badge ${p.action}">${p.action}</span>
                    <div class="proposal-card-actions">
                        <button class="proposal-btn accept" title="Accept">✓</button>
                        <button class="proposal-btn reject" title="Reject">✗</button>
                        <button class="proposal-btn preview" title="Preview">👁</button>
                    </div>
                </div>
                <div class="proposal-atom-id">@${p.atom_id}</div>
                <div class="proposal-atom-name">${escapeHtml(p.content?.human_name || p.atom_id)}</div>
                <div class="proposal-impact">${escapeHtml(p.impact_summary || '')}</div>
            `;

            dom.proposalsList.appendChild(card);
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
        state.proposals = [];
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
        getProposals: () => state.proposals,
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
```

## Expected Outcome

After this step:
1. User can type a message and press `Ctrl+Enter` or click Send
2. The message appears in the chat with a user avatar bubble
3. A typing indicator (three bouncing dots) appears while waiting
4. The model's response renders as markdown in a model avatar bubble
5. If the response contains proposals, basic proposal cards appear in the right panel
6. Conversation history is maintained in-memory and sent with each request
7. Auto-growing textarea expands as user types multiline
8. New Session button clears all state
9. Suggestion chips populate the input on click
10. Error responses show inline with red styling

## Key Design Decisions

- **IIFE pattern**: Encapsulates all Spec Builder logic; exposes `window.SpecBuilder` for cross-module access
- **State object**: Single source of truth for messages, proposals, actions, context
- **`atdContextSent` dedup set**: Prevents re-sending the same atom's full content on subsequent messages
- **Basic proposal rendering**: Step 4 will enhance with accept/reject handlers and diff preview
