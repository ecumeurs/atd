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
