package registry

import (
	"context"
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
		name         string
		fields       fields
		args         args
		expected     bool
		withCallback bool
	}{
		{
			name: "Successfully removed with no callback",
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
			expected: false,
		},
		{
			name: "Successfully removed with callback",
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
			expected:     false,
			withCallback: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				lock:                        tt.fields.lock,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}

			var called bool
			if tt.withCallback {
				s.SetEvictedCallback(
					func(expiredService Service) {
						called = true
						if !reflect.DeepEqual(expiredService, tt.args.service) {
							t.Errorf("Store.expire() expected %v but got %v", tt.args.service, expiredService)
						}
					},
				)
			}

			s.expire(tt.args.service)

			if _, ok := s.reg[tt.args.service.UUID]; ok != tt.expected {
				t.Errorf("Store.expire() expected %t but was %t", tt.expected, ok)
			}

			if tt.withCallback && !called {
				t.Error("Store.expire() callback to have fired but was has not")
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

			if _, ok := s.reg[tt.args.service.UUID]; ok != tt.expected {
				t.Errorf("Store.expire() service expected %t but was %t", tt.expected, ok)
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

func TestStore_DeleteExpired(t *testing.T) {
	type fields struct {
		cfg                         Config
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	tests := []struct {
		name         string
		fields       fields
		expected     []string
		unexpected   []string
		withCallback bool
	}{
		{
			name: "Successfully cleaned up expired services",
			fields: fields{
				cfg: Config{},
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						ExpiresAt: time.Now().Add(5 * time.Second),
						Expiry:    5 * time.Second,
					},
					"some-service-expired-service": {
						UUID:      "some-service-expired-service",
						ExpiresAt: time.Now().Add(-5 * time.Second),
						Expiry:    5 * time.Second,
					},
				},
				UnimplementedRegistryServer: proto.UnimplementedRegistryServer{},
			},
			expected:   []string{"some-service"},
			unexpected: []string{"some-service-expired-service"},
		},
		{
			name: "Successfully cleaned up expired services with callback",
			fields: fields{
				cfg: Config{},
				reg: map[string]Service{
					"some-service": {
						UUID:      "some-service",
						ExpiresAt: time.Now().Add(5 * time.Second),
						Expiry:    5 * time.Second,
					},
					"some-service-expired-service": {
						UUID:      "some-service-expired-service",
						ExpiresAt: time.Now().Add(-5 * time.Second),
						Expiry:    5 * time.Second,
					},
				},
				UnimplementedRegistryServer: proto.UnimplementedRegistryServer{},
			},
			expected:     []string{"some-service"},
			unexpected:   []string{"some-service-expired-service"},
			withCallback: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				cfg:                         tt.fields.cfg,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}

			var called bool
			if tt.withCallback {
				s.SetEvictedCallback(
					func(expiredService Service) {
						called = true
						matched := false
						for _, e := range tt.unexpected {
							if expiredService.UUID == e {
								matched = true
							}
						}
						if !matched {
							t.Errorf("Store.DeleteExpired() expected evicted callback to be called but was not")
						}
					},
				)
			}

			s.DeleteExpired()

			for _, each := range tt.expected {
				if _, ok := s.reg[each]; !ok {
					t.Errorf("Store.DeleteExpired() expected %s but was not found", each)
				}
			}

			for _, each := range tt.unexpected {
				if _, ok := s.reg[each]; ok {
					t.Errorf("Store.DeleteExpired() %s not expected but was found", each)
				}
			}

			if tt.withCallback && !called {
				t.Errorf("Store.DeleteExpired() expected eviction callback to be called but was not")
			}
		})
	}
}

func TestStore_GetByUUID(t *testing.T) {
	now := time.Now()

	type fields struct {
		cfg                         Config
		lock                        sync.RWMutex
		reg                         map[string]Service
		evictedClb                  EvictedClb
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		ctx context.Context
		req *proto.GetByUUIDReq
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.GetByUUIDResp
		wantErr bool
	}{
		{
			name: "service found",
			fields: fields{
				cfg:  Config{CleanupInterval: 5 * time.Second},
				lock: sync.RWMutex{},
				reg: map[string]Service{
					"my-service1": {
						UUID:      "my-service1",
						Name:      "my-service1",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(10 * time.Second),
					},
					"my-service2": {
						UUID:      "my-service2",
						Name:      "my-service2",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(12 * time.Second),
					},
					"my-service3": {
						UUID:      "my-service3",
						Name:      "my-service3",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(13 * time.Second),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetByUUIDReq{Uuid: "my-service2"},
			},
			want: &proto.GetByUUIDResp{
				Service: Service{
					UUID:      "my-service2",
					Name:      "my-service2",
					Expiry:    10 * time.Second,
					ExpiresAt: now.Add(12 * time.Second),
				}.ToPB(),
			},
			wantErr: false,
		},
		{
			name: "service not found",
			fields: fields{
				cfg:  Config{CleanupInterval: 5 * time.Second},
				lock: sync.RWMutex{},
				reg: map[string]Service{
					"my-service1": {
						UUID:      "my-service1",
						Name:      "my-service1",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(10 * time.Second),
					},
					"my-service2": {
						UUID:      "my-service2",
						Name:      "my-service2",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(12 * time.Second),
					},
					"my-service3": {
						UUID:      "my-service3",
						Name:      "my-service3",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(13 * time.Second),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetByUUIDReq{Uuid: "does-not-exist"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "service expired",
			fields: fields{
				cfg:  Config{CleanupInterval: 5 * time.Second},
				lock: sync.RWMutex{},
				reg: map[string]Service{
					"my-service1": {
						UUID:      "my-service1",
						Name:      "my-service1",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(10 * time.Second),
					},
					"my-service2": {
						UUID:      "my-service2",
						Name:      "my-service2",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(-12 * time.Second),
					},
					"my-service3": {
						UUID:      "my-service3",
						Name:      "my-service3",
						Expiry:    10 * time.Second,
						ExpiresAt: now.Add(13 * time.Second),
					},
				},
			},
			args: args{
				ctx: context.Background(),
				req: &proto.GetByUUIDReq{Uuid: "my-service2"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				cfg:        tt.fields.cfg,
				reg:        tt.fields.reg,
				evictedClb: tt.fields.evictedClb,
			}
			got, err := s.GetByUUID(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.GetByUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.GetByUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
