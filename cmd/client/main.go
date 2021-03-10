package main

import (
	"context"
	"flag"
	"io"
	"os/signal"

	"github.com/jenmud/registry"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background())
	defer cancel()

	addr := flag.String("addr", ":8000", "RPC service to connect to")
	flag.Parse()

	conn, err := grpc.DialContext(ctx, *addr, grpc.WithInsecure())
	if err != nil {
		logrus.Fatal(err)
	}

	client := registry.NewRegistryServiceClient(conn)

	stream, err := client.Events(ctx, &registry.EventReq{})
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Connected to %q and waiting for events", *addr)

	for {
		select {
		case <-ctx.Done():
			logrus.Info(ctx.Err().Error())
		default:
			event, err := stream.Recv()

			if err == io.EOF {
				logrus.Info("We are done with receiving events, closing...")
				cancel()
			}

			if err != nil {
				logrus.Fatal(err)
			}

			logrus.Infof("Event: %s", event)
		}
	}
}
