package registry

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/google/uuid"
)

// New returns a new empty registry.
func New() *Store {
	return &Store{
		uids: make(map[string]string),
		reg: make(map[string]map[string]*Node),
	}
}

// Store used for storing and querying for nodes.
type Store struct {
	UnimplementedRegistryServiceServer
	uids map[string]string
	reg map[string]map[string]*Node
}

// Register registers a new node.
func (s *Store) Register(ctx context.Context, node *Node) (*Node, error) {
	resp := new(Node)

	resp.Uid = uuid.New().String()
	resp.Name = node.GetName()
	resp.Address = node.GetAddress()
	resp.Metadata = node.GetMetadata()

	logrus.Infof("Adding new node (%q), %s", resp.GetUid(), resp)
	return resp, nil
}
