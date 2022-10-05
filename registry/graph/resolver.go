package graph

import "github.com/jenmud/submitSD/registry/store"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{ store *store.Registry }

// NewResolver returns a new resolver.
func NewResolver() *Resolver {
	return &Resolver{store: store.New()}
}
