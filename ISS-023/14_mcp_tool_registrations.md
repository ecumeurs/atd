# Task 14 — MCP Tool Registrations

**Depends on:** Task 13 (MCP server infrastructure)  
**Produces:** Populated `cmd/mcp_tools.go` — all ATD subcommands exposed as MCP tools  
**Note:** Only wire up tools that are already migrated. If a subcommand from Phase 2–3 is not yet done, skip it and add a TODO comment.

---

## Context

Each MCP tool wraps an existing Cobra subcommand's **internal logic** (not the CLI binary). We reuse the exact same functions already tested in Phase 2 and Phase 3. The binding pattern is:

1. Extract arguments from the `map[string]any` that MCP sends
2. Call the extracted `runX(...)` helper function (see refactoring pattern below)
3. Return the result as a `string` (JSON already marshalled, or plain text)

**Registry API** (from `internal/mcp/registry.go`):
```go
r.Register(mcp.Tool{Name: "...", Description: "...", InputSchema: map[string]any{...}}, handlerFunc)
```
The type is `mcp.Tool` (not `mcp.ToolSchema`). `HandlerFunc` signature: `func(args map[string]any) (string, error)`.

---

## Helper: `argString` and `argBool`

Add these private helpers at the top of `mcp_tools.go` to safely extract arguments:

```go
func argString(args map[string]any, key, fallback string) string {
    if v, ok := args[key]; ok {
        if s, ok := v.(string); ok {
            return s
        }
    }
    return fallback
}

func argBool(args map[string]any, key string) bool {
    if v, ok := args[key]; ok {
        if b, ok := v.(bool); ok {
            return b
        }
    }
    return false
}
```

---

## MCP Schema Convention

Every tool's `InputSchema` follows JSON Schema draft-07:

```go
map[string]any{
    "type": "object",
    "properties": map[string]any{
        "field_name": map[string]any{
            "type":        "string",
            "description": "What this field does.",
        },
    },
    "required": []string{"field_name"}, // only truly required fields
}
```

Optional fields are listed in `properties` but omitted from `required`.

---

## 14.1 — `atd_query`

**Source:** logic in `cmd/query.go`  
**What it does:** Deterministic search through atom frontmatter fields.

```go
r.Register(mcp.Tool{
    Name:        "atd_query",
    Description: "Search ATD atoms by frontmatter field value. Returns JSON array of matching atoms.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "field":  map[string]any{"type": "string", "description": "Frontmatter field to search (e.g. 'type', 'status', 'id')."},
            "search": map[string]any{"type": "string", "description": "Value to match (case-insensitive substring)."},
        },
        "required": []string{"search"},
    },
}, func(args map[string]any) (string, error) {
    field  := argString(args, "field", "")
    search := argString(args, "search", "")
    // Call the same logic used by queryCmd.RunE, but return string instead of printing.
    // The query logic lives in cmd/query.go: call runQuery(field, search) if extracted to a function,
    // OR replicate the 30-line walk here directly.
    return runQuery(field, search)
})
```

> **Action required:** Refactor `cmd/query.go` to extract its core logic into `func runQuery(field, search string) (string, error)` that returns the JSON string instead of printing it. The Cobra handler becomes a one-liner: `text, err := runQuery(...); fmt.Println(text)`.

---

## 14.2 — `atd_crawl`

**Source:** logic in `cmd/crawl.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_crawl",
    Description: "Crawl ATD docs and source code. Returns a dependency graph JSON. Set gaps=true to list STABLE atoms with no implementations.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "src":  map[string]any{"type": "string", "description": "Path to source code directory (optional)."},
            "gaps": map[string]any{"type": "boolean", "description": "If true, return only orphaned STABLE atoms."},
            "docs": map[string]any{"type": "string", "description": "Override docs directory path."},
        },
    },
}, func(args map[string]any) (string, error) {
    src  := argString(args, "src", "")
    docs := argString(args, "docs", config.DocsDir())
    gaps := argBool(args, "gaps")
    return runCrawl(src, docs, gaps)
})
```

> **Action required:** Refactor `cmd/crawl.go` → extract `func runCrawl(src, docs string, gaps bool) (string, error)`.

---

## 14.3 — `atd_weave`

**Source:** logic in `cmd/weave.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_weave",
    Description: "Populate the dependents[] array in ATD atoms by scanning parents references. Bi-directional link weaving.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{},
    },
}, func(args map[string]any) (string, error) {
    return runWeave()
})
```

> **Action required:** Refactor `cmd/weave.go` → extract `func runWeave() (string, error)` that returns a summary string (e.g. `"Weaved 14 links across 8 atoms"`).

---

## 14.4 — `atd_update`

**Source:** logic in `cmd/update.go`  
This is the most important write-path tool. MCP schema validation prevents bad flag combinations.

