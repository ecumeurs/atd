# WebUI Remaining Work — Roadmap

## 1. ATD Type/Layer Audit Script

**Problem**: Many atoms have incorrect `type` and `layer` (e.g., `mechanic_atd_init` was `CUSTOMER` layer, `service_atd_tiered_provider` was `CUSTOMER`). The user has been manually fixing these.

**Solution**: Create a script that:
1. Dumps all ATDs as a JSON table: `id`, `human_name`, `type`, `layer`, `parents`, `intent`
2. Sends this to a local LLM with a prompt asking it to verify correct type/layer per the ATD type reference table
3. Outputs a diff of proposed corrections

**Implementation**:
- Temporary python script at `scripts/audit_layers.py`
- Mostly needs to parse the ATD files yaml frontmatter and the intent section
- Print everything in an output file and then we will manually fix the issues

---

## 2. Explorer UX Fixes (Quick)

### 2a. Remove module grouping
**File**: `webui/static/js/explorer.js`
- Remove the `groupByModule()` function and its usage in `renderWaterfall()`
- Render all atoms flat within each lane (sorted by type or name)

### 2b. Unhighlight on re-click / ESC
**File**: `webui/static/js/explorer.js`
- In `createAtomCard()` click handler: if clicking the same atom that's already selected → call `clearHighlights()` and `setCurrentAtom(null)`
- Add ESC key listener in `initExplorer()` to clear highlights and close detail panel

### 2c. Three-color highlighting (selected / ancestors / descendants)
**File**: `webui/static/js/explorer.js` + `webui/static/styles.css`
- Current atom: distinct highlight (e.g., bright blue border + glow)
- Ancestors (parents upward): one color (e.g., purple/magenta tint matching CUSTOMER lane)
- Descendants (dependents downward): another color (e.g., green tint matching IMPL lane)
- CSS classes: `.card-selected`, `.card-ancestor`, `.card-descendant` (replacing the single `.card-highlighted`)

### 2d. Layer badge in detail panel
**File**: `webui/static/js/details.js` + `webui/static/index.html`
- Add a `<div class="badge-layer" id="detail-layer">` next to the type/status badges in the detail header
- In `showDetails()`: set `dom.dLayer.textContent = atom.layer` with layer-specific color (purple/blue/green)

### 2e. Priority: numeric 1-5 input
**File**: `webui/static/index.html` + `webui/static/js/details.js`
- Replace the priority `<select>` (CORE/SECONDARY/EXPERIMENTAL/FLAVOR) with `<input type="number" min="1" max="5" id="edit-priority">`
- In `enterEditMode()`: set `.value = atom.priority || 3`
- Per `domain_atd_structure`: priority is `integer 1 (low) to 5 (highest)`

---

## 3. ISS-054: Proper ATD Summary CLI/MCP Command

**Status**: Open — the CLI tool `atd summary` does not exist yet.

**Spec** (from ISS-054):
- **CLI**: `atd summary <id> [--length short|default|extended|long]`
- **MCP**: `atd_summary` tool
- **Behavior**:
  1. Run `atd trace <id>` to get graph slice (parents, dependents, code_links)
  2. If LLM available (tiered provider): call `atd assemble --starts <id,id,...> --purpose "summarize"` to get stitched content
  3. Fallback (LLM-light): aggregate intents by layer, structured output
  4. Append metadata: involved atom IDs with file paths, layers, types
  5. Output uses `[[atom_id]]` notation for traceability

**Files to create/modify**:
- `scripts/cmd/atd/cmd/summary.go` — new CLI subcommand
- `scripts/cmd/atd/cmd/assemble.go` — check out how it works, it seems to be doing most of the work already but might not handle multiple starts ids correctly. Also it seems to need a snapshot parameter to do something but it's unclear as to what it does. 
- `scripts/cmd/atd-serve/tools.go` — register `atd_summary` MCP tool
- ATD atom: `service_atd_summary.atom.md`

---

## 4. Refactor WebUI Summary to Use ATD CLI

**Problem**: `handlers_atd.go` has a hand-rolled 100-line implementation that:
- Reimplements graph traversal (duplicates `atd trace`)
- Hardcodes `llama3.2` model (ignores tiered provider)
- Calls Ollama directly (bypasses `.atd` configuration)

**Fix**: Replace `handleSummary()` and `tryOllamaSummary()` with:
```go
func handleSummary(c *gin.Context) {
    id := c.Param("id")
    length := c.DefaultQuery("length", "default")
    toolPath := filepath.Join(expandPath(AppConfig.ToolkitPath), "atd")

    cmd := exec.Command(toolPath, "summary", id, "--length", length, "--json", "--structured")
    output, err := cmd.CombinedOutput()
    if err != nil {
        // fallback to basic intent extraction
    }
    // parse and return JSON output
}
```

This depends on #3 (the CLI command) being implemented first.

---

## 5. Ctrl+K Document Generation Feature

**Concept**: Extend the existing Ctrl+K search overlay with a "Generate Document" mode.

### User Flow
1. User presses Ctrl+K → search overlay opens
2. User types a natural language query, e.g.: *"tell me about the summary feature in the UI, model selection"*
3. User clicks a "📄 Generate Document" button (or prefix query with `/doc `)
4. Backend:
   a. Calls `atd search --query "<user query>"` (semantic search) to find relevant atoms
   b. Calls `atd trace <id>` on top results to get graph context
   c. Calls `atd assemble --starts <ids> --purpose "<user query>"` to stitch content
   d. Optionally sends to LLM for narrative polish
5. Returns a markdown document → rendered as HTML in a modal
6. User can download as `.md` file

### Caching (max 5)
- Backend stores up to 5 generated documents in temp files
- Each has: query, timestamp, markdown content, involved atoms
- A dropdown menu in the header (or next to the search button) shows recent documents
- When a 6th is generated, the oldest is evicted (LRU)

### Implementation

**Backend** (`handlers_atd.go`):
- New endpoint: `POST /api/generate-document` body: `{"query": "..."}`
- New endpoint: `GET /api/documents` — list cached docs
- New endpoint: `GET /api/documents/:idx` — retrieve specific cached doc
- Uses ATD CLI tools: `atd search`, `atd trace`, `atd assemble`

**Frontend** (`webui/static/js/search.js`):
- Add "📄 Generate Document" button below search results
- On click: POST query to `/api/generate-document`, show loading spinner
- On response: open a document viewer modal with rendered HTML + download button

**Frontend** (new: `webui/static/js/documents.js`):
- Document viewer modal component
- Documents dropdown in the header showing cached documents
- Download button generating `.md` blob

**ATD atoms**:
- `mechanic_webui_document_generation.atom.md` — the pipeline
- `ui_webui_document_viewer.atom.md` — the modal UI

---

## Execution Order

| # | Task | Depends On | Scope |
|---|------|-----------|-------|
| 1 | Explorer UX fixes (2a-2e) | Nothing | WebUI only |
| 2 | ATD audit script (1) | Nothing | CLI script |
| 3 | `atd summary` CLI/MCP (3) | Nothing | CLI + MCP |
| 4 | Refactor webui summary (4) | Task 3 | WebUI backend |
| 5 | Document generation (5) | Task 3 | WebUI full-stack |

Tasks 1, 2, 3 can proceed in parallel. Tasks 4 and 5 require Task 3.
