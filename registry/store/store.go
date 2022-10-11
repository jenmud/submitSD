package store

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jenmud/submitSD/registry/graph/model"
)

// DefaultConfig is a config with sensible values.
var DefaultConfig = Config{
	TTL:             30 * time.Second,
	CleanupInterval: time.Minute,
}

// Config is the registry configurations.
type Config struct {
	// TTL is the default TTL of a service is a TTL is not provided.
	TTL time.Duration
	// CleanupInterval is how often the cleanup routine will run.
	CleanupInterval time.Duration
	// Callback is a callback function which is called when a service is expired and removed.
	Callback Callback
}

// New returns a new empty registry store.
func New(cfg Config) *Registry {
	r := &Registry{
		items:  make(map[string]model.Service),
		Config: cfg,
		done:   make(chan bool),
	}

	go r.init()

	return r
}

// Callback is a callback function which is called when a service is expired.
type Callback func(event model.Event)

type Registry struct {
	lock   sync.RWMutex
	Config Config
	items  map[string]model.Service
	done   chan bool
}

// init starts the cleanup routine and sets up the registry.
func (r *Registry) init() {
	ticker := time.NewTicker(r.Config.CleanupInterval)
	for {
		select {
		case <-r.done:
			ticker.Stop()
			log.Print("cleanup ticker stopped")
			return
		case <-ticker.C:
			log.Print("cleanup run")
			r.Items()
		}
	}
}

// SetCallback overrides the config callback function with the provided.
func (r *Registry) SetCallback(clb Callback) {
	r.Config.Callback = clb
}

// publish publishes a event.
func (r *Registry) publish(action model.Action, service model.Service) {
	if r.Config.Callback != nil {
		r.Config.Callback(
			model.Event{
				Timestamp: time.Now(),
				Event:     action,
				Service:   &service,
			},
		)
	}
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
		Config:    input.Config,
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
		service.TTL = r.Config.TTL.String()
	}

	expiry, err := time.ParseDuration(service.TTL)
	if err != nil {
		return model.Service{}, err
	}

	service.ExpiresAt = time.Now().Add(expiry)
	r.items[service.ID] = service

	r.publish(model.ActionCreated, service)
	return service, nil
}

func (r *Registry) remove(id string) (model.Service, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	service, ok := r.items[id]
	if !ok {
		return model.Service{}, fmt.Errorf("could not find service with ID: %s", id)
	}

	delete(r.items, id)
	return service, nil
}

// Remove removes the service from the registery.
func (r *Registry) Remove(id string) error {
	service, err := r.remove(id)
	if err != nil {
		return err
	}

	r.publish(model.ActionExpired, service)
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

	r.publish(model.ActionRenewed, service)
	return service, nil
}

// Closed closed down the registry.
func (r *Registry) Close() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.done <- true
	return nil
}
