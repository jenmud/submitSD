package registry

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// NewExpiryNode returns a new expiry node.
func NewExpiryNode(n *Node, f ExpiryCallback) *ExpiryNode {
	return &ExpiryNode{Node: n, callback: f}
}

// ExpiryNode is `registry.Node` wrapper with an expiry timer.
// `.Start` needs to be called to start the expiry timer.
type ExpiryNode struct {
	*Node
	callback ExpiryCallback
	timer    *time.Timer
}

// Start starts the expiry timer.
func (n *ExpiryNode) Start(d time.Duration) {
	now := time.Now().UTC()
	n.Node.Expiry = now.Add(d).Format(time.RFC3339)
	n.timer = time.AfterFunc(d, func() { n.expire() })
	logrus.Infof("Node %s started at %s and expires at %s (duration: %s)", n, now, n.Node.Expiry, d)
}

func (n *ExpiryNode) expire() {
	n.Node.Expired = true
	n.Node.Expiry = time.Now().UTC().Format(time.RFC3339)
	logrus.Infof("Node %s has expired at %s", n, n.Node.Expiry)
	n.callback(n)
}

// Expire expires the node.
func (n *ExpiryNode) Expire() {
	if n.GetExpired() {
		return
	}

	if !n.timer.Stop() {
		<-n.timer.C
	}
	n.expire()
}

// Reset resets the expiry timer.
func (n *ExpiryNode) Reset(d time.Duration) {
	if !n.timer.Stop() {
		<-n.timer.C
	}
	n.timer.Reset(d)
	oldExpiry := n.Node.Expiry
	n.Node.Expired = false
	n.Node.Expiry = time.Now().UTC().Add(d).Format(time.RFC3339)
	logrus.Infof("Reset node %s expiry %s -> %s", n, oldExpiry, n.Node.GetExpiry())
}

// Close closes the node by first expiring it.
func (n *ExpiryNode) Close() error {
	n.Expire()
	return nil
}

func (n *ExpiryNode) String() string {
	return fmt.Sprintf("%s (uid: %s)", n.GetName(), n.GetUid())
}
