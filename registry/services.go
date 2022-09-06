package registry

import (
	"net/netip"
	"time"

	"github.com/jenmud/submitSD/registry/proto"
)

// Service is a service registered as a service
type Service struct {
	UUID        string         `json:"uuid"`
	Description string         `json:"description"`
	Version     string         `json:"version"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	IP          netip.AddrPort `json:"ip"`
	CreatedAt   time.Time      `json:"created_at"`
	ExpiresAt   time.Time      `json:"expires_at"`
	Expiry      time.Duration  `json:"expiry"`
}

// FromPB takes a service proto message and updates the service with the fields.
func (s Service) FromPB(service *proto.Service) error {
	ip, err := netip.ParseAddrPort(service.GetIp())
	if err != nil {
		return err
	}

	createdAt, err := time.Parse(time.RFC3339, service.GetCreatedAt())
	if err != nil {
		return err
	}

	expiry, err := time.ParseDuration(service.GetExpiry())
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(expiry)

	s.UUID = service.GetUuid()
	s.Description = service.GetDescription()
	s.Version = service.GetVersion()
	s.Name = service.GetName()
	s.Type = service.GetType()
	s.IP = ip
	s.CreatedAt = createdAt
	s.ExpiresAt = expiresAt
	s.Expiry = expiry

	return nil
}

// ToPB returns the service as a proto message.
func (s Service) ToPB() *proto.Service {
	return &proto.Service{
		Uuid:        s.UUID,
		Name:        s.Name,
		Description: s.Description,
		Version:     s.Version,
		Type:        s.Type,
		Ip:          s.IP.String(),
		CreatedAt:   s.CreatedAt.Format(time.RFC3339),
		ExpiresAt:   s.ExpiresAt.Format(time.RFC3339),
		Expiry:      s.Expiry.String(),
	}
}
