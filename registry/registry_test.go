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

func TestStore_Add(t *testing.T) {
	type fields struct {
		lock                        sync.RWMutex
		reg                         map[string]Service
		UnimplementedRegistryServer proto.UnimplementedRegistryServer
	}
	type args struct {
		ctx context.Context
		req *proto.AddReq
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *proto.AddResp
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Store{
				lock:                        tt.fields.lock,
				reg:                         tt.fields.reg,
				UnimplementedRegistryServer: tt.fields.UnimplementedRegistryServer,
			}
			got, err := s.Add(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Add() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Store.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
