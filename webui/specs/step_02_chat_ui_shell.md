# Step 2: Chat UI Shell (HTML + CSS Tab System)

## Objective
Add a "Spec Builder" tab to the existing WebUI that shows a two-column layout: chat panel (left) and proposals panel (right). No JS logic yet — this step is purely structural.

## Prerequisites
- Step 1 completed (backend endpoint exists)
- Existing files: `index.html` (169 lines), `styles.css` (745 lines), `app.js` (524 lines)

## Current HTML Structure (key parts)

The existing `index.html` has this layout:
```
<body>
  <div class="app-container">
    <header class="app-header">
      <div class="logo-area"> ... </div>
      <div class="header-tools"> ... buttons ... </div>
    </header>
    <div class="bulk-bar" id="bulk-bar"> ... </div>
    <main class="main-content">
      <section class="viz-section"> ... treemap + treeview ... </section>
      <aside class="details-panel" id="details-panel"> ... atom details ... </aside>
    </main>
  </div>
  <script src="/static/app.js"></script>
</body>
```

## Files to Modify

### 1. `webui/static/index.html`

#### A. Add tab switcher buttons in the header

Replace the `<div class="header-tools">` section. The current content is:
```html
<div class="header-tools">
    <button class="btn btn-outline" id="toggle-layout" title="Toggle Display Ratio">⬌</button>
    <button class="btn btn-outline" id="toggle-select-mode">Select Mode</button>
    <button class="btn btn-outline" id="refresh-btn">Refresh Data</button>
</div>
```

Replace with:
```html
<div class="header-tabs">
    <button class="tab-btn active" id="tab-explorer" data-tab="explorer">
        <span class="tab-icon">⊞</span> Explorer
    </button>
    <button class="tab-btn" id="tab-spec-builder" data-tab="spec-builder">
        <span class="tab-icon">💬</span> Spec Builder
    </button>
</div>
<div class="header-tools">
    <button class="btn btn-outline" id="toggle-layout" title="Toggle Display Ratio">⬌</button>
    <button class="btn btn-outline" id="toggle-select-mode">Select Mode</button>
    <button class="btn btn-outline" id="refresh-btn">Refresh Data</button>
</div>
```

#### B. Add the Spec Builder tab content

After the closing `</main>` tag of the existing content (but before `</div>` of `.app-container`), add a new section. Wrap the existing `<main>` and `<div class="bulk-bar">` with a container div:

