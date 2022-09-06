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
// true is returned if the service has expired and been removed.
func (s *Store) expireAndRemove(service Service) (bool, error) {
	if service.ExpiresAt.After(time.Now()) {
		return false, nil
	}

	if err := s.expire(service); err != nil {
		return false, err
	}

	return true, nil
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
