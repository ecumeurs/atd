package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"atd-tools/config"
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
		Result  json.RawMessage `json:"result"`
		Error   *RPCError       `json:"error"`
	}
	if err := json.Unmarshal(raw, &msg); err != nil {
		return errResponse(nil, CodeParseError, "parse error")
	}

	// Notifications have no id — do not reply.
	if msg.ID == nil && msg.Method != "" {
		r.handleNotification(msg.Method, msg.Params)
		return nil
	}

	// Is it a response from the client?
	if msg.Method == "" && msg.ID != nil {
		r.mu.Lock()
		var idKey any
		switch v := msg.ID.(type) {
		case float64:
			idKey = int64(v)
		default:
			idKey = v
		}

		if ch, ok := r.pending[idKey]; ok {
			delete(r.pending, idKey)
			r.mu.Unlock()

			var resp Response
			json.Unmarshal(raw, &resp)
			ch <- &resp
		} else {
			r.mu.Unlock()
		}
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

	if p.Capabilities != nil {
		if _, ok := p.Capabilities["roots"]; ok {
			r.mu.Lock()
			r.clientSupportsRoots = true
			r.mu.Unlock()
		}
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
	if method == "notifications/initialized" {
		r.mu.Lock()
		supportsRoots := r.clientSupportsRoots
		r.mu.Unlock()

		if supportsRoots {
			go func() {
				resp, err := r.SendRequest("roots/list", nil)
				if err == nil && resp.Result != nil {
					b, _ := json.Marshal(resp.Result)
					var lr RootsListResult
					if err := json.Unmarshal(b, &lr); err == nil {
						if len(lr.Roots) > 0 {
							uri := lr.Roots[0].Uri
							if strings.HasPrefix(uri, "file://") {
								path := strings.TrimPrefix(uri, "file://")
								config.LoadFromDir(path)
							}
						}
					}
				}
			}()
		}
	}
}

// --- helpers ---

func okResponse(id any, result any) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Result: result}
}

func errResponse(id any, code int, msg string) *Response {
	return &Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: msg}}
}
