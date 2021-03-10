package registry

import (
	"context"
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
		reg:           make(map[string]*Node),
		timers:        make(map[string]*time.Timer),
		eventChannels: make(map[string]chan *EventResp),
	}
}

// Store used for storing and querying for nodes.
type Store struct {
	lock          sync.RWMutex
	reg           map[string]*Node
	timers        map[string]*time.Timer
	eventChannels map[string]chan *EventResp
	settings      Settings
	UnimplementedRegistryServiceServer
}

// Register registers a new node.
func (s *Store) Register(ctx context.Context, node *Node) (*Node, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

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
		If we get to this point, generate a uuid and add to node
		to the registry.
	*/
	resp := &Node{
		Uid:            uuid.New().String(),
		Name:           node.GetName(),
		Address:        node.GetAddress(),
		Expiry:         time.Now().UTC().Format(time.RFC3339),
		ExpiryDuration: expiry.String(),
		Metadata:       node.GetMetadata(),
		Expired:        false,
	}

	logrus.Infof("Adding new node %q (%s), %s", resp.GetName(), resp.GetUid(), resp)
	s.reg[resp.GetUid()] = resp
	s.timers[resp.GetUid()] = time.AfterFunc(expiry, func() { s.remove(resp) })

	s.sendEvent(ctx, &EventResp{Uid: resp.Uid, Event: "register", Datetime: time.Now().UTC().Format(time.RFC3339)})
	return resp, nil
}

func (s *Store) resetExpiry(node *Node, expiry time.Duration) (*Node, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	/*
		Fetch the nodes timer to update and reset
	*/
	timer, ok := s.timers[node.GetUid()]
	if !ok {
		return node, fmt.Errorf("Node with UID %q timer was not found", node.GetUid())
	}

	newExpiry := time.Now().Add(expiry)
	logrus.Infof("Resetting node %q expiry to %s", node.GetName(), newExpiry)

	if !timer.Stop() {
		logrus.Infof("Waiting for node %q timer to stop", node.GetName())
		<-timer.C
	}

	timer.Reset(expiry)
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
	node, ok := s.reg[req.GetUid()]
	if !ok {
		return resp, fmt.Errorf("Node with UID %q was not found", req.GetUid())
	}

	node, err := s.resetExpiry(node, expiry)
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
	s.lock.Lock()
	defer s.lock.Unlock()

	if n, ok := s.reg[node.GetUid()]; ok {
		logrus.Infof("Removing node %s", n.GetUid())
		delete(s.reg, n.GetUid())
	}

	if timer, ok := s.timers[node.GetUid()]; ok {
		logrus.Infof("Stopping node %s expiry timer", node.GetUid())
		timer.Stop()
		delete(s.timers, node.GetUid())
	}

	return nil
}

// Unregister unregisters a node from the registry.
func (s *Store) Unregister(ctx context.Context, node *Node) (*Node, error) {
	if node.GetUid() == "" {
		return nil, fmt.Errorf("Node UID field is required")
	}

	/*
		First find the node and remove it.
	*/
	n, ok := s.reg[node.GetUid()]
	if !ok {
		return node, fmt.Errorf("Could not find node %q to unregister", node.GetUid())
	}

	logrus.Infof("Unregistering node %q at %s", n.GetUid(), time.Now().UTC().Format(time.RFC3339))
	delete(s.reg, node.GetUid())

	/*
		Find the nodes timer and stop it.
	*/
	timer, ok := s.timers[node.GetUid()]
	if !ok {
		return node, fmt.Errorf("Could not find node timer %q to unregister", node.GetUid())
	}

	logrus.Infof("Unregistering node %q timer at %s", node.GetUid(), time.Now().UTC().Format(time.RFC3339))
	timer.Stop()
	delete(s.timers, node.GetUid())

	s.sendEvent(ctx, &EventResp{Uid: node.GetUid(), Event: "unregister", Datetime: time.Now().UTC().Format(time.RFC3339)})
	return node, nil
}

// Get queries and fetches the node from the registry.
func (s *Store) Get(ctx context.Context, req *GetReq) (*Node, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if n, ok := s.reg[req.GetUid()]; ok {
		return n, nil
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
			nodes = append(nodes, node)
		case node.GetName():
			nodes = append(nodes, node)
		}
	}

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
	s.lock.RLock()
	nodes := make([]*Node, 0, len(s.reg))
	s.lock.RUnlock()

	for _, node := range s.reg {
		nodes = append(nodes, node)
	}

	for _, node := range nodes {
		s.remove(node)
	}

	for _, eventChannel := range s.eventChannels {
		close(eventChannel)
	}

	return nil
}
