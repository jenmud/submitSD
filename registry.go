package registry

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// New returns a new empty registry.
func New() *Store {
	return &Store{
		reg: make(map[string]*Node),
	}
}

// Store used for storing and querying for nodes.
type Store struct {
	UnimplementedRegistryServiceServer
	reg map[string]*Node
}

// Register registers a new node.
func (s *Store) Register(ctx context.Context, node *Node) (*Node, error) {
	resp := new(Node)

	/*
		If the node has a UID, then update the existing UID.
	*/
	if n, ok := s.reg[node.GetUid()]; ok {
		logrus.Infof("Updating node %s with %s", n, node)
		s.reg[node.GetUid()] = node
		return node, nil
	}

	/*
		If we get to this point, generate a uuid and add to node
		to the registry.
	*/
	resp.Uid = uuid.New().String()
	resp.Name = node.GetName()
	resp.Address = node.GetAddress()
	resp.Metadata = node.GetMetadata()

	logrus.Infof("Adding new node (%q), %s", resp.GetUid(), resp)
	s.reg[resp.GetUid()] = resp
	return resp, nil
}

// Unregister unregisters a node from the registry.
func (s *Store) Unregister(ctx context.Context, node *Node) (*UnregisterResp, error) {
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
func (s *Store) Get(ctx context.Context, req *GetReq) (*Node, error) {
	if node, ok := s.reg[req.GetUid()]; ok {
		return node, nil
	}

	return nil, fmt.Errorf("Node %s was not found", req.GetUid())
}

// Search searches the registry for node with matching names.
func (s *Store) Search(ctx context.Context, req *SearchReq) (*SearchResp, error) {
	nodes := []*Node{}

	for _, node := range s.reg {
		if node.GetName() == req.GetName() {
			nodes = append(nodes, node)
		}
	}

	return &SearchResp{Nodes: nodes}, nil
}
