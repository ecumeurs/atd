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
