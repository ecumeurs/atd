package mcp

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

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

	mu                  sync.Mutex
	pending             map[any]chan *Response
	nextReqID           int64
	sender              func(any)
	clientSupportsRoots bool
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		tools:   make(map[string]entry),
		pending: make(map[any]chan *Response),
	}
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

// SetSender sets the callback used to send raw messages to the client.
func (r *Registry) SetSender(sender func(any)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sender = sender
}

// SendRequest sends a JSON-RPC request to the client and waits for the response.
func (r *Registry) SendRequest(method string, params any) (*Response, error) {
	r.mu.Lock()
	r.nextReqID++
	id := r.nextReqID
	ch := make(chan *Response, 1)
	r.pending[id] = ch
	sender := r.sender
	r.mu.Unlock()

	if sender == nil {
		r.mu.Lock()
		delete(r.pending, id)
		r.mu.Unlock()
		return nil, fmt.Errorf("no sender configured")
	}

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}
	if params != nil {
		b, err := json.Marshal(params)
		if err == nil {
			req.Params = b
		}
	}

	sender(req)

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(5 * time.Second):
		r.mu.Lock()
		delete(r.pending, id)
		r.mu.Unlock()
		return nil, fmt.Errorf("timeout waiting for response")
	}
}
