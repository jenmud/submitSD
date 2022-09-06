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
func New() *Store {
	return &Store{reg: make(map[string]Service)}
}

// Store storing registered services
type Store struct {
	lock sync.RWMutex
	reg  map[string]Service
	proto.UnimplementedRegistryServer
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
func (s *Store) expire(service Service) error {
	s.lock.Lock()
	delete(s.reg, service.UUID)
	s.lock.Unlock()
	return nil
}

// expireAndRemove checks is a service has expired and will remove it if it has expired.
func (s *Store) expireAndRemove(service Service) error {
	if service.ExpiresAt.Before(time.Now()) {
		return nil
	}
	return s.expire(service)
}

// fetch fetches the service from the store expiring it if the service has already expired.
func (s *Store) fetch(service Service) (Service, error) {
	s.lock.RUnlock()
	defer s.lock.Unlock()

	service, ok := s.reg[service.UUID]
	if !ok {
		return Service{}, errors.New("no service found")
	}

	if err := s.expireAndRemove(service); err != nil {
		return Service{}, err
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
