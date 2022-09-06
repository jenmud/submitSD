package registry

import (
	"net/netip"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jenmud/submitSD/registry/proto"
)

func TestStore_add(t *testing.T) {
	type fields struct {
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		service Service
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Successful",
			fields: fields{reg: make(map[string]Service)},
			args: args{Service{
				UUID:        "some-uuid",
				Description: "some test service",
				Version:     "v1.2.3",
				Name:        "myTestService",
				Type:        "MockService",
				IP:          netip.MustParseAddrPort("10.2.3.4:8080"),
				CreatedAt:   time.Date(2022, 9, 6, 22, 12, 0, 0, time.Local),
				ExpiresAt:   time.Date(2022, 9, 6, 22, 12, 0, 5, time.Local),
				Expiry:      5 * time.Second,
			}},
			wantErr: false,
		},
		{
			name:   "Failed",
			fields: fields{reg: map[string]Service{"some-uuid": {UUID: "some-uuid", Name: "service-already-exists"}}},
			args: args{Service{
				UUID:        "some-uuid",
				Description: "some test service",
				Version:     "v1.2.3",
				Name:        "myTestService",
				Type:        "MockService",
				IP:          netip.MustParseAddrPort("10.2.3.4:8080"),
				CreatedAt:   time.Date(2022, 9, 6, 22, 12, 0, 0, time.Local),
				ExpiresAt:   time.Date(2022, 9, 6, 22, 12, 0, 5, time.Local),
				Expiry:      5 * time.Second,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}
			if err := s.add(tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("Store.add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStore_expire(t *testing.T) {
	type fields struct {
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		service Service
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Successfully removed",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						ExpiresAt: time.Now().Add(-5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service",
					ExpiresAt: time.Now().Add(-5 * time.Second),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				lock:                        tt.fields.lock,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}
			if err := s.expire(tt.args.service); (err != nil) != tt.wantErr {
				t.Errorf("Store.expire() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if _, ok := s.reg[tt.args.service.UUID]; ok {
					t.Errorf("Store.expire() expected service to have been removed but was found")
				}
			}
		})
	}
}

func TestStore_expireAndRemove(t *testing.T) {
	type fields struct {
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		service Service
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		expired  bool
		expected bool
	}{
		{
			name: "Service expired and successfully removed",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						ExpiresAt: time.Now().Add(-5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service",
					ExpiresAt: time.Now().Add(-5 * time.Second),
				},
			},
			wantErr:  false,
			expired:  true,
			expected: false,
		},
		{
			name: "Service has not yet expired",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						ExpiresAt: time.Now().Add(5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service",
					ExpiresAt: time.Now().Add(5 * time.Second),
				},
			},
			wantErr:  false,
			expected: true,
			expired:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				lock:                        tt.fields.lock,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}
			expired, err := s.expireAndRemove(tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.expireAndRemove() error = %v, wantErr %v", err, tt.wantErr)
			}

			if expired != tt.expired {
				t.Errorf("Store.expireAndRemove() service to be expired %t but was %t", tt.expired, expired)
			}

			if !tt.wantErr {
				if _, ok := s.reg[tt.args.service.UUID]; ok != tt.expected {
					t.Errorf("Store.expire() service expected %t but was %t", tt.expected, ok)
				}
			}
		})
	}
}

func TestStore_fetch(t *testing.T) {
	now := time.Now()

	type fields struct {
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		service Service
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		want     Service
		wantErr  bool
		expected bool
	}{
		{
			name: "Service successfully found and not expired",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
						ExpiresAt: now.Add(5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service",
					IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
					ExpiresAt: now.Add(5 * time.Second),
				},
			},
			wantErr:  false,
			expected: true,
			want: Service{
				UUID:      "some-service",
				IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
				ExpiresAt: now.Add(5 * time.Second),
			},
		},
		{
			name: "Service found and has expired",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
						ExpiresAt: now.Add(-5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service",
					IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
					ExpiresAt: now.Add(-5 * time.Second),
				},
			},
			wantErr:  true,
			expected: false,
			want:     Service{},
		},
		{
			name: "Service not found",
			fields: fields{
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
						ExpiresAt: now.Add(5 * time.Second),
					},
				},
			},
			args: args{
				service: Service{
					UUID:      "some-service-other",
					IP:        netip.MustParseAddrPort("10.1.2.3:8080"),
					ExpiresAt: now.Add(5 * time.Second),
				},
			},
			wantErr:  true,
			expected: false,
			want:     Service{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				lock:                        tt.fields.lock,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}
			got, err := s.fetch(tt.args.service)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.fetch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.fetch() = %v, want %v", got, tt.want)
			}
			if _, ok := s.reg[tt.args.service.UUID]; ok != tt.expected {
				t.Errorf("Store.fetch() = service expected %t but was %t (services: %d)", tt.expected, ok, len(s.reg))
			}
		})
	}
}
