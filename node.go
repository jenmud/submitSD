package registry

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// NewExpiryNode returns a new expiry node.
func NewExpiryNode(n *Node, d time.Duration, f ExpiryCallback) *ExpiryNode {
	xn := &ExpiryNode{Node: n, callback: f, expired: make(chan time.Time)}
	xn.StartExpiryTimer(d)
	return xn
}

// ExpiryNode is `registry.Node` wrapper with an expiry timer.
type ExpiryNode struct {
	*Node
	callback ExpiryCallback
	timer    *time.Ticker
	expired  chan time.Time
}

// StartExpiryTimer starts the expiry timer.
func (n *ExpiryNode) StartExpiryTimer(d time.Duration) {
	now := time.Now().UTC()
	n.Node.Expiry = now.Add(d).Format(time.RFC3339)
	logrus.Infof("Node %s started at %s and expires at %s (duration: %s)", n, now, n.Node.Expiry, d)
	n.timer = time.NewTicker(d)
	go n.watcher()
}

func (n *ExpiryNode) watcher() {
	for {
		select {
		case dt := <-n.expired:
			logrus.Infof("Node %s forcefully expired at %s", n, dt.UTC())
			n.Node.Expired = true
			n.timer.Stop()
			n.callback(n)
			return
		case dt := <-n.timer.C:
			logrus.Infof("Node %s timer expired at %s", n, dt.UTC())
			n.Node.Expired = true
			n.callback(n)
			return
		}
	}
}

// Expire expires the node.
func (n *ExpiryNode) Expire() {
	n.timer.Stop()
	n.expired <- time.Now()
}

// Reset resets the expiry timer.
func (n *ExpiryNode) Reset(d time.Duration) {
	n.timer.Reset(d)
	oldExpiry := n.Node.Expiry
	n.Node.Expired = false
	n.Node.Expiry = time.Now().UTC().Add(d).Format(time.RFC3339)
	logrus.Infof("Reset node %s expiry %s -> %s", n, oldExpiry, n.Node.GetExpiry())
}

// Close closes the node by first expiring it.
func (n *ExpiryNode) Close() error {
	n.Expire()
	close(n.expired)
	return nil
}

func (n *ExpiryNode) String() string {
	return fmt.Sprintf("%s (uid: %s)", n.GetName(), n.GetUid())
}
