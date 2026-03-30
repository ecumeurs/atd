# Step 7: Context Drift Warning

## Objective
Display a warning after 10 user↔model exchanges recommending a session restart to avoid hallucinations and context window overflow.

## Prerequisites
- Steps 1-6 completed
- `state.exchangeCount` incremented in `sendMessage` (Step 3)
- `checkContextDrift` function exists (Step 3)
- `resetSession` function exists (Step 3)
- CSS class `.drift-warning` exists (Step 2)

## Current Implementation (Step 3)

Step 3 already includes a basic drift check:

```javascript
function checkContextDrift() {
    if (state.exchangeCount === 10) {
        // renders warning
    }
}
```

## Enhancement: Progressive Warnings

Replace the simple check with a progressive system:

### Update in `spec-builder.js`

```javascript
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
        <div class="drift-text">
            <strong>${mainText}</strong>
            <div style="font-size: 12px; margin-top: 2px; opacity: 0.8;">${subText}</div>
        </div>
        <button class="btn btn-outline btn-sm drift-restart-btn">🔄 New Session</button>
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
```

### CSS additions (add to `spec-builder.css`)

```css
/* --- Progressive Drift Warnings --- */
.drift-warning.drift-caution {
    background: rgba(245, 124, 0, 0.12);
    border-color: rgba(245, 124, 0, 0.3);
    color: var(--color-orange);
}

.drift-warning.drift-warning {
    background: rgba(251, 192, 45, 0.12);
    border-color: rgba(251, 192, 45, 0.3);
    color: var(--color-yellow);
}

.drift-warning.drift-critical {
    background: rgba(211, 47, 47, 0.12);
    border-color: rgba(211, 47, 47, 0.3);
    color: var(--color-red-light);
}

.drift-text {
    flex: 1;
}

.drift-restart-btn {
    white-space: nowrap;
}

/* Exchange Counter Badge */
.exchange-badge {
    font-size: 11px;
    font-weight: 600;
    padding: 2px 8px;
    border-radius: 10px;
    background: var(--bg-card);
    color: var(--text-muted);
    border: 1px solid var(--border);
}

.exchange-badge.warning {
    background: rgba(245, 124, 0, 0.2);
    color: var(--color-orange);
    border-color: var(--color-orange);
}

.exchange-badge.critical {
    background: rgba(211, 47, 47, 0.2);
    color: var(--color-red-light);
    border-color: var(--color-red);
}
```

## Expected Outcome

1. After 10 exchanges: orange caution banner appears in chat
2. After 15 exchanges: yellow stronger warning appears
3. After 20 exchanges: red critical warning appears
4. Each warning has a "New Session" button that resets everything
5. An exchange counter badge appears in the chat header (e.g., "7/10")
6. The badge color changes from neutral → orange → red as exchanges increase
7. Session reset clears the counter and removes all warnings
