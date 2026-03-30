# Step 6: Action History Forwarding

## Objective
Ensure that when the user accepts or rejects proposals, this action history is forwarded to Gemini on the next message so the model knows which proposals were validated and which were discarded.

## Prerequisites
- Steps 1-5 completed
- `state.actions` array exists in `spec-builder.js` (from Step 3)
- `handleProposalAction` pushes to `state.actions` (from Step 4)
- `buildChatPayload` includes `actions: state.actions` (from Step 3)

## How It Works (already mostly wired)

The architecture from Steps 1-5 already handles most of this:

1. **Step 4** records each accept/reject in `state.actions`:
   ```javascript
   state.actions.push({
       proposal_id: proposal._id,
       atom_id: proposal.atom_id,
       action: 'ACCEPTED' or 'REJECTED',
       summary: `CREATE module_login: Handles user authentication flow`,
   });
   ```

2. **Step 3** `buildChatPayload()` sends `state.actions` in the request body.

3. **Step 1** backend injects `actions` into the system prompt:
   ```
   --- USER ACTION HISTORY ---
   - Proposal for atom 'module_login': ACCEPTED. CREATE module_login: Handles user authentication flow
   - Proposal for atom 'rule_password_policy': REJECTED. CREATE rule_password_policy: Enforce minimum 8 chars
   ```

## Enhancement: Visual Action History in Chat

Add a visual indicator in the chat when the user takes an action, so the conversation reads naturally:

### Add to `spec-builder.js`

```javascript
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
```

### CSS for action notices

Add to `spec-builder.css`:

```css
/* Action History Notices in Chat */
.chat-action-notice {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 16px;
    margin: 4px 0;
    border-radius: 8px;
    background: var(--bg-card);
    font-size: 12px;
    color: var(--text-muted);
    animation: msgFadeIn 0.2s ease;
}

.action-accept {
    color: var(--color-green-light);
}

.action-reject {
    color: var(--color-red-light);
}

.action-detail {
    color: var(--text-main);
    font-size: 12px;
}

.action-detail strong {
    font-weight: 600;
    font-size: 10px;
    text-transform: uppercase;
    opacity: 0.8;
    margin-right: 4px;
}
```

### Update `handleProposalAction` (Step 4)

After the line that pushes to `state.actions`, call the renderer:

```javascript
// Inside handleProposalAction, after state.actions.push(...)
renderActionInChat(proposal, action);
```

## Summary of Action Flow

```
User accepts/rejects proposal
    ↓
1. state.actions gets new entry
2. Visual notice appears in chat feed ("✅ Accepted: CREATE @module_login")
3. Proposal card changes to green/red
    ↓
User sends next message
    ↓
4. buildChatPayload() includes all actions
5. Backend injects action history into system prompt
6. Gemini sees: "User accepted module_login, rejected rule_password_policy"
7. Gemini can adjust subsequent proposals based on what user liked/disliked
```

## Expected Outcome

1. Every accept/reject action shows an inline notice in the chat stream
2. The notice shows the action taken, the proposal type, and the atom ID
3. All action history is forwarded to Gemini on the next message
4. Gemini can reference prior decisions in its reasoning
5. The chat reads as a natural conversation with action milestones interspersed
