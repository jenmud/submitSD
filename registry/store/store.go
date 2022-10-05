package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jenmud/submitSD/registry/graph/model"
)

const DefaultTTL = "30s"

// New returns a new empty registry store.
func New() *Registry {
	return &Registry{items: make(map[string]model.Service)}
}

type Registry struct {
	lock  sync.RWMutex
	items map[string]model.Service
	// subscribers []chan
}

// has returns true if the there is a item with the provided ID.
func (r *Registry) has(id string) bool {
	_, err := r.Get(id)
	return err == nil
}

// Get fetches the service that has the provided id.
func (r *Registry) Get(id string) (model.Service, error) {
	r.lock.RLock()

	service, ok := r.items[id]
	if !ok {
		r.lock.RUnlock()
		return model.Service{}, fmt.Errorf("could not find service with ID: %s", id)
	}

	r.lock.RUnlock()

	/*
		Check that the service has expired and remove it if it has expired.
	*/
	if time.Now().After(service.ExpiresAt) {
		if err := r.Expire(service.ID); err != nil {
			return model.Service{}, err
		}
		return model.Service{}, fmt.Errorf("service has expired at: %s", service.ExpiresAt.Format(time.RFC3339))
	}

	return service, nil
}

// Items returns all the registered items in the registry automatically expiring services that has expired.
func (r *Registry) Items() []model.Service {
	items := []model.Service{}
	expired := []model.Service{}

	r.lock.RLock()
	for _, service := range r.items {
		if time.Now().After(service.ExpiresAt) {
			expired = append(expired, service)
		} else {
			items = append(items, service)
		}
	}
	r.lock.RUnlock()

	/*
		Clean up all the expired items
	*/
	for _, service := range expired {
		r.Expire(service.ID)
	}

	return items
}

func (r *Registry) Add(ctx context.Context, input model.NewServiceInput) (model.Service, error) {
	var id string

	if input.ID != nil {
		id = *input.ID
	} else {
		id = uuid.NewString()
	}

	if r.has(id) {
		return model.Service{}, fmt.Errorf("duplicate ID found: %s", id)
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	service := model.Service{
		ID:        id,
		Name:      input.Name,
		Address:   input.Address,
		CreatedAt: time.Now(),
		TTL:       *input.TTL,
	}

	if input.Description != nil {
		service.Description = *input.Description
	}

	if input.Version != nil {
		service.Version = *input.Version
	}

	if input.Type != nil {
		service.Type = *input.Type
	}

	if input.TTL != nil {
		service.TTL = *input.TTL
	} else {
		service.TTL = DefaultTTL
	}

	expiry, err := time.ParseDuration(service.TTL)
	if err != nil {
		return model.Service{}, err
	}

	service.ExpiresAt = time.Now().Add(expiry)
	r.items[service.ID] = service
	return service, nil
}

// Remove removes the service from the registery.
func (r *Registry) Remove(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	delete(r.items, id)
	return nil
}

// Expire expires the service and removes it from the registry.
func (r *Registry) Expire(id string) error {
	return r.Remove(id)
}

// Renew returns the services expiry.
func (r *Registry) Renew(id string, ttl time.Duration) (model.Service, error) {
	service, err := r.Get(id)
	if err != nil {
		return model.Service{}, fmt.Errorf("could not find service with ID: %s", id)
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	service.TTL = ttl.String()
	service.ExpiresAt = time.Now().Add(ttl)
	r.items[service.ID] = service
	return service, nil
}
