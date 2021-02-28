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
func (s *Store) Register(ctx context.Context, n *Node) (*ExpiryNode, error) {
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
		logrus.Infof("Updating node %q (%s) with %q (%s), expiry: %s", xn.GetName(), xn.GetUid(), n.GetName(), n.GetUid(), expiry)
		s.reg[xn.GetUid()] = xn
		return xn, nil
	}

	/*
		If we get to this point, generate a uuid and add to node
		to the registry.
	*/
	n.Uid = uuid.New().String()
	resp := NewExpiryNode(n, expiry)
	logrus.Infof("Adding new node %q (%s), %s", resp.GetName(), resp.GetUid(), resp)
	s.reg[resp.GetUid()] = resp
	return resp, nil
}

// Unregister unregisters a node from the registry.
func (s *Store) Unregister(ctx context.Context, node *Node) (*UnregisterResp, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if node.GetUid() == "" {
		return nil, fmt.Errorf("Node UID field is required")
	}

	if n, ok := s.reg[node.GetUid()]; ok {
		delete(s.reg, node.GetUid())
		return &UnregisterResp{Uid: n.GetUid(), DeletedAt: time.Now().UTC().Format(time.RFC3339)}, nil
	}

	return nil, fmt.Errorf("Node %s was not found", node.GetUid())
}

// Get queries and fetches the node from the registry.
func (s *Store) Get(ctx context.Context, req *GetReq) (*ExpiryNode, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if node, ok := s.reg[req.GetUid()]; ok {
		return node, nil
	}

	return nil, fmt.Errorf("Node %s was not found", req.GetUid())
}

// Search searches the registry for node with matching names.
func (s *Store) Search(ctx context.Context, req *SearchReq) ([]*ExpiryNode, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	nodes := []*ExpiryNode{}

	for _, node := range s.reg {
		switch req.GetName() {
		case "*":
			nodes = append(nodes, node)
		case node.GetName():
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}
