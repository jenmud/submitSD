package registry

import (
	"context"
	reflect "reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStore_Register(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name    string
		args    []*Node
		wantErr bool
	}

	tests := []testCase{
		{
			name: "Add new node",
			args: []*Node{
				{Name: "testNodeA", Address: "0.0.0.0:1234"},
				{Name: "testNodeB", Address: "0.0.0.0:1234"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		reg := New()

		t.Run(tt.name, func(t *testing.T) {
			actual, err := reg.Register(ctx, tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Store.Register() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assert.NotEmptyf(t, actual.GetUid(), "Store.Register() expected UID but got %s", actual.GetName())
			_, hasUID := reg.uids[tt.args.GetUid()]
			_, hasReg := reg.uids[tt.args.GetUid()]
			assert.Equal()
			assert.True(t, reflect.DeepEqual(tt.expectedUids, reg.uids), "Store.Register() expected: %v but got %v", tt.expectedUids, reg.uids)
		})
	}
}
