package main

import (
	"context"
	"flag"
	"net"
	"time"

	"github.com/jenmud/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	name := flag.String("service-name", "registry.srv", "Service registry name.")
	addr := flag.String("addr", ":8000", "Address to listen and accept client connections.")
	expiry := flag.Duration("expiry-duration", registry.DefaultExpiry, "Default node expiry duration.")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
	}

	var options []grpc.ServerOption
	srv := grpc.NewServer(options...)
	reg := registry.New(registry.Settings{ExpiryDuration: *expiry})
	node, err := reg.Register(context.Background(), &registry.Node{Name: *name, Address: *addr})
	if err != nil {
		logrus.Fatal(err)
	}

	// Run a background updating worker
	go func() {
		xd := *expiry - time.Second
		ticker := time.NewTicker(xd)
		for {
			<-ticker.C
			n, err := reg.Register(context.Background(), node)
			if err != nil {
				logrus.Errorf("Error updating node %s", n, err)
				continue
			}
			node = n
		}
	}()

	registry.RegisterRegistryServiceServer(srv, reg)

	logrus.Infof("RPC registry service (%s, %s) listen and accepting client connection on %q", node.GetName(), node.GetUid(), listener.Addr())
	if err := srv.Serve(listener); err != nil {
		logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
	}
}
