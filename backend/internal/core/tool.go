// Package core defines the contracts shared by every tool module.
//
// A tool is a self-contained unit of functionality (formatter, generator,
// converter, ...) that exposes its own REST routes. The API layer only
// depends on these interfaces, never on concrete tool implementations,
// which keeps the dependency direction pointing inwards.
package core

import "github.com/go-chi/chi/v5"

// Category groups tools in the UI and the discovery endpoint.
type Category string

const (
	CategoryFormatter Category = "Formatter"
	CategoryGenerator Category = "Generators"
)

// Tool is the contract every tool module must satisfy to be mounted
// under the versioned API.
type Tool interface {
	// ID is the URL-safe, unique identifier of the tool
	// (e.g. "json", "hash", "qrcode"). Routes are mounted at
	// /api/v1/tools/{ID}.
	ID() string

	// Name is the human-readable name shown in discovery metadata.
	Name() string

	// Category buckets the tool for navigation purposes.
	Category() Category

	// Description is a short, one-line summary of the tool.
	Description() string

	// RegisterRoutes attaches the tool's endpoints to the given
	// sub-router, which is already scoped to /api/v1/tools/{ID}.
	RegisterRoutes(r chi.Router)
}

// Registry collects tools in a deterministic order so the API layer can
// mount and describe them without knowing their concrete types.
type Registry struct {
	tools []Tool
	byID  map[string]Tool
}

// NewRegistry returns an empty tool registry.
func NewRegistry() *Registry {
	return &Registry{byID: make(map[string]Tool)}
}

// MustRegister adds a tool to the registry and panics on a duplicate ID.
// Registration happens once at startup, so failing fast is the right
// behaviour: a duplicate ID is always a programming error.
func (reg *Registry) MustRegister(t Tool) {
	if _, exists := reg.byID[t.ID()]; exists {
		panic("core: duplicate tool ID registered: " + t.ID())
	}
	reg.byID[t.ID()] = t
	reg.tools = append(reg.tools, t)
}

// Tools returns the registered tools in registration order.
func (reg *Registry) Tools() []Tool {
	out := make([]Tool, len(reg.tools))
	copy(out, reg.tools)
	return out
}
