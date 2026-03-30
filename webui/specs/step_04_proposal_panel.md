# Step 4: Proposal Cards — Accept/Reject & Apply via `atd update`

## Objective
Enhance the proposal cards in the side panel with full accept/reject flow, detail expansion, and backend application via the existing `/api/atd/:id/update` endpoint and `atd update` CLI.

## Prerequisites
- Steps 1-3 completed
- `window.SpecBuilder.state.proposals` array populated by Step 3
- Existing backend endpoints: `POST /api/atd/:id/update` (in `main.go` line 157)
- Existing bulk update: `POST /api/bulk-update` (in `main.go` line 119)
- `atd` binary available at path defined in `config.json` → `toolkit_path`

## Changes Required

### 1. Add a new backend endpoint for proposal application

In `webui/main.go`, add inside the `api` group:

```go
api.POST("/gemini/apply-proposal", func(c *gin.Context) {
    var proposal struct {
        Action  string                 `json:"action"`
        AtomID  string                 `json:"atom_id"`
        Content map[string]interface{} `json:"content"`
    }
    if err := c.ShouldBindJSON(&proposal); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")
    atdDir := filepath.Join(AppConfig.ProjectPath, AppConfig.ATDPath)

    switch proposal.Action {
    case "CREATE":
        // Build the atom file path
        filePath := filepath.Join(atdDir, proposal.AtomID+".atom.md")

        // Build atd update args to create a new atom
        args := []string{"update", "--file", filePath}

        // Map content fields to --set and section flags
        if v, ok := proposal.Content["human_name"].(string); ok && v != "" {
            args = append(args, "--set", "human_name="+v)
        }
        if v, ok := proposal.Content["type"].(string); ok && v != "" {
            args = append(args, "--set", "type="+v)
        }
        if v, ok := proposal.Content["layer"].(string); ok && v != "" {
            args = append(args, "--set", "layer="+v)
        }
        if v, ok := proposal.Content["status"].(string); ok && v != "" {
            args = append(args, "--set", "status="+v)
        }
        if v, ok := proposal.Content["priority"].(string); ok && v != "" {
            args = append(args, "--set", "priority="+v)
        }
        if v, ok := proposal.Content["intent"].(string); ok && v != "" {
            args = append(args, "--intent", v)
        }
        if v, ok := proposal.Content["logic"].(string); ok && v != "" {
            args = append(args, "--logic", v)
        }
        if v, ok := proposal.Content["technical_interface"].(string); ok && v != "" {
            args = append(args, "--interface", v)
        }
        if v, ok := proposal.Content["expectation"].(string); ok && v != "" {
            args = append(args, "--expectation", v)
        }

        cmd := exec.Command(toolPath, args...)
        output, err := cmd.CombinedOutput()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to create atom",
                "details": string(output),
            })
            return
        }

    case "UPDATE":
        atom, exists := Atoms[proposal.AtomID]
        if !exists {
            c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found: " + proposal.AtomID})
            return
        }

        args := []string{"update", "--file", atom.FilePath}
        // Same field mapping as CREATE
        if v, ok := proposal.Content["human_name"].(string); ok && v != "" {
            args = append(args, "--set", "human_name="+v)
        }
        if v, ok := proposal.Content["type"].(string); ok && v != "" {
            args = append(args, "--set", "type="+v)
        }
        if v, ok := proposal.Content["status"].(string); ok && v != "" {
            args = append(args, "--set", "status="+v)
        }
        if v, ok := proposal.Content["intent"].(string); ok && v != "" {
            args = append(args, "--intent", v)
        }
        if v, ok := proposal.Content["logic"].(string); ok && v != "" {
            args = append(args, "--logic", v)
        }
        if v, ok := proposal.Content["technical_interface"].(string); ok && v != "" {
            args = append(args, "--interface", v)
        }
        if v, ok := proposal.Content["expectation"].(string); ok && v != "" {
            args = append(args, "--expectation", v)
        }

        cmd := exec.Command(toolPath, args...)
        output, err := cmd.CombinedOutput()
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{
                "error":   "Failed to update atom",
                "details": string(output),
            })
            return
        }

    case "DELETE":
        atom, exists := Atoms[proposal.AtomID]
        if !exists {
            c.JSON(http.StatusNotFound, gin.H{"error": "Atom not found: " + proposal.AtomID})
            return
        }
        // Delete the file
        if err := os.Remove(atom.FilePath); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete atom file"})
            return
        }
    }

    // Refresh atoms after any change
    refreshAtoms()
    c.JSON(http.StatusOK, gin.H{"message": "Proposal applied successfully", "atom_id": proposal.AtomID})
})
```

### 2. Enhance proposal rendering in `spec-builder.js`