The full structure should become:
```html
<!-- Explorer Tab -->
<div class="tab-content active" id="content-explorer">
    <div class="bulk-bar" id="bulk-bar" style="display: none;"> ... existing ... </div>
    <main class="main-content">
        ... existing viz-section + details-panel ...
    </main>
</div>

<!-- Spec Builder Tab -->
<div class="tab-content" id="content-spec-builder">
    <div class="spec-builder-layout">
        <!-- Chat Column -->
        <div class="chat-column">
            <div class="chat-header">
                <div class="chat-header-left">
                    <h2>Spec Builder</h2>
                    <div class="model-selector">
                        <label for="model-select" class="model-label">Model:</label>
                        <select id="model-select" class="model-select">
                            <option value="gemini-2.5-flash" selected>gemini-2.5-flash</option>
                            <option value="gemini-2.5-pro">gemini-2.5-pro</option>
                            <option value="gemini-2.5-flash-lite">gemini-2.5-flash-lite</option>
                            <option value="gemini-3-flash-preview">gemini-3-flash-preview</option>
                            <option value="gemini-3.1-pro-preview">gemini-3.1-pro-preview</option>
                            <option value="gemini-3.1-flash-lite-preview">gemini-3.1-flash-lite-preview</option>
                        </select>
                    </div>
                </div>
                <div class="chat-controls">
                    <button class="btn btn-outline btn-sm" id="btn-new-session" title="Start new session">
                        🔄 New Session
                    </button>
                    <button class="btn btn-outline btn-sm" id="btn-export-chat" title="Export conversation">
                        📥 Export
                    </button>
                </div>
            </div>
            <div class="chat-messages" id="chat-messages">
                <div class="chat-welcome">
                    <div class="welcome-icon">🏗️</div>
                    <h3>ATD Spec Builder</h3>
                    <p>Describe the feature or system you want to specify. I'll help you create properly structured ATD atoms following the atomic documentation rules.</p>
                    <div class="welcome-suggestions">
                        <button class="suggestion-chip" data-prompt="I want to define a new feature for">Define a feature</button>
                        <button class="suggestion-chip" data-prompt="Help me decompose this requirement:">Decompose a requirement</button>
                        <button class="suggestion-chip" data-prompt="Review and improve these existing atoms:">Review existing atoms</button>
                    </div>
                </div>
            </div>
            <div class="chat-context-bar" id="chat-context-bar" style="display: none;">
                <span class="context-label">📎 ATD Context:</span>
                <div class="context-chips" id="context-chips"></div>
                <button class="btn-icon" id="btn-add-context" title="Add ATD context">+</button>
            </div>
            <div class="chat-input-area">
                <textarea id="chat-input" placeholder="Describe what you want to specify..." rows="1"></textarea>
                <button class="btn btn-send" id="btn-send" title="Send (Ctrl+Enter)">
                    <span>➤</span>
                </button>
            </div>
        </div>

        <!-- Proposals Column -->
        <div class="proposals-column">
            <div class="proposals-header">
                <h2>Proposals</h2>
                <span class="proposal-count" id="proposal-count">0 pending</span>
            </div>
            <div class="proposals-list" id="proposals-list">
                <div class="proposals-empty">
                    <p>Proposals from the AI will appear here for your review.</p>
                    <p class="proposals-hint">You can accept ✓ or reject ✗ each proposal individually.</p>
                </div>
            </div>
            <div class="proposals-actions" id="proposals-actions" style="display: none;">
                <button class="btn btn-primary btn-sm" id="btn-accept-all">Accept All</button>
                <button class="btn btn-outline btn-sm" id="btn-reject-all">Reject All</button>
            </div>
        </div>
    </div>
</div>
```

#### C. Add the new JS file reference

Before the closing `</body>` tag, after the existing `app.js` script:
```html
<script src="/static/spec-builder.js"></script>
```

### 2. Create `webui/static/spec-builder.css`

Create this new file with the Spec Builder styles:

