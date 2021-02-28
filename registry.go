package registry

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// New returns a new empty registry.
func New(settings Settings) *Store {
	return &Store{
		settings: settings,
		reg:      make(map[string]*ExpiryNode),
	}
}

// Store used for storing and querying for nodes.
type Store struct {
	lock     sync.RWMutex
	reg      map[string]*ExpiryNode
	settings Settings
	UnimplementedRegistryServiceServer
}

// Register registers a new node.
func (s *Store) Register(ctx context.Context, n *Node) (*Node, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// parse the expiry else use the default
	expiry := DefaultExpiry
	if n.GetExpiryDuration() != "" {
		e, err := time.ParseDuration(n.GetExpiryDuration())
		if err != nil {
			logrus.Errorf("Error parsing node expiry, reverting to default %s: %s", DefaultExpiry, err)
		} else {
			expiry = e
		}
	}

	/*
		If the node has a UID, then update the existing UID.
	*/
	if xn, ok := s.reg[n.GetUid()]; ok {
		xn.Reset(expiry)
		xd := time.Now().UTC().Add(-expiry)
		logrus.Infof("Updating node %s with %q (%s), expiry: %s", xn, n.GetName(), n.GetUid(), xd)
		s.reg[xn.GetUid()] = xn
		return xn.Node, nil
	}

	/*
		If we get to this point, generate a uuid and add to node
		to the registry.
	*/
	n.Uid = uuid.New().String()
	resp := NewExpiryNode(n, expiry)
	logrus.Infof("Adding new node %q (%s), %s", resp.GetName(), resp.GetUid(), resp)
	s.reg[resp.GetUid()] = resp
	return resp.Node, nil
}

// Unregister unregisters a node from the registry.
func (s *Store) Unregister(ctx context.Context, node *Node) (*Node, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if node.GetUid() == "" {
		return nil, fmt.Errorf("Node UID field is required")
	}

	if n, ok := s.reg[node.GetUid()]; ok {
		n.Expire()
		delete(s.reg, node.GetUid())
		return n.Node, nil
	}

	return nil, fmt.Errorf("Node %s was not found", node.GetUid())
}

// Get queries and fetches the node from the registry.
func (s *Store) Get(ctx context.Context, req *GetReq) (*Node, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if n, ok := s.reg[req.GetUid()]; ok {
		return n.Node, nil
	}

	return nil, fmt.Errorf("Node %s was not found", req.GetUid())
}

// Search searches the registry for node with matching names.
func (s *Store) Search(ctx context.Context, req *SearchReq) (*SearchResp, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	nodes := []*Node{}

	for _, node := range s.reg {
		switch req.GetName() {
		case "*":
			nodes = append(nodes, node.Node)
		case node.GetName():
			nodes = append(nodes, node.Node)
		}
	}

	return &SearchResp{Nodes: nodes}, nil
}
