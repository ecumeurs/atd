# Step 5: ATD Context Injection & Pre-Send Confirmation

## Objective
Allow users to attach ATD atoms as context before sending a message. Show recommended atoms based on conversation topic. Ask for confirmation before injecting ATD content into the Gemini request.

## Prerequisites
- Steps 1-4 completed
- `GET /api/gemini/atoms?q=search` endpoint exists (Step 1)
- `state.atdContext` and `state.atdContextSent` exist (Step 3)
- Context bar HTML exists: `#chat-context-bar`, `#context-chips`, `#btn-add-context` (Step 2)

## Changes to `spec-builder.js`

### 1. ATD Context Search Modal

Add the search modal open/close logic:

```javascript
// ============================================
// ATD CONTEXT MANAGEMENT
// ============================================

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
                item.addEventListener('click', () => {
                    addAtdContext(atom);
                    closeContextSearch();
                });
            }

            resultsContainer.appendChild(item);
        });
    } catch (err) {
        resultsContainer.innerHTML = '<div style="padding: 20px; color: var(--color-red-light);">Failed to search atoms</div>';
    }
}

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
    if (state.atdContext.length === 0) {
        dom.contextBar.style.display = 'none';
        return;
    }

    dom.contextBar.style.display = 'flex';
    dom.contextChips.innerHTML = '';

    state.atdContext.forEach(atom => {
        const chip = document.createElement('span');
        chip.className = 'context-chip';

        const alreadySent = state.atdContextSent.has(atom.id);
        if (alreadySent) chip.style.opacity = '0.5';

        chip.innerHTML = `
            <span>${escapeHtml(atom.human_name || atom.id)}</span>
            ${!alreadySent ? '<span class="chip-remove" title="Remove">×</span>' : '<span title="Already sent" style="font-size:10px">✓</span>'}
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
```

### 2. Wire the add context button

In `init()`, add:

```javascript
if (dom.addContextBtn) {
    dom.addContextBtn.addEventListener('click', openContextSearch);
}
```

### 3. Pre-send confirmation for ATD context

Modify the `sendMessage` function to show a confirmation dialog before sending if new ATD context is attached:

```javascript
async function sendMessage() {
    const text = dom.chatInput.value.trim();
    if (!text || state.isLoading) return;

    // Check if there's new ATD context to confirm
    const newContext = state.atdContext.filter(atd => !state.atdContextSent.has(atd.id));
    if (newContext.length > 0) {
        const confirmed = await showContextConfirmation(newContext);
        if (!confirmed) return;
    }

    // ... rest of sendMessage from Step 3 unchanged ...
}
```

### 4. Context confirmation dialog

```javascript
function showContextConfirmation(newContext) {
    return new Promise((resolve) => {
        const overlay = document.createElement('div');
        overlay.className = 'context-search-modal'; // Reuse modal styles
        overlay.style.zIndex = '101';

        const panel = document.createElement('div');
        panel.className = 'context-search-panel';
        panel.style.width = '420px';

        panel.innerHTML = `
            <div style="padding: 20px;">
                <h3 style="font-size: 16px; margin-bottom: 12px;">📎 ATD Context Injection</h3>
                <p style="font-size: 13px; color: var(--text-muted); margin-bottom: 16px;">
                    The following atoms will be sent as context with your message. Their full content will be included once.
                </p>
                <div id="confirm-context-list" style="max-height: 250px; overflow-y: auto;"></div>
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
            item.style.cssText = 'display: flex; align-items: center; gap: 10px; padding: 8px 12px; border-radius: 6px; background: var(--bg-card); margin-bottom: 6px;';
            item.innerHTML = `
                <input type="checkbox" checked data-atom-id="${atom.id}" style="accent-color: var(--accent);">
                <span style="font-size: 10px; color: var(--accent); font-weight: 600;">${atom.type || ''}</span>
                <span style="font-size: 13px;">${escapeHtml(atom.human_name || atom.id)}</span>
            `;
            list.appendChild(item);
        });

        panel.querySelector('#confirm-send').addEventListener('click', () => {
            // Remove unchecked atoms from context for this send
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
```

### 5. Auto-recommend relevant atoms

After each model response, check if any mentioned atom IDs exist in the atom list and suggest attaching them:

```javascript
function autoRecommendContext(modelMessage) {
    // Simple heuristic: look for atom-like patterns in the response
    const atomIdPattern = /\b([a-z]+_[a-z_]+)\b/g;
    const matches = modelMessage.match(atomIdPattern) || [];

    // Check which matches correspond to real atoms not yet in context
    const suggestions = [];
    matches.forEach(async id => {
        if (state.atdContext.some(a => a.id === id)) return;
        if (state.atdContextSent.has(id)) return;
        // Would need to verify against the atoms list
        // For now, just note it in the chat
    });
}
```

## Backend Changes

### `GET /api/gemini/atoms` enhancement

If not already done in Step 1, ensure the endpoint returns the full atom content for selected atoms. Add an endpoint to fetch full atom data:

```go
api.GET("/gemini/atom/:id", func(c *gin.Context) {
    id := c.Param("id")
    atom, exists := Atoms[id]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found"})
        return
    }
    c.JSON(http.StatusOK, gin.H{
        "id":         atom.ID,
        "human_name": atom.HumanName,
        "type":       atom.Type,
        "status":     atom.Status,
        "content":    atom.Content,
        "parents":    atom.Parents,
        "dependents": atom.Dependents,
    })
})
```

## Expected Outcome

1. Clicking the "+" button in the context bar opens a search modal
2. User can search atoms by name or ID with instant filtering
3. Clicking an atom adds it as a context chip in the context bar
4. Before sending a message with new context, a confirmation dialog shows which atoms will be attached
5. User can uncheck specific atoms to exclude them before sending
6. Already-sent atoms show a ✓ indicator and are grayed out
7. Context chips can be removed by clicking the × button
