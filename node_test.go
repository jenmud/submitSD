package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExpiryNode__StartExpiryTimer(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		10*time.Millisecond,
	)
	assert.False(t, n.Expired())
	time.Sleep(20 * time.Millisecond)
	assert.True(t, n.Expired())
}

func TestExpiryNode__Expired(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		10*time.Millisecond,
	)
	time.Sleep(20 * time.Millisecond)
	assert.True(t, n.Expired())
}

func TestExpiryNode__Expired__not_expired(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		10*time.Millisecond,
	)
	time.Sleep(5 * time.Millisecond)
	assert.False(t, n.Expired())
}

func TestExpiryNode__Expire(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		time.Minute,
	)

	time.Sleep(1 * time.Second)
	assert.False(t, n.Expired())
	n.Expire()
	assert.True(t, n.Expired())
}

func TestExpiryNode__Expire_already_expired(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		time.Minute,
	)

	time.Sleep(1 * time.Second)
	n.Expire()
	n.Expire()
	assert.True(t, n.Expired())
}

func TestExpiryNode__Reset(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		20*time.Millisecond,
	)

	// sleep for a short period
	time.Sleep(10 * time.Millisecond)

	// reset the timer and wait for another short period
	// before checking it expired.
	n.Reset(20 * time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, n.Expired())

	// Do the reset test again to make sure the reset is working
	n.Reset(20 * time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, n.Expired())

	// now wait till the node has expired
	// and check that the expiry timer still expires the node
	time.Sleep(1 * time.Second)
	assert.True(t, n.Expired())
}

func TestExpiryNode__Close(t *testing.T) {
	n := NewExpiryNode(
		&Node{Uid: "abc123", Name: "TestNode", Address: "0.0.0.0:8000"},
		20*time.Millisecond,
	)

	time.Sleep(10 * time.Millisecond)
	n.Close()
	assert.True(t, n.Expired())
}
