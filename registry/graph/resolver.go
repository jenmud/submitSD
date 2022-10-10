package graph

import (
	"sync"

	"github.com/jenmud/submitSD/registry/graph/model"
	"github.com/jenmud/submitSD/registry/store"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	lock        sync.RWMutex
	store       *store.Registry
	subscribers map[string]chan<- *model.Event
}

func (r *Resolver) Publish(event *model.Event) {
	r.lock.RLock()
	defer r.lock.RUnlock()
	for _, ch := range r.subscribers {
		ch <- event
	}
}

// NewResolver returns a new resolver.
func NewResolver() *Resolver {
	return &Resolver{
		store:       store.New(),
		subscribers: make(map[string]chan<- *model.Event),
	}
}
