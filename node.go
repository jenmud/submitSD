package registry

import (
	"fmt"
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
	timer *time.Timer
}

// StartExpiryTimer starts the expiry timer.
func (n *ExpiryNode) StartExpiryTimer(d time.Duration) {
	logrus.Infof("Node %s expires in %s", n, d)
	n.timer = time.NewTimer(d)
	go func() {
		<-n.timer.C
		logrus.Infof("Node %s expired at %s", n, time.Now().UTC())
		n.Node.Expired = true
	}()
}

// Expire expires the node.
func (n *ExpiryNode) Expire() {
	if !n.GetExpired() && !n.timer.Stop() {
		logrus.Infof("Waiting for node %s timer to stop", n)
		<-n.timer.C
	}

	n.Node.Expired = true
	logrus.Infof("Node %s has expired", n)
}

// Reset resets the expiry timer.
func (n *ExpiryNode) Reset(d time.Duration) {
	if !n.timer.Stop() {
		logrus.Infof("Waiting for node %s timer to stop", n)
		<-n.timer.C
	}

	n.timer.Reset(d)
	n.Node.Expired = false
	logrus.Infof("Reset node %s expiry by %s", n, d)
}

// Close closes the node by first expiring it.
func (n *ExpiryNode) Close() error {
	n.Expire()
	return nil
}

func (n *ExpiryNode) String() string {
	return fmt.Sprintf("%s (uid: %s)", n.GetName(), n.GetUid())
}