```go
r.Register(mcp.Tool{
    Name:        "atd_update",
    Description: "Surgically update fields in an ATD atom file without rewriting it. Pass set as 'key=value' pairs.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "file":      map[string]any{"type": "string", "description": "Absolute or relative path to the .atom.md file."},
            "set":       map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Frontmatter edits as 'key=value' strings, e.g. [\"status=STABLE\",\"priority=CORE\"]."},
            "intent":    map[string]any{"type": "string", "description": "New INTENT section text."},
            "logic":     map[string]any{"type": "string", "description": "New THE RULE / LOGIC section text."},
            "interface": map[string]any{"type": "string", "description": "New TECHNICAL INTERFACE section text."},
            "spec_link": map[string]any{"type": "string", "description": "Atom ID to prepend as @spec-link in a source file (requires spec_link_file)."},
            "spec_link_file": map[string]any{"type": "string", "description": "Source file path for --spec-link injection."},
        },
        "required": []string{"file"},
    },
}, func(args map[string]any) (string, error) {
    file      := argString(args, "file", "")
    intent    := argString(args, "intent", "")
    logic     := argString(args, "logic", "")
    iface     := argString(args, "interface", "")
    specLink  := argString(args, "spec_link", "")
    specFile  := argString(args, "spec_link_file", "")

    var setPairs []string
    if raw, ok := args["set"]; ok {
        if arr, ok := raw.([]any); ok {
            for _, item := range arr {
                if s, ok := item.(string); ok {
                    setPairs = append(setPairs, s)
                }
            }
        }
    }

    return runUpdate(file, setPairs, intent, logic, iface, specLink, specFile)
})
```

> **Action required:** Refactor `cmd/update.go` → extract `func runUpdate(file string, set []string, intent, logic, iface, specLink, specLinkFile string) (string, error)`.

---

## 14.5 — `atd_roadmap`

**Source:** logic in `cmd/roadmap.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_roadmap",
    Description: "Scan a source directory and build a complexity roadmap JSON identifying high-density files.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "dir": map[string]any{"type": "string", "description": "Directory to scan."},
            "out": map[string]any{"type": "string", "description": "Optional output file path for roadmap.json."},
        },
        "required": []string{"dir"},
    },
}, func(args map[string]any) (string, error) {
    dir := argString(args, "dir", "")
    out := argString(args, "out", "")
    return runRoadmap(dir, out)
})
```

> **Action required:** Refactor `cmd/roadmap.go` → extract `func runRoadmap(dir, out string) (string, error)`.

---

## 14.6 — `atd_verify`

**Source:** logic in `cmd/verify.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_verify",
    Description: "Run git diff, extract @spec-link tags, and produce an audit prompt for the IDE Agent.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{},
    },
}, func(args map[string]any) (string, error) {
    return runVerify()
})
```

> **Action required:** Refactor `cmd/verify.go` → extract `func runVerify() (string, error)`.

---

## 14.7 — `atd_assemble`

**Source:** logic in `cmd/assemble.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_assemble",
    Description: "Assemble ATD atoms into a readable document following dependency links from start atom IDs.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "starts":  map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "description": "Atom IDs to start assembly from."},
            "purpose": map[string]any{"type": "string", "description": "Purpose context for the assembled document."},
        },
        "required": []string{"starts"},
    },
}, func(args map[string]any) (string, error) {
    var starts []string
    if raw, ok := args["starts"]; ok {
        if arr, ok := raw.([]any); ok {
            for _, item := range arr {
                if s, ok := item.(string); ok {
                    starts = append(starts, s)
                }
            }
        }
    }
    purpose := argString(args, "purpose", "")
    return runAssemble(starts, purpose)
})
```

> **Action required:** Refactor `cmd/assemble.go` → extract `func runAssemble(starts []string, purpose string) (string, error)`.

---

## 14.8 — `atd_test_links`

**Source:** logic in `cmd/test_links.go`  

```go
r.Register(mcp.Tool{
    Name:        "atd_test_links",
    Description: "Discover @test-link tags in source files and map them to atom IDs.",
    InputSchema: map[string]any{
        "type": "object",
        "properties": map[string]any{
            "atom": map[string]any{"type": "string", "description": "Filter by atom ID (optional)."},
            "src":  map[string]any{"type": "string", "description": "Source directory to scan."},
        },
        "required": []string{"src"},
    },
}, func(args map[string]any) (string, error) {
    atom := argString(args, "atom", "")
    src  := argString(args, "src", "")
    return runTestLinks(atom, src)
})
```

> **Action required:** Refactor `cmd/test_links.go` → extract `func runTestLinks(atom, src string) (string, error)`.

---

## Refactoring Pattern (Apply to Every Command Above)

For each command `cmd/X.go`, the refactor is always the same shape:

**Before:**
```go
var xCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        flag, _ := cmd.Flags().GetString("flag")
        // ... 40 lines of logic ...
        fmt.Println(string(output))
        return nil
    },
}
```

**After:**
```go
// runX contains the testable core logic.
func runX(flag string) (string, error) {
    // ... same 40 lines, return string(output) instead of fmt.Println ...
}

var xCmd = &cobra.Command{
    RunE: func(cmd *cobra.Command, args []string) error {
        flag, _ := cmd.Flags().GetString("flag")
        text, err := runX(flag)
        if err != nil { return err }
        fmt.Println(text)
        return nil
    },
}
```

This pattern makes every command independently testable and MCP-wrappable with no duplication.

---

## Acceptance Criteria

- [ ] `go build ./...` from `scripts/cmd/atd/` succeeds
- [ ] `POST /tools/list` returns at least 8 tools after `RegisterMCPTools` is complete
- [ ] `POST /tools/call` with `{"name":"atd_query","arguments":{"search":"movement"}}` returns JSON atom results
- [ ] `POST /tools/call` with `{"name":"atd_crawl","arguments":{"gaps":true}}` returns gap report JSON
- [ ] `POST /tools/call` with `{"name":"atd_update","arguments":{"file":"path/to/atom.md","set":["status=STABLE"]}}` modifies the file and returns success message
- [ ] All previously passing `go test` suites still pass after the refactors
