# Step 9: Polish, Animations & Error Handling

## Objective
Add final visual polish: loading skeletons, smooth transitions, error state handling, token usage display, and responsive layout adjustments.

## Prerequisites
- Steps 1-8 completed

## CSS Additions to `spec-builder.css`

### 1. Loading State for Send Button

```css
/* Send button loading state */
.btn-send.loading {
    position: relative;
    color: transparent;
}

.btn-send.loading::after {
    content: '';
    position: absolute;
    width: 20px;
    height: 20px;
    border: 2px solid transparent;
    border-top-color: white;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
}

@keyframes spin {
    to { transform: rotate(360deg); }
}
```

### 2. Message Timestamps

```css
.msg-time {
    font-size: 10px;
    color: var(--text-muted);
    margin-top: 4px;
    opacity: 0.6;
}

.chat-msg.user .msg-time {
    text-align: right;
}
```

### 3. Proposal Card Hover Enhancement

```css
.proposal-card {
    position: relative;
    overflow: hidden;
}

.proposal-card::before {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    width: 3px;
    height: 100%;
    background: var(--border);
    transition: background 0.2s;
}

.proposal-card:hover::before {
    background: var(--accent);
}

.proposal-card.accepted::before {
    background: var(--color-green);
}

.proposal-card.rejected::before {
    background: var(--color-red);
}
```

### 4. Empty States Enhancement

```css
.chat-welcome,
.proposals-empty {
    opacity: 0;
    animation: fadeInUp 0.5s ease 0.2s forwards;
}

@keyframes fadeInUp {
    from { opacity: 0; transform: translateY(20px); }
    to { opacity: 1; transform: translateY(0); }
}
```

### 5. Responsive Layout

```css
/* Responsive: collapse proposals into overlay on small screens */
@media (max-width: 900px) {
    .spec-builder-layout {
        flex-direction: column;
    }

    .proposals-column {
        width: 100%;
        max-height: 40vh;
        border-left: none;
        border-top: 1px solid var(--border);
    }

    .chat-column {
        border-right: none;
    }
}

@media (max-width: 600px) {
    .chat-header {
        padding: 12px 16px;
    }

    .chat-messages {
        padding: 16px;
    }

    .chat-input-area {
        padding: 12px 16px;
    }

    .welcome-suggestions {
        flex-direction: column;
    }

    .header-tabs {
        gap: 2px;
    }

    .tab-btn {
        padding: 4px 10px;
        font-size: 12px;
    }
}
```

### 6. Smooth Scrollbar for Chat

```css
.chat-messages::-webkit-scrollbar {
    width: 4px;
}

.chat-messages::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 4px;
}

.chat-messages::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.2);
}
```

## JS Enhancements in `spec-builder.js`

### 1. Timestamps on Messages

Modify `renderMessage` to add timestamps:

```javascript
// After creating bubble element:
const time = document.createElement('div');
time.className = 'msg-time';
time.textContent = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
bubble.appendChild(time);
```

### 2. Send Button Loading State

In `sendMessage`, update the button state:

```javascript
// Before fetch:
dom.sendBtn.classList.add('loading');

// In finally block:
dom.sendBtn.classList.remove('loading');
```

### 3. API Error Handling Improvement

Enhance `renderErrorMessage` with retry capability:

```javascript
function renderErrorMessage(errorText, retryPayload) {
    const msgEl = document.createElement('div');
    msgEl.className = 'chat-msg model';

    const avatar = document.createElement('div');
    avatar.className = 'msg-avatar';
    avatar.style.background = 'var(--color-red)';
    avatar.textContent = '⚠';

    const bubble = document.createElement('div');
    bubble.className = 'msg-bubble';
    bubble.style.borderColor = 'rgba(211, 47, 47, 0.3)';

    bubble.innerHTML = `
        <div style="color: var(--color-red-light); font-weight: 500; margin-bottom: 4px;">Error</div>
        <div style="font-size: 13px; color: var(--text-muted);">${escapeHtml(errorText)}</div>
    `;

    // Retry button
    if (retryPayload) {
        const retryBtn = document.createElement('button');
        retryBtn.className = 'btn btn-outline btn-sm';
        retryBtn.style.marginTop = '8px';
        retryBtn.textContent = '🔄 Retry';
        retryBtn.addEventListener('click', () => {
            msgEl.remove();
            // Re-attempt the last message
            state.messages.pop(); // Remove the user message that failed
            dom.chatInput.value = retryPayload;
            sendMessage();
        });
        bubble.appendChild(retryBtn);
    }

    msgEl.append(avatar, bubble);
    dom.messagesContainer.appendChild(msgEl);
    scrollToBottom();
}
```

### 4. Connection Check on Tab Switch

When switching to the Spec Builder tab, verify the backend is reachable:

```javascript
// In app.js tab switching, or in spec-builder.js init
async function checkBackendHealth() {
    try {
        const resp = await fetch('/api/info');
        if (!resp.ok) throw new Error('Backend unreachable');
    } catch (err) {
        renderErrorMessage('Cannot connect to WebUI backend. Please ensure the server is running.');
    }
}
```

## Expected Outcome

1. Send button shows a spinning animation while waiting for response
2. Messages display timestamps (e.g., "14:32")
3. Proposal cards have a colored left border accent on hover
4. Empty states fade in with a smooth animation
5. Error messages show a retry button to resend the last message
6. Layout adapts to smaller screens (proposals collapse below chat)
7. Custom thin scrollbar in chat area
8. Backend health check on tab switch
