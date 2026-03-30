# Step 8: Session Management

## Objective
Add full session management: conversation export to Markdown, keyboard shortcuts, and session metadata display.

## Prerequisites
- Steps 1-7 completed
- `resetSession` exists (Step 3)
- `#btn-export-chat` and `#btn-new-session` exist in HTML (Step 2)

## Changes to `spec-builder.js`

### 1. Export Conversation as Markdown

Wire the export button (add to `init()`):

```javascript
if (dom.exportBtn) {
    dom.exportBtn.addEventListener('click', exportConversation);
}
```

Implement the export function:

```javascript
function exportConversation() {
    if (state.messages.length === 0) {
        alert('No conversation to export.');
        return;
    }

    let md = `# ATD Spec Builder Session\n`;
    md += `**Date:** ${new Date().toISOString()}\n`;
    md += `**Exchanges:** ${state.exchangeCount}\n`;
    md += `**Proposals:** ${state.proposals.length} (${state.proposals.filter(p => p._status === 'accepted').length} accepted, ${state.proposals.filter(p => p._status === 'rejected').length} rejected)\n\n`;
    md += `---\n\n`;

    // Messages
    md += `## Conversation\n\n`;
    state.messages.forEach(msg => {
        const role = msg.role === 'user' ? '**User**' : '**Spec Builder**';
        md += `### ${role}\n\n${msg.content}\n\n`;
    });

    // Proposals summary
    if (state.proposals.length > 0) {
        md += `---\n\n## Proposals Summary\n\n`;
        md += `| Action | Atom ID | Status | Human Name |\n`;
        md += `|--------|---------|--------|------------|\n`;
        state.proposals.forEach(p => {
            md += `| ${p.action} | \`${p.atom_id}\` | ${p._status} | ${p.content?.human_name || '-'} |\n`;
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
```

### 2. Keyboard Shortcuts

Add global keyboard handler in `init()`:

```javascript
// Global keyboard shortcuts
document.addEventListener('keydown', (e) => {
    // Only active when Spec Builder tab is visible
    const specTab = document.getElementById('content-spec-builder');
    if (!specTab || !specTab.classList.contains('active')) return;

    // Ctrl+Enter → Send message (already handled in textarea keydown)
    // Escape → Close any open modal
    if (e.key === 'Escape') {
        closeContextSearch();
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
        if (document.activeElement !== dom.chatInput && !document.getElementById('context-search-modal')) {
            dom.chatInput.focus();
        }
    }
});
```

### 3. Session Info Display

Add a session timestamp and exchange counter to the chat header. Modify `resetSession` to record session start time:

```javascript
function resetSession() {
    if (state.messages.length > 0 && !confirm('Start a new session? All conversation history will be cleared.')) return;

    state.messages = [];
    state.proposals = [];
    state.actions = [];
    state.atdContext = [];
    state.atdContextSent.clear();
    state.exchangeCount = 0;
    state.sessionStart = new Date();

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
```

## Expected Outcome

1. **Export**: Downloads a Markdown file with full conversation, proposals table, and context list
2. **Shortcuts**:
   - `Ctrl+Enter` → Send message
   - `Ctrl+Shift+N` → New session
   - `Ctrl+Shift+E` → Export conversation
   - `Ctrl+Shift+A` → Open ATD context search
   - `Escape` → Close modals
3. **Session reset**: Fully recreates the welcome screen and clears all state
4. **Auto-focus**: Typing anywhere auto-focuses the chat input
