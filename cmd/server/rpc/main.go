package main

import (
	"flag"
	"net"

	"github.com/jenmud/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":8000", "Address to listen and accept client connections.")
	flag.Parse()

	listener, err := net.Listen("tcp", *addr)
	if err != nil {
		logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
	}

	var options []grpc.ServerOption
	srv := grpc.NewServer(options...)

	registry.RegisterRegistryServiceServer(srv, &registry.Store{})

	logrus.Infof("RPC registry service listen and accepting client connection on %q", listener.Addr())
	if err := srv.Serve(listener); err != nil {
		logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
	}
}