```css
/* ============================================
   SPEC BUILDER TAB STYLES
   ============================================ */

/* --- Tab System --- */
.header-tabs {
    display: flex;
    gap: 4px;
    background-color: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 8px;
    padding: 3px;
}

.tab-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 6px 16px;
    border-radius: 6px;
    border: none;
    background: transparent;
    color: var(--text-muted);
    font-family: var(--font-family);
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    white-space: nowrap;
}

.tab-btn:hover {
    color: var(--text-main);
    background-color: rgba(255, 255, 255, 0.05);
}

.tab-btn.active {
    color: var(--accent);
    background-color: var(--bg-panel);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.tab-icon {
    font-size: 14px;
}

.tab-content {
    display: none;
    flex: 1;
    overflow: hidden;
}

.tab-content.active {
    display: flex;
    flex-direction: column;
}

/* --- Spec Builder Layout --- */
.spec-builder-layout {
    display: flex;
    flex: 1;
    overflow: hidden;
}

/* --- Chat Column --- */
.chat-column {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-right: 1px solid var(--border);
    min-width: 0;
}

.chat-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 24px;
    border-bottom: 1px solid var(--border);
    background-color: var(--bg-panel);
}

.chat-header h2 {
    font-size: 18px;
    font-weight: 600;
}

.chat-controls {
    display: flex;
    gap: 8px;
}

/* --- Model Selector --- */
.chat-header-left {
    display: flex;
    align-items: center;
    gap: 16px;
}

.model-selector {
    display: flex;
    align-items: center;
    gap: 6px;
}

.model-label {
    font-size: 11px;
    color: var(--text-muted);
    text-transform: uppercase;
    font-weight: 600;
}

.model-select {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 6px;
    color: var(--text-main);
    font-family: var(--font-family);
    font-size: 12px;
    padding: 4px 8px;
    cursor: pointer;
    transition: border-color 0.2s;
}

.model-select:focus {
    outline: none;
    border-color: var(--accent);
}

.model-select:hover {
    border-color: var(--accent);
}

/* --- Chat Messages Area --- */
.chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    display: flex;
    flex-direction: column;
    gap: 16px;
}

/* Welcome State */
.chat-welcome {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 60px 40px;
    flex: 1;
}

.welcome-icon {
    font-size: 48px;
    margin-bottom: 16px;
}

.chat-welcome h3 {
    font-size: 20px;
    font-weight: 600;
    margin-bottom: 8px;
}

.chat-welcome p {
    color: var(--text-muted);
    font-size: 14px;
    max-width: 400px;
    line-height: 1.6;
}

.welcome-suggestions {
    display: flex;
    gap: 8px;
    margin-top: 24px;
    flex-wrap: wrap;
    justify-content: center;
}

.suggestion-chip {
    padding: 8px 16px;
    border-radius: 20px;
    border: 1px solid var(--border);
    background: var(--bg-card);
    color: var(--text-main);
    font-family: var(--font-family);
    font-size: 13px;
    cursor: pointer;
    transition: all 0.2s;
}

.suggestion-chip:hover {
    border-color: var(--accent);
    background: rgba(76, 139, 245, 0.1);
    color: var(--accent);
}

/* Message Bubbles */
.chat-msg {
    display: flex;
    gap: 12px;
    max-width: 85%;
    animation: msgFadeIn 0.3s ease;
}

@keyframes msgFadeIn {
    from { opacity: 0; transform: translateY(8px); }
    to { opacity: 1; transform: translateY(0); }
}

.chat-msg.user {
    align-self: flex-end;
    flex-direction: row-reverse;
}

.chat-msg.model {
    align-self: flex-start;
}

.msg-avatar {
    width: 32px;
    height: 32px;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    flex-shrink: 0;
}

.chat-msg.user .msg-avatar {
    background: var(--accent);
}

.chat-msg.model .msg-avatar {
    background: linear-gradient(135deg, #8b5cf6, #6366f1);
}

.msg-bubble {
    padding: 12px 16px;
    border-radius: 16px;
    font-size: 14px;
    line-height: 1.6;
}

.chat-msg.user .msg-bubble {
    background: var(--accent);
    color: white;
    border-bottom-right-radius: 4px;
}

.chat-msg.model .msg-bubble {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-bottom-left-radius: 4px;
}

.msg-bubble .markdown-body {
    max-height: none;
    padding-right: 0;
}

/* Typing Indicator */
.typing-indicator {
    display: flex;
    gap: 4px;
    padding: 12px 16px;
    align-items: center;
}

.typing-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--text-muted);
    animation: typingBounce 1.4s infinite;
}

.typing-dot:nth-child(2) { animation-delay: 0.2s; }
.typing-dot:nth-child(3) { animation-delay: 0.4s; }

@keyframes typingBounce {
    0%, 60%, 100% { transform: translateY(0); opacity: 0.4; }
    30% { transform: translateY(-6px); opacity: 1; }
}

/* --- Context Bar --- */
.chat-context-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 24px;
    border-top: 1px solid var(--border);
    background: var(--bg-card);
    overflow-x: auto;
}

.context-label {
    font-size: 12px;
    color: var(--text-muted);
    white-space: nowrap;
}

.context-chips {
    display: flex;
    gap: 6px;
    overflow-x: auto;
}

.context-chip {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 2px 8px;
    border-radius: 12px;
    background: rgba(76, 139, 245, 0.15);
    color: var(--accent);
    font-size: 11px;
    white-space: nowrap;
}

.context-chip .chip-remove {
    cursor: pointer;
    opacity: 0.6;
    font-size: 14px;
}

.context-chip .chip-remove:hover {
    opacity: 1;
}

.btn-icon {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    border: 1px dashed var(--border);
    background: transparent;
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    flex-shrink: 0;
}

.btn-icon:hover {
    border-color: var(--accent);
    color: var(--accent);
}

/* --- Chat Input --- */
.chat-input-area {
    display: flex;
    gap: 12px;
    padding: 16px 24px;
    border-top: 1px solid var(--border);
    background: var(--bg-panel);
    align-items: flex-end;
}

#chat-input {
    flex: 1;
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 12px;
    padding: 12px 16px;
    color: var(--text-main);
    font-family: var(--font-family);
    font-size: 14px;
    line-height: 1.5;
    resize: none;
    max-height: 150px;
    overflow-y: auto;
}

#chat-input:focus {
    outline: none;
    border-color: var(--accent);
    box-shadow: 0 0 0 3px rgba(76, 139, 245, 0.15);
}

#chat-input::placeholder {
    color: var(--text-muted);
}

.btn-send {
    width: 44px;
    height: 44px;
    border-radius: 50%;
    background: var(--accent);
    border: none;
    color: white;
    font-size: 18px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
    flex-shrink: 0;
}

.btn-send:hover {
    background: #3b76db;
    transform: scale(1.05);
}

.btn-send:disabled {
    background: var(--bg-card);
    color: var(--text-muted);
    cursor: not-allowed;
    transform: none;
}

/* --- Proposals Column --- */
.proposals-column {
    width: 380px;
    flex-shrink: 0;
    display: flex;
    flex-direction: column;
    background: var(--bg-panel);
    overflow: hidden;
}

.proposals-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid var(--border);
}

.proposals-header h2 {
    font-size: 16px;
    font-weight: 600;
}

.proposal-count {
    font-size: 12px;
    color: var(--text-muted);
    background: var(--bg-card);
    padding: 2px 10px;
    border-radius: 10px;
}

.proposals-list {
    flex: 1;
    overflow-y: auto;
    padding: 16px;
    display: flex;
    flex-direction: column;
    gap: 12px;
}

.proposals-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    padding: 40px 20px;
    color: var(--text-muted);
    font-size: 13px;
    flex: 1;
}

.proposals-hint {
    margin-top: 8px;
    font-size: 12px;
    opacity: 0.7;
}

/* Proposal Card */
.proposal-card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 16px;
    transition: all 0.2s;
    animation: cardSlideIn 0.3s ease;
}

@keyframes cardSlideIn {
    from { opacity: 0; transform: translateX(20px); }
    to { opacity: 1; transform: translateX(0); }
}

.proposal-card:hover {
    border-color: rgba(76, 139, 245, 0.3);
}

.proposal-card.accepted {
    border-color: var(--color-green);
    background: rgba(46, 125, 50, 0.08);
}

.proposal-card.rejected {
    border-color: var(--color-red);
    background: rgba(211, 47, 47, 0.08);
    opacity: 0.6;
}

.proposal-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
}

.proposal-action-badge {
    font-size: 10px;
    font-weight: 700;
    padding: 2px 8px;
    border-radius: 4px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

.proposal-action-badge.CREATE {
    background: rgba(46, 125, 50, 0.2);
    color: var(--color-green-light);
}

.proposal-action-badge.UPDATE {
    background: rgba(251, 192, 45, 0.2);
    color: var(--color-yellow);
}

.proposal-action-badge.DELETE {
    background: rgba(211, 47, 47, 0.2);
    color: var(--color-red-light);
}

.proposal-card-actions {
    display: flex;
    gap: 4px;
}

.proposal-btn {
    width: 28px;
    height: 28px;
    border-radius: 6px;
    border: 1px solid var(--border);
    background: transparent;
    color: var(--text-muted);
    font-size: 14px;
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.2s;
}

.proposal-btn.accept:hover {
    border-color: var(--color-green);
    color: var(--color-green-light);
    background: rgba(46, 125, 50, 0.15);
}

.proposal-btn.reject:hover {
    border-color: var(--color-red);
    color: var(--color-red-light);
    background: rgba(211, 47, 47, 0.15);
}

.proposal-btn.preview:hover {
    border-color: var(--accent);
    color: var(--accent);
    background: rgba(76, 139, 245, 0.15);
}

.proposal-atom-id {
    font-family: monospace;
    font-size: 12px;
    color: var(--accent);
    margin-bottom: 4px;
}

.proposal-atom-name {
    font-size: 14px;
    font-weight: 500;
    margin-bottom: 8px;
}

.proposal-impact {
    font-size: 12px;
    color: var(--text-muted);
    line-height: 1.5;
}

.proposal-detail {
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--border);
    font-size: 13px;
}

.proposal-detail .field-row {
    display: flex;
    gap: 8px;
    margin-bottom: 6px;
}

.proposal-detail .field-label {
    color: var(--text-muted);
    font-size: 11px;
    text-transform: uppercase;
    font-weight: 600;
    min-width: 60px;
}

.proposal-detail .field-value {
    color: var(--text-main);
    font-size: 12px;
}

/* --- Proposals Bottom Actions --- */
.proposals-actions {
    display: flex;
    gap: 8px;
    padding: 12px 16px;
    border-top: 1px solid var(--border);
}

.proposals-actions .btn {
    flex: 1;
    margin-bottom: 0;
}

/* --- Context Drift Warning --- */
.drift-warning {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 16px;
    margin: 8px 0;
    background: rgba(245, 124, 0, 0.12);
    border: 1px solid rgba(245, 124, 0, 0.3);
    border-radius: 10px;
    font-size: 13px;
    color: var(--color-orange);
    animation: msgFadeIn 0.3s ease;
}

.drift-warning .drift-icon {
    font-size: 18px;
}

.drift-warning button {
    margin-left: auto;
    white-space: nowrap;
}

/* --- ATD Context Search Modal --- */
.context-search-modal {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    animation: modalFadeIn 0.2s ease;
}

@keyframes modalFadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
}

.context-search-panel {
    width: 480px;
    max-height: 500px;
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: 12px;
    overflow: hidden;
    box-shadow: 0 16px 48px rgba(0, 0, 0, 0.4);
}

.context-search-input {
    width: 100%;
    padding: 16px 20px;
    border: none;
    border-bottom: 1px solid var(--border);
    background: var(--bg-panel);
    color: var(--text-main);
    font-family: var(--font-family);
    font-size: 15px;
}

.context-search-input:focus {
    outline: none;
}

.context-search-results {
    max-height: 400px;
    overflow-y: auto;
    padding: 8px;
}

.context-search-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    border-radius: 6px;
    cursor: pointer;
    transition: background 0.15s;
    gap: 12px;
}

.context-search-item:hover {
    background: rgba(255, 255, 255, 0.05);
}

.context-search-item .item-type {
    font-size: 10px;
    font-weight: 600;
    color: var(--accent);
    opacity: 0.8;
    min-width: 70px;
}

.context-search-item .item-name {
    font-size: 13px;
    flex: 1;
}

.context-search-item .item-id {
    font-size: 11px;
    font-family: monospace;
    color: var(--text-muted);
}
```

