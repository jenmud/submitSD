package registry

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jenmud/submitSD/registry/proto"
)

// New returns a empty new store
func New(cfg Config) *Store {
	return &Store{cfg: cfg, reg: make(map[string]Service)}
}

// Config is the configuration for the store
type Config struct {
	// CleanupInterval is how frequently the cleanup runs over services expiring those that are expired.
	CleanupInterval time.Duration
}

// EvictedClb is a callback that is called with the service that was evicted from the store
// including when it is removed from the store manually.
type EvictedClb func(Service)

// Store storing registered services
type Store struct {
	cfg        Config
	lock       sync.RWMutex
	reg        map[string]Service
	evictedClb EvictedClb
	proto.UnimplementedRegistryServer
}

// SetEvictedCallback sets the callback to be called when a service has been evicted
func (s *Store) SetEvictedCallback(callback EvictedClb) {
	s.evictedClb = callback
}

func (s *Store) add(service Service) error {
	s.lock.RLock()
	if service, ok := s.reg[service.UUID]; ok {
		s.lock.RUnlock()
		return fmt.Errorf("duplicate service %s", service.UUID)
	}
	s.lock.RUnlock()

	s.lock.Lock()
	s.reg[service.UUID] = service
	s.lock.Unlock()
	return nil
}

// DeleteExpired removes all expired services.
func (s *Store) DeleteExpired() {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, service := range s.reg {
		if service.HasExpired() {
			s.expire(service)
		}
	}
}

// Add adds a new service to the store
func (s *Store) Add(ctx context.Context, req *proto.AddReq) (*proto.AddResp, error) {
	serviceReq := req.GetService()

	service := Service{}
	if err := service.FromPB(serviceReq); err != nil {
		return nil, err
	}

	if err := s.add(service); err != nil {
		return nil, err
	}

	return &proto.AddResp{Service: service.ToPB()}, nil
}

// expire expires a service and removes it from the store.
func (s *Store) expire(service Service) {
	delete(s.reg, service.UUID)

	if s.evictedClb != nil {
		s.evictedClb(service)
	}
}

// expireAndRemove checks is a service has expired and will remove it if it has expired.
// true is returned if the service has expired and been removed.
func (s *Store) expireAndRemove(service Service) (bool, error) {
	if service.HasExpired() {
		s.lock.Lock()
		s.expire(service)
		s.lock.Unlock()
		return true, nil
	}

	return false, nil
}

// fetch fetches the service from the store expiring it if the service has already expired.
func (s *Store) fetch(service Service) (Service, error) {
	s.lock.RLock()

	service, ok := s.reg[service.UUID]
	if !ok {
		s.lock.RUnlock()
		return Service{}, errors.New("no service found")
	}

	s.lock.RUnlock()

	ok, err := s.expireAndRemove(service)
	if err != nil {
		return Service{}, err
	}

	if ok {
		return Service{}, errors.New("service has expired")
	}

	return service, nil
}

// GetByUUID fetches a service by uuid
func (s *Store) GetByUUID(ctx context.Context, req *proto.GetByUUIDReq) (*proto.GetByUUIDResp, error) {
	service, err := s.fetch(Service{UUID: req.GetUuid()})
	if err != nil {
		return nil, err
	}

	return &proto.GetByUUIDResp{Service: service.ToPB()}, nil
}
