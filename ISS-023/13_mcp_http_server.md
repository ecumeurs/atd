# Task 13 — MCP Server Infrastructure

**Depends on:** All previous tasks complete (Phase 1–4)  
**Produces:**
- `internal/mcp/types.go` — JSON-RPC 2.0 + MCP protocol types
- `internal/mcp/registry.go` — Tool registration system
- `internal/mcp/stdio.go` — stdio transport (primary)
- `internal/mcp/http.go` — Streamable HTTP transport (secondary)
- `cmd/serve.go` — `atd serve` Cobra subcommand
- `cmd/mcp_tools.go` — empty scaffold (populated in Task 14)

**No new Go dependencies** — stdlib only (`net/http`, `encoding/json`, `bufio`, `os`).

---

## Protocol Overview (MCP 2025-11-25)

MCP uses **JSON-RPC 2.0** as its wire format over two transport options:

### Wire Format

```
// Request (client → server)
{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}

// Result response (server → client)
{"jsonrpc":"2.0","id":1,"result":{...}}

// Error response (server → client)
{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"method not found"}}

// Notification (no id, no reply expected)
{"jsonrpc":"2.0","method":"notifications/initialized"}
```

### MCP Methods we implement

| Method | Direction | Description |
|---|---|---|
| `initialize` | client→server | Handshake, exchange capabilities |
| `notifications/initialized` | client→server | Client acknowledges ready |
| `tools/list` | client→server | Get all available tools |
| `tools/call` | client→server | Invoke a tool by name |
| `ping` | either | Keepalive |

### Transports

**1. stdio (primary, simpler):** Server reads JSON-RPC messages from stdin, one per line. Server writes responses to stdout, one per line. Errors written to stderr.

**2. Streamable HTTP (secondary):** Single endpoint at `/mcp` supporting:
- `POST /mcp` — client sends one JSON-RPC message; server replies with `application/json` (simple case) or opens an SSE stream (`text/event-stream`)
- `GET /mcp` — client opens a long-lived SSE stream to receive server-initiated messages
- `DELETE /mcp` — client terminates a session  
- Session tracked via `MCP-Session-Id` header (UUID, server-generated at initialization)

For the ATD use case, **stdio is easiest and sufficient for VS Code / Claude Desktop**. HTTP is provided for clients that need it.

---

## 13.1 — Create `internal/mcp/types.go`

Create file at: `scripts/internal/mcp/types.go`

```go
package mcp

import "encoding/json"

// --- JSON-RPC 2.0 wire types ---

// Request is a JSON-RPC 2.0 request message.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`       // must be "2.0"
	ID      any             `json:"id"`            // string | number, never null
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 result response.
type Response struct {
	JSONRPC string `json:"jsonrpc"` // "2.0"
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

// RPCError is the error object inside an error response.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Notification is a JSON-RPC 2.0 notification (no id, no reply).
type Notification struct {
	JSONRPC string `json:"jsonrpc"` // "2.0"
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternalError  = -32603
)

// --- MCP-specific types ---

// InitializeParams is the params for the "initialize" method.
type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

// ClientInfo describes the connecting client.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializeResult is the result for the "initialize" method.
type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
	Instructions    string         `json:"instructions,omitempty"`
}

// ServerInfo describes this server.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool describes one callable tool for tools/list.
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ToolsListResult is the result for tools/list.
type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

// ToolCallParams is the params for tools/call.
type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

// ContentBlock is a single output element in a tool result.
type ContentBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

// ToolCallResult is the result for tools/call.
type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}
```

---

## 13.2 — Create `internal/mcp/registry.go`

Create file at: `scripts/internal/mcp/registry.go`

```go
package mcp

import "fmt"

// HandlerFunc is the signature every registered tool must implement.
// args is the decoded arguments map. Returns output text or an error.
type HandlerFunc func(args map[string]any) (string, error)

// entry is one registered tool.
type entry struct {
	Tool    Tool
	Handler HandlerFunc
}

// Registry holds all registered tools.
type Registry struct {
	tools map[string]entry
	order []string
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]entry)}
}