Replace the `handleProposals` and `renderProposalsList` functions from Step 3 with full implementations:

```javascript
// ============================================
// PROPOSALS — Full Implementation
// ============================================
function handleProposals(proposals) {
    proposals.forEach(p => {
        p._id = `proposal_${Date.now()}_${Math.random().toString(36).substr(2, 5)}`;
        p._status = 'pending';
        state.proposals.push(p);
        renderProposalCard(p);
    });
    updateProposalCount();
}

function renderProposalCard(p) {
    // Remove empty state if present
    const emptyEl = dom.proposalsList.querySelector('.proposals-empty');
    if (emptyEl) emptyEl.remove();

    const card = document.createElement('div');
    card.className = 'proposal-card';
    card.dataset.proposalId = p._id;

    // Header with action badge and buttons
    const header = document.createElement('div');
    header.className = 'proposal-card-header';

    const badge = document.createElement('span');
    badge.className = `proposal-action-badge ${p.action}`;
    badge.textContent = p.action;

    const actions = document.createElement('div');
    actions.className = 'proposal-card-actions';

    const acceptBtn = createProposalBtn('accept', '✓', () => handleProposalAction(p, 'accepted', card));
    const rejectBtn = createProposalBtn('reject', '✗', () => handleProposalAction(p, 'rejected', card));
    const previewBtn = createProposalBtn('preview', '👁', () => toggleProposalDetail(card, p));

    actions.append(acceptBtn, rejectBtn, previewBtn);
    header.append(badge, actions);

    // Body
    const atomId = document.createElement('div');
    atomId.className = 'proposal-atom-id';
    atomId.textContent = `@${p.atom_id}`;

    const atomName = document.createElement('div');
    atomName.className = 'proposal-atom-name';
    atomName.textContent = p.content?.human_name || p.atom_id;

    const impact = document.createElement('div');
    impact.className = 'proposal-impact';
    impact.textContent = p.impact_summary || '';

    card.append(header, atomId, atomName, impact);
    dom.proposalsList.appendChild(card);

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

async function handleProposalAction(proposal, action, cardEl) {
    proposal._status = action;

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
                proposal._status = 'pending';
                cardEl.className = 'proposal-card';
                return;
            }
            // Show success indicator on card
            const successTag = document.createElement('div');
            successTag.style.cssText = 'color: var(--color-green-light); font-size: 11px; margin-top: 8px;';
            successTag.textContent = '✓ Applied successfully';
            cardEl.appendChild(successTag);
        } catch (err) {
            alert('Network error applying proposal: ' + err.message);
        }
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
    const pending = state.proposals.filter(p => p._status === 'pending').length;
    const accepted = state.proposals.filter(p => p._status === 'accepted').length;
    const rejected = state.proposals.filter(p => p._status === 'rejected').length;

    let text = `${pending} pending`;
    if (accepted > 0) text += ` · ${accepted} applied`;
    if (rejected > 0) text += ` · ${rejected} rejected`;
    dom.proposalCount.textContent = text;

    if (pending === 0 && dom.proposalsActions) {
        dom.proposalsActions.style.display = 'none';
    }
}
```

### 3. Wire Accept All / Reject All buttons

In the `init()` function in `spec-builder.js`, add:

```javascript
if (dom.acceptAllBtn) {
    dom.acceptAllBtn.addEventListener('click', async () => {
        const pendingCards = dom.proposalsList.querySelectorAll('.proposal-card:not(.accepted):not(.rejected)');
        for (const card of pendingCards) {
            const id = card.dataset.proposalId;
            const proposal = state.proposals.find(p => p._id === id);
            if (proposal) await handleProposalAction(proposal, 'accepted', card);
        }
    });
}

if (dom.rejectAllBtn) {
    dom.rejectAllBtn.addEventListener('click', () => {
        const pendingCards = dom.proposalsList.querySelectorAll('.proposal-card:not(.accepted):not(.rejected)');
        pendingCards.forEach(card => {
            const id = card.dataset.proposalId;
            const proposal = state.proposals.find(p => p._id === id);
            if (proposal) handleProposalAction(proposal, 'rejected', card);
        });
    });
}
```

## Expected Outcome

1. Each proposal card has ✓ (accept), ✗ (reject), and 👁 (preview) buttons
2. Clicking ✓ calls `POST /api/gemini/apply-proposal` which executes `atd update` to apply changes
3. Clicking 👁 toggles an expandable detail section showing all proposed fields
4. Accepted cards turn green with success confirmation; rejected cards turn red and fade
5. Action buttons become disabled after accepting/rejecting
6. "Accept All" / "Reject All" buttons process all pending proposals
7. Accepted/rejected actions are recorded in `state.actions` for forwarding to Gemini in Step 6
8. Pending/applied/rejected counts update in the header badge
