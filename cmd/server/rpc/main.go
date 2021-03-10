package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/jenmud/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

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

	node, err := reg.Register(ctx, &registry.Node{Name: *name, Address: *addr})
	if err != nil {
		logrus.Fatal(err)
	}

	// Run a background updating worker
	go func(ctx context.Context, uid string, expiry time.Duration) {
		ticker := time.NewTicker(time.Second)

		for {
			select {
			case <-ctx.Done():
				logrus.Infof("Stopping heartbeat pulser: %s", ctx.Err())
				return
			case <-ticker.C:
				resp, err := reg.Heartbeat(ctx, &registry.HeartbeatReq{Uid: uid, Duration: expiry.String()})
				if err != nil {
					logrus.Errorf("Error sending heartbeat pulse for node %q: %s", uid, err)
					cancel()
				}
				logrus.Infof("Node received new heartbeat: %s", resp)
			}
		}
	}(ctx, node.GetUid(), *expiry)

	registry.RegisterRegistryServiceServer(srv, reg)

	go func() {
		if err := srv.Serve(listener); err != nil {
			logrus.Fatalf("Error starting registry RPC service %q: %s", *addr, err)
		}
	}()

	logrus.Infof("RPC registry service (%s, %s) listen and accepting client connection on %q", node.GetName(), node.GetUid(), listener.Addr())
	<-ctx.Done()
	logrus.Infof("Shutting down RPC registry service....")
	reg.Close()
	srv.GracefulStop()
}
