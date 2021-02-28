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
	assert.True(t, actual.GetExpired())
}

func TestStore_Register__Update_node_expiry(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)

	node := &Node{Name: "TestNode", Address: "tcp://localhost:1234", ExpiryDuration: "20ms"}
	actual, err := store.Register(ctx, node)
	assert.Nil(t, err)
	time.Sleep(10 * time.Millisecond)

	updated := &Node{Uid: actual.GetUid(), Name: "TestNode", Address: "tcp://localhost:1234", ExpiryDuration: "30ms"}
	newNode, err := store.Register(ctx, updated)
	assert.Nil(t, err)
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, actual.GetUid(), newNode.GetUid())
	assert.False(t, newNode.GetExpired())

	time.Sleep(40 * time.Millisecond)
	assert.True(t, actual.GetExpired())
}

func TestStore_Register__Update_Existing_Node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node, err := store.Register(ctx, &Node{Address: "tcp://localhost:1234"})
	node.Address = "udp://localhost:1234"
	actual, err := store.Register(ctx, node)
	assert.Nil(t, err)
	assert.Equal(t, node.GetUid(), actual.GetUid())
	assert.Equal(t, "udp://localhost:1234", actual.GetAddress())
}

func TestStore_Unregister__Existing_node(t *testing.T) {
	ctx := context.Background()
	settings := Settings{ExpiryDuration: DefaultExpiry}
	store := New(settings)
	node, err := store.Register(ctx, &Node{Address: "tcp://localhost:1234"})
	actual, err := store.Unregister(ctx, node)
	assert.Nil(t, err)
	assert.Equal(t, node.GetUid(), actual.GetUid())
	assert.NotEmpty(t, actual.GetDeletedAt())
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
