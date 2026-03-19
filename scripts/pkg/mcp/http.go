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
	flusher, _ := w.(http.Flusher)
	if flusher != nil {
		flusher.Flush()
	}

	sendCh := make(chan any, 100)
	r.SetSender(func(msg any) {
		sendCh <- msg
	})

	// Keep connection open and forward server messages
	for {
		select {
		case <-req.Context().Done():
			r.SetSender(nil)
			return
		case msg := <-sendCh:
			b, err := json.Marshal(msg)
			if err == nil {
				fmt.Fprintf(w, "event: message\ndata: %s\n\n", string(b))
				if flusher != nil {
					flusher.Flush()
				}
			}
		}
	}
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
