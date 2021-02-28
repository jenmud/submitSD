package registry

import (
	"time"

	"github.com/sirupsen/logrus"
)

// NewExpiryNode returns a new expiry node.
func NewExpiryNode(n *Node, d time.Duration) *ExpiryNode {
	xn := &ExpiryNode{Node: n}
	xn.StartExpiryTimer(d)
	return xn
}

// ExpiryNode is `registry.Node` wrapper with an expiry timer.
type ExpiryNode struct {
	*Node
	expired bool
	timer   *time.Timer
}

// StartExpiryTimer starts the expiry timer.
func (n *ExpiryNode) StartExpiryTimer(d time.Duration) {
	logrus.Infof("Node %q (%s) expires in %s", n.GetName(), n.GetUid(), d)
	n.timer = time.NewTimer(d)
	go func() {
		<-n.timer.C
		logrus.Infof("Node %q (%s) expired at %s", n.GetName(), n.GetUid(), time.Now().UTC())
		n.expired = true
	}()
}

// Expired returns true is the node is expired.
func (n *ExpiryNode) Expired() bool {
	return n.expired
}

// Expire expires the node.
func (n *ExpiryNode) Expire() {
	if !n.expired && !n.timer.Stop() {
		logrus.Infof("Waiting for node %q (%s) timer to stop", n.GetName(), n.GetUid())
		<-n.timer.C
	}

	logrus.Infof("Node %q (%s) has expired", n.GetName(), n.GetUid())
	n.expired = true
}

// Reset resets the expiry timer.
func (n *ExpiryNode) Reset(d time.Duration) {
	n.Expire()
	n.timer.Reset(d)
	logrus.Infof("Reset node %q (%s) expiry by %s", n.GetName(), n.GetUid(), d)
	n.expired = false
}

// Close closes the node by first expiring it.
func (n *ExpiryNode) Close() error {
	n.Expire()
	return nil
}