### 3. `webui/static/styles.css`

Add at the **very end** of the file a single import line:

```css
/* No changes needed — spec-builder.css is a separate file loaded in index.html */
```

Actually, add the CSS link in `index.html` instead. In the `<head>` section, after the existing stylesheet link:

```html
<link rel="stylesheet" href="/static/spec-builder.css">
```

### 4. `webui/static/app.js`

Add tab switching logic at the end of the `DOMContentLoaded` callback (before the final `});`):

```javascript
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
    });
});
```

## Expected Outcome

After this step:
1. Two tabs appear in the header: "Explorer" (active by default) and "Spec Builder"
2. Clicking "Spec Builder" reveals the two-column layout with chat panel and proposals panel
3. The chat panel shows a welcome screen with suggestion chips
4. The proposals panel shows an empty state message
5. All styles follow the existing dark theme with consistent typography and spacing
6. Clicking "Explorer" returns to the original treemap/treeview UI
7. No JavaScript chat logic yet — that's Step 3

## Key Design Decisions

- **Separate CSS file** (`spec-builder.css`): Keeps the Spec Builder styles isolated from the existing 745-line `styles.css`
- **Separate JS file** (`spec-builder.js`): Will be created in Step 3; keeps Spec Builder logic separate from explorer logic
- **Two-column layout**: Chat fills remaining space, proposals panel is fixed at 380px width
- **Welcome state**: Shows quick-start suggestion chips to guide first interaction