// Register adds a tool. name must be unique and match Tool.Name.
func (r *Registry) Register(t Tool, h HandlerFunc) {
	r.tools[t.Name] = entry{Tool: t, Handler: h}
	r.order = append(r.order, t.Name)
}

// List returns all tools in registration order.
func (r *Registry) List() []Tool {
	out := make([]Tool, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.tools[name].Tool)
	}
	return out
}

// Call executes a registered tool by name.
func (r *Registry) Call(name string, args map[string]any) (string, error) {
	e, ok := r.tools[name]
	if !ok {
		return "", fmt.Errorf("unknown tool: %q", name)
	}
	return e.Handler(args)
}
```

---

## 13.3 — Create `internal/mcp/handler.go`

This file contains the shared method dispatch logic used by both transports.

Create file at: `scripts/internal/mcp/handler.go`

```go
package mcp

import (
	"encoding/json"
	"fmt"
)

// Handler processes one JSON-RPC request and returns a Response.
// Returns nil for notifications (no reply expected).
func (r *Registry) Handle(raw []byte) *Response {
	// Try to decode as a generic message to find id and method.
	var msg struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      any             `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return errResponse(nil, CodeParseError, "parse error")
	}

	// Notifications have no id — do not reply.
	if msg.ID == nil && msg.Method != "" {
		r.handleNotification(msg.Method, msg.Params)
		return nil
	}

	return r.dispatch(msg.ID, msg.Method, msg.Params)
}

func (r *Registry) dispatch(id any, method string, rawParams json.RawMessage) *Response {
	switch method {
	case "initialize":
		return r.handleInitialize(id, rawParams)
	case "tools/list":
		return okResponse(id, ToolsListResult{Tools: r.List()})
	case "tools/call":
		return r.handleToolsCall(id, rawParams)
	case "ping":
		return okResponse(id, map[string]any{})
	default:
		return errResponse(id, CodeMethodNotFound, fmt.Sprintf("method not found: %s", method))
	}
}

func (r *Registry) handleInitialize(id any, rawParams json.RawMessage) *Response {
	var p InitializeParams
	if err := json.Unmarshal(rawParams, &p); err != nil {
		return errResponse(id, CodeInvalidParams, "invalid initialize params")
	}
	result := InitializeResult{
		ProtocolVersion: "2025-11-25",
		Capabilities: map[string]any{
			"tools": map[string]any{},
		},
		ServerInfo: ServerInfo{Name: "atd", Version: "1.0.0"},
		Instructions: "ATD MCP server. Use tools/list to discover available ATD operations.",
	}
	return okResponse(id, result)
}

func (r *Registry) handleToolsCall(id any, rawParams json.RawMessage) *Response {
	var p ToolCallParams
	if err := json.Unmarshal(rawParams, &p); err != nil {
		return errResponse(id, CodeInvalidParams, "invalid tools/call params")
	}
	text, err := r.Call(p.Name, p.Arguments)
	if err != nil {
		return okResponse(id, ToolCallResult{
			IsError: true,
			Content: []ContentBlock{{Type: "text", Text: err.Error()}},
		})
	}
	return okResponse(id, ToolCallResult{
		Content: []ContentBlock{{Type: "text", Text: text}},
	})
}

func (r *Registry) handleNotification(method string, _ json.RawMessage) {
	// notifications/initialized — no action needed for basic server.
	// Future: could log or trigger post-init hooks.
}

// --- helpers ---

func okResponse(id any, result any) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func errResponse(id any, code int, msg string) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}
```

---

## 13.4 — Create `internal/mcp/stdio.go`

stdio is the **primary transport**. The server reads newline-delimited JSON from stdin and writes to stdout.

Create file at: `scripts/internal/mcp/stdio.go`

```go
package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// ServeStdio runs the MCP server over stdin/stdout.
// This is the primary transport — compatible with VS Code, Claude Desktop,
// and any MCP host that launches the server as a subprocess.
// Blocking — returns only on EOF or read error.
func (r *Registry) ServeStdio() error {
	scanner := bufio.NewScanner(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		resp := r.Handle(line)
		if resp == nil {
			// Notification — no reply.
			continue
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "[mcp/stdio] encode error: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return err
	}
	return nil
}
```

---

## 13.5 — Create `internal/mcp/http.go`

HTTP Streamable transport. Single `/mcp` endpoint for `POST` and `GET`.

Create file at: `scripts/internal/mcp/http.go`

```go
package mcp

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// sessions tracks active session IDs (in-memory, no persistence).
var sessions = map[string]bool{}

func newSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// ServeHTTP starts a blocking HTTP server on addr (e.g. ":7474").
// Registers the single /mcp endpoint for POST and GET.
func (r *Registry) ServeHTTP(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", r.mcpHandler)

	log.Printf("[mcp/http] Listening on http://localhost%s/mcp", addr)
	return http.ListenAndServe(addr, mux)
}

func (r *Registry) mcpHandler(w http.ResponseWriter, req *http.Request) {
	// Security: validate Origin header to prevent DNS rebinding.
	origin := req.Header.Get("Origin")
	if origin != "" && origin != "null" {
		// Allow localhost origins for local dev. Reject others.
		if !isAllowedOrigin(origin) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
	}

	switch req.Method {
	case http.MethodPost:
		r.handlePost(w, req)
	case http.MethodGet:
		r.handleGet(w, req)
	case http.MethodDelete:
		r.handleDelete(w, req)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Registry) handlePost(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(io.LimitReader(req.Body, 1<<20))
	if err != nil {
		writeJSONRPCError(w, http.StatusBadRequest, nil, CodeInvalidRequest, "cannot read body")
		return
	}

	resp := r.Handle(body)

	// Notification: no response body.
	if resp == nil {
		w.WriteHeader(http.StatusAccepted)
		return
	}

	// For initialize: assign and return session ID header.
	if isInitializeResponse(resp) {
		sid := newSessionID()
		sessions[sid] = true
		w.Header().Set("MCP-Session-Id", sid)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}

// handleGet opens an SSE stream. For basic servers this can return 405.
// We implement a minimal SSE endpoint so spec-compliant clients don't error.
func (r *Registry) handleGet(w http.ResponseWriter, req *http.Request) {
	// Check the client accepts SSE.
	accept := req.Header.Get("Accept")
	if accept != "" && accept != "*/*" && !containsType(accept, "text/event-stream") {
		http.Error(w, "not acceptable", http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	// Prime the client with an event ID for reconnection.
	fmt.Fprintf(w, "id: 0\ndata: \n\n")
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Keep connection open until client disconnects.
	<-req.Context().Done()
}

func (r *Registry) handleDelete(w http.ResponseWriter, req *http.Request) {
	sid := req.Header.Get("MCP-Session-Id")
	if sid != "" {
		delete(sessions, sid)
	}
	w.WriteHeader(http.StatusOK)
}

// --- helpers ---

func isInitializeResponse(resp *Response) bool {
	if resp == nil || resp.Result == nil {
		return false
	}
	m, ok := resp.Result.(map[string]any)
	if !ok {
		// Check if it's an InitializeResult struct
		_, isInit := resp.Result.(InitializeResult)
		return isInit
	}
	_, hasProtoVersion := m["protocolVersion"]
	return hasProtoVersion
}

func isAllowedOrigin(origin string) bool {
	allowed := []string{
		"http://localhost", "http://127.0.0.1",
		"https://localhost", "https://127.0.0.1",
	}
	for _, a := range allowed {
		if len(origin) >= len(a) && origin[:len(a)] == a {
			return true
		}
	}
	return false
}

func containsType(header, contentType string) bool {
	return len(header) > 0 && (header == contentType ||
		len(header) > len(contentType) && header[:len(contentType)] == contentType)
}

func writeJSONRPCError(w http.ResponseWriter, httpStatus int, id any, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(errResponse(id, code, msg))
}

// io.LimitReader helper shim.
func init() {
	_ = io.LimitReader // ensure import is used
}
```

---

## 13.6 — Create `cmd/serve.go`

Create file at: `scripts/cmd/atd/cmd/serve.go`

```go
package cmd

import (
	"fmt"

	"atd-tools/config"
	"atd-tools/internal/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the ATD MCP server",
	Long: `Start an MCP-compatible server exposing all ATD tools.

By default, the server uses stdio transport — suitable for VS Code, Claude Desktop,
and any MCP host that launches the server as a subprocess.

To use HTTP transport instead (Streamable HTTP, spec 2025-11-25):
  atd serve --http --port 7474

Example .mcp.json for VS Code (stdio, recommended):
  {
    "servers": {
      "atd": {
        "type": "stdio",
        "command": "/path/to/atd",
        "args": ["serve"]
      }
    }
  }

Example .mcp.json for VS Code (HTTP):
  {
    "servers": {
      "atd": {
        "type": "http",
        "url": "http://localhost:7474/mcp"
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		useHTTP, _ := cmd.Flags().GetBool("http")
		port, _ := cmd.Flags().GetInt("port")

		// Config must be loaded so all tool handlers can call config.DocsDir() etc.
		if err := config.Load(); err != nil {
			return fmt.Errorf("config: %w", err)
		}

		r := mcp.NewRegistry()
		RegisterMCPTools(r)

		if useHTTP {
			addr := fmt.Sprintf(":%d", port)
			return r.ServeHTTP(addr)
		}

		return r.ServeStdio()
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().Bool("http", false, "Use Streamable HTTP transport instead of stdio")
	serveCmd.Flags().IntP("port", "p", 7474, "HTTP port (only used with --http)")
}
```

---

## 13.7 — Create `cmd/mcp_tools.go` scaffold

Create file at: `scripts/cmd/atd/cmd/mcp_tools.go`

```go
package cmd

import "atd-tools/internal/mcp"

// RegisterMCPTools registers all ATD subcommands as MCP tools.
// Populated in Task 14.
func RegisterMCPTools(r *mcp.Registry) {
	// Task 14 adds registrations here.
}
```

---

## 13.8 — Verify build

From `scripts/cmd/atd/`:
```bash
go build ./...
```

Expected: compiles with no errors. `atd serve` appears in `atd --help`.

---

## 13.9 — Smoke test stdio transport

```bash
# Build
go build -o /tmp/atd .

# Send initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' \
  | /tmp/atd serve

# Expected output (one JSON line):
# {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{"tools":{}},"serverInfo":{"name":"atd","version":"1.0.0"},"instructions":"..."}}

# Then tools/list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | /tmp/atd serve
# Expected: {"jsonrpc":"2.0","id":2,"result":{"tools":[]}}
```

---

## 13.10 — Smoke test HTTP transport

```bash
# Terminal 1
/tmp/atd serve --http --port 7474

# Terminal 2
curl -s -X POST http://localhost:7474/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"curl-test","version":"1.0"}}}' | jq .

# Expected: {"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25",...}}
# Response headers should include: MCP-Session-Id: <uuid>

curl -s -X POST http://localhost:7474/mcp \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  -d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | jq .
# Expected: {"jsonrpc":"2.0","id":2,"result":{"tools":[]}}
```

---

## Acceptance Criteria

- [ ] `go build ./...` succeeds from `scripts/cmd/atd/`
- [ ] `atd serve --help` shows usage with `--http` and `--port` flags  
- [ ] stdio: `initialize` request returns correct `InitializeResult` with `protocolVersion: "2025-11-25"`
- [ ] stdio: `tools/list` returns `{"tools":[]}`
- [ ] stdio: unknown method returns JSON-RPC error `{"code":-32601,...}`
- [ ] stdio: notification `notifications/initialized` produces no output (no reply)
- [ ] HTTP: `POST /mcp` with `initialize` returns `MCP-Session-Id` header
- [ ] HTTP: `GET /mcp` returns `Content-Type: text/event-stream`
- [ ] HTTP: `DELETE /mcp` returns 200
- [ ] HTTP: invalid `Origin` header returns 403
