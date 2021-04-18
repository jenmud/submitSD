package registry

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// New returns a new empty registry.
func New(settings Settings) *Store {
	return &Store{
		settings:      settings,
		reg:           new(sync.Map),
		eventChannels: make(map[string]chan *EventResp),
	}
}

// Store used for storing and querying for nodes.
type Store struct {
	reg           *sync.Map
	eventChannels map[string]chan *EventResp
	settings      Settings
	UnimplementedRegistryServiceServer
}

// Register registers a new node.
func (s *Store) Register(ctx context.Context, node *Node) (*Node, error) {
	/*
		parse the expiry duration else use the default
	*/
	expiry := DefaultExpiry
	if node.GetExpiryDuration() != "" {
		e, err := time.ParseDuration(node.GetExpiryDuration())
		if err != nil {
			logrus.Errorf("Error parsing node expiry, reverting to default %s: %s", DefaultExpiry, err)
		} else {
			expiry = e
		}
	}

	/*
		If the node exists and is not expired, update it
	*/
	if n, err := s.get(node.GetUid()); err == nil {
		logrus.Infof("Node %q exists, updating touching it", n)
		_, err := s.Heartbeat(ctx, &HeartbeatReq{Uid: n.GetUid(), Duration: expiry.String()})
		return n, err
	}

	/*
		If we get to this point, generate a uuid and add to node
		to the registry.
	*/
	resp := &Node{
		Uid:            uuid.New().String(),
		Name:           node.GetName(),
		Address:        node.GetAddress(),
		Expiry:         time.Now().Add(expiry).Format(time.RFC3339),
		ExpiryDuration: expiry.String(),
		Metadata:       node.GetMetadata(),
		Expired:        false,
	}

	logrus.Infof("Adding new node %q (%s), %s", resp.GetName(), resp.GetUid(), resp)
	s.reg.Store(resp.GetUid(), resp)

	s.sendEvent(ctx, &EventResp{Uid: resp.Uid, Event: "register", Datetime: time.Now().UTC().Format(time.RFC3339)})
	return resp, nil
}

func (s *Store) resetExpiry(node *Node, expiry time.Duration) (*Node, error) {
	newExpiry := time.Now().Add(expiry)
	logrus.Infof("Resetting node %q expiry to %s", node.GetName(), newExpiry)

	node.ExpiryDuration = expiry.String()
	node.Expiry = newExpiry.UTC().Format(time.RFC3339)
	return node, nil
}

// Heartbeat does a single heartbeat update resetting the expiry.
// Note that the `UID` field is required else a error is returned.
func (s *Store) Heartbeat(ctx context.Context, req *HeartbeatReq) (*HeartbeatResp, error) {
	resp := &HeartbeatResp{}

	/*
		parse the expiry duration else use the default
	*/
	expiry := DefaultExpiry
	if req.GetDuration() != "" {
		e, err := time.ParseDuration(req.GetDuration())
		if err != nil {
			logrus.Errorf("Error parsing expiry duration, reverting to default %s: %s", expiry, err)
		} else {
			expiry = e
		}
	}

	/*
		Fetch the node in question
	*/
	cached, ok := s.reg.Load(req.GetUid())
	if !ok {
		return resp, fmt.Errorf("Node with UID %q was not found", req.GetUid())
	}

	node, err := s.resetExpiry(cached.(*Node), expiry)
	if err != nil {
		return resp, err
	}

	resp.Uid = node.GetUid()
	resp.Expiry = node.GetExpiry()

	s.sendEvent(ctx, &EventResp{Uid: resp.Uid, Event: "heartbeat", Datetime: time.Now().UTC().Format(time.RFC3339)})
	return resp, nil
}

// Heartbeats opens a bidirectional heartbeat stream used for continues heartbeat updates.
// Note that the `UID` field is required else a error is returned.
func (s *Store) Heartbeats(stream RegistryService_HeartbeatsServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		resp, err := s.Heartbeat(context.Background(), req)
		if err != nil {
			return err
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

func (s *Store) remove(node *Node) error {
	cached, ok := s.reg.LoadAndDelete(node.GetUid())

	if ok {
		n := cached.(*Node)
		logrus.Infof("Removing node %s", n.GetUid())
		n.Expired = true
		return nil
	}

	return fmt.Errorf("could not find node %q to unregister", node.GetUid())
}

// Unregister unregisters a node from the registry.
func (s *Store) Unregister(ctx context.Context, node *Node) (*Node, error) {
	n, err := s.get(node.GetUid())
	if err != nil {
		return nil, err
	}

	logrus.Infof("Unregistering node %q at %s", n.GetUid(), time.Now().UTC().Format(time.RFC3339))

	s.remove(n)
	s.sendEvent(ctx, &EventResp{Uid: n.GetUid(), Event: "unregister", Datetime: time.Now().UTC().Format(time.RFC3339)})
	return n, nil
}

func (s *Store) validate_node(n *Node) error {
	expiry, err := time.Parse(time.RFC3339, n.GetExpiry())
	if err != nil {
		return err
	}

	if time.Now().UTC().After(expiry) {
		msg := fmt.Sprintf("Node %q has expired %s ago", n.GetUid(), time.Since(expiry))
		logrus.Infof(msg)

		if err := s.remove(n); err != nil {
			return err
		}

		return errors.New(msg)
	}

	return nil
}

func (s *Store) get(uid string) (*Node, error) {
	cached, ok := s.reg.Load(uid)

	if ok {
		node := cached.(*Node)
		return node, s.validate_node(node)
	}

	return nil, fmt.Errorf("Could not find node with UID %q", uid)
}

// Get queries and fetches the node from the registry.
func (s *Store) Get(ctx context.Context, req *GetReq) (*Node, error) {
	return s.get(req.GetUid())
}

// Search searches the registry for node with matching names.
func (s *Store) Search(ctx context.Context, req *SearchReq) (*SearchResp, error) {
	nodes := []*Node{}

	s.reg.Range(func(key, value interface{}) bool {
		node := value.(*Node)
		if err := s.validate_node(node); err == nil {
			switch req.GetName() {
			case "*":
				nodes = append(nodes, node)
			case node.GetName():
				nodes = append(nodes, node)
			}
		}
		return true
	})

	return &SearchResp{Nodes: nodes}, nil
}

func (s *Store) sendEvent(ctx context.Context, event *EventResp) {
	for key, eventChannel := range s.eventChannels {
		logrus.Infof("Sending event %s to event channel %q", event, key)
		eventChannel <- event
	}
}

// Events subscribes a client to a stream of events.
func (s *Store) Events(req *EventReq, stream RegistryService_EventsServer) error {
	evID := uuid.New().String()
	eventChan := make(chan *EventResp, 1)
	s.eventChannels[evID] = eventChan
	logrus.Infof("Created and registered event channel %q", evID)

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				delete(s.eventChannels, evID)
				logrus.Infof("Closed %q and unregistered event channel", evID)
			}

			if event == nil {
				delete(s.eventChannels, evID)
				return fmt.Errorf("No event was received or event channel %q was closed", evID)
			}

			if err := stream.Send(event); err != nil {
				delete(s.eventChannels, evID)
				logrus.Errorf("Error sending event to channel %q: %s", evID, err)
				return err
			}
		}
	}
}

// Close closes the registry.
func (s *Store) Close() error {
	nodes := []*Node{}

	s.reg.Range(func(key, value interface{}) bool {
		node := value.(*Node)
		nodes = append(nodes, node)
		return true
	})

	for _, node := range nodes {
		s.remove(node)
	}

	for _, eventChannel := range s.eventChannels {
		close(eventChannel)
	}

	return nil
}
