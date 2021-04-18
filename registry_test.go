package registry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStore_Register__New_Node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node := &Node{Address: "tcp://localhost:1234"}
	actual, err := store.Register(ctx, node)
	assert.Nil(t, err)
	assert.NotEmpty(t, actual.GetUid())
	assert.Equal(t, "tcp://localhost:1234", actual.GetAddress())
}

func TestStore_Register__New_Node_short_expiry(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node := &Node{Name: "TestNode", Address: "tcp://localhost:1234", ExpiryDuration: "20ms"}
	actual, err := store.Register(ctx, node)
	assert.Nil(t, err)
	time.Sleep(30 * time.Millisecond)
	node, err = store.Get(ctx, &GetReq{Uid: actual.GetUid()})
	assert.NotNil(t, err)
	assert.True(t, node.GetExpired())
}

func TestStore_Heartbeat__Update_node_expiry(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)

	node := &Node{Name: "TestNode", Address: "tcp://localhost:1234", ExpiryDuration: "1s"}
	actual, err := store.Register(ctx, node)
	assert.Nil(t, err)
	time.Sleep(1 / 2 * time.Second)

	resp, err := store.Heartbeat(ctx, &HeartbeatReq{Uid: actual.GetUid(), Duration: "2s"})
	assert.Nil(t, err)
	time.Sleep(time.Second)

	updated, err := store.Get(ctx, &GetReq{Uid: actual.GetUid()})
	assert.Nil(t, err)
	assert.Equal(t, actual.GetUid(), resp.GetUid())
	assert.Equal(t, actual.GetUid(), updated.GetUid())
	assert.False(t, updated.GetExpired())
}

func TestStore_Unregister__Existing_node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)

	node, err := store.Register(ctx, &Node{Address: "tcp://localhost:1234"})

	actual, err := store.Unregister(ctx, node)
	assert.Nil(t, err)
	assert.Equal(t, node.GetUid(), actual.GetUid())
	assert.True(t, actual.GetExpired())

	n, err := store.Get(ctx, &GetReq{Uid: node.GetUid()})
	assert.Nil(t, n)
	assert.NotNil(t, err)
}

func TestStore_Unregister__Missing_node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node := &Node{Uid: "missing", Address: "tcp://localhost:1234"}
	actual, err := store.Unregister(ctx, node)
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestStore_Unregister__Missing_uid(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node := &Node{Address: "tcp://localhost:1234"}
	actual, err := store.Unregister(ctx, node)
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestStore_Get__Existing_node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node, err := store.Register(ctx, &Node{Address: "tcp://localhost:1234"})
	actual, err := store.Get(ctx, &GetReq{Uid: node.GetUid()})
	assert.Nil(t, err)
	assert.Equal(t, node, actual)
}

func TestStore_Get__Missing_node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	store.Register(ctx, &Node{Address: "tcp://localhost:1234"})
	actual, err := store.Get(ctx, &GetReq{Uid: "missing"})
	assert.NotNil(t, err)
	assert.Nil(t, actual)
}

func TestStore_Search__Found(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node1, err := store.Register(ctx, &Node{Name: "web.srv", Address: "tcp://localhost:1234"})
	store.Register(ctx, &Node{Name: "mail.srv", Address: "tcp://localhost:2345"})
	node3, err := store.Register(ctx, &Node{Name: "web.srv", Address: "tcp://localhost:3456"})
	actual, err := store.Search(ctx, &SearchReq{Name: "web.srv"})
	assert.Nil(t, err)
	assert.ElementsMatch(t, []*Node{node1, node3}, actual.GetNodes())
}

func TestStore_Search__all_nodes(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node1, err := store.Register(ctx, &Node{Name: "web.srv", Address: "tcp://localhost:1234"})
	node2, err := store.Register(ctx, &Node{Name: "mail.srv", Address: "tcp://localhost:2345"})
	node3, err := store.Register(ctx, &Node{Name: "web.srv", Address: "tcp://localhost:3456"})
	actual, err := store.Search(ctx, &SearchReq{Name: "*"})
	assert.Nil(t, err)
	assert.ElementsMatch(t, []*Node{node1, node2, node3}, actual.GetNodes())
}
